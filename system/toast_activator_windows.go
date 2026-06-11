//go:build windows

package system

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	clsctxLocalServer = 0x4

	coinitMultiThreaded = 0x0

	regclsMultipleUse = 0x1

	hresultOK               = 0x00000000
	hresultNoInterface      = 0x80004002
	hresultPointer          = 0x80004003
	hresultClassNoAggregate = 0x80040110
)

var (
	procCoRegisterClassObject = ole32.NewProc("CoRegisterClassObject")
	procCoRevokeClassObject   = ole32.NewProc("CoRevokeClassObject")

	iidIUnknown = windows.GUID{
		Data1: 0x00000000,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	iidIClassFactory = windows.GUID{
		Data1: 0x00000001,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	iidNotificationActivationCallback = windows.GUID{
		Data1: 0x53e31837,
		Data2: 0x6600,
		Data3: 0x4a81,
		Data4: [8]byte{0x93, 0x95, 0x75, 0xcf, 0xfe, 0x74, 0x6f, 0x94},
	}

	windowsToastClassFactoryQueryInterfaceProc = syscall.NewCallback(windowsToastClassFactoryQueryInterface)
	windowsToastClassFactoryAddRefProc         = syscall.NewCallback(windowsToastClassFactoryAddRef)
	windowsToastClassFactoryReleaseProc        = syscall.NewCallback(windowsToastClassFactoryRelease)
	windowsToastClassFactoryCreateInstanceProc = syscall.NewCallback(windowsToastClassFactoryCreateInstance)
	windowsToastClassFactoryLockServerProc     = syscall.NewCallback(windowsToastClassFactoryLockServer)
	windowsToastClassFactoryVtblValue          = windowsToastClassFactoryVtbl{
		QueryInterface: windowsToastClassFactoryQueryInterfaceProc,
		AddRef:         windowsToastClassFactoryAddRefProc,
		Release:        windowsToastClassFactoryReleaseProc,
		CreateInstance: windowsToastClassFactoryCreateInstanceProc,
		LockServer:     windowsToastClassFactoryLockServerProc,
	}

	windowsToastActivationQueryInterfaceProc = syscall.NewCallback(windowsToastActivationQueryInterface)
	windowsToastActivationAddRefProc         = syscall.NewCallback(windowsToastActivationAddRef)
	windowsToastActivationReleaseProc        = syscall.NewCallback(windowsToastActivationRelease)
	windowsToastActivationActivateProc       = syscall.NewCallback(windowsToastActivationActivate)
	windowsToastActivationVtblValue          = windowsToastActivationCallbackVtbl{
		QueryInterface: windowsToastActivationQueryInterfaceProc,
		AddRef:         windowsToastActivationAddRefProc,
		Release:        windowsToastActivationReleaseProc,
		Activate:       windowsToastActivationActivateProc,
	}
)

type windowsToastClassFactory struct {
	lpVtbl *windowsToastClassFactoryVtbl
	refs   int32
	server *windowsToastActivatorServer
}

type windowsToastClassFactoryVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	CreateInstance uintptr
	LockServer     uintptr
}

type windowsToastActivationCallback struct {
	lpVtbl *windowsToastActivationCallbackVtbl
	refs   int32
	server *windowsToastActivatorServer
}

type windowsToastActivationCallbackVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Activate       uintptr
}

type windowsNotificationUserInputData struct {
	key   *uint16
	value *uint16
}

type windowsToastActivatorServer struct {
	clsid   windows.GUID
	handler func(ToastActivationEvent)
	factory *windowsToastClassFactory
	mu      sync.Mutex
	objects map[*windowsToastActivationCallback]struct{}
}

func (windowsDriver) startToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) (toastActivatorHandle, error) {
	parsed, err := windowsToastShortcutActivatorCLSID(clsid)
	if err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	ready := make(chan error, 1)
	server := newWindowsToastActivatorServer(parsed, fn)

	go server.run(runCtx, ready, done)

	select {
	case err := <-ready:
		if err != nil {
			cancel()
			<-done
			return nil, err
		}
		return &simpleToastActivatorHandle{cancel: cancel, done: done}, nil
	case <-ctx.Done():
		cancel()
		<-done
		return nil, ctx.Err()
	}
}

func newWindowsToastActivatorServer(clsid windows.GUID, fn func(ToastActivationEvent)) *windowsToastActivatorServer {
	server := &windowsToastActivatorServer{
		clsid:   clsid,
		handler: fn,
		objects: make(map[*windowsToastActivationCallback]struct{}),
	}
	server.factory = &windowsToastClassFactory{
		lpVtbl: &windowsToastClassFactoryVtblValue,
		refs:   1,
		server: server,
	}
	return server
}

func (s *windowsToastActivatorServer) run(ctx context.Context, ready chan<- error, done chan<- struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(done)

	initialized, err := initCOMMTA()
	if err != nil {
		ready <- fmt.Errorf("system: %s: initialize toast activator COM: %w", CapabilityNotification, err)
		return
	}
	if initialized {
		defer windows.CoUninitialize()
	}

	var cookie uint32
	r, _, callErr := procCoRegisterClassObject.Call(
		uintptr(unsafe.Pointer(&s.clsid)),
		uintptr(unsafe.Pointer(s.factory)),
		clsctxLocalServer,
		regclsMultipleUse,
		uintptr(unsafe.Pointer(&cookie)),
	)
	if err := hresultError(r); err != nil {
		if callErr != syscall.Errno(0) {
			ready <- fmt.Errorf("system: %s: register toast activator COM class: %w: %v", CapabilityNotification, ErrUnavailable, callErr)
			return
		}
		ready <- fmt.Errorf("system: %s: register toast activator COM class: %w: %v", CapabilityNotification, ErrUnavailable, err)
		return
	}
	ready <- nil

	<-ctx.Done()
	if cookie != 0 {
		procCoRevokeClassObject.Call(uintptr(cookie))
	}
}

