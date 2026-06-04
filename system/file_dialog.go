package system

import (
	"context"
	"fmt"
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

// FileDialogOption configures native file dialogs.
type FileDialogOption func(*fileDialogOptions)

type fileDialogOptions struct {
	title            string
	defaultDir       string
	defaultName      string
	filters          []FileFilter
	owner            uintptr
	allowCreateDirs  bool
	allowMissingPath bool
	overwritePrompt  bool
}

type fileDialogDriver interface {
	openFileDialog(ctx context.Context, mode FileDialogMode, opts fileDialogOptions) (FileDialogResult, error)
}

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

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	fd, ok := d.(fileDialogDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityFileDialog) {
		return FileDialogResult{}, fmt.Errorf("system: %s: %w", CapabilityFileDialog, ErrUnsupported)
	}
	return fd.openFileDialog(ctx, mode, opts)
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
