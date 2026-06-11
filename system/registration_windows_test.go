//go:build windows

package system

import (
	"path/filepath"
	"testing"
)

func TestWindowsRegistrationPaths(t *testing.T) {
	if got := windowsProtocolRegistryPath("fluxui"); got != `Software\Classes\fluxui` {
		t.Fatalf("unexpected protocol path: %q", got)
	}
	if got := windowsFileExtensionRegistryPath(".flux"); got != `Software\Classes\.flux` {
		t.Fatalf("unexpected extension path: %q", got)
	}
	if got := windowsProgIDRegistryPath("FluxUI.Document"); got != `Software\Classes\FluxUI.Document` {
		t.Fatalf("unexpected progID path: %q", got)
	}
	if got := windowsToastActivatorRegistryPath("{01234567-89AB-CDEF-0123-456789ABCDEF}"); got != `Software\Classes\CLSID\{01234567-89AB-CDEF-0123-456789ABCDEF}` {
		t.Fatalf("unexpected toast activator path: %q", got)
	}
}

func TestWindowsRegistrationCOMConstants(t *testing.T) {
	if iidPropertyStore.Data1 != 0x886d8eeb ||
		iidPropertyStore.Data2 != 0x8cf2 ||
		iidPropertyStore.Data3 != 0x4446 ||
		iidPropertyStore.Data4 != [8]byte{0x8d, 0x02, 0xcd, 0xba, 0x1d, 0xbd, 0xcf, 0x99} {
		t.Fatalf("unexpected IPropertyStore IID: %#v", iidPropertyStore)
	}
	if pkeyAppUserModelID.pid != 5 || pkeyAppUserModelID.fmtid.Data1 != 0x9f4c2855 {
		t.Fatalf("unexpected AppUserModelID property key: %#v", pkeyAppUserModelID)
	}
	if pkeyAppUserModelToastActivatorCLSID.pid != 26 || pkeyAppUserModelToastActivatorCLSID.fmtid != pkeyAppUserModelID.fmtid {
		t.Fatalf("unexpected ToastActivatorCLSID property key: %#v", pkeyAppUserModelToastActivatorCLSID)
	}
	if vtCLSID != 72 {
		t.Fatalf("unexpected VT_CLSID value: %d", vtCLSID)
	}
}

func TestWindowsRegistrationDisplayNames(t *testing.T) {
	if got := protocolDisplayName("fluxui", registrationOptions{}); got != "URL:fluxui Protocol" {
		t.Fatalf("unexpected protocol display name: %q", got)
	}
	if got := protocolDisplayName("fluxui", registrationOptions{displayName: "FluxUI"}); got != "FluxUI" {
		t.Fatalf("unexpected custom protocol display name: %q", got)
	}
	if got := fileAssociationDisplayName(".flux", "FluxUI.Document", registrationOptions{}); got != "FluxUI.Document file (.flux)" {
		t.Fatalf("unexpected file association display name: %q", got)
	}
}

func TestWindowsToastShortcutNameNormalization(t *testing.T) {
	got, err := normalizeToastShortcutName("FluxUI")
	if err != nil {
		t.Fatalf("unexpected normalize error: %v", err)
	}
	if got != "FluxUI.lnk" {
		t.Fatalf("unexpected shortcut name: %q", got)
	}

	got, err = normalizeToastShortcutName("FluxUI.lnk")
	if err != nil {
		t.Fatalf("unexpected normalize .lnk error: %v", err)
	}
	if got != "FluxUI.lnk" {
		t.Fatalf("unexpected normalized .lnk name: %q", got)
	}

	if _, err := normalizeToastShortcutName(`bad/name`); err == nil {
		t.Fatal("expected invalid shortcut name to fail")
	}
}

func TestWindowsToastShortcutArguments(t *testing.T) {
	args := windowsToastShortcutArguments([]string{
		"plain",
		"two words",
		`quote"here`,
		`C:\Path With Space\`,
		"",
	})
	want := `plain "two words" "quote\"here" "C:\Path With Space\\" ""`
	if args != want {
		t.Fatalf("unexpected quoted arguments:\nwant %q\n got %q", want, args)
	}
}

func TestWindowsToastShortcutIconLocation(t *testing.T) {
	tests := []struct {
		name      string
		icon      string
		wantPath  string
		wantIndex int32
	}{
		{name: "path", icon: `C:\Program Files\FluxUI\app.exe`, wantPath: `C:\Program Files\FluxUI\app.exe`},
		{name: "resource", icon: `C:\Program Files\FluxUI\app.exe,4`, wantPath: `C:\Program Files\FluxUI\app.exe`, wantIndex: 4},
		{name: "quoted resource", icon: `"C:\Program Files\FluxUI\app.exe",2`, wantPath: `C:\Program Files\FluxUI\app.exe`, wantIndex: 2},
		{name: "comma path resource", icon: `"C:\Program Files\Flux,UI\app.exe",7`, wantPath: `C:\Program Files\Flux,UI\app.exe`, wantIndex: 7},
		{name: "invalid index", icon: `C:\Program Files\FluxUI\app.exe,index`, wantPath: `C:\Program Files\FluxUI\app.exe,index`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotIndex := windowsToastShortcutIconLocation(tt.icon)
			if gotPath != tt.wantPath || gotIndex != tt.wantIndex {
				t.Fatalf("unexpected icon location: path=%q index=%d", gotPath, gotIndex)
			}
		})
	}
}

func TestWindowsToastShortcutActivatorCLSID(t *testing.T) {
	clsid, err := windowsToastShortcutActivatorCLSID("{01234567-89AB-CDEF-0123-456789ABCDEF}")
	if err != nil {
		t.Fatalf("unexpected CLSID parse error: %v", err)
	}
	if clsid.Data1 != 0x01234567 || clsid.Data2 != 0x89ab || clsid.Data3 != 0xcdef ||
		clsid.Data4 != [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef} {
		t.Fatalf("unexpected parsed CLSID: %#v", clsid)
	}

	if _, err := windowsToastShortcutActivatorCLSID("not-a-guid"); err == nil || !IsInvalidTarget(err) {
		t.Fatalf("expected invalid CLSID to return ErrInvalidTarget, got %v", err)
	}
}

func TestWindowsToastShortcutExecutableMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.exe")
	_, err := windowsResolveToastShortcutExecutable(missing)
	if err == nil {
		t.Fatal("expected missing executable to fail")
	}
	if !IsUnavailable(err) || !IsTargetNotFound(err) {
		t.Fatalf("expected unavailable target-not-found, got %v", err)
	}
}
