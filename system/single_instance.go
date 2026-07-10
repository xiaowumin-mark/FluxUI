package system

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	singleInstancePortBase       = 49152
	singleInstancePortRange      = 10000
	singleInstanceIOTimeout      = 1500 * time.Millisecond
	singleInstanceRetryInterval  = 25 * time.Millisecond
	singleInstanceMaxBytes       = 64 * 1024
	singleInstanceMaxConnections = 32
	singleInstanceSecretBytes    = 32
	singleInstanceStateVersion   = 2
	singleInstanceProtocol       = "fluxui-single-instance-v2"
)

var (
	errSingleInstanceAuthentication  = errors.New("single-instance authentication failed")
	errSingleInstanceHandshakeAbsent = errors.New("single-instance handshake unavailable")
	errSingleInstanceMessageRejected = errors.New("single-instance message rejected")
	errSingleInstanceMessageTooLarge = errors.New("single-instance message too large")
)

// SingleInstanceEvent is delivered to the primary instance when a secondary
// launch forwards its startup data.
type SingleInstanceEvent struct {
	ID      string
	Args    []string
	Payload string
}

// SingleInstanceOption configures single-instance acquisition.
type SingleInstanceOption func(*singleInstanceOptions)

type singleInstanceOptions struct {
	id             string
	args           []string
	payload        string
	port           int
	onSecondLaunch func(SingleInstanceEvent)
}

type singleInstanceDriver interface {
	acquireSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error)
}

type singleInstanceHandle interface {
	events() <-chan SingleInstanceEvent
	close() error
}

// SingleInstance owns the primary instance coordination channel.
type SingleInstance struct {
	handle singleInstanceHandle
}

// SingleInstanceID sets the stable application identity used for coordination.
//
// Applications should pass an explicit reverse-DNS style value, for example
// "com.example.myapp". If omitted, FluxUI derives a best-effort ID from the
// current executable name.
func SingleInstanceID(id string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.id = strings.TrimSpace(id)
	}
}

// SingleInstanceArgs sets the startup arguments forwarded by secondary launches.
//
// When omitted, os.Args[1:] is used.
func SingleInstanceArgs(args ...string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.args = append([]string(nil), args...)
	}
}

// SingleInstancePayload sets an optional string payload forwarded by secondary
// launches. Protocol activation handlers can use this for the raw activation URI.
func SingleInstancePayload(payload string) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.payload = payload
	}
}

// SingleInstanceOnSecondLaunch registers a callback for forwarded secondary
// launches. Events are also available through (*SingleInstance).Events().
func SingleInstanceOnSecondLaunch(fn func(SingleInstanceEvent)) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.onSecondLaunch = fn
	}
}

// SingleInstancePort sets the preferred loopback port instead of the port derived
// from ID. If that port is occupied by an unauthenticated process, FluxUI falls
// back to an ephemeral loopback port published through its per-user state file.
//
// This is mainly useful when two applications would otherwise collide on the
// same derived port. Values must be in the TCP user port range.
func SingleInstancePort(port int) SingleInstanceOption {
	return func(opts *singleInstanceOptions) {
		opts.port = port
	}
}

// AcquireSingleInstance becomes the primary instance or forwards startup data to
// the current primary instance.
//
// On success, the caller owns the primary instance and should close the returned
// handle during shutdown. If another instance is already running, the startup
// data is forwarded and the function returns an error wrapping ErrAlreadyRunning.
//
// Forwarding is authenticated by a random capability stored in the current
// user's cache directory. This prevents injection by processes that cannot read
// that state, but it does not authenticate the calling executable. In particular,
// Windows processes running as the same account can normally share LocalAppData
// across UAC integrity levels. To prevent lower-integrity command injection,
// elevated Windows primaries do not forward Args or Payload; secondary instances
// still receive ErrAlreadyRunning. Applications must always validate forwarded
// data before performing privileged work.
func AcquireSingleInstance(ctx context.Context, options ...SingleInstanceOption) (*SingleInstance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	opts := defaultSingleInstanceOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	if err := normalizeSingleInstanceOptions(&opts); err != nil {
		return nil, err
	}

	d, supported := currentDriverFor(CapabilitySingleInstance)
	sd, ok := d.(singleInstanceDriver)
	if !ok || !supported {
		return nil, fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrUnsupported)
	}

	handle, err := sd.acquireSingleInstance(ctx, opts)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("system: %s: acquire: %w", CapabilitySingleInstance, ErrUnavailable)
	}
	return &SingleInstance{handle: handle}, nil
}

