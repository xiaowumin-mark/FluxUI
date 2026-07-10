//go:build windows

package system

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"net/netip"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	singleInstanceTCPTableOwnerPIDListener = 3
	singleInstanceMIBTCPStateListen        = 2
	singleInstanceTCPRowOwnerPIDBytes      = 24
)

var singleInstanceGetExtendedTCPTable = windows.NewLazySystemDLL("iphlpapi.dll").NewProc("GetExtendedTcpTable")

func lockSingleInstanceState(ctx context.Context, path string) (func() error, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline := singleInstanceDeadline(ctx)
	file, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	var overlapped windows.Overlapped
	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		err = windows.LockFileEx(
			windows.Handle(file.Fd()),
			windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
			0,
			1,
			0,
			&overlapped,
		)
		if err == nil {
			break
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			_ = file.Close()
			return nil, err
		}
		if err := waitSingleInstanceRetry(ctx, deadline); err != nil {
			_ = file.Close()
			return nil, err
		}
	}
	return func() error {
		unlockErr := windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &overlapped)
		closeErr := file.Close()
		if unlockErr != nil {
			return unlockErr
		}
		return closeErr
	}, nil
}

type windowsSingleInstanceStateVerifier struct {
	listenerProcessID      func(string) (int, error)
	processElevated        func(int) (bool, error)
	processImageIdentity   func(int) (windowsSingleInstanceImageIdentity, error)
	currentProcessID       func() int
	currentProcessElevated func() bool
}

type windowsSingleInstanceImageIdentity struct {
	volumeSerialNumber uint32
	fileIndexHigh      uint32
	fileIndexLow       uint32
}

func singleInstanceProcessAlive(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return false, nil
		}
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return true, nil
		}
		return false, fmt.Errorf("open process %d: %w", pid, err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	status, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false, fmt.Errorf("query process %d: %w", pid, err)
	}
	switch status {
	case windows.WAIT_OBJECT_0:
		return false, nil
	case uint32(windows.WAIT_TIMEOUT):
		return true, nil
	default:
		return false, fmt.Errorf("query process %d: unexpected wait status %d", pid, status)
	}
}

func singleInstanceForwardingAllowed() bool {
	return !windows.GetCurrentProcessToken().IsElevated()
}

func singleInstancePublishedStateTrusted(state loopbackSingleInstanceState) (bool, error) {
	return verifyWindowsSingleInstancePublishedState(state, windowsSingleInstanceStateVerifier{
		listenerProcessID:      windowsSingleInstanceListenerProcessID,
		processElevated:        windowsSingleInstanceProcessElevated,
		processImageIdentity:   windowsSingleInstanceProcessImageIdentity,
		currentProcessID:       os.Getpid,
		currentProcessElevated: func() bool { return windows.GetCurrentProcessToken().IsElevated() },
	})
}

func verifyWindowsSingleInstancePublishedState(state loopbackSingleInstanceState, verifier windowsSingleInstanceStateVerifier) (bool, error) {
	if verifier.listenerProcessID == nil || verifier.processElevated == nil || verifier.processImageIdentity == nil ||
		verifier.currentProcessID == nil || verifier.currentProcessElevated == nil {
		return false, errors.New("single-instance state verifier is incomplete")
	}
	listenerPID, err := verifier.listenerProcessID(state.Address)
	if err != nil {
		return false, fmt.Errorf("query listener owner: %w", err)
	}
	if listenerPID <= 0 || listenerPID != state.PID {
		return false, nil
	}
	if !verifier.currentProcessElevated() {
		return true, nil
	}
	targetElevated, err := verifier.processElevated(state.PID)
	if err != nil || !targetElevated {
		// The state controls the target PID. Treat target-specific query failures
		// as an untrusted claim so they cannot become a pre-created elevated DoS.
		return false, nil
	}
	currentIdentity, err := verifier.processImageIdentity(verifier.currentProcessID())
	if err != nil {
		return false, fmt.Errorf("query current process image identity: %w", err)
	}
	targetIdentity, err := verifier.processImageIdentity(state.PID)
	if err != nil {
		return false, nil
	}
	return targetIdentity == currentIdentity, nil
}

func windowsSingleInstanceProcessElevated(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return false, nil
		}
		return false, fmt.Errorf("open process %d: %w", pid, err)
	}
	defer func() { _ = windows.CloseHandle(process) }()

	var token windows.Token
	if err := windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token); err != nil {
		return false, fmt.Errorf("open process %d token: %w", pid, err)
	}
	defer token.Close()
	return token.IsElevated(), nil
}

