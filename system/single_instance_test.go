package system

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoopbackSingleInstanceRejectsUnauthenticatedMessage(t *testing.T) {
	handle, secret := startTestLoopbackSingleInstance(t, "com.example.auth")
	conn, challenge := dialTestSingleInstance(t, handle.listener.Addr().String())
	defer conn.Close()

	wrongSecret := bytes.Repeat([]byte{0x5a}, singleInstanceSecretBytes)
	msg := loopbackSingleInstanceMessage{
		Protocol: singleInstanceProtocol,
		ID:       handle.id,
		Args:     []string{"--forged"},
		Payload:  "fluxui://forged",
	}
	msg.MAC = singleInstanceRequestMAC(wrongSecret, challenge.Nonce, msg)
	if err := writeSingleInstanceFrame(conn, msg); err != nil {
		t.Fatalf("write forged message: %v", err)
	}

	var response loopbackSingleInstanceResponse
	if err := readSingleInstanceFrame(conn, &response); err != nil {
		t.Fatalf("read rejection: %v", err)
	}
	if response.Status != "unauthorized" {
		t.Fatalf("expected unauthorized response, got %#v", response)
	}
	if !verifySingleInstanceResponseMAC(secret, challenge.Nonce, response) {
		t.Fatal("expected authenticated rejection from primary")
	}
	select {
	case event := <-handle.events():
		t.Fatalf("unauthenticated message dispatched event: %#v", event)
	default:
	}
}

func TestLoopbackSingleInstanceRejectsOversizedMessage(t *testing.T) {
	handle, secret := startTestLoopbackSingleInstance(t, "com.example.oversized")
	conn, challenge := dialTestSingleInstance(t, handle.listener.Addr().String())
	defer conn.Close()

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], singleInstanceMaxBytes+1)
	if _, err := conn.Write(header[:]); err != nil {
		t.Fatalf("write oversized header: %v", err)
	}

	var response loopbackSingleInstanceResponse
	if err := readSingleInstanceFrame(conn, &response); err != nil {
		t.Fatalf("read oversized rejection: %v", err)
	}
	if response.Status != "error" {
		t.Fatalf("expected error response, got %#v", response)
	}
	if !verifySingleInstanceResponseMAC(secret, challenge.Nonce, response) {
		t.Fatal("expected authenticated oversized-message rejection")
	}
	select {
	case event := <-handle.events():
		t.Fatalf("oversized message dispatched event: %#v", event)
	default:
	}
}

func TestLoopbackSingleInstanceRejectsUnauthenticatedPrimary(t *testing.T) {
	client, server := net.Pipe()
	secret := bytes.Repeat([]byte{0x41}, singleInstanceSecretBytes)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer server.Close()
		challengeBytes := bytes.Repeat([]byte{0x33}, singleInstanceSecretBytes)
		challenge := loopbackSingleInstanceChallenge{
			Protocol: singleInstanceProtocol,
			Nonce:    encodeTestSingleInstanceBytes(challengeBytes),
		}
		if err := writeSingleInstanceFrame(server, challenge); err != nil {
			return
		}
		var msg loopbackSingleInstanceMessage
		if err := readSingleInstanceFrame(server, &msg); err != nil {
			return
		}
		_ = writeSingleInstanceFrame(server, loopbackSingleInstanceResponse{
			Status: "ok",
			MAC:    encodeTestSingleInstanceBytes(bytes.Repeat([]byte{0x7f}, singleInstanceSecretBytes)),
		})
	}()

	err := writeLoopbackSingleInstanceMessage(client, singleInstanceOptions{id: "com.example.client"}, secret)
	if err == nil || !strings.Contains(err.Error(), "response authentication failed") {
		t.Fatalf("expected unauthenticated primary rejection, got %v", err)
	}
	<-done
}

func TestLoopbackSingleInstancePrivilegedPrimaryDoesNotForwardPayload(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for single-instance test: %v", err)
	}
	secret := bytes.Repeat([]byte{0x44}, singleInstanceSecretBytes)
	handle := newLoopbackSingleInstanceHandle(
		listener,
		singleInstanceOptions{id: "com.example.privileged"},
		secret,
		"",
		"",
		false,
	)
	go handle.acceptLoop()
	t.Cleanup(func() {
		if !handle.closed.Load() {
			_ = handle.close()
		}
	})

	err = forwardLoopbackSingleInstance(context.Background(), listener.Addr().String(), singleInstanceOptions{
		id:      handle.id,
		args:    []string{"--privileged-action"},
		payload: "fluxui://privileged",
	}, secret)
	if !IsAlreadyRunning(err) {
		t.Fatalf("expected privileged primary to report already running, got %v", err)
	}
	select {
	case event := <-handle.events():
		t.Fatalf("privileged primary dispatched untrusted launch data: %#v", event)
	default:
	}
}

