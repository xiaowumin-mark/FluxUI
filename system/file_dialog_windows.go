//go:build windows

package system

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	clsctxInprocServer = 0x1

	coinitApartmentThreaded = 0x2

	sFalse = 0x1

	hresultCancelled = 0x800704c7

	sigdnFileSysPath = 0x80058000

	fosOverwritePrompt  = 0x2
	fosStrictFileTypes  = 0x4
	fosNoChangeDir      = 0x8
	fosPickFolders      = 0x20
	fosForceFilesystem  = 0x40
	fosAllowMultiSelect = 0x200
	fosPathMustExist    = 0x800
	fosFileMustExist    = 0x1000
	fosCreatePrompt     = 0x2000
	fosNoReadOnlyReturn = 0x8000
)

var (
	ole32                  = windows.NewLazySystemDLL("ole32.dll")
	shell32                = windows.NewLazySystemDLL("shell32.dll")
	procCoCreateInstance   = ole32.NewProc("CoCreateInstance")
	procSHCreateItemFromPN = shell32.NewProc("SHCreateItemFromParsingName")

	clsidFileOpenDialog = windows.GUID{Data1: 0xdc1c5a9c, Data2: 0xe88a, Data3: 0x4dde, Data4: [8]byte{0xa5, 0xa1, 0x60, 0xf8, 0x2a, 0x20, 0xae, 0xf7}}
	clsidFileSaveDialog = windows.GUID{Data1: 0xc0b4e2f3, Data2: 0xba21, Data3: 0x4773, Data4: [8]byte{0x8d, 0xba, 0x33, 0x5e, 0xc9, 0x46, 0xeb, 0x8b}}
	iidFileOpenDialog   = windows.GUID{Data1: 0xd57c7288, Data2: 0xd4ad, Data3: 0x4768, Data4: [8]byte{0xbe, 0x02, 0x9d, 0x96, 0x95, 0x32, 0xd9, 0x60}}
	iidFileSaveDialog   = windows.GUID{Data1: 0x84bccd23, Data2: 0x5fde, Data3: 0x4cdb, Data4: [8]byte{0xae, 0xa4, 0xaf, 0x64, 0xb8, 0x3d, 0x78, 0xab}}
	iidShellItem        = windows.GUID{Data1: 0x43826d1e, Data2: 0xe718, Data3: 0x42ee, Data4: [8]byte{0xbc, 0x55, 0xa1, 0xe2, 0x61, 0xc3, 0x7b, 0xfe}}
)

type comInterface struct {
	lpVtbl *comInterfaceVtbl
}

type comInterfaceVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type comdlgFilterSpec struct {
	Name *uint16
	Spec *uint16
}

type iFileDialog struct {
	lpVtbl *iFileDialogVtbl
}

type iFileDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
}

type iFileOpenDialog struct {
	lpVtbl *iFileOpenDialogVtbl
}

type iFileOpenDialogVtbl struct {
	iFileDialogVtbl
	GetResults       uintptr
	GetSelectedItems uintptr
}

type iShellItem struct {
	lpVtbl *iShellItemVtbl
}

type iShellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

type iShellItemArray struct {
	lpVtbl *iShellItemArrayVtbl
}

type iShellItemArrayVtbl struct {
	QueryInterface             uintptr
	AddRef                     uintptr
	Release                    uintptr
	BindToHandler              uintptr
	GetPropertyStore           uintptr
	GetPropertyDescriptionList uintptr
	GetAttributes              uintptr
	GetCount                   uintptr
	GetItemAt                  uintptr
	EnumItems                  uintptr
}

func (windowsDriver) openFileDialog(ctx context.Context, mode FileDialogMode, opts fileDialogOptions) (FileDialogResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}

	switch mode {
	case FileDialogOpenFile, FileDialogOpenFiles, FileDialogSaveFile, FileDialogPickFolder:
	default:
		return FileDialogResult{}, fmt.Errorf("system: unknown file dialog mode %q", mode)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	initialized, err := initCOMApartment()
	if err != nil {
		return FileDialogResult{}, fmt.Errorf("system: initialize file dialog COM apartment: %w", err)
	}
	if initialized {
		defer windows.CoUninitialize()
	}

	dialog, err := newWindowsFileDialog(mode)
	if err != nil {
		return FileDialogResult{}, err
	}
	defer dialog.Release()

	if err := configureWindowsFileDialog(dialog, mode, opts); err != nil {
		return FileDialogResult{}, err
	}

	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}

	stopCancelWatcher := watchWindowsFileDialogContext(ctx, dialog)

	if err := dialog.Show(opts.owner); err != nil {
		stopCancelWatcher()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return FileDialogResult{}, ctxErr
		}
		if isHRESULT(err, hresultCancelled) {
			return FileDialogResult{Cancelled: true}, nil
		}
		return FileDialogResult{}, fmt.Errorf("system: show file dialog: %w", err)
	}
	stopCancelWatcher()
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}

	paths, err := windowsFileDialogPaths(dialog, mode)
	if err != nil {
		return FileDialogResult{}, err
	}
	if len(paths) == 0 {
		return FileDialogResult{Cancelled: true}, nil
	}
	return FileDialogResult{Paths: paths}, nil
}