// Events returns secondary launch events delivered to the primary instance.
func (s *SingleInstance) Events() <-chan SingleInstanceEvent {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.events()
}

// Close releases the primary instance coordination channel.
func (s *SingleInstance) Close() error {
	if s == nil || s.handle == nil {
		return singleInstanceClosedError()
	}
	return s.handle.close()
}

func defaultSingleInstanceOptions() singleInstanceOptions {
	return singleInstanceOptions{
		id:   defaultSingleInstanceID(),
		args: append([]string(nil), os.Args[1:]...),
	}
}

func normalizeSingleInstanceOptions(opts *singleInstanceOptions) error {
	if opts == nil {
		return fmt.Errorf("system: %s: missing options: %w", CapabilitySingleInstance, ErrUnavailable)
	}
	opts.id = strings.TrimSpace(opts.id)
	if opts.id == "" {
		return fmt.Errorf("system: %s: id is empty", CapabilitySingleInstance)
	}
	opts.args = append([]string(nil), opts.args...)
	if opts.port == 0 {
		opts.port = singleInstancePortForID(opts.id)
	}
	if opts.port < 1024 || opts.port > 65535 {
		return fmt.Errorf("system: %s: invalid port %d", CapabilitySingleInstance, opts.port)
	}
	return nil
}

func defaultSingleInstanceID() string {
	exe, err := os.Executable()
	if err == nil && exe != "" {
		return "fluxui." + filepath.Base(exe)
	}
	if len(os.Args) > 0 && os.Args[0] != "" {
		return "fluxui." + filepath.Base(os.Args[0])
	}
	return "fluxui.app"
}

func singleInstancePortForID(id string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return singleInstancePortBase + int(h.Sum32()%singleInstancePortRange)
}

func acquireLoopbackSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error) {
	statePath, err := singleInstanceStatePath(opts)
	if err != nil {
		return nil, fmt.Errorf("system: %s: state path: %w: %v", CapabilitySingleInstance, ErrUnavailable, err)
	}
	return acquireLoopbackSingleInstanceWithState(ctx, opts, statePath)
}

func singleInstanceAddress(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}

type loopbackSingleInstanceMessage struct {
	Protocol string   `json:"protocol"`
	ID       string   `json:"id"`
	Args     []string `json:"args,omitempty"`
	Payload  string   `json:"payload,omitempty"`
	MAC      string   `json:"mac"`
}

type loopbackSingleInstanceResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	MAC    string `json:"mac"`
}

type loopbackSingleInstanceChallenge struct {
	Protocol string `json:"protocol"`
	Nonce    string `json:"nonce"`
}

type loopbackSingleInstanceState struct {
	Version  int    `json:"version"`
	Protocol string `json:"protocol"`
	ID       string `json:"id"`
	Address  string `json:"address"`
	Secret   string `json:"secret"`
	Owner    string `json:"owner"`
	PID      int    `json:"pid"`
}

type loopbackSingleInstanceHandle struct {
	listener   net.Listener
	eventsCh   chan SingleInstanceEvent
	callback   func(SingleInstanceEvent)
	id         string
	secret     []byte
	statePath  string
	stateOwner string
	closed     atomic.Bool
	acceptDone chan struct{}
	connSlots  chan struct{}
	connWG     sync.WaitGroup
	forwarding bool
}

func newLoopbackSingleInstanceHandle(listener net.Listener, opts singleInstanceOptions, secret []byte, statePath, stateOwner string, forwarding bool) *loopbackSingleInstanceHandle {
	return &loopbackSingleInstanceHandle{
		listener:   listener,
		eventsCh:   make(chan SingleInstanceEvent, 16),
		callback:   opts.onSecondLaunch,
		id:         opts.id,
		secret:     append([]byte(nil), secret...),
		statePath:  statePath,
		stateOwner: stateOwner,
		acceptDone: make(chan struct{}),
		connSlots:  make(chan struct{}, singleInstanceMaxConnections),
		forwarding: forwarding,
	}
}

func (h *loopbackSingleInstanceHandle) events() <-chan SingleInstanceEvent {
	if h == nil {
		return nil
	}
	return h.eventsCh
}

