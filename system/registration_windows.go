//go:build windows

package system

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	windowsClassesRootPath = `Software\Classes`
	windowsStartupRunPath  = `Software\Microsoft\Windows\CurrentVersion\Run`

	vtLPWSTR = 31
	vtCLSID  = 72
)

var (
	clsidShellLink     = windows.GUID{Data1: 0x00021401, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidShellLinkW      = windows.GUID{Data1: 0x000214f9, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidPersistFile     = windows.GUID{Data1: 0x0000010b, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidPropertyStore   = windows.GUID{Data1: 0x886d8eeb, Data2: 0x8cf2, Data3: 0x4446, Data4: [8]byte{0x8d, 0x02, 0xcd, 0xba, 0x1d, 0xbd, 0xcf, 0x99}}
	pkeyAppUserModelID = propertyKey{
		fmtid: windows.GUID{Data1: 0x9f4c2855, Data2: 0x9f79, Data3: 0x4b39, Data4: [8]byte{0xa8, 0xd0, 0xe1, 0xd4, 0x2d, 0xe1, 0xd5, 0xf3}},
		pid:   5,
	}
	pkeyAppUserModelToastActivatorCLSID = propertyKey{
		fmtid: windows.GUID{Data1: 0x9f4c2855, Data2: 0x9f79, Data3: 0x4b39, Data4: [8]byte{0xa8, 0xd0, 0xe1, 0xd4, 0x2d, 0xe1, 0xd5, 0xf3}},
		pid:   26,
	}
)

type iShellLinkW struct {
	lpVtbl *iShellLinkWVtbl
}

type iShellLinkWVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	GetPath             uintptr
	GetIDList           uintptr
	SetIDList           uintptr
	GetDescription      uintptr
	SetDescription      uintptr
	GetWorkingDirectory uintptr
	SetWorkingDirectory uintptr
	GetArguments        uintptr
	SetArguments        uintptr
	GetHotkey           uintptr
	SetHotkey           uintptr
	GetShowCmd          uintptr
	SetShowCmd          uintptr
	GetIconLocation     uintptr
	SetIconLocation     uintptr
	SetRelativePath     uintptr
	Resolve             uintptr
	SetPath             uintptr
}

type iPersistFile struct {
	lpVtbl *iPersistFileVtbl
}

type iPersistFileVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetClassID     uintptr
	IsDirty        uintptr
	Load           uintptr
	Save           uintptr
	SaveCompleted  uintptr
	GetCurFile     uintptr
}

type iPropertyStore struct {
	lpVtbl *iPropertyStoreVtbl
}

type iPropertyStoreVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetCount       uintptr
	GetAt          uintptr
	GetValue       uintptr
	SetValue       uintptr
	Commit         uintptr
}

type propertyKey struct {
	fmtid windows.GUID
	pid   uint32
}

type propVariant struct {
	vt         uint16
	reserved1  uint16
	reserved2  uint16
	reserved3  uint16
	value      uintptr
	valueExtra uintptr
}

func (windowsDriver) registerProtocolHandler(ctx context.Context, scheme, command string, opts registrationOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	keyPath := windowsProtocolRegistryPath(scheme)
	if err := windowsSetDefaultRegistryValue(keyPath, protocolDisplayName(scheme, opts)); err != nil {
		return err
	}
	if err := windowsSetRegistryStringValue(keyPath, "URL Protocol", ""); err != nil {
		return err
	}
	if opts.icon != "" {
		if err := windowsSetDefaultRegistryValue(keyPath+`\DefaultIcon`, opts.icon); err != nil {
			return err
		}
	}
	return windowsSetDefaultRegistryValue(keyPath+`\shell\open\command`, command)
}

func (windowsDriver) unregisterProtocolHandler(ctx context.Context, scheme string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return windowsDeleteRegistryTree(registry.CURRENT_USER, windowsProtocolRegistryPath(scheme))
}