func watchWindowsFileDialogContext(ctx context.Context, dialog *iFileDialog) func() {
	if ctx == nil || ctx.Done() == nil || dialog == nil {
		return func() {}
	}

	done := make(chan struct{})
	stopped := make(chan struct{})
	var stopOnce sync.Once
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
			_ = dialog.Close(hresultCancelled)
		case <-done:
		}
	}()

	return func() {
		stopOnce.Do(func() {
			close(done)
			<-stopped
		})
	}
}

func initCOMApartment() (bool, error) {
	err := windows.CoInitializeEx(0, coinitApartmentThreaded)
	if err == nil {
		return true, nil
	}
	if errno, ok := err.(syscall.Errno); ok && uintptr(errno) == sFalse {
		return true, nil
	}
	return false, err
}

func newWindowsFileDialog(mode FileDialogMode) (*iFileDialog, error) {
	var clsid *windows.GUID
	var iid *windows.GUID
	switch mode {
	case FileDialogSaveFile:
		clsid = &clsidFileSaveDialog
		iid = &iidFileSaveDialog
	default:
		clsid = &clsidFileOpenDialog
		iid = &iidFileOpenDialog
	}

	var dialog *iFileDialog
	r, _, callErr := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(clsid)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&dialog)),
	)
	if failedHRESULT(r) {
		if callErr != syscall.Errno(0) {
			return nil, fmt.Errorf("system: create file dialog: %w", callErr)
		}
		return nil, fmt.Errorf("system: create file dialog: HRESULT 0x%08x", uint32(r))
	}
	if dialog == nil {
		return nil, fmt.Errorf("system: create file dialog: %w", ErrUnavailable)
	}
	return dialog, nil
}

func configureWindowsFileDialog(dialog *iFileDialog, mode FileDialogMode, opts fileDialogOptions) error {
	flags := uint32(fosForceFilesystem | fosNoChangeDir)
	switch mode {
	case FileDialogOpenFile:
		flags |= fosNoReadOnlyReturn
		if !opts.allowMissingPath {
			flags |= fosPathMustExist | fosFileMustExist
		}
	case FileDialogOpenFiles:
		flags |= fosAllowMultiSelect | fosNoReadOnlyReturn
		if !opts.allowMissingPath {
			flags |= fosPathMustExist | fosFileMustExist
		}
	case FileDialogSaveFile:
		if opts.overwritePrompt {
			flags |= fosOverwritePrompt
		}
		if opts.allowCreateDirs {
			flags |= fosCreatePrompt
		}
		if !opts.allowMissingPath {
			flags |= fosPathMustExist
		}
	case FileDialogPickFolder:
		flags |= fosPickFolders
		if !opts.allowMissingPath {
			flags |= fosPathMustExist
		}
	}
	if len(opts.filters) > 0 {
		flags |= fosStrictFileTypes
	}
	if err := dialog.SetOptions(flags); err != nil {
		return fmt.Errorf("system: configure file dialog options: %w", err)
	}

	if opts.title != "" {
		if err := dialog.SetTitle(opts.title); err != nil {
			return fmt.Errorf("system: configure file dialog title: %w", err)
		}
	}
	if opts.defaultName != "" && mode != FileDialogPickFolder {
		if err := dialog.SetFileName(opts.defaultName); err != nil {
			return fmt.Errorf("system: configure file dialog default name: %w", err)
		}
	}
	if opts.defaultExtension != "" && mode != FileDialogPickFolder {
		if err := dialog.SetDefaultExtension(opts.defaultExtension); err != nil {
			return fmt.Errorf("system: configure file dialog default extension: %w", err)
		}
	}
	if len(opts.filters) > 0 && mode != FileDialogPickFolder {
		if err := dialog.SetFileTypes(opts.filters); err != nil {
			return fmt.Errorf("system: configure file dialog filters: %w", err)
		}
	}
	if opts.defaultDir != "" {
		folder, err := shellItemFromPath(opts.defaultDir)
		if err != nil {
			return fmt.Errorf("system: configure file dialog default dir: %w", &FileDialogError{
				Kind: FileDialogErrorDefaultDir,
				Path: opts.defaultDir,
				Err:  err,
			})
		}
		defer folder.Release()

		if err := dialog.SetFolder(folder); err != nil {
			return fmt.Errorf("system: configure file dialog default dir: %w", err)
		}
	}
	return nil
}