func (h *loopbackSingleInstanceHandle) close() error {
	if h == nil || !h.closed.CompareAndSwap(false, true) {
		return singleInstanceClosedError()
	}

	var closeErr error
	if err := h.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		closeErr = err
	}
	if _, err := removeSingleInstanceStateIfOwner(h.statePath, h.stateOwner); err != nil && closeErr == nil {
		closeErr = err
	}
	<-h.acceptDone
	h.connWG.Wait()
	close(h.eventsCh)
	clear(h.secret)
	if closeErr != nil {
		return fmt.Errorf("system: %s: close: %w", CapabilitySingleInstance, closeErr)
	}
	return nil
}

func (h *loopbackSingleInstanceHandle) acceptLoop() {
	defer close(h.acceptDone)

	for {
		conn, err := h.listener.Accept()
		if err != nil {
			return
		}
		select {
		case h.connSlots <- struct{}{}:
			h.connWG.Add(1)
			go func() {
				defer h.connWG.Done()
				defer func() { <-h.connSlots }()
				h.handleConn(conn)
			}()
		default:
			_ = conn.Close()
		}
	}
}

func (h *loopbackSingleInstanceHandle) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(singleInstanceIOTimeout))

	challengeBytes, err := singleInstanceRandomBytes(singleInstanceSecretBytes)
	if err != nil {
		return
	}
	challenge := loopbackSingleInstanceChallenge{
		Protocol: singleInstanceProtocol,
		Nonce:    base64.RawURLEncoding.EncodeToString(challengeBytes),
	}
	clear(challengeBytes)
	if err := writeSingleInstanceFrame(conn, challenge); err != nil {
		return
	}

	var msg loopbackSingleInstanceMessage
	if err := readSingleInstanceFrame(conn, &msg); err != nil {
		h.writeResponse(conn, challenge.Nonce, "error", "decode")
		return
	}
	if msg.Protocol != singleInstanceProtocol || msg.ID != h.id {
		h.writeResponse(conn, challenge.Nonce, "mismatch", "id mismatch")
		return
	}
	if !verifySingleInstanceRequestMAC(h.secret, challenge.Nonce, msg) {
		h.writeResponse(conn, challenge.Nonce, "unauthorized", "authentication failed")
		return
	}
	if !h.forwarding {
		h.writeResponse(conn, challenge.Nonce, "already-running", "forwarding disabled for privileged primary")
		return
	}

	event := SingleInstanceEvent{
		ID:      msg.ID,
		Args:    append([]string(nil), msg.Args...),
		Payload: msg.Payload,
	}
	if !h.dispatch(event) {
		h.writeResponse(conn, challenge.Nonce, "closed", "primary is closing")
		return
	}
	h.writeResponse(conn, challenge.Nonce, "ok", "")
}

func (h *loopbackSingleInstanceHandle) dispatch(event SingleInstanceEvent) bool {
	if h.closed.Load() {
		return false
	}
	if h.callback != nil {
		go h.callback(event)
	}
	select {
	case h.eventsCh <- event:
	default:
	}
	return true
}

func (h *loopbackSingleInstanceHandle) writeResponse(conn net.Conn, challenge, status, detail string) {
	response := loopbackSingleInstanceResponse{
		Status: status,
		Error:  detail,
	}
	response.MAC = singleInstanceResponseMAC(h.secret, challenge, response)
	_ = writeSingleInstanceFrame(conn, response)
}

func forwardLoopbackSingleInstance(ctx context.Context, addr string, opts singleInstanceOptions, secret []byte) error {
	deadline := singleInstanceDeadline(ctx)
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		conn, err := (&net.Dialer{Deadline: deadline}).DialContext(ctx, "tcp", addr)
		if err == nil {
			err = writeLoopbackSingleInstanceMessageBefore(conn, opts, secret, deadline)
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if IsAlreadyRunning(err) || !retryableSingleInstanceHandshakeError(err) {
				return err
			}
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			break
		}
		if err := waitSingleInstanceRetry(ctx, deadline); err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			break
		}
	}
	return fmt.Errorf("contact primary: %w", lastErr)
}

func writeLoopbackSingleInstanceMessage(conn net.Conn, opts singleInstanceOptions, secret []byte) error {
	return writeLoopbackSingleInstanceMessageBefore(conn, opts, secret, time.Now().Add(singleInstanceIOTimeout))
}