func (windowsDriver) registerFileAssociation(ctx context.Context, extension, progID, command string, opts registrationOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	extensionPath := windowsFileExtensionRegistryPath(extension)
	if err := windowsSetDefaultRegistryValue(extensionPath, progID); err != nil {
		return err
	}
	progPath := windowsProgIDRegistryPath(progID)
	if err := windowsSetDefaultRegistryValue(progPath, fileAssociationDisplayName(extension, progID, opts)); err != nil {
		return err
	}
	if opts.icon != "" {
		if err := windowsSetDefaultRegistryValue(progPath+`\DefaultIcon`, opts.icon); err != nil {
			return err
		}
	}
	return windowsSetDefaultRegistryValue(progPath+`\shell\open\command`, command)
}

func (windowsDriver) unregisterFileAssociation(ctx context.Context, extension, progID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	extensionPath := windowsFileExtensionRegistryPath(extension)
	if current, err := windowsReadDefaultRegistryValue(extensionPath); err == nil && current == progID {
		if err := windowsDeleteRegistryTree(registry.CURRENT_USER, extensionPath); err != nil {
			return err
		}
	}
	return windowsDeleteRegistryTree(registry.CURRENT_USER, windowsProgIDRegistryPath(progID))
}

func (windowsDriver) registerStartupTask(ctx context.Context, name, command string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return windowsSetRegistryStringValue(windowsStartupRunPath, name, command)
}

func (windowsDriver) unregisterStartupTask(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, windowsStartupRunPath, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return windowsRegistrationUnavailable("open startup key", err)
	}
	defer key.Close()
	if err := key.DeleteValue(name); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return windowsRegistrationUnavailable("delete startup value", err)
	}
	return nil
}

func (windowsDriver) registerToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts toastShortcutOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	executable, err := windowsResolveToastShortcutExecutable(executable)
	if err != nil {
		return err
	}
	shortcutPath, err := windowsToastShortcutPath(shortcutName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(shortcutPath), 0755); err != nil {
		return windowsRegistrationUnavailable("create toast shortcut directory", err)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	initialized, err := initCOMApartment()
	if err != nil {
		return windowsRegistrationUnavailable("initialize toast shortcut COM", err)
	}
	if initialized {
		defer windows.CoUninitialize()
	}

	link, err := newWindowsShellLink()
	if err != nil {
		return windowsRegistrationUnavailable("create toast shortcut", err)
	}
	defer link.Release()
	if err := link.SetPath(executable); err != nil {
		return windowsRegistrationUnavailable("set toast shortcut target", err)
	}
	if len(opts.arguments) > 0 {
		if err := link.SetArguments(windowsToastShortcutArguments(opts.arguments)); err != nil {
			return windowsRegistrationUnavailable("set toast shortcut arguments", err)
		}
	}
	if opts.icon != "" {
		iconPath, iconIndex := windowsToastShortcutIconLocation(opts.icon)
		if err := link.SetIconLocation(iconPath, iconIndex); err != nil {
			return windowsRegistrationUnavailable("set toast shortcut icon", err)
		}
	}
	if err := link.SetAppUserModelID(appID); err != nil {
		return windowsRegistrationUnavailable("set toast shortcut AppUserModelID", err)
	}
	if opts.activatorCLSID != "" {
		clsid, err := windowsToastShortcutActivatorCLSID(opts.activatorCLSID)
		if err != nil {
			return err
		}
		if err := link.SetToastActivatorCLSID(clsid); err != nil {
			return windowsRegistrationUnavailable("set toast shortcut activator CLSID", err)
		}
	}
	if err := link.Save(shortcutPath); err != nil {
		return windowsRegistrationUnavailable("save toast shortcut", err)
	}
	return nil
}