func TestForwardLoopbackSingleInstanceRetriesTransientHandshakeEOF(t *testing.T) {
	handle, secret := startTestLoopbackSingleInstance(t, "com.example.retry-handshake")
	blockers := make([]net.Conn, 0, singleInstanceMaxConnections)
	for i := 0; i < singleInstanceMaxConnections; i++ {
		conn, _ := dialTestSingleInstance(t, handle.listener.Addr().String())
		blockers = append(blockers, conn)
	}
	defer func() {
		for _, conn := range blockers {
			_ = conn.Close()
		}
	}()

	released := make(chan struct{})
	go func() {
		timer := time.NewTimer(100 * time.Millisecond)
		defer timer.Stop()
		<-timer.C
		_ = blockers[0].Close()
		close(released)
	}()

	opts := singleInstanceOptions{
		id:      handle.id,
		args:    []string{"--retry"},
		payload: "fluxui://retry-handshake",
	}
	err := forwardLoopbackSingleInstance(context.Background(), handle.listener.Addr().String(), opts, secret)
	if !IsAlreadyRunning(err) {
		t.Fatalf("expected retry to reach primary, got %v", err)
	}
	<-released

	select {
	case event := <-handle.events():
		if event.Payload != opts.payload {
			t.Fatalf("unexpected retried event: %#v", event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event after transient handshake EOF")
	}
}

func TestLoopbackSingleInstanceCloseWaitsForConnections(t *testing.T) {
	handle, secret := startTestLoopbackSingleInstance(t, "com.example.close-race")

	const connectionCount = 24
	type pendingMessage struct {
		conn net.Conn
		msg  loopbackSingleInstanceMessage
	}
	pending := make([]pendingMessage, 0, connectionCount)
	for i := 0; i < connectionCount; i++ {
		conn, challenge := dialTestSingleInstance(t, handle.listener.Addr().String())
		msg := loopbackSingleInstanceMessage{
			Protocol: singleInstanceProtocol,
			ID:       handle.id,
			Args:     []string{"--open"},
			Payload:  "fluxui://close-race",
		}
		msg.MAC = singleInstanceRequestMAC(secret, challenge.Nonce, msg)
		pending = append(pending, pendingMessage{conn: conn, msg: msg})
	}

	start := make(chan struct{})
	var writers sync.WaitGroup
	for _, request := range pending {
		request := request
		writers.Add(1)
		go func() {
			defer writers.Done()
			defer request.conn.Close()
			<-start
			_ = request.conn.SetDeadline(time.Now().Add(2 * singleInstanceIOTimeout))
			if err := writeSingleInstanceFrame(request.conn, request.msg); err != nil {
				return
			}
			var response loopbackSingleInstanceResponse
			_ = readSingleInstanceFrame(request.conn, &response)
		}()
	}

	closeDone := make(chan error, 1)
	go func() {
		<-start
		runtime.Gosched()
		closeDone <- handle.close()
	}()
	close(start)
	writers.Wait()

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("close single-instance handle: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrent close")
	}
	for range handle.events() {
	}
}

func TestLoopbackSingleInstanceFallsBackFromOccupiedPort(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupy preferred port: %v", err)
	}
	defer occupied.Close()
	preferredPort := occupied.Addr().(*net.TCPAddr).Port

	opts := singleInstanceOptions{
		id:   "com.example.port-fallback",
		port: preferredPort,
	}
	statePath := filepath.Join(t.TempDir(), "single-instance.json")
	handleValue, err := acquireLoopbackSingleInstanceWithState(context.Background(), opts, statePath)
	if err != nil {
		t.Fatalf("acquire with occupied preferred port: %v", err)
	}
	handle := handleValue.(*loopbackSingleInstanceHandle)
	if got := handle.listener.Addr().(*net.TCPAddr).Port; got == preferredPort {
		t.Fatalf("expected ephemeral fallback port, still using %d", got)
	}
	if err := handle.close(); err != nil {
		t.Fatalf("close fallback handle: %v", err)
	}
}