func windowsSingleInstanceProcessImageIdentity(pid int) (windowsSingleInstanceImageIdentity, error) {
	if pid <= 0 {
		return windowsSingleInstanceImageIdentity{}, errors.New("process ID is invalid")
	}
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return windowsSingleInstanceImageIdentity{}, fmt.Errorf("open process %d: %w", pid, err)
	}
	defer func() { _ = windows.CloseHandle(process) }()

	path, err := windowsSingleInstanceProcessImagePath(process)
	if err != nil {
		return windowsSingleInstanceImageIdentity{}, fmt.Errorf("query process %d image path: %w", pid, err)
	}
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windowsSingleInstanceImageIdentity{}, fmt.Errorf("encode process %d image path: %w", pid, err)
	}
	file, err := windows.CreateFile(
		pathPtr,
		windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return windowsSingleInstanceImageIdentity{}, fmt.Errorf("open process %d image: %w", pid, err)
	}
	defer func() { _ = windows.CloseHandle(file) }()

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(file, &info); err != nil {
		return windowsSingleInstanceImageIdentity{}, fmt.Errorf("query process %d image file identity: %w", pid, err)
	}
	return windowsSingleInstanceImageIdentity{
		volumeSerialNumber: info.VolumeSerialNumber,
		fileIndexHigh:      info.FileIndexHigh,
		fileIndexLow:       info.FileIndexLow,
	}, nil
}

func windowsSingleInstanceProcessImagePath(process windows.Handle) (string, error) {
	const maxPathChars = 32 * 1024
	for size := uint32(windows.MAX_PATH); size <= maxPathChars; {
		buffer := make([]uint16, int(size))
		length := size
		err := windows.QueryFullProcessImageName(process, 0, &buffer[0], &length)
		if err == nil {
			if length == 0 || length > uint32(len(buffer)) {
				return "", errors.New("process image path is empty or truncated")
			}
			return windows.UTF16ToString(buffer[:length]), nil
		}
		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) || size == maxPathChars {
			return "", err
		}
		size *= 2
		if size > maxPathChars {
			size = maxPathChars
		}
	}
	return "", errors.New("process image path exceeds Windows limit")
}

func windowsSingleInstanceListenerProcessID(address string) (int, error) {
	target, err := netip.ParseAddrPort(address)
	if err != nil || !target.Addr().Unmap().Is4() {
		return 0, fmt.Errorf("invalid IPv4 listener address %q", address)
	}
	table, err := windowsSingleInstanceTCPListenerTable()
	if err != nil {
		return 0, err
	}
	if len(table) < 4 {
		return 0, errors.New("TCP listener table is truncated")
	}
	count := int(binary.LittleEndian.Uint32(table[:4]))
	if count > (len(table)-4)/singleInstanceTCPRowOwnerPIDBytes {
		return 0, errors.New("TCP listener table row count is invalid")
	}
	targetAddress := target.Addr().Unmap()
	targetPort := target.Port()
	for index := 0; index < count; index++ {
		offset := 4 + index*singleInstanceTCPRowOwnerPIDBytes
		row := table[offset : offset+singleInstanceTCPRowOwnerPIDBytes]
		if binary.LittleEndian.Uint32(row[0:4]) != singleInstanceMIBTCPStateListen {
			continue
		}
		addressValue := binary.LittleEndian.Uint32(row[4:8])
		localAddress := netip.AddrFrom4([4]byte{
			byte(addressValue),
			byte(addressValue >> 8),
			byte(addressValue >> 16),
			byte(addressValue >> 24),
		})
		localPort := bits.ReverseBytes16(uint16(binary.LittleEndian.Uint32(row[8:12])))
		if localAddress == targetAddress && localPort == targetPort {
			return int(binary.LittleEndian.Uint32(row[20:24])), nil
		}
	}
	return 0, nil
}

func windowsSingleInstanceTCPListenerTable() ([]byte, error) {
	if err := singleInstanceGetExtendedTCPTable.Find(); err != nil {
		return nil, err
	}
	var size uint32
	status, _, _ := singleInstanceGetExtendedTCPTable.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
		1,
		windows.AF_INET,
		singleInstanceTCPTableOwnerPIDListener,
		0,
	)
	if status != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
		if status == 0 && size == 0 {
			return []byte{0, 0, 0, 0}, nil
		}
		return nil, syscall.Errno(status)
	}

	for attempts := 0; attempts < 3; attempts++ {
		if size < 4 {
			size = 4
		}
		table := make([]byte, int(size))
		status, _, _ = singleInstanceGetExtendedTCPTable.Call(
			uintptr(unsafe.Pointer(&table[0])),
			uintptr(unsafe.Pointer(&size)),
			1,
			windows.AF_INET,
			singleInstanceTCPTableOwnerPIDListener,
			0,
		)
		if status == 0 {
			return table, nil
		}
		if status != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
			return nil, syscall.Errno(status)
		}
	}
	return nil, errors.New("TCP listener table changed repeatedly")
}