func initCOMMTA() (bool, error) {
	err := windows.CoInitializeEx(0, coinitMultiThreaded)
	if err == nil {
		return true, nil
	}
	if errno, ok := err.(syscall.Errno); ok && uintptr(errno) == sFalse {
		return false, nil
	}
	return false, err
}

func (s *windowsToastActivatorServer) newActivationCallback() *windowsToastActivationCallback {
	obj := &windowsToastActivationCallback{
		lpVtbl: &windowsToastActivationVtblValue,
		server: s,
	}
	s.mu.Lock()
	s.objects[obj] = struct{}{}
	s.mu.Unlock()
	return obj
}

func (s *windowsToastActivatorServer) removeActivationCallback(obj *windowsToastActivationCallback) {
	s.mu.Lock()
	delete(s.objects, obj)
	s.mu.Unlock()
}

func windowsToastClassFactoryQueryInterface(this *windowsToastClassFactory, riid *windows.GUID, out *unsafe.Pointer) uintptr {
	if out == nil {
		return hresultPointer
	}
	*out = nil
	if riid == nil || (!windowsGUIDEqual(riid, &iidIUnknown) && !windowsGUIDEqual(riid, &iidIClassFactory)) {
		return hresultNoInterface
	}
	*out = unsafe.Pointer(this)
	windowsToastClassFactoryAddRef(this)
	return hresultOK
}

func windowsToastClassFactoryAddRef(this *windowsToastClassFactory) uintptr {
	if this == nil {
		return 0
	}
	return uintptr(atomic.AddInt32(&this.refs, 1))
}

func windowsToastClassFactoryRelease(this *windowsToastClassFactory) uintptr {
	if this == nil {
		return 0
	}
	next := atomic.AddInt32(&this.refs, -1)
	if next < 0 {
		atomic.StoreInt32(&this.refs, 0)
		return 0
	}
	return uintptr(next)
}

func windowsToastClassFactoryCreateInstance(this *windowsToastClassFactory, outer uintptr, riid *windows.GUID, out *unsafe.Pointer) uintptr {
	if out == nil {
		return hresultPointer
	}
	*out = nil
	if outer != 0 {
		return hresultClassNoAggregate
	}
	if this == nil || this.server == nil {
		return hresultNoInterface
	}
	obj := this.server.newActivationCallback()
	hr := windowsToastActivationQueryInterface(obj, riid, out)
	if hr != hresultOK {
		this.server.removeActivationCallback(obj)
	}
	return hr
}

func windowsToastClassFactoryLockServer(this *windowsToastClassFactory, lock uintptr) uintptr {
	if lock != 0 {
		windowsToastClassFactoryAddRef(this)
	} else {
		windowsToastClassFactoryRelease(this)
	}
	return hresultOK
}

func windowsToastActivationQueryInterface(this *windowsToastActivationCallback, riid *windows.GUID, out *unsafe.Pointer) uintptr {
	if out == nil {
		return hresultPointer
	}
	*out = nil
	if riid == nil || (!windowsGUIDEqual(riid, &iidIUnknown) && !windowsGUIDEqual(riid, &iidNotificationActivationCallback)) {
		return hresultNoInterface
	}
	*out = unsafe.Pointer(this)
	windowsToastActivationAddRef(this)
	return hresultOK
}

func windowsToastActivationAddRef(this *windowsToastActivationCallback) uintptr {
	if this == nil {
		return 0
	}
	return uintptr(atomic.AddInt32(&this.refs, 1))
}

func windowsToastActivationRelease(this *windowsToastActivationCallback) uintptr {
	if this == nil {
		return 0
	}
	next := atomic.AddInt32(&this.refs, -1)
	if next <= 0 {
		atomic.StoreInt32(&this.refs, 0)
		if this.server != nil {
			this.server.removeActivationCallback(this)
		}
		return 0
	}
	return uintptr(next)
}

func windowsToastActivationActivate(this *windowsToastActivationCallback, appIDPtr *uint16, argsPtr *uint16, inputPtr *windowsNotificationUserInputData, inputCount uintptr) uintptr {
	if this == nil || this.server == nil || this.server.handler == nil {
		return hresultOK
	}
	event := ToastActivationEvent{
		AppID:     windowsUTF16PtrToString(appIDPtr),
		Arguments: windowsUTF16PtrToString(argsPtr),
		UserInput: windowsToastActivationUserInput(inputPtr, uint32(inputCount)),
	}
	handler := this.server.handler
	go func() {
		defer func() {
			_ = recover()
		}()
		handler(event)
	}()
	return hresultOK
}

func windowsToastActivationUserInput(inputPtr *windowsNotificationUserInputData, count uint32) map[string]string {
	if inputPtr == nil || count == 0 {
		return nil
	}
	if count > 128 {
		count = 128
	}
	items := unsafe.Slice(inputPtr, int(count))
	values := make(map[string]string, len(items))
	for _, item := range items {
		key := windowsUTF16PtrToString(item.key)
		if key == "" {
			continue
		}
		values[key] = windowsUTF16PtrToString(item.value)
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func windowsUTF16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	return windows.UTF16PtrToString(ptr)
}

func windowsGUIDEqual(a, b *windows.GUID) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Data1 == b.Data1 &&
		a.Data2 == b.Data2 &&
		a.Data3 == b.Data3 &&
		a.Data4 == b.Data4
}