func (windowsDriver) unregisterToastShortcut(ctx context.Context, shortcutName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	shortcutPath, err := windowsToastShortcutPath(shortcutName)
	if err != nil {
		return err
	}
	if err := os.Remove(shortcutPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return windowsRegistrationUnavailable("delete toast shortcut", err)
	}
	return nil
}

func (windowsDriver) registerToastActivator(ctx context.Context, clsid, command string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	parsed, err := windowsToastShortcutActivatorCLSID(clsid)
	if err != nil {
		return err
	}
	keyPath := windowsToastActivatorRegistryPath(parsed.String())
	if err := windowsSetDefaultRegistryValue(keyPath, "FluxUI Toast Activator"); err != nil {
		return err
	}
	return windowsSetDefaultRegistryValue(keyPath+`\LocalServer32`, command)
}

func (windowsDriver) unregisterToastActivator(ctx context.Context, clsid string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	parsed, err := windowsToastShortcutActivatorCLSID(clsid)
	if err != nil {
		return err
	}
	return windowsDeleteRegistryTree(registry.CURRENT_USER, windowsToastActivatorRegistryPath(parsed.String()))
}

func windowsProtocolRegistryPath(scheme string) string {
	return windowsClassesRootPath + `\` + scheme
}

func windowsFileExtensionRegistryPath(extension string) string {
	return windowsClassesRootPath + `\` + extension
}

func windowsProgIDRegistryPath(progID string) string {
	return windowsClassesRootPath + `\` + progID
}

func windowsToastActivatorRegistryPath(clsid string) string {
	return windowsClassesRootPath + `\CLSID\` + clsid
}

func protocolDisplayName(scheme string, opts registrationOptions) string {
	if opts.displayName != "" {
		return opts.displayName
	}
	return "URL:" + scheme + " Protocol"
}

func fileAssociationDisplayName(extension, progID string, opts registrationOptions) string {
	if opts.displayName != "" {
		return opts.displayName
	}
	return progID + " file (" + extension + ")"
}

func windowsSetDefaultRegistryValue(path, value string) error {
	return windowsSetRegistryStringValue(path, "", value)
}

func windowsSetRegistryStringValue(path, name, value string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return windowsRegistrationUnavailable("create registry key", err)
	}
	defer key.Close()
	if err := key.SetStringValue(name, value); err != nil {
		return windowsRegistrationUnavailable("set registry value", err)
	}
	return nil
}

func windowsReadDefaultRegistryValue(path string) (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()
	value, _, err := key.GetStringValue("")
	return value, err
}

func windowsDeleteRegistryTree(root registry.Key, path string) error {
	key, err := registry.OpenKey(root, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return windowsRegistrationUnavailable("open registry key", err)
	}
	subKeys, err := key.ReadSubKeyNames(-1)
	key.Close()
	if err != nil {
		return windowsRegistrationUnavailable("read registry subkeys", err)
	}
	for _, subKey := range subKeys {
		if err := windowsDeleteRegistryTree(root, path+`\`+subKey); err != nil {
			return err
		}
	}
	if err := registry.DeleteKey(root, path); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return windowsRegistrationUnavailable("delete registry key", err)
	}
	return nil
}

func windowsRegistrationUnavailable(operation string, err error) error {
	if err == nil {
		err = ErrUnavailable
	}
	return fmt.Errorf("system: %s: %s: %w: %v", CapabilitySystemRegistration, operation, ErrUnavailable, err)
}

func windowsResolveToastShortcutExecutable(executable string) (string, error) {
	abs, err := filepath.Abs(executable)
	if err != nil {
		return "", fmt.Errorf("system: %s: toast shortcut executable %q is invalid: %w: %w", CapabilitySystemRegistration, executable, ErrInvalidTarget, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("system: %s: toast shortcut executable %q was not found: %w: %w: %w", CapabilitySystemRegistration, abs, ErrUnavailable, ErrTargetNotFound, err)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("system: %s: toast shortcut executable %q is denied: %w: %w: %w", CapabilitySystemRegistration, abs, ErrUnavailable, ErrAccessDenied, err)
		}
		return "", fmt.Errorf("system: %s: toast shortcut executable %q is unavailable: %w: %w", CapabilitySystemRegistration, abs, ErrUnavailable, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("system: %s: toast shortcut executable %q is a directory: %w", CapabilitySystemRegistration, abs, ErrInvalidTarget)
	}
	return abs, nil
}

func windowsToastShortcutPath(shortcutName string) (string, error) {
	appData := strings.TrimSpace(os.Getenv("APPDATA"))
	if appData == "" {
		return "", fmt.Errorf("system: %s: APPDATA is unavailable: %w", CapabilitySystemRegistration, ErrUnavailable)
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", shortcutName), nil
}

func windowsToastShortcutIconLocation(icon string) (string, int32) {
	icon = strings.TrimSpace(icon)
	trimmed := strings.Trim(icon, `"`)
	comma := strings.LastIndex(trimmed, ",")
	if comma < 0 {
		return icon, 0
	}
	path := strings.Trim(strings.TrimSpace(trimmed[:comma]), `"`)
	index, err := strconv.ParseInt(strings.TrimSpace(trimmed[comma+1:]), 10, 32)
	if err != nil {
		return icon, 0
	}
	if path == "" {
		return icon, 0
	}
	return path, int32(index)
}

