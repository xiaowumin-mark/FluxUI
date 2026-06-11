//go:build windows

package system

import (
	"context"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsToastActivatorConstants(t *testing.T) {
	if iidNotificationActivationCallback.Data1 != 0x53e31837 ||
		iidNotificationActivationCallback.Data2 != 0x6600 ||
		iidNotificationActivationCallback.Data3 != 0x4a81 ||
		iidNotificationActivationCallback.Data4 != [8]byte{0x93, 0x95, 0x75, 0xcf, 0xfe, 0x74, 0x6f, 0x94} {
		t.Fatalf("unexpected INotificationActivationCallback IID: %#v", iidNotificationActivationCallback)
	}
}

func TestWindowsToastActivationUserInput(t *testing.T) {
	key, err := windows.UTF16PtrFromString("reply")
	if err != nil {
		t.Fatal(err)
	}
	value, err := windows.UTF16PtrFromString("ok")
	if err != nil {
		t.Fatal(err)
	}
	input := []windowsNotificationUserInputData{{
		key:   key,
		value: value,
	}}

	got := windowsToastActivationUserInput(&input[0], uint32(len(input)))
	if got["reply"] != "ok" {
		t.Fatalf("unexpected user input: %#v", got)
	}
}

func TestWindowsToastActivatorClassFactoryDispatch(t *testing.T) {
	clsid, err := windows.GUIDFromString("{01234567-89AB-CDEF-0123-456789ABCDEF}")
	if err != nil {
		t.Fatal(err)
	}
	events := make(chan ToastActivationEvent, 1)
	server := newWindowsToastActivatorServer(clsid, func(event ToastActivationEvent) {
		events <- event
	})

	var raw unsafe.Pointer
	hr := windowsToastClassFactoryCreateInstance(
		server.factory,
		0,
		&iidNotificationActivationCallback,
		&raw,
	)
	if hr != hresultOK || raw == nil {
		t.Fatalf("expected class factory to create activation callback, hr=0x%08x raw=%p", uint32(hr), raw)
	}
	callback := (*windowsToastActivationCallback)(raw)
	defer windowsToastActivationRelease(callback)

	appID, err := windows.UTF16PtrFromString("com.example.FluxUI")
	if err != nil {
		t.Fatal(err)
	}
	args, err := windows.UTF16PtrFromString("fluxui:action:open")
	if err != nil {
		t.Fatal(err)
	}
	hr = windowsToastActivationActivate(callback, appID, args, nil, 0)
	if hr != hresultOK {
		t.Fatalf("unexpected activation HRESULT: 0x%08x", uint32(hr))
	}

	select {
	case event := <-events:
		if event.AppID != "com.example.FluxUI" || event.Arguments != "fluxui:action:open" {
			t.Fatalf("unexpected activation event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for activation event")
	}
}

func TestWindowsStartToastActivatorLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	activator, err := StartToastActivator(ctx, "{12345678-89AB-CDEF-0123-456789ABCDEF}", func(ToastActivationEvent) {})
	if err != nil {
		t.Fatalf("unexpected activator start error: %v", err)
	}
	if activator == nil {
		t.Fatal("expected activator")
	}
	if err := activator.Close(); err != nil {
		t.Fatalf("unexpected activator close error: %v", err)
	}
	select {
	case <-activator.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for activator shutdown")
	}
}

func TestWindowsStartToastActivatorCoCreateDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clsidText := "{87654321-89AB-CDEF-0123-456789ABCDEF}"
	events := make(chan ToastActivationEvent, 1)
	activator, err := StartToastActivator(ctx, clsidText, func(event ToastActivationEvent) {
		events <- event
	})
	if err != nil {
		t.Fatalf("unexpected activator start error: %v", err)
	}
	defer activator.Close()

	initialized, err := initCOMMTA()
	if err != nil {
		t.Fatalf("initialize test COM apartment: %v", err)
	}
	if initialized {
		defer windows.CoUninitialize()
	}

	clsid, err := windows.GUIDFromString(clsidText)
	if err != nil {
		t.Fatal(err)
	}
	var callback *windowsToastActivationCallback
	r, _, callErr := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsid)),
		0,
		clsctxLocalServer,
		uintptr(unsafe.Pointer(&iidNotificationActivationCallback)),
		uintptr(unsafe.Pointer(&callback)),
	)
	if err := hresultError(r); err != nil {
		t.Fatalf("CoCreateInstance failed: hr=0x%08x err=%v callErr=%v", uint32(r), err, callErr)
	}
	if callback == nil {
		t.Fatal("expected activation callback instance")
	}
	defer windowsToastActivationRelease(callback)

	appID, err := windows.UTF16PtrFromString("com.example.FluxUI")
	if err != nil {
		t.Fatal(err)
	}
	args, err := windows.UTF16PtrFromString("fluxui:click:reports")
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ = syscall.SyscallN(
		callback.lpVtbl.Activate,
		uintptr(unsafe.Pointer(callback)),
		uintptr(unsafe.Pointer(appID)),
		uintptr(unsafe.Pointer(args)),
		0,
		0,
	)
	if err := hresultError(r); err != nil {
		t.Fatalf("Activate failed: hr=0x%08x err=%v", uint32(r), err)
	}

	select {
	case event := <-events:
		if event.AppID != "com.example.FluxUI" || event.Arguments != "fluxui:click:reports" {
			t.Fatalf("unexpected activation event: %#v", event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for COM activation event")
	}
}
