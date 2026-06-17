package system

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// FileDialogMode identifies the native file dialog variant to open.
type FileDialogMode string

const (
	FileDialogOpenFile   FileDialogMode = "open_file"
	FileDialogOpenFiles  FileDialogMode = "open_files"
	FileDialogSaveFile   FileDialogMode = "save_file"
	FileDialogPickFolder FileDialogMode = "pick_folder"
)

// FileFilter describes one visible filter in a native file dialog.
type FileFilter struct {
	Name     string
	Patterns []string
}

// FileDialogResult is returned by all file dialog APIs.
type FileDialogResult struct {
	Paths     []string
	Cancelled bool
}

// FileDialogErrorKind identifies structured file dialog error categories.
type FileDialogErrorKind string

const (
	FileDialogErrorDefaultDir   FileDialogErrorKind = "default_dir"
	FileDialogErrorSelectedPath FileDialogErrorKind = "selected_path"
	FileDialogErrorPath         FileDialogErrorKind = "path"
)

// FileDialogError wraps file-dialog errors with path-oriented context.
type FileDialogError struct {
	Kind FileDialogErrorKind
	Path string
	Err  error
}

func (e *FileDialogError) Error() string {
	if e == nil {
		return ""
	}
	if e.Path != "" {
		return fmt.Sprintf("system: file dialog %s %q: %v", e.Kind, e.Path, e.Err)
	}
	return fmt.Sprintf("system: file dialog %s: %v", e.Kind, e.Err)
}

func (e *FileDialogError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsFileDialogErrorKind reports whether err contains a FileDialogError of kind.
func IsFileDialogErrorKind(err error, kind FileDialogErrorKind) bool {
	var dialogErr *FileDialogError
	return errors.As(err, &dialogErr) && dialogErr.Kind == kind
}

// FileDialogOption configures native file dialogs.
type FileDialogOption func(*fileDialogOptions)

type fileDialogOptions struct {
	title            string
	defaultDir       string
	defaultName      string
	defaultExtension string
	filters          []FileFilter
	owner            uintptr
	allowCreateDirs  bool
	allowMissingPath bool
	overwritePrompt  bool
	rememberDirKey   string
}

type fileDialogDriver interface {
	openFileDialog(ctx context.Context, mode FileDialogMode, opts fileDialogOptions) (FileDialogResult, error)
}

var (
	fileDialogRememberedDirsMu sync.Mutex
	fileDialogRememberedDirs   = make(map[string]string)
)

func defaultFileDialogOptions() fileDialogOptions {
	return fileDialogOptions{
		allowCreateDirs: true,
		overwritePrompt: true,
	}
}

// FileDialogTitle sets the native dialog title.
func FileDialogTitle(value string) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.title = value
	}
}

// FileDialogDefaultDir sets the initial directory for the native dialog.
func FileDialogDefaultDir(value string) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.defaultDir = value
	}
}

// FileDialogDefaultName sets the initial file name for save dialogs.
func FileDialogDefaultName(value string) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.defaultName = value
	}
}

// FileDialogDefaultExtension sets the default extension used by save dialogs.
func FileDialogDefaultExtension(value string) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.defaultExtension = normalizeFileDialogExtension(value)
	}
}

// FileDialogFilters sets the visible file type filters.
func FileDialogFilters(filters ...FileFilter) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.filters = cloneFileFilters(filters)
	}
}

// FileDialogOwner sets the native owner window handle when the caller already has one.
//
// On Windows this value is interpreted as HWND. Passing 0 keeps the dialog ownerless.
func FileDialogOwner(owner uintptr) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.owner = owner
	}
}

// FileDialogAllowCreateDirs controls whether the dialog may prompt to create missing folders.
func FileDialogAllowCreateDirs(allow bool) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.allowCreateDirs = allow
	}
}

// FileDialogAllowMissingPath controls whether selected paths must already exist.
func FileDialogAllowMissingPath(allow bool) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.allowMissingPath = allow
	}
}

// FileDialogOverwritePrompt controls whether save dialogs ask before replacing an existing file.
func FileDialogOverwritePrompt(prompt bool) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.overwritePrompt = prompt
	}
}

// FileDialogRememberDir remembers the last successful directory for the given key.
func FileDialogRememberDir(key string) FileDialogOption {
	return func(opts *fileDialogOptions) {
		opts.rememberDirKey = strings.TrimSpace(key)
	}
}