func TestWriteSingleInstanceStateExclusivePublishesCompleteRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	state := loopbackSingleInstanceState{
		Version:  singleInstanceStateVersion,
		Protocol: singleInstanceProtocol,
		ID:       "com.example.atomic-state",
		Address:  "127.0.0.1:54321",
		Secret:   encodeTestSingleInstanceBytes(bytes.Repeat([]byte{1}, singleInstanceSecretBytes)),
		Owner:    encodeTestSingleInstanceBytes(bytes.Repeat([]byte{2}, singleInstanceSecretBytes)),
		PID:      123,
	}

	claimed, err := writeSingleInstanceStateExclusive(path, state)
	if err != nil || !claimed {
		t.Fatalf("publish state: claimed=%v err=%v", claimed, err)
	}
	got, err := readSingleInstanceStateRecord(path)
	if err != nil {
		t.Fatalf("read published state: %v", err)
	}
	if got != state {
		t.Fatalf("published state mismatch: got %#v want %#v", got, state)
	}
	claimed, err = writeSingleInstanceStateExclusive(path, state)
	if err != nil || claimed {
		t.Fatalf("second state claim should lose atomically: claimed=%v err=%v", claimed, err)
	}
}

func TestSingleInstanceStateLockSerializesPublish(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("cross-process state locking is unavailable on this platform")
	}
	path := filepath.Join(t.TempDir(), "state.json")
	state := loopbackSingleInstanceState{
		Version:  singleInstanceStateVersion,
		Protocol: singleInstanceProtocol,
		ID:       "com.example.locked-state",
		Address:  "127.0.0.1:54321",
		Secret:   encodeTestSingleInstanceBytes(bytes.Repeat([]byte{1}, singleInstanceSecretBytes)),
		Owner:    encodeTestSingleInstanceBytes(bytes.Repeat([]byte{2}, singleInstanceSecretBytes)),
		PID:      123,
	}

	unlock, err := lockSingleInstanceState(context.Background(), path)
	if err != nil {
		t.Fatalf("lock state: %v", err)
	}
	locked := true
	defer func() {
		if locked {
			_ = unlock()
		}
	}()

	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		claimed, err := writeSingleInstanceStateExclusive(path, state)
		if err == nil && !claimed {
			err = errors.New("state claim unexpectedly lost")
		}
		done <- err
	}()
	<-started
	select {
	case err := <-done:
		t.Fatalf("state publish bypassed mutation lock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if err := unlock(); err != nil {
		t.Fatalf("unlock state: %v", err)
	}
	locked = false
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("publish after unlock: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("state publish did not resume after unlock")
	}
}

func TestSingleInstanceStateLockHonorsContextDeadline(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("cross-process state locking is unavailable on this platform")
	}
	path := filepath.Join(t.TempDir(), "state.json")
	unblock, err := lockSingleInstanceState(context.Background(), path)
	if err != nil {
		t.Fatalf("lock state: %v", err)
	}
	defer unblock()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = writeSingleInstanceStateExclusiveContext(ctx, path, loopbackSingleInstanceState{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline while state lock is held, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("state lock ignored context deadline for %s", elapsed)
	}
}

func TestRemoveSingleInstanceStateDoesNotDeleteReplacementOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	oldState := loopbackSingleInstanceState{
		Version:  singleInstanceStateVersion,
		Protocol: singleInstanceProtocol,
		ID:       "com.example.replaced-state",
		Address:  "127.0.0.1:54321",
		Secret:   encodeTestSingleInstanceBytes(bytes.Repeat([]byte{1}, singleInstanceSecretBytes)),
		Owner:    encodeTestSingleInstanceBytes(bytes.Repeat([]byte{2}, singleInstanceSecretBytes)),
		PID:      123,
	}
	newState := oldState
	newState.Owner = encodeTestSingleInstanceBytes(bytes.Repeat([]byte{3}, singleInstanceSecretBytes))

	claimed, err := writeSingleInstanceStateExclusive(path, oldState)
	if err != nil || !claimed {
		t.Fatalf("publish old state: claimed=%v err=%v", claimed, err)
	}
	removed, err := removeSingleInstanceStateIfOwner(path, oldState.Owner)
	if err != nil || !removed {
		t.Fatalf("remove old state: removed=%v err=%v", removed, err)
	}
	claimed, err = writeSingleInstanceStateExclusive(path, newState)
	if err != nil || !claimed {
		t.Fatalf("publish replacement state: claimed=%v err=%v", claimed, err)
	}

	removed, err = removeSingleInstanceStateIfOwner(path, oldState.Owner)
	if err != nil {
		t.Fatalf("remove stale owner: %v", err)
	}
	if removed {
		t.Fatal("stale owner removed replacement state")
	}
	got, err := readSingleInstanceStateRecord(path)
	if err != nil {
		t.Fatalf("read replacement state: %v", err)
	}
	if got.Owner != newState.Owner {
		t.Fatalf("replacement owner changed: got %q want %q", got.Owner, newState.Owner)
	}
}