func writeLoopbackSingleInstanceMessageBefore(conn net.Conn, opts singleInstanceOptions, secret []byte, deadline time.Time) error {
	defer conn.Close()
	_ = conn.SetDeadline(deadline)

	var challenge loopbackSingleInstanceChallenge
	if err := readSingleInstanceFrame(conn, &challenge); err != nil {
		return fmt.Errorf("read challenge: %v: %w", err, errSingleInstanceHandshakeAbsent)
	}
	if err := validateSingleInstanceChallenge(challenge); err != nil {
		return fmt.Errorf("validate challenge: %v: %w", err, errSingleInstanceAuthentication)
	}
	msg := loopbackSingleInstanceMessage{
		Protocol: singleInstanceProtocol,
		ID:       opts.id,
		Args:     append([]string(nil), opts.args...),
		Payload:  opts.payload,
	}
	msg.MAC = singleInstanceRequestMAC(secret, challenge.Nonce, msg)
	if err := writeSingleInstanceFrame(conn, msg); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	var response loopbackSingleInstanceResponse
	if err := readSingleInstanceFrame(conn, &response); err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if !verifySingleInstanceResponseMAC(secret, challenge.Nonce, response) {
		return fmt.Errorf("primary response authentication failed: %w", errSingleInstanceAuthentication)
	}
	if response.Status == "already-running" {
		return fmt.Errorf("system: %s: primary did not accept forwarded launch data: %w", CapabilitySingleInstance, ErrAlreadyRunning)
	}
	if response.Status != "ok" {
		if response.Error == "" {
			response.Error = response.Status
		}
		return fmt.Errorf("primary rejected payload: %s: %w", response.Error, errSingleInstanceMessageRejected)
	}
	return fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrAlreadyRunning)
}

func retryableSingleInstanceHandshakeError(err error) bool {
	return errors.Is(err, errSingleInstanceHandshakeAbsent)
}

func singleInstanceDeadline(ctx context.Context) time.Time {
	deadline := time.Now().Add(singleInstanceIOTimeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		return contextDeadline
	}
	return deadline
}

func acquireLoopbackSingleInstanceWithState(ctx context.Context, opts singleInstanceOptions, statePath string) (singleInstanceHandle, error) {
	return acquireLoopbackSingleInstanceWithStateVerifier(ctx, opts, statePath, singleInstancePublishedStateTrusted)
}

type singleInstanceStateVerifier func(loopbackSingleInstanceState) (bool, error)