func windowsToastShortcutActivatorCLSID(clsid string) (windows.GUID, error) {
	parsed, err := windows.GUIDFromString(strings.TrimSpace(clsid))
	if err != nil {
		return windows.GUID{}, fmt.Errorf("system: %s: toast shortcut activator CLSID %q is invalid: %w: %v", CapabilitySystemRegistration, clsid, ErrInvalidTarget, err)
	}
	return parsed, nil
}

func newWindowsShellLink() (*iShellLinkW, error) {
	var link *iShellLinkW
	r, _, callErr := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidShellLink)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(&iidShellLinkW)),
		uintptr(unsafe.Pointer(&link)),
	)
	if failedHRESULT(r) {
		if callErr != syscall.Errno(0) {
			return nil, callErr
		}
		return nil, fmt.Errorf("HRESULT 0x%08x", uint32(r))
	}
	if link == nil {
		return nil, ErrUnavailable
	}
	return link, nil
}

func (l *iShellLinkW) Release() {
	if l != nil {
		syscall.SyscallN(l.lpVtbl.Release, uintptr(unsafe.Pointer(l)))
	}
}

func (l *iShellLinkW) SetPath(path string) error {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(l.lpVtbl.SetPath, uintptr(unsafe.Pointer(l)), uintptr(unsafe.Pointer(path16)))
	runtime.KeepAlive(path16)
	return hresultError(r)
}

func (l *iShellLinkW) SetArguments(args string) error {
	args16, err := windows.UTF16PtrFromString(args)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(l.lpVtbl.SetArguments, uintptr(unsafe.Pointer(l)), uintptr(unsafe.Pointer(args16)))
	runtime.KeepAlive(args16)
	return hresultError(r)
}

func (l *iShellLinkW) SetIconLocation(icon string, index int32) error {
	icon16, err := windows.UTF16PtrFromString(icon)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(l.lpVtbl.SetIconLocation, uintptr(unsafe.Pointer(l)), uintptr(unsafe.Pointer(icon16)), uintptr(index))
	runtime.KeepAlive(icon16)
	return hresultError(r)
}

func (l *iShellLinkW) Save(path string) error {
	var file *iPersistFile
	if err := comQueryInterface(unsafe.Pointer(l), &iidPersistFile, unsafe.Pointer(&file)); err != nil {
		return err
	}
	defer file.Release()
	return file.Save(path, true)
}