func TestAcquireLoopbackSingleInstanceRemovesUntrustedPublishedState(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for untrusted state: %v", err)
	}
	defer occupied.Close()

	statePath := filepath.Join(t.TempDir(), "state.json")
	state := loopbackSingleInstanceState{
		Version:  singleInstanceStateVersion,
		Protocol: singleInstanceProtocol,
		ID:       "com.example.untrusted-state",
		Address:  occupied.Addr().String(),
		Secret:   encodeTestSingleInstanceBytes(bytes.Repeat([]byte{3}, singleInstanceSecretBytes)),
		Owner:    encodeTestSingleInstanceBytes(bytes.Repeat([]byte{4}, singleInstanceSecretBytes)),
		PID:      os.Getpid(),
	}
	claimed, err := writeSingleInstanceStateExclusive(statePath, state)
	if err != nil || !claimed {
		t.Fatalf("publish untrusted state: claimed=%v err=%v", claimed, err)
	}

	verifyCalls := 0
	handleValue, err := acquireLoopbackSingleInstanceWithStateVerifier(
		context.Background(),
		singleInstanceOptions{id: state.ID, port: occupied.Addr().(*net.TCPAddr).Port},
		statePath,
		func(got loopbackSingleInstanceState) (bool, error) {
			verifyCalls++
			if got.Owner != state.Owner {
				t.Fatalf("verified unexpected state owner %q", got.Owner)
			}
			return false, nil
		},
	)
	if err != nil {
		t.Fatalf("acquire after untrusted state: %v", err)
	}
	handle := handleValue.(*loopbackSingleInstanceHandle)
	defer func() {
		if !handle.closed.Load() {
			_ = handle.close()
		}
	}()
	if verifyCalls != 1 {
		t.Fatalf("unexpected verifier calls: %d", verifyCalls)
	}
	if handle.listener.Addr().String() == occupied.Addr().String() {
		t.Fatal("primary reused the untrusted listener")
	}
}

func startTestLoopbackSingleInstance(t *testing.T, id string) (*loopbackSingleInstanceHandle, []byte) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for single-instance test: %v", err)
	}
	secret := bytes.Repeat([]byte{0x42}, singleInstanceSecretBytes)
	handle := newLoopbackSingleInstanceHandle(listener, singleInstanceOptions{id: id}, secret, "", "", true)
	go handle.acceptLoop()
	t.Cleanup(func() {
		if !handle.closed.Load() {
			_ = handle.close()
		}
	})
	return handle, secret
}

func dialTestSingleInstance(t *testing.T, address string) (net.Conn, loopbackSingleInstanceChallenge) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, singleInstanceIOTimeout)
	if err != nil {
		t.Fatalf("dial single-instance primary: %v", err)
	}
	if err := conn.SetDeadline(time.Now().Add(2 * singleInstanceIOTimeout)); err != nil {
		conn.Close()
		t.Fatalf("set single-instance test deadline: %v", err)
	}
	var challenge loopbackSingleInstanceChallenge
	if err := readSingleInstanceFrame(conn, &challenge); err != nil {
		conn.Close()
		t.Fatalf("read single-instance challenge: %v", err)
	}
	if err := validateSingleInstanceChallenge(challenge); err != nil {
		conn.Close()
		t.Fatalf("validate single-instance challenge: %v", err)
	}
	return conn, challenge
}

func encodeTestSingleInstanceBytes(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