func acquireLoopbackSingleInstanceWithStateVerifier(ctx context.Context, opts singleInstanceOptions, statePath string, verify singleInstanceStateVerifier) (singleInstanceHandle, error) {
	if verify == nil {
		verify = singleInstancePublishedStateTrusted
	}
	deadline := singleInstanceDeadline(ctx)
	acquireCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	ctx = acquireCtx
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		state, secret, err := readSingleInstanceState(statePath, opts)
		if err == nil {
			trusted, verifyErr := verify(state)
			if verifyErr != nil {
				clear(secret)
				return nil, fmt.Errorf("system: %s: verify published primary: %w: %v", CapabilitySingleInstance, ErrUnavailable, verifyErr)
			}
			if !trusted {
				clear(secret)
				removed, removeErr := removeSingleInstanceStateIfOwnerContext(ctx, statePath, state.Owner)
				if removeErr != nil {
					return nil, fmt.Errorf("system: %s: remove untrusted state: %w: %v", CapabilitySingleInstance, ErrUnavailable, removeErr)
				}
				if removed {
					continue
				}
				if err := waitSingleInstanceRetry(ctx, deadline); err != nil {
					return nil, fmt.Errorf("system: %s: untrusted state changed while acquiring: %w: %v", CapabilitySingleInstance, ErrUnavailable, err)
				}
				continue
			}
			forwardErr := forwardLoopbackSingleInstance(ctx, state.Address, opts, secret)
			clear(secret)
			if IsAlreadyRunning(forwardErr) {
				return nil, forwardErr
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			currentState, currentErr := readSingleInstanceStateRecord(statePath)
			if errors.Is(currentErr, os.ErrNotExist) || (currentErr == nil && currentState.Owner != state.Owner) {
				continue
			}
			if currentErr != nil {
				return nil, fmt.Errorf("system: %s: re-read state: %w: %v", CapabilitySingleInstance, ErrUnavailable, currentErr)
			}

			alive, aliveErr := singleInstanceProcessAlive(state.PID)
			if aliveErr != nil {
				return nil, fmt.Errorf("system: %s: check primary process: %w: forward=%v process=%v", CapabilitySingleInstance, ErrUnavailable, forwardErr, aliveErr)
			}
			if alive {
				return nil, fmt.Errorf("system: %s: authenticated primary unavailable: %w: %v", CapabilitySingleInstance, ErrUnavailable, forwardErr)
			}
			removed, removeErr := removeSingleInstanceStateIfOwnerContext(ctx, statePath, state.Owner)
			if removeErr != nil {
				return nil, fmt.Errorf("system: %s: remove stale state: %w: %v", CapabilitySingleInstance, ErrUnavailable, removeErr)
			}
			if removed {
				continue
			}
			if err := waitSingleInstanceRetry(ctx, deadline); err != nil {
				return nil, fmt.Errorf("system: %s: state changed while acquiring: %w: %v", CapabilitySingleInstance, ErrUnavailable, err)
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			if time.Now().Before(deadline) {
				waitErr := waitSingleInstanceRetry(ctx, deadline)
				if waitErr == nil {
					continue
				}
				if ctxErr := ctx.Err(); ctxErr != nil {
					return nil, ctxErr
				}
			}
			return nil, fmt.Errorf("system: %s: read state: %w: %v", CapabilitySingleInstance, ErrUnavailable, err)
		}

		listener, listenErr := listenLoopbackSingleInstance(ctx, opts.port)
		if listenErr != nil {
			return nil, fmt.Errorf("system: %s: listen: %w: %v", CapabilitySingleInstance, ErrUnavailable, listenErr)
		}
		secret, randomErr := singleInstanceRandomBytes(singleInstanceSecretBytes)
		if randomErr != nil {
			_ = listener.Close()
			return nil, fmt.Errorf("system: %s: generate secret: %w: %v", CapabilitySingleInstance, ErrUnavailable, randomErr)
		}
		ownerBytes, randomErr := singleInstanceRandomBytes(singleInstanceSecretBytes)
		if randomErr != nil {
			clear(secret)
			_ = listener.Close()
			return nil, fmt.Errorf("system: %s: generate owner token: %w: %v", CapabilitySingleInstance, ErrUnavailable, randomErr)
		}
		owner := base64.RawURLEncoding.EncodeToString(ownerBytes)
		clear(ownerBytes)
		state = loopbackSingleInstanceState{
			Version:  singleInstanceStateVersion,
			Protocol: singleInstanceProtocol,
			ID:       opts.id,
			Address:  listener.Addr().String(),
			Secret:   base64.RawURLEncoding.EncodeToString(secret),
			Owner:    owner,
			PID:      os.Getpid(),
		}
		claimed, claimErr := writeSingleInstanceStateExclusiveContext(ctx, statePath, state)
		if claimErr != nil {
			clear(secret)
			_ = listener.Close()
			return nil, fmt.Errorf("system: %s: publish state: %w: %v", CapabilitySingleInstance, ErrUnavailable, claimErr)
		}
		if !claimed {
			clear(secret)
			_ = listener.Close()
			continue
		}

		handle := newLoopbackSingleInstanceHandle(listener, opts, secret, statePath, owner, singleInstanceForwardingAllowed())
		clear(secret)
		go handle.acceptLoop()
		return handle, nil
	}
}

func listenLoopbackSingleInstance(ctx context.Context, port int) (net.Listener, error) {
	preferred := singleInstanceAddress(port)
	listener, preferredErr := new(net.ListenConfig).Listen(ctx, "tcp", preferred)
	if preferredErr == nil {
		return listener, nil
	}
	listener, fallbackErr := new(net.ListenConfig).Listen(ctx, "tcp", "127.0.0.1:0")
	if fallbackErr != nil {
		return nil, fmt.Errorf("preferred %s: %v; fallback: %w", preferred, preferredErr, fallbackErr)
	}
	return listener, nil
}

func singleInstanceStatePath(opts singleInstanceOptions) (string, error) {
	// The state file is a per-user capability, not a process-identity credential.
	// File modes and the inherited user-profile ACL keep it out of other users'
	// reach, but same-account processes that can read the cache can reuse the key.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(cacheDir) == "" {
		return "", errors.New("user cache directory is empty")
	}
	stateDir := filepath.Join(cacheDir, "fluxui-single-instance")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return "", err
	}
	info, err := os.Lstat(stateDir)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", fmt.Errorf("state directory %q is not a regular directory", stateDir)
	}
	if err := os.Chmod(stateDir, 0o700); err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(opts.id + "\x00" + strconv.Itoa(opts.port)))
	return filepath.Join(stateDir, hex.EncodeToString(sum[:])+".json"), nil
}