func (l *iShellLinkW) SetAppUserModelID(appID string) error {
	var store *iPropertyStore
	if err := comQueryInterface(unsafe.Pointer(l), &iidPropertyStore, unsafe.Pointer(&store)); err != nil {
		return err
	}
	defer store.Release()
	return store.SetString(pkeyAppUserModelID, appID)
}

func (l *iShellLinkW) SetToastActivatorCLSID(clsid windows.GUID) error {
	var store *iPropertyStore
	if err := comQueryInterface(unsafe.Pointer(l), &iidPropertyStore, unsafe.Pointer(&store)); err != nil {
		return err
	}
	defer store.Release()
	return store.SetGUID(pkeyAppUserModelToastActivatorCLSID, clsid)
}

func (f *iPersistFile) Release() {
	if f != nil {
		syscall.SyscallN(f.lpVtbl.Release, uintptr(unsafe.Pointer(f)))
	}
}

func (f *iPersistFile) Save(path string, remember bool) error {
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var rememberValue uintptr
	if remember {
		rememberValue = 1
	}
	r, _, _ := syscall.SyscallN(f.lpVtbl.Save, uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(path16)), rememberValue)
	runtime.KeepAlive(path16)
	return hresultError(r)
}

func (s *iPropertyStore) Release() {
	if s != nil {
		syscall.SyscallN(s.lpVtbl.Release, uintptr(unsafe.Pointer(s)))
	}
}

func (s *iPropertyStore) SetString(key propertyKey, value string) error {
	value16, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	prop := propVariant{
		vt:    vtLPWSTR,
		value: uintptr(unsafe.Pointer(value16)),
	}
	r, _, _ := syscall.SyscallN(s.lpVtbl.SetValue, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&key)), uintptr(unsafe.Pointer(&prop)))
	runtime.KeepAlive(value16)
	if err := hresultError(r); err != nil {
		return err
	}
	r, _, _ = syscall.SyscallN(s.lpVtbl.Commit, uintptr(unsafe.Pointer(s)))
	return hresultError(r)
}

func (s *iPropertyStore) SetGUID(key propertyKey, value windows.GUID) error {
	prop := propVariant{
		vt:    vtCLSID,
		value: uintptr(unsafe.Pointer(&value)),
	}
	r, _, _ := syscall.SyscallN(s.lpVtbl.SetValue, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&key)), uintptr(unsafe.Pointer(&prop)))
	runtime.KeepAlive(&value)
	if err := hresultError(r); err != nil {
		return err
	}
	r, _, _ = syscall.SyscallN(s.lpVtbl.Commit, uintptr(unsafe.Pointer(s)))
	return hresultError(r)
}

func comQueryInterface(obj unsafe.Pointer, iid *windows.GUID, out unsafe.Pointer) error {
	if obj == nil {
		return ErrUnavailable
	}
	unknown := (*comInterface)(obj)
	r, _, _ := syscall.SyscallN(unknown.lpVtbl.QueryInterface, uintptr(obj), uintptr(unsafe.Pointer(iid)), uintptr(out))
	return hresultError(r)
}

func windowsToastShortcutArguments(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, windowsQuoteCommandLineArgument(arg))
	}
	return strings.Join(quoted, " ")
}

func windowsQuoteCommandLineArgument(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, " \t\n\v\"") {
		return arg
	}
	var b strings.Builder
	b.WriteByte('"')
	backslashes := 0
	for _, r := range arg {
		if r == '\\' {
			backslashes++
			continue
		}
		if r == '"' {
			b.WriteString(strings.Repeat(`\`, backslashes*2+1))
			b.WriteRune(r)
			backslashes = 0
			continue
		}
		if backslashes > 0 {
			b.WriteString(strings.Repeat(`\`, backslashes))
			backslashes = 0
		}
		b.WriteRune(r)
	}
	if backslashes > 0 {
		b.WriteString(strings.Repeat(`\`, backslashes*2))
	}
	b.WriteByte('"')
	return b.String()
}
