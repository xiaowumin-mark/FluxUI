package ui

import (
	"image/color"
	"time"

	internal "github.com/xiaowumin-mark/FluxUI/internal"
	widget "github.com/xiaowumin-mark/FluxUI/widget"
)

type Axis = widget.Axis

const (
	Horizontal Axis = widget.Horizontal
	Vertical   Axis = widget.Vertical
)

type ImageSource = widget.ImageSource
type ImageFit = widget.ImageFit

const (
	ImageFitContain = widget.ImageFitContain
	ImageFitCover   = widget.ImageFitCover
	ImageFitFill    = widget.ImageFitFill
	ImageFitNone    = widget.ImageFitNone
)

type DividerOption = widget.DividerOption
type ImageOption = widget.ImageOption
type IconOption = widget.IconOption
type CardOption = widget.CardOption
type RadioGroupOption = widget.RadioGroupOption
type RadioItem = widget.RadioItem
type ProgressOption = widget.ProgressOption
type TabsOption = widget.TabsOption
type TabItem = widget.TabItem
type DialogOption = widget.DialogOption
type PopupOption = widget.PopupOption
type ToastOption = widget.ToastOption
type ToastType = widget.ToastType
type ToastPosition = widget.ToastPosition
type ScrollOption = widget.ScrollOption
type ScrollRef = widget.ScrollRef
type ClickAreaOption = widget.ClickAreaOption
type RadioGroupRef = widget.RadioGroupRef
type TabsRef = widget.TabsRef
type DialogRef = widget.DialogRef
type BottomNavRef = widget.BottomNavRef
type ListOption = widget.ListOption
type GridOption = widget.GridOption
type AppBarOption = widget.AppBarOption
type BottomNavOption = widget.BottomNavOption
type BottomNavAlignment = widget.BottomNavAlignment
type NavItem = widget.NavItem
type SelectValue = string

const (
	ToastInfo    ToastType = widget.ToastInfo
	ToastSuccess ToastType = widget.ToastSuccess
	ToastWarning ToastType = widget.ToastWarning
	ToastError   ToastType = widget.ToastError
)

const (
	ToastTop    ToastPosition = widget.ToastTop
	ToastCenter ToastPosition = widget.ToastCenter
	ToastBottom ToastPosition = widget.ToastBottom
)

const (
	BottomNavAlignStart       BottomNavAlignment = widget.BottomNavAlignStart
	BottomNavAlignCenter      BottomNavAlignment = widget.BottomNavAlignCenter
	BottomNavAlignEnd         BottomNavAlignment = widget.BottomNavAlignEnd
	BottomNavAlignSpaceEvenly BottomNavAlignment = widget.BottomNavAlignSpaceEvenly
)

type SelectOptionItem[T comparable] = widget.SelectOptionItem[T]
type SelectOption[T comparable] = widget.SelectOption[T]
type SelectRef[T comparable] = widget.SelectRef[T]

func Spacer(width, height float32) Widget {
	return widget.Spacer(width, height)
}

func ClickArea(child Widget, onClick func(ctx *Context), opts ...ClickAreaOption) Widget {
	return widget.ClickArea(child, onClick, opts...)
}

func NewClickAreaRef() *ClickAreaRef {
	return widget.NewClickAreaRef()
}

func ClickAreaAttachRef(ref *ClickAreaRef) ClickAreaOption {
	return widget.ClickAreaAttachRef(ref)
}

func FixedWidth(width float32, child Widget) Widget {
	return widget.FixedWidth(width, child)
}

func FixedHeight(height float32, child Widget) Widget {
	return widget.FixedHeight(height, child)
}

func FixedSize(width, height float32, child Widget) Widget {
	return widget.FixedSize(width, height, child)
}

func FillWidth(child Widget) Widget {
	return widget.FillWidth(child)
}

func FillHeight(child Widget) Widget {
	return widget.FillHeight(child)
}

func Fill(child Widget) Widget {
	return widget.Fill(child)
}

func HSpacer(width float32) Widget {
	return widget.HSpacer(width)
}

func VSpacer(height float32) Widget {
	return widget.VSpacer(height)
}

func Divider(opts ...DividerOption) Widget {
	return widget.Divider(opts...)
}

func DividerVertical(vertical bool) DividerOption {
	return widget.DividerVertical(vertical)
}

func DividerThickness(thickness float32) DividerOption {
	return widget.DividerThickness(thickness)
}

func DividerColor(col color.NRGBA) DividerOption {
	return widget.DividerColor(col)
}