func writeSingleInstanceStateExclusive(path string, state loopbackSingleInstanceState) (claimed bool, err error) {
	return writeSingleInstanceStateExclusiveContext(context.Background(), path, state)
}

func writeSingleInstanceStateExclusiveContext(ctx context.Context, path string, state loopbackSingleInstanceState) (claimed bool, err error) {
	unlock, err := lockSingleInstanceState(ctx, path)
	if err != nil {
		return false, err
	}
	defer func() {
		if unlockErr := unlock(); err == nil && unlockErr != nil && !claimed {
			err = unlockErr
		}
	}()
	claimed, err = writeSingleInstanceStateExclusiveLocked(path, state)
	return claimed, err
}

func writeSingleInstanceStateExclusiveLocked(path string, state loopbackSingleInstanceState) (bool, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return false, err
	}
	if len(data) > singleInstanceMaxBytes {
		return false, errSingleInstanceMessageTooLarge
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".fluxui-single-instance-*.tmp")
	if err != nil {
		return false, err
	}
	tempPath := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(tempPath)
	}()
	if err := file.Chmod(0o600); err != nil {
		return false, err
	}
	n, err := file.Write(data)
	if err != nil {
		return false, err
	}
	if n != len(data) {
		return false, io.ErrShortWrite
	}
	if err := file.Sync(); err != nil {
		return false, err
	}
	if err := file.Close(); err != nil {
		return false, err
	}
	if err := os.Link(tempPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func readSingleInstanceState(path string, opts singleInstanceOptions) (loopbackSingleInstanceState, []byte, error) {
	state, err := readSingleInstanceStateRecord(path)
	if err != nil {
		return loopbackSingleInstanceState{}, nil, err
	}
	if state.Version != singleInstanceStateVersion || state.Protocol != singleInstanceProtocol {
		return loopbackSingleInstanceState{}, nil, errors.New("unsupported state protocol")
	}
	if state.ID != opts.id {
		return loopbackSingleInstanceState{}, nil, errors.New("state ID mismatch")
	}
	if state.PID <= 0 || state.Owner == "" {
		return loopbackSingleInstanceState{}, nil, errors.New("invalid state owner")
	}
	owner, err := base64.RawURLEncoding.DecodeString(state.Owner)
	if err != nil || len(owner) != singleInstanceSecretBytes {
		clear(owner)
		return loopbackSingleInstanceState{}, nil, errors.New("invalid state owner token")
	}
	clear(owner)
	secret, err := base64.RawURLEncoding.DecodeString(state.Secret)
	if err != nil || len(secret) != singleInstanceSecretBytes {
		clear(secret)
		return loopbackSingleInstanceState{}, nil, errors.New("invalid state secret")
	}
	if err := validateSingleInstanceAddress(state.Address); err != nil {
		clear(secret)
		return loopbackSingleInstanceState{}, nil, err
	}
	return state, secret, nil
}

func readSingleInstanceStateRecord(path string) (loopbackSingleInstanceState, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return loopbackSingleInstanceState{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return loopbackSingleInstanceState{}, errors.New("state path is not a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return loopbackSingleInstanceState{}, errors.New("state file permissions are too broad")
	}
	file, err := os.Open(path)
	if err != nil {
		return loopbackSingleInstanceState{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, singleInstanceMaxBytes+1))
	if err != nil {
		return loopbackSingleInstanceState{}, err
	}
	if len(data) > singleInstanceMaxBytes {
		return loopbackSingleInstanceState{}, errSingleInstanceMessageTooLarge
	}
	var state loopbackSingleInstanceState
	if err := decodeSingleInstanceJSON(data, &state); err != nil {
		return loopbackSingleInstanceState{}, err
	}
	return state, nil
}

func removeSingleInstanceStateIfOwner(path, owner string) (removed bool, err error) {
	return removeSingleInstanceStateIfOwnerContext(context.Background(), path, owner)
}

func removeSingleInstanceStateIfOwnerContext(ctx context.Context, path, owner string) (removed bool, err error) {
	if path == "" || owner == "" {
		return false, nil
	}
	unlock, err := lockSingleInstanceState(ctx, path)
	if err != nil {
		return false, err
	}
	defer func() {
		if unlockErr := unlock(); err == nil && unlockErr != nil && !removed {
			err = unlockErr
		}
	}()
	removed, err = removeSingleInstanceStateIfOwnerLocked(path, owner)
	return removed, err
}

func removeSingleInstanceStateIfOwnerLocked(path, owner string) (bool, error) {
	state, err := readSingleInstanceStateRecord(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if state.Owner != owner {
		return false, nil
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateSingleInstanceAddress(address string) error {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return errors.New("invalid state address")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("state address is not loopback")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port <= 0 || port > 65535 {
		return errors.New("invalid state port")
	}
	return nil
}

func validateSingleInstanceChallenge(challenge loopbackSingleInstanceChallenge) error {
	if challenge.Protocol != singleInstanceProtocol {
		return errors.New("protocol mismatch")
	}
	nonce, err := base64.RawURLEncoding.DecodeString(challenge.Nonce)
	if err != nil || len(nonce) != singleInstanceSecretBytes {
		clear(nonce)
		return errors.New("invalid nonce")
	}
	clear(nonce)
	return nil
}

func singleInstanceRequestMAC(secret []byte, challenge string, msg loopbackSingleInstanceMessage) string {
	payload := struct {
		Challenge string   `json:"challenge"`
		Protocol  string   `json:"protocol"`
		ID        string   `json:"id"`
		Args      []string `json:"args,omitempty"`
		Payload   string   `json:"payload,omitempty"`
	}{
		Challenge: challenge,
		Protocol:  msg.Protocol,
		ID:        msg.ID,
		Args:      msg.Args,
		Payload:   msg.Payload,
	}
	data, _ := json.Marshal(payload)
	return singleInstanceMAC(secret, data)
}

func verifySingleInstanceRequestMAC(secret []byte, challenge string, msg loopbackSingleInstanceMessage) bool {
	expected := singleInstanceRequestMAC(secret, challenge, msg)
	return equalSingleInstanceMAC(expected, msg.MAC)
}

func singleInstanceResponseMAC(secret []byte, challenge string, response loopbackSingleInstanceResponse) string {
	payload := struct {
		Challenge string `json:"challenge"`
		Status    string `json:"status"`
		Error     string `json:"error,omitempty"`
	}{
		Challenge: challenge,
		Status:    response.Status,
		Error:     response.Error,
	}
	data, _ := json.Marshal(payload)
	return singleInstanceMAC(secret, data)
}

func verifySingleInstanceResponseMAC(secret []byte, challenge string, response loopbackSingleInstanceResponse) bool {
	expected := singleInstanceResponseMAC(secret, challenge, response)
	return equalSingleInstanceMAC(expected, response.MAC)
}

func singleInstanceMAC(secret, data []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(data)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func equalSingleInstanceMAC(expected, actual string) bool {
	expectedBytes, expectedErr := base64.RawURLEncoding.DecodeString(expected)
	actualBytes, actualErr := base64.RawURLEncoding.DecodeString(actual)
	defer clear(expectedBytes)
	defer clear(actualBytes)
	if expectedErr != nil || actualErr != nil {
		return false
	}
	return hmac.Equal(expectedBytes, actualBytes)
}

func writeSingleInstanceFrame(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(data) > singleInstanceMaxBytes {
		return errSingleInstanceMessageTooLarge
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))
	if err := writeSingleInstanceBytes(w, header[:]); err != nil {
		return err
	}
	return writeSingleInstanceBytes(w, data)
}

func writeSingleInstanceBytes(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func readSingleInstanceFrame(r io.Reader, value any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size == 0 {
		return errors.New("empty single-instance message")
	}
	if size > singleInstanceMaxBytes {
		return errSingleInstanceMessageTooLarge
	}
	data := make([]byte, int(size))
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}
	return decodeSingleInstanceJSON(data, value)
}

func decodeSingleInstanceJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func singleInstanceRandomBytes(size int) ([]byte, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		clear(data)
		return nil, err
	}
	return data, nil
}

func waitSingleInstanceRetry(ctx context.Context, deadline time.Time) error {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return context.DeadlineExceeded
	}
	delay := singleInstanceRetryInterval
	if remaining < delay {
		delay = remaining
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func singleInstanceClosedError() error {
	return fmt.Errorf("system: %s: %w", CapabilitySingleInstance, ErrClosed)
}