func windowsFileDialogPaths(dialog *iFileDialog, mode FileDialogMode) ([]string, error) {
	if mode == FileDialogOpenFiles {
		openDialog := (*iFileOpenDialog)(unsafe.Pointer(dialog))
		items, err := openDialog.GetResults()
		if err != nil {
			return nil, fmt.Errorf("system: read selected files: %w", err)
		}
		defer items.Release()
		return items.Paths()
	}

	item, err := dialog.GetResult()
	if err != nil {
		return nil, fmt.Errorf("system: read selected path: %w", err)
	}
	defer item.Release()

	path, err := item.Path()
	if err != nil {
		return nil, err
	}
	return []string{path}, nil
}

func (d *iFileDialog) Release() {
	if d == nil {
		return
	}
	syscall.SyscallN(d.lpVtbl.Release, uintptr(unsafe.Pointer(d)))
}

func (d *iFileDialog) Show(owner uintptr) error {
	r, _, _ := syscall.SyscallN(d.lpVtbl.Show, uintptr(unsafe.Pointer(d)), owner)
	return hresultError(r)
}

func (d *iFileDialog) SetOptions(flags uint32) error {
	r, _, _ := syscall.SyscallN(d.lpVtbl.SetOptions, uintptr(unsafe.Pointer(d)), uintptr(flags))
	return hresultError(r)
}

func (d *iFileDialog) SetFolder(folder *iShellItem) error {
	r, _, _ := syscall.SyscallN(d.lpVtbl.SetFolder, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(folder)))
	return hresultError(r)
}

func (d *iFileDialog) SetFileName(name string) error {
	name16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(d.lpVtbl.SetFileName, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(name16)))
	return hresultError(r)
}

func (d *iFileDialog) SetDefaultExtension(extension string) error {
	extension16, err := windows.UTF16PtrFromString(extension)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(d.lpVtbl.SetDefaultExtension, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(extension16)))
	return hresultError(r)
}

func (d *iFileDialog) Close(hr uintptr) error {
	r, _, _ := syscall.SyscallN(d.lpVtbl.Close, uintptr(unsafe.Pointer(d)), hr)
	return hresultError(r)
}

func (d *iFileDialog) SetTitle(title string) error {
	title16, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	r, _, _ := syscall.SyscallN(d.lpVtbl.SetTitle, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(title16)))
	return hresultError(r)
}

func (d *iFileDialog) SetFileTypes(filters []FileFilter) error {
	specs, keepAlive, err := windowsFilterSpecs(filters)
	if err != nil {
		return err
	}
	_ = keepAlive

	r, _, _ := syscall.SyscallN(
		d.lpVtbl.SetFileTypes,
		uintptr(unsafe.Pointer(d)),
		uintptr(len(specs)),
		uintptr(unsafe.Pointer(&specs[0])),
	)
	runtime.KeepAlive(keepAlive)
	return hresultError(r)
}

func (d *iFileDialog) GetResult() (*iShellItem, error) {
	var item *iShellItem
	r, _, _ := syscall.SyscallN(d.lpVtbl.GetResult, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&item)))
	if err := hresultError(r); err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("system: selected item missing: %w", ErrUnavailable)
	}
	return item, nil
}

func (d *iFileOpenDialog) GetResults() (*iShellItemArray, error) {
	var items *iShellItemArray
	r, _, _ := syscall.SyscallN(d.lpVtbl.GetResults, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&items)))
	if err := hresultError(r); err != nil {
		return nil, err
	}
	if items == nil {
		return nil, fmt.Errorf("system: selected item array missing: %w", ErrUnavailable)
	}
	return items, nil
}

func (item *iShellItem) Release() {
	if item == nil {
		return
	}
	syscall.SyscallN(item.lpVtbl.Release, uintptr(unsafe.Pointer(item)))
}