// OpenFileDialog opens a native dialog for selecting one file.
func OpenFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error) {
	return openFileDialog(ctx, FileDialogOpenFile, opts...)
}

// OpenFilesDialog opens a native dialog for selecting multiple files.
func OpenFilesDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error) {
	return openFileDialog(ctx, FileDialogOpenFiles, opts...)
}

// SaveFileDialog opens a native dialog for selecting one save target.
func SaveFileDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error) {
	return openFileDialog(ctx, FileDialogSaveFile, opts...)
}

// PickFolderDialog opens a native dialog for selecting one folder.
func PickFolderDialog(ctx context.Context, opts ...FileDialogOption) (FileDialogResult, error) {
	return openFileDialog(ctx, FileDialogPickFolder, opts...)
}

func openFileDialog(ctx context.Context, mode FileDialogMode, options ...FileDialogOption) (FileDialogResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}

	opts := defaultFileDialogOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	opts.filters = cloneFileFilters(opts.filters)
	if opts.defaultDir == "" && opts.rememberDirKey != "" {
		opts.defaultDir = rememberedFileDialogDir(opts.rememberDirKey)
	}

	d, supported := currentDriverFor(CapabilityFileDialog)
	fd, ok := d.(fileDialogDriver)
	if !ok || !supported {
		return FileDialogResult{}, fmt.Errorf("system: %s: %w", CapabilityFileDialog, ErrUnsupported)
	}
	result, err := fd.openFileDialog(ctx, mode, opts)
	if err != nil || result.Cancelled {
		return result, err
	}
	result, err = normalizeFileDialogResultPaths(result)
	if err != nil {
		return FileDialogResult{}, err
	}
	result = finalizeFileDialogResult(mode, opts, result)
	rememberFileDialogResultDir(mode, opts.rememberDirKey, result)
	return result, nil
}

func cloneFileFilters(filters []FileFilter) []FileFilter {
	if len(filters) == 0 {
		return nil
	}

	cloned := make([]FileFilter, len(filters))
	for i, filter := range filters {
		cloned[i] = FileFilter{
			Name:     filter.Name,
			Patterns: append([]string(nil), filter.Patterns...),
		}
	}
	return cloned
}

func normalizeFileDialogExtension(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "*")
	value = strings.TrimPrefix(value, ".")
	return value
}

func normalizeFileDialogResultPaths(result FileDialogResult) (FileDialogResult, error) {
	if result.Cancelled {
		return result, nil
	}
	if len(result.Paths) == 0 {
		return FileDialogResult{}, &FileDialogError{
			Kind: FileDialogErrorPath,
			Err:  fmt.Errorf("no selected paths returned"),
		}
	}

	paths := make([]string, 0, len(result.Paths))
	for _, path := range result.Paths {
		if strings.TrimSpace(path) == "" {
			return FileDialogResult{}, &FileDialogError{
				Kind: FileDialogErrorPath,
				Path: path,
				Err:  fmt.Errorf("empty selected path"),
			}
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return FileDialogResult{}, &FileDialogError{
				Kind: FileDialogErrorPath,
				Path: path,
				Err:  err,
			}
		}
		paths = append(paths, abs)
	}
	result.Paths = paths
	return result, nil
}

func finalizeFileDialogResult(mode FileDialogMode, opts fileDialogOptions, result FileDialogResult) FileDialogResult {
	if mode != FileDialogSaveFile || opts.defaultExtension == "" || len(result.Paths) == 0 {
		return result
	}
	paths := append([]string(nil), result.Paths...)
	if filepath.Ext(paths[0]) == "" {
		paths[0] += "." + opts.defaultExtension
	}
	result.Paths = paths
	return result
}

func rememberedFileDialogDir(key string) string {
	if key == "" {
		return ""
	}
	fileDialogRememberedDirsMu.Lock()
	dir := fileDialogRememberedDirs[key]
	fileDialogRememberedDirsMu.Unlock()
	return dir
}

func rememberFileDialogResultDir(mode FileDialogMode, key string, result FileDialogResult) {
	if key == "" || len(result.Paths) == 0 {
		return
	}
	dir := result.Paths[0]
	if mode != FileDialogPickFolder {
		dir = filepath.Dir(dir)
	}
	if dir == "." || dir == "" {
		return
	}
	fileDialogRememberedDirsMu.Lock()
	fileDialogRememberedDirs[key] = dir
	fileDialogRememberedDirsMu.Unlock()
}