func DividerLength(length float32) DividerOption {
	return widget.DividerLength(length)
}

func DividerMargin(insets Insets) DividerOption {
	return widget.DividerMargin(insets)
}

func Image(src ImageSource, opts ...ImageOption) Widget {
	return widget.Image(src, opts...)
}

func ImageWidth(width float32) ImageOption {
	return widget.ImageWidth(width)
}

func ImageHeight(height float32) ImageOption {
	return widget.ImageHeight(height)
}

func ImageFitMode(fit ImageFit) ImageOption {
	return widget.ImageFitMode(fit)
}

func ImageRadius(radius float32) ImageOption {
	return widget.ImageRadius(radius)
}

func ImageBackground(col color.NRGBA) ImageOption {
	return widget.ImageBackground(col)
}

func ImageOnClick(fn func(ctx *Context)) ImageOption {
	return widget.ImageOnClick(fn)
}

func ImageAttachRef(ref *ButtonRef) ImageOption {
	return widget.ImageAttachRef(ref)
}

func Icon(name string, opts ...IconOption) Widget {
	return widget.Icon(name, opts...)
}

func IconSize(size float32) IconOption {
	return widget.IconSize(size)
}

func IconColor(col color.NRGBA) IconOption {
	return widget.IconColor(col)
}

func IconOnClick(fn func(ctx *Context)) IconOption {
	return widget.IconOnClick(fn)
}

func IconAttachRef(ref *ButtonRef) IconOption {
	return widget.IconAttachRef(ref)
}

func Card(child Widget, opts ...CardOption) Widget {
	return widget.Card(child, opts...)
}

func CardPadding(insets Insets) CardOption {
	return widget.CardPadding(insets)
}

func CardRadius(radius float32) CardOption {
	return widget.CardRadius(radius)
}

func CardBackground(col color.NRGBA) CardOption {
	return widget.CardBackground(col)
}

func CardBorder(col color.NRGBA, width float32) CardOption {
	return widget.CardBorder(col, width)
}

func CardShadow(level int) CardOption {
	return widget.CardShadow(level)
}

func CardOnClick(fn func(ctx *Context)) CardOption {
	return widget.CardOnClick(fn)
}

func CardAttachRef(ref *ButtonRef) CardOption {
	return widget.CardAttachRef(ref)
}

func RadioGroup(value string, items []RadioItem, opts ...RadioGroupOption) Widget {
	return widget.RadioGroup(value, items, opts...)
}

func RadioGroupDirection(axis Axis) RadioGroupOption {
	return widget.RadioGroupDirection(axis)
}

func RadioGroupDisabled(disabled bool) RadioGroupOption {
	return widget.RadioGroupDisabled(disabled)
}

func RadioGroupOnChange(fn func(ctx *Context, value string)) RadioGroupOption {
	return widget.RadioGroupOnChange(fn)
}

func RadioGroupSize(size float32) RadioGroupOption {
	return widget.RadioGroupSize(size)
}

func RadioGroupColor(col color.NRGBA) RadioGroupOption {
	return widget.RadioGroupColor(col)
}

func NewRadioGroupRef() *RadioGroupRef {
	return widget.NewRadioGroupRef()
}

func RadioGroupAttachRef(ref *RadioGroupRef) RadioGroupOption {
	return widget.RadioGroupAttachRef(ref)
}