func (item *iShellItem) Path() (string, error) {
	var raw *uint16
	r, _, _ := syscall.SyscallN(
		item.lpVtbl.GetDisplayName,
		uintptr(unsafe.Pointer(item)),
		sigdnFileSysPath,
		uintptr(unsafe.Pointer(&raw)),
	)
	if err := hresultError(r); err != nil {
		return "", fmt.Errorf("system: read shell item path: %w", err)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))

	path := windows.UTF16PtrToString(raw)
	if path == "" {
		return "", &FileDialogError{Kind: FileDialogErrorSelectedPath, Err: ErrUnavailable}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

func (items *iShellItemArray) Release() {
	if items == nil {
		return
	}
	syscall.SyscallN(items.lpVtbl.Release, uintptr(unsafe.Pointer(items)))
}

func (items *iShellItemArray) Count() (uint32, error) {
	var count uint32
	r, _, _ := syscall.SyscallN(items.lpVtbl.GetCount, uintptr(unsafe.Pointer(items)), uintptr(unsafe.Pointer(&count)))
	if err := hresultError(r); err != nil {
		return 0, err
	}
	return count, nil
}

func (items *iShellItemArray) ItemAt(index uint32) (*iShellItem, error) {
	var item *iShellItem
	r, _, _ := syscall.SyscallN(
		items.lpVtbl.GetItemAt,
		uintptr(unsafe.Pointer(items)),
		uintptr(index),
		uintptr(unsafe.Pointer(&item)),
	)
	if err := hresultError(r); err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("system: selected item missing: %w", ErrUnavailable)
	}
	return item, nil
}

func (items *iShellItemArray) Paths() ([]string, error) {
	count, err := items.Count()
	if err != nil {
		return nil, fmt.Errorf("system: read selected file count: %w", err)
	}
	paths := make([]string, 0, count)
	for i := uint32(0); i < count; i++ {
		item, err := items.ItemAt(i)
		if err != nil {
			return nil, fmt.Errorf("system: read selected file %d: %w", i, err)
		}
		path, pathErr := item.Path()
		item.Release()
		if pathErr != nil {
			return nil, pathErr
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func shellItemFromPath(path string) (*iShellItem, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	path16, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return nil, err
	}

	var item *iShellItem
	r, _, callErr := procSHCreateItemFromPN.Call(
		uintptr(unsafe.Pointer(path16)),
		0,
		uintptr(unsafe.Pointer(&iidShellItem)),
		uintptr(unsafe.Pointer(&item)),
	)
	if failedHRESULT(r) {
		if callErr != syscall.Errno(0) {
			return nil, callErr
		}
		return nil, fmt.Errorf("HRESULT 0x%08x", uint32(r))
	}
	if item == nil {
		return nil, fmt.Errorf("shell item missing: %w", ErrUnavailable)
	}
	return item, nil
}

func windowsFilterSpecs(filters []FileFilter) ([]comdlgFilterSpec, [][]uint16, error) {
	specs := make([]comdlgFilterSpec, 0, len(filters))
	keepAlive := make([][]uint16, 0, len(filters)*2)
	for _, filter := range filters {
		patterns := normalizeFilterPatterns(filter.Patterns)
		if len(patterns) == 0 {
			continue
		}
		name := strings.TrimSpace(filter.Name)
		if name == "" {
			name = strings.Join(patterns, ", ")
		}

		name16, err := windows.UTF16FromString(name)
		if err != nil {
			return nil, nil, err
		}
		spec16, err := windows.UTF16FromString(strings.Join(patterns, ";"))
		if err != nil {
			return nil, nil, err
		}
		keepAlive = append(keepAlive, name16, spec16)
		specs = append(specs, comdlgFilterSpec{
			Name: &keepAlive[len(keepAlive)-2][0],
			Spec: &keepAlive[len(keepAlive)-1][0],
		})
	}
	if len(specs) == 0 {
		return nil, nil, fmt.Errorf("system: file dialog filters have no patterns")
	}
	return specs, keepAlive, nil
}

func normalizeFilterPatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if !strings.ContainsAny(pattern, "*?") {
			if strings.HasPrefix(pattern, ".") {
				pattern = "*" + pattern
			} else {
				pattern = "*." + pattern
			}
		}
		normalized = append(normalized, pattern)
	}
	return normalized
}

func hresultError(hr uintptr) error {
	if !failedHRESULT(hr) {
		return nil
	}
	return syscall.Errno(hr)
}

func failedHRESULT(hr uintptr) bool {
	return uint32(hr)&0x80000000 != 0
}

func isHRESULT(err error, hr uintptr) bool {
	if errno, ok := err.(syscall.Errno); ok {
		return uintptr(errno) == hr
	}
	return false
}