func Select[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget {
	return widget.Select(value, options, opts...)
}

func SelectPlaceholder[T comparable](text string) SelectOption[T] {
	return widget.SelectPlaceholder[T](text)
}

func SelectDisabled[T comparable](disabled bool) SelectOption[T] {
	return widget.SelectDisabled[T](disabled)
}

func SelectSearchable[T comparable](searchable bool) SelectOption[T] {
	return widget.SelectSearchable[T](searchable)
}

func SelectMaxHeight[T comparable](height float32) SelectOption[T] {
	return widget.SelectMaxHeight[T](height)
}

func SelectOnChange[T comparable](fn func(ctx *Context, value T)) SelectOption[T] {
	return widget.SelectOnChange[T](fn)
}

func SelectOnOpenChange[T comparable](fn func(ctx *Context, opened bool)) SelectOption[T] {
	return widget.SelectOnOpenChange[T](fn)
}

func NewSelectRef[T comparable]() *SelectRef[T] {
	return widget.NewSelectRef[T]()
}

func SelectAttachRef[T comparable](ref *SelectRef[T]) SelectOption[T] {
	return widget.SelectAttachRef[T](ref)
}

func ProgressBar(value float32, opts ...ProgressOption) Widget {
	return widget.ProgressBar(value, opts...)
}

func CircularProgress(value float32, opts ...ProgressOption) Widget {
	return widget.CircularProgress(value, opts...)
}

func ProgressMin(min float32) ProgressOption {
	return widget.ProgressMin(min)
}

func ProgressMax(max float32) ProgressOption {
	return widget.ProgressMax(max)
}

func ProgressIndeterminate(indeterminate bool) ProgressOption {
	return widget.ProgressIndeterminate(indeterminate)
}

func ProgressThickness(thickness float32) ProgressOption {
	return widget.ProgressThickness(thickness)
}

func ProgressTrackColor(col color.NRGBA) ProgressOption {
	return widget.ProgressTrackColor(col)
}

func ProgressFillColor(col color.NRGBA) ProgressOption {
	return widget.ProgressFillColor(col)
}

func ProgressSize(size float32) ProgressOption {
	return widget.ProgressSize(size)
}

func Tabs(active string, items []TabItem, opts ...TabsOption) Widget {
	return widget.Tabs(active, items, opts...)
}

func TabsOnChange(fn func(ctx *Context, key string)) TabsOption {
	return widget.TabsOnChange(fn)
}

func TabsScrollable(scrollable bool) TabsOption {
	return widget.TabsScrollable(scrollable)
}

func TabsIndicatorColor(col color.NRGBA) TabsOption {
	return widget.TabsIndicatorColor(col)
}

func TabsTextColor(col color.NRGBA) TabsOption {
	return widget.TabsTextColor(col)
}

func TabsActiveTextColor(col color.NRGBA) TabsOption {
	return widget.TabsActiveTextColor(col)
}

func NewTabsRef() *TabsRef {
	return widget.NewTabsRef()
}

func TabsAttachRef(ref *TabsRef) TabsOption {
	return widget.TabsAttachRef(ref)
}

func Dialog(open bool, child Widget, opts ...DialogOption) Widget {
	return widget.Dialog(open, child, opts...)
}

func DialogTitle(title string) DialogOption {
	return widget.DialogTitle(title)
}

func DialogWidth(width float32) DialogOption {
	return widget.DialogWidth(width)
}

func DialogRadius(radius float32) DialogOption {
	return widget.DialogRadius(radius)
}

func DialogMaskClosable(maskClosable bool) DialogOption {
	return widget.DialogMaskClosable(maskClosable)
}

func DialogOnOpenChange(fn func(ctx *Context, open bool)) DialogOption {
	return widget.DialogOnOpenChange(fn)
}

func DialogOnConfirm(fn func(ctx *Context)) DialogOption {
	return widget.DialogOnConfirm(fn)
}

func DialogOnCancel(fn func(ctx *Context)) DialogOption {
	return widget.DialogOnCancel(fn)
}

func NewDialogRef() *DialogRef {
	return widget.NewDialogRef()
}

func DialogAttachRef(ref *DialogRef) DialogOption {
	return widget.DialogAttachRef(ref)
}

func DialogConfirmText(text string) DialogOption {
	return widget.DialogConfirmText(text)
}

func DialogCancelText(text string) DialogOption {
	return widget.DialogCancelText(text)
}

func Popup(open bool, child Widget, opts ...PopupOption) Widget {
	return widget.Popup(open, child, opts...)
}

func PopupWidth(width float32) PopupOption {
	return widget.PopupWidth(width)
}

func PopupRadius(radius float32) PopupOption {
	return widget.PopupRadius(radius)
}

func PopupMaskClosable(maskClosable bool) PopupOption {
	return widget.PopupMaskClosable(maskClosable)
}

func PopupBackground(bg color.NRGBA) PopupOption {
	return widget.PopupBackground(bg)
}

func PopupPadding(insets Insets) PopupOption {
	return widget.PopupPadding(insets)
}

func PopupOnOpenChange(fn func(ctx *Context, open bool)) PopupOption {
	return widget.PopupOnOpenChange(fn)
}

func PopupAttachRef(ref *DialogRef) PopupOption {
	return widget.PopupAttachRef(ref)
}

func Toast(message string, opts ...ToastOption) Widget {
	return widget.Toast(message, opts...)
}

func ToastTypeOf(kind ToastType) ToastOption {
	return widget.ToastTypeOf(kind)
}

func ToastDuration(duration time.Duration) ToastOption {
	return widget.ToastDuration(duration)
}

func ToastPositionOf(position ToastPosition) ToastOption {
	return widget.ToastPositionOf(position)
}

func ToastOnClose(fn func(ctx *Context)) ToastOption {
	return widget.ToastOnClose(fn)
}

func ScrollView(child Widget, opts ...ScrollOption) Widget {
	return widget.ScrollView(child, opts...)
}

func ScrollVertical(vertical bool) ScrollOption {
	return widget.ScrollVertical(vertical)
}

func ScrollHorizontal(horizontal bool) ScrollOption {
	return widget.ScrollHorizontal(horizontal)
}

func ScrollBarVisible(visible bool) ScrollOption {
	return widget.ScrollBarVisible(visible)
}

func ScrollOnChange(fn func(ctx *Context, x, y float32)) ScrollOption {
	return widget.ScrollOnChange(fn)
}

func NewScrollRef() *ScrollRef {
	return widget.NewScrollRef()
}

func ScrollAttachRef(ref *ScrollRef) ScrollOption {
	return widget.ScrollAttachRef(ref)
}

func ScrollAutoToEnd(enabled bool) ScrollOption {
	return widget.ScrollAutoToEnd(enabled)
}

func ScrollAutoToEndKey(key any) ScrollOption {
	return widget.ScrollAutoToEndKey(key)
}

func ListView(count int, itemBuilder func(ctx *Context, index int) Widget, opts ...ListOption) Widget {
	return widget.ListView(count, func(ctx *internal.Context, index int) widget.Widget {
		return itemBuilder(ctx, index)
	}, opts...)
}

func ListAxis(axis Axis) ListOption {
	return widget.ListAxis(axis)
}

func ListVirtualized(virtualized bool) ListOption {
	return widget.ListVirtualized(virtualized)
}

func ListItemSpacing(spacing float32) ListOption {
	return widget.ListItemSpacing(spacing)
}

func ListPadding(insets Insets) ListOption {
	return widget.ListPadding(insets)
}

func ListOnReachEnd(fn func(ctx *Context)) ListOption {
	return widget.ListOnReachEnd(fn)
}

func Grid(columns int, children ...Widget) Widget {
	return widget.Grid(columns, children...)
}

func GridView(count int, columns int, itemBuilder func(ctx *Context, index int) Widget, opts ...GridOption) Widget {
	return widget.GridView(count, columns, func(ctx *internal.Context, index int) widget.Widget {
		return itemBuilder(ctx, index)
	}, opts...)
}

func GridGap(rowGap, colGap float32) GridOption {
	return widget.GridGap(rowGap, colGap)
}

func GridPadding(insets Insets) GridOption {
	return widget.GridPadding(insets)
}

func GridMinItemWidth(width float32) GridOption {
	return widget.GridMinItemWidth(width)
}

func GridOnReachEnd(fn func(ctx *Context)) GridOption {
	return widget.GridOnReachEnd(fn)
}

func AppBar(title Widget, opts ...AppBarOption) Widget {
	return widget.AppBar(title, opts...)
}

func AppBarLeading(leading Widget) AppBarOption {
	return widget.AppBarLeading(leading)
}

func AppBarActions(actions ...Widget) AppBarOption {
	return widget.AppBarActions(actions...)
}

func AppBarHeight(height float32) AppBarOption {
	return widget.AppBarHeight(height)
}

func AppBarBackground(col color.NRGBA) AppBarOption {
	return widget.AppBarBackground(col)
}

func BottomNavigation(active string, items []NavItem, opts ...BottomNavOption) Widget {
	return widget.BottomNavigation(active, items, opts...)
}

func BottomNavOnChange(fn func(ctx *Context, key string)) BottomNavOption {
	return widget.BottomNavOnChange(fn)
}

func BottomNavBackground(col color.NRGBA) BottomNavOption {
	return widget.BottomNavBackground(col)
}

func BottomNavActiveColor(col color.NRGBA) BottomNavOption {
	return widget.BottomNavActiveColor(col)
}

func BottomNavInactiveColor(col color.NRGBA) BottomNavOption {
	return widget.BottomNavInactiveColor(col)
}

func BottomNavAlignmentOf(alignment BottomNavAlignment) BottomNavOption {
	return widget.BottomNavAlignmentOf(alignment)
}

func NewBottomNavRef() *BottomNavRef {
	return widget.NewBottomNavRef()
}

func BottomNavAttachRef(ref *BottomNavRef) BottomNavOption {
	return widget.BottomNavAttachRef(ref)
}
