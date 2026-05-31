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
type PressableOption = widget.PressableOption
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
type PopupRef = widget.PopupRef
type MenuItem = widget.MenuItem
type MenuOption = widget.MenuOption
type DropdownMenuOption = widget.DropdownMenuOption
type ListItemOption = widget.ListItemOption
type IconButtonOption = widget.IconButtonOption
type FloatingActionButtonOption = widget.FloatingActionButtonOption
type NavigationRailOption = widget.NavigationRailOption
type NavigationDrawerOption = widget.NavigationDrawerOption
type SnackbarOption = widget.SnackbarOption
type TooltipOption = widget.TooltipOption
type BadgeOption = widget.BadgeOption
type ChipOption = widget.ChipOption
type SearchBarOption = widget.SearchBarOption

// ElementNavItem 描述 BottomNavigationElement 的单项配置。
type ElementNavItem struct {
	Key   string
	Label string
	Icon  Element
}

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

type singleChildElement struct {
	kind     string
	child    Element
	renderFn func(child Widget) Widget
}

type multiChildElement struct {
	kind     string
	children []Element
	renderFn func(children []Widget) Widget
}

type appBarElement struct {
	title   Element
	leading Element
	actions []Element
	opts    []AppBarOption
}

type withFontElement struct {
	font  FontSpec
	child Element
}

type flexElement struct {
	kind   string
	weight float32
	child  Element
}

type listViewElement struct {
	count   int
	builder func(ctx *Context, index int) Element
	opts    []ListOption
}

type gridElement struct {
	columns  int
	children []Element
}

type gridViewElement struct {
	count   int
	columns int
	builder func(ctx *Context, index int) Element
	opts    []GridOption
}

func Spacer(width, height float32) Widget {
	return widget.Spacer(width, height)
}

// TextElement 创建可参与 reconciler 的静态文本 Element。
func TextElement(content string, opts ...TextOption) Element {
	return FromWidget(widget.Text(content, opts...))
}

// SpacerElement 创建可参与 reconciler 的静态空白 Element。
func SpacerElement(width, height float32) Element {
	return FromWidget(widget.Spacer(width, height))
}

// DividerElement 创建可参与 reconciler 的静态分割线 Element。
func DividerElement(opts ...DividerOption) Element {
	return FromWidget(widget.Divider(opts...))
}

// ColumnElement 创建可参与 reconciler 的纵向布局 Element。
func ColumnElement(children ...Element) Element {
	return &layoutElement{kind: "column", children: append([]Element(nil), children...)}
}

// RowElement 创建可参与 reconciler 的横向布局 Element。
func RowElement(children ...Element) Element {
	return &layoutElement{kind: "row", children: append([]Element(nil), children...)}
}

// StackElement 创建可参与 reconciler 的堆叠布局 Element。
func StackElement(children ...Element) Element {
	return &layoutElement{kind: "stack", children: append([]Element(nil), children...)}
}

// CenterElement 创建可参与 reconciler 的居中布局 Element。
func CenterElement(child Element) Element {
	return &layoutElement{kind: "center", child: child}
}

// PaddingElement 创建可参与 reconciler 的内边距布局 Element。
func PaddingElement(insets Insets, child Element) Element {
	return &layoutElement{kind: "padding", insets: insets, child: child}
}

// Deprecated: ContainerElement 已被 ContainerDecorationElement 取代。
func ContainerElement(st Style, child Element) Element {
	return &layoutElement{kind: "container", style: st, child: child}
}

// ContainerDecorationElement 创建可参与 reconciler 的装饰容器 Element。
func ContainerDecorationElement(d Decoration, child Element, opts ...ContainerDecorationOption) Element {
	return &layoutElement{kind: "container-decoration", decoration: d, decoOpts: append([]ContainerDecorationOption(nil), opts...), child: child}
}

// ButtonElement 创建可参与 reconciler 的按钮 Element。
func ButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "button", child: child, renderFn: func(child Widget) Widget {
		return widget.Button(child, opts...)
	}}
}

func FilledButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "filled-button", child: child, renderFn: func(child Widget) Widget {
		return widget.FilledButton(child, opts...)
	}}
}

func FilledTonalButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "filled-tonal-button", child: child, renderFn: func(child Widget) Widget {
		return widget.FilledTonalButton(child, opts...)
	}}
}

func OutlinedButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "outlined-button", child: child, renderFn: func(child Widget) Widget {
		return widget.OutlinedButton(child, opts...)
	}}
}

func TextButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "text-button", child: child, renderFn: func(child Widget) Widget {
		return widget.TextButton(child, opts...)
	}}
}

func ElevatedButtonElement(child Element, opts ...ButtonOption) Element {
	return &singleChildElement{kind: "elevated-button", child: child, renderFn: func(child Widget) Widget {
		return widget.ElevatedButton(child, opts...)
	}}
}

// ClickAreaElement 创建可参与 reconciler 的可点击区域 Element。
//
// Deprecated: 请使用 PressableElement。
func ClickAreaElement(child Element, onClick func(ctx *Context), opts ...ClickAreaOption) Element {
	return PressableElement(child, onClick, opts...)
}

// PressableElement 创建无固定视觉样式、可参与 reconciler 的通用可点击区域 Element。
func PressableElement(child Element, onClick func(ctx *Context), opts ...PressableOption) Element {
	return &singleChildElement{kind: "pressable", child: child, renderFn: func(child Widget) Widget {
		return widget.Pressable(child, onClick, opts...)
	}}
}

// CheckboxElement 创建可参与 reconciler 的复选框 Element。
func CheckboxElement(label string, checked bool, opts ...CheckboxOption) Element {
	return FromWidget(widget.Checkbox(label, checked, opts...))
}

// SwitchElement 创建可参与 reconciler 的开关 Element。
func SwitchElement(checked bool, opts ...SwitchOption) Element {
	return FromWidget(widget.Switch(checked, opts...))
}

// TextFieldElement 创建可参与 reconciler 的受控输入框 Element。
func TextFieldElement(value string, opts ...InputOption) Element {
	return FromWidget(widget.TextField(value, opts...))
}

func OutlinedTextFieldElement(value string, opts ...InputOption) Element {
	return FromWidget(widget.OutlinedTextField(value, opts...))
}

func FilledTextFieldElement(value string, opts ...InputOption) Element {
	return FromWidget(widget.FilledTextField(value, opts...))
}

// SliderElement 创建可参与 reconciler 的受控滑块 Element。
func SliderElement(value float32, opts ...SliderOption) Element {
	return FromWidget(widget.Slider(value, opts...))
}

// RadioGroupElement 创建可参与 reconciler 的受控单选组 Element。
func RadioGroupElement(value string, items []RadioItem, opts ...RadioGroupOption) Element {
	return FromWidget(widget.RadioGroup(value, items, opts...))
}

// SelectElement 创建可参与 reconciler 的受控下拉选择 Element。
func SelectElement[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Element {
	return FromWidget(widget.Select(value, options, opts...))
}

func MenuElement(items []MenuItem, opts ...MenuOption) Element {
	return FromWidget(widget.Menu(items, opts...))
}

func DropdownMenuElement(open bool, trigger Element, items []MenuItem, opts ...DropdownMenuOption) Element {
	return &singleChildElement{kind: "dropdown-menu", child: trigger, renderFn: func(child Widget) Widget {
		return widget.DropdownMenu(open, child, items, opts...)
	}}
}

func ListItemElement(headline string, opts ...ListItemOption) Element {
	return FromWidget(widget.ListItem(headline, opts...))
}

func ListItemElementWithSlots(headline, supporting, leading, trailing Element, opts ...ListItemOption) Element {
	return &multiChildElement{kind: "list-item", children: []Element{headline, supporting, leading, trailing}, renderFn: func(children []Widget) Widget {
		var h, s, l, t Widget
		if len(children) > 0 {
			h = children[0]
		}
		if len(children) > 1 {
			s = children[1]
		}
		if len(children) > 2 {
			l = children[2]
		}
		if len(children) > 3 {
			t = children[3]
		}
		return widget.ListItemWithSlots(h, s, l, t, opts...)
	}}
}

func IconButtonElement(child Element, opts ...IconButtonOption) Element {
	return &singleChildElement{kind: "icon-button", child: child, renderFn: func(child Widget) Widget {
		return widget.IconButton(child, opts...)
	}}
}

func FilledIconButtonElement(child Element, opts ...IconButtonOption) Element {
	return &singleChildElement{kind: "filled-icon-button", child: child, renderFn: func(child Widget) Widget {
		return widget.FilledIconButton(child, opts...)
	}}
}

func FilledTonalIconButtonElement(child Element, opts ...IconButtonOption) Element {
	return &singleChildElement{kind: "filled-tonal-icon-button", child: child, renderFn: func(child Widget) Widget {
		return widget.FilledTonalIconButton(child, opts...)
	}}
}

func OutlinedIconButtonElement(child Element, opts ...IconButtonOption) Element {
	return &singleChildElement{kind: "outlined-icon-button", child: child, renderFn: func(child Widget) Widget {
		return widget.OutlinedIconButton(child, opts...)
	}}
}

func FloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element {
	return &singleChildElement{kind: "floating-action-button", child: icon, renderFn: func(child Widget) Widget {
		return widget.FloatingActionButton(child, opts...)
	}}
}

func SmallFloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element {
	return &singleChildElement{kind: "small-floating-action-button", child: icon, renderFn: func(child Widget) Widget {
		return widget.SmallFloatingActionButton(child, opts...)
	}}
}

func LargeFloatingActionButtonElement(icon Element, opts ...FloatingActionButtonOption) Element {
	return &singleChildElement{kind: "large-floating-action-button", child: icon, renderFn: func(child Widget) Widget {
		return widget.LargeFloatingActionButton(child, opts...)
	}}
}

func ExtendedFloatingActionButtonElement(icon, label Element, opts ...FloatingActionButtonOption) Element {
	return &multiChildElement{kind: "extended-floating-action-button", children: []Element{icon, label}, renderFn: func(children []Widget) Widget {
		var iconWidget, labelWidget Widget
		if len(children) > 0 {
			iconWidget = children[0]
		}
		if len(children) > 1 {
			labelWidget = children[1]
		}
		return widget.ExtendedFloatingActionButton(iconWidget, labelWidget, opts...)
	}}
}

// ProgressBarElement 创建可参与 reconciler 的线性进度 Element。
func ProgressBarElement(value float32, opts ...ProgressOption) Element {
	return FromWidget(widget.ProgressBar(value, opts...))
}

func LinearProgressIndicatorElement(value float32, opts ...ProgressOption) Element {
	return FromWidget(widget.LinearProgressIndicator(value, opts...))
}

// CircularProgressElement 创建可参与 reconciler 的环形进度 Element。
func CircularProgressElement(value float32, opts ...ProgressOption) Element {
	return FromWidget(widget.CircularProgress(value, opts...))
}

func CircularProgressIndicatorElement(value float32, opts ...ProgressOption) Element {
	return FromWidget(widget.CircularProgressIndicator(value, opts...))
}

// ImageElement 创建可参与 reconciler 的图片 Element。
func ImageElement(src ImageSource, opts ...ImageOption) Element {
	return FromWidget(widget.Image(src, opts...))
}

// IconElement 创建可参与 reconciler 的图标 Element。
func IconElement(name string, opts ...IconOption) Element {
	return FromWidget(widget.Icon(name, opts...))
}

// CardElement 创建可参与 reconciler 的卡片 Element。
func CardElement(child Element, opts ...CardOption) Element {
	return &singleChildElement{kind: "card", child: child, renderFn: func(child Widget) Widget {
		return widget.Card(child, opts...)
	}}
}

func FilledCardElement(child Element, opts ...CardOption) Element {
	return &singleChildElement{kind: "filled-card", child: child, renderFn: func(child Widget) Widget {
		return widget.FilledCard(child, opts...)
	}}
}

func ElevatedCardElement(child Element, opts ...CardOption) Element {
	return &singleChildElement{kind: "elevated-card", child: child, renderFn: func(child Widget) Widget {
		return widget.ElevatedCard(child, opts...)
	}}
}

func OutlinedCardElement(child Element, opts ...CardOption) Element {
	return &singleChildElement{kind: "outlined-card", child: child, renderFn: func(child Widget) Widget {
		return widget.OutlinedCard(child, opts...)
	}}
}

// TabsElement 创建可参与 reconciler 的标签栏 Element。
func TabsElement(active string, items []TabItem, opts ...TabsOption) Element {
	return FromWidget(widget.Tabs(active, items, opts...))
}

// DialogElement 创建可参与 reconciler 的对话框 Element。
func DialogElement(open bool, child Element, opts ...DialogOption) Element {
	return &singleChildElement{kind: "dialog", child: child, renderFn: func(child Widget) Widget {
		return widget.Dialog(open, child, opts...)
	}}
}

// PopupElement 创建可参与 reconciler 的自定义弹窗 Element。
func PopupElement(open bool, child Element, opts ...PopupOption) Element {
	return &singleChildElement{kind: "popup", child: child, renderFn: func(child Widget) Widget {
		return widget.Popup(open, child, opts...)
	}}
}

// ToastElement 创建可参与 reconciler 的吐司 Element。
func ToastElement(message string, opts ...ToastOption) Element {
	return FromWidget(widget.Toast(message, opts...))
}

func SnackbarElement(message string, opts ...SnackbarOption) Element {
	return FromWidget(widget.Snackbar(message, opts...))
}

func TooltipElement(label string, child Element, opts ...TooltipOption) Element {
	return &singleChildElement{kind: "tooltip", child: child, renderFn: func(child Widget) Widget {
		return widget.Tooltip(label, child, opts...)
	}}
}

func BadgeElement(child Element, label string, opts ...BadgeOption) Element {
	return &singleChildElement{kind: "badge", child: child, renderFn: func(child Widget) Widget {
		return widget.Badge(child, label, opts...)
	}}
}

func AssistChipElement(label string, opts ...ChipOption) Element {
	return FromWidget(widget.AssistChip(label, opts...))
}

func FilterChipElement(label string, opts ...ChipOption) Element {
	return FromWidget(widget.FilterChip(label, opts...))
}

func InputChipElement(label string, opts ...ChipOption) Element {
	return FromWidget(widget.InputChip(label, opts...))
}

func SuggestionChipElement(label string, opts ...ChipOption) Element {
	return FromWidget(widget.SuggestionChip(label, opts...))
}

func ChipElementWithSlots(label Element, opts ...ChipOption) Element {
	return &singleChildElement{kind: "chip", child: label, renderFn: func(child Widget) Widget {
		return widget.ChipWithSlots(child, opts...)
	}}
}

func SearchBarElement(value string, opts ...SearchBarOption) Element {
	return FromWidget(widget.SearchBar(value, opts...))
}

// ScrollViewElement 创建可参与 reconciler 的滚动容器 Element。
func ScrollViewElement(child Element, opts ...ScrollOption) Element {
	return &singleChildElement{kind: "scroll-view", child: child, renderFn: func(child Widget) Widget {
		return widget.ScrollView(child, opts...)
	}}
}

// ListViewElement 创建可参与 reconciler 的列表 Element。
func ListViewElement(count int, itemBuilder func(ctx *Context, index int) Element, opts ...ListOption) Element {
	return &listViewElement{count: count, builder: itemBuilder, opts: append([]ListOption(nil), opts...)}
}

// GridElement 创建可参与 reconciler 的固定网格 Element。
func GridElement(columns int, children ...Element) Element {
	return &gridElement{columns: columns, children: append([]Element(nil), children...)}
}

// GridViewElement 创建可参与 reconciler 的动态网格列表 Element。
func GridViewElement(count int, columns int, itemBuilder func(ctx *Context, index int) Element, opts ...GridOption) Element {
	return &gridViewElement{count: count, columns: columns, builder: itemBuilder, opts: append([]GridOption(nil), opts...)}
}

// AppBarElement 创建可参与 reconciler 的顶部导航栏 Element。
func AppBarElement(title Element, opts ...AppBarOption) Element {
	return &appBarElement{title: title, opts: append([]AppBarOption(nil), opts...)}
}

// AppBarElementWithSlots 创建包含 leading/actions Element 插槽的顶部导航栏。
func AppBarElementWithSlots(title Element, leading Element, actions []Element, opts ...AppBarOption) Element {
	return &appBarElement{
		title:   title,
		leading: leading,
		actions: append([]Element(nil), actions...),
		opts:    append([]AppBarOption(nil), opts...),
	}
}

// BottomNavigationElement 创建可参与 reconciler 的底部导航 Element。
func BottomNavigationElement(active string, items []ElementNavItem, opts ...BottomNavOption) Element {
	children := make([]Element, 0, len(items))
	for _, item := range items {
		children = append(children, item.Icon)
	}
	return &multiChildElement{kind: "bottom-navigation", children: children, renderFn: func(children []Widget) Widget {
		legacyItems := make([]widget.NavItem, 0, len(items))
		for idx, item := range items {
			var icon Widget
			if idx < len(children) {
				icon = children[idx]
			}
			legacyItems = append(legacyItems, widget.NavItem{Key: item.Key, Label: item.Label, Icon: icon})
		}
		return widget.BottomNavigation(active, legacyItems, opts...)
	}}
}

func NavigationRailElement(active string, items []ElementNavItem, opts ...NavigationRailOption) Element {
	children := make([]Element, 0, len(items))
	for _, item := range items {
		children = append(children, item.Icon)
	}
	return &multiChildElement{kind: "navigation-rail", children: children, renderFn: func(children []Widget) Widget {
		legacyItems := make([]widget.NavItem, 0, len(items))
		for idx, item := range items {
			var icon Widget
			if idx < len(children) {
				icon = children[idx]
			}
			legacyItems = append(legacyItems, widget.NavItem{Key: item.Key, Label: item.Label, Icon: icon})
		}
		return widget.NavigationRail(active, legacyItems, opts...)
	}}
}

func NavigationDrawerElement(active string, items []ElementNavItem, opts ...NavigationDrawerOption) Element {
	children := make([]Element, 0, len(items))
	for _, item := range items {
		children = append(children, item.Icon)
	}
	return &multiChildElement{kind: "navigation-drawer", children: children, renderFn: func(children []Widget) Widget {
		legacyItems := make([]widget.NavItem, 0, len(items))
		for idx, item := range items {
			var icon Widget
			if idx < len(children) {
				icon = children[idx]
			}
			legacyItems = append(legacyItems, widget.NavItem{Key: item.Key, Label: item.Label, Icon: icon})
		}
		return widget.NavigationDrawer(active, legacyItems, opts...)
	}}
}

// WithFontElement 在 Element 子树中覆盖默认字体。
func WithFontElement(font FontSpec, child Element) Element {
	if child == nil {
		return nil
	}
	return &withFontElement{font: font, child: child}
}

// FlexedElement 创建带权重的弹性 Element 子项。
func FlexedElement(weight float32, child Element) Element {
	if child == nil {
		return nil
	}
	return &flexElement{kind: "flexed", weight: weight, child: child}
}

// ExpandedElement 创建权重为 1 的弹性 Element 子项。
func ExpandedElement(child Element) Element {
	if child == nil {
		return nil
	}
	return &flexElement{kind: "expanded", weight: 1, child: child}
}

// FixedWidthElement 固定 Element 子树宽度。
func FixedWidthElement(width float32, child Element) Element {
	return &singleChildElement{kind: "fixed-width", child: child, renderFn: func(child Widget) Widget {
		return widget.FixedWidth(width, child)
	}}
}

// FixedHeightElement 固定 Element 子树高度。
func FixedHeightElement(height float32, child Element) Element {
	return &singleChildElement{kind: "fixed-height", child: child, renderFn: func(child Widget) Widget {
		return widget.FixedHeight(height, child)
	}}
}

// FixedSizeElement 固定 Element 子树尺寸。
func FixedSizeElement(width, height float32, child Element) Element {
	return &singleChildElement{kind: "fixed-size", child: child, renderFn: func(child Widget) Widget {
		return widget.FixedSize(width, height, child)
	}}
}

// FillWidthElement 让 Element 子树填满可用宽度。
func FillWidthElement(child Element) Element {
	return &singleChildElement{kind: "fill-width", child: child, renderFn: widget.FillWidth}
}

// FillHeightElement 让 Element 子树填满可用高度。
func FillHeightElement(child Element) Element {
	return &singleChildElement{kind: "fill-height", child: child, renderFn: widget.FillHeight}
}

// FillElement 让 Element 子树填满可用空间。
func FillElement(child Element) Element {
	return &singleChildElement{kind: "fill", child: child, renderFn: widget.Fill}
}

// HSpacerElement 创建水平空白 Element。
func HSpacerElement(width float32) Element {
	return FromWidget(widget.HSpacer(width))
}

// VSpacerElement 创建垂直空白 Element。
func VSpacerElement(height float32) Element {
	return FromWidget(widget.VSpacer(height))
}

func ClickArea(child Widget, onClick func(ctx *Context), opts ...ClickAreaOption) Widget {
	return Pressable(child, onClick, opts...)
}

func Pressable(child Widget, onClick func(ctx *Context), opts ...PressableOption) Widget {
	return widget.Pressable(child, onClick, opts...)
}

func NewClickAreaRef() *ClickAreaRef {
	return widget.NewClickAreaRef()
}

func NewPressableRef() *PressableRef {
	return widget.NewPressableRef()
}

func ClickAreaAttachRef(ref *ClickAreaRef) ClickAreaOption {
	return widget.ClickAreaAttachRef(ref)
}

func PressableAttachRef(ref *PressableRef) PressableOption {
	return widget.PressableAttachRef(ref)
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

func DividerDecoration(d Decoration) DividerOption {
	return widget.DividerDecoration(d)
}

type themeProviderElement struct {
	theme *Theme
	child Element
}

// ThemeProviderElement 为子树提供局部主题。子树的 Widget 通过 UseTheme 读取到该主题，
// 而非运行时全局主题。可用于实现局部深色模式等场景。
func ThemeProviderElement(th *Theme, child Element) Element {
	if child == nil {
		return nil
	}
	return &themeProviderElement{theme: th, child: child}
}

func (e *themeProviderElement) render() widget.Widget {
	ctx := &Context{}
	ctx.SetThemeOverride(e.theme)
	return renderElement(e.child)
}

func (e *themeProviderElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *themeProviderElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "theme-provider", ChildCount: 1}
}

func (e *themeProviderElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return []Element{e.child}
}

func (e *themeProviderElement) ChildContext(ctx *Context) *Context {
	if ctx == nil || e == nil || e.theme == nil {
		return ctx
	}
	return ctx.WithTheme(e.theme)
}

func (e *themeProviderElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	return firstWidget(children)
}

type layoutElement struct {
	kind       string
	style      Style
	insets     Insets
	decoration Decoration
	decoOpts   []ContainerDecorationOption
	child      Element
	children   []Element
}

func (e *layoutElement) render() widget.Widget {
	switch e.kind {
	case "column":
		return widget.Column(renderElements(e.children)...)
	case "row":
		return widget.Row(renderElements(e.children)...)
	case "stack":
		return widget.Stack(renderElements(e.children)...)
	case "center":
		return widget.Center(renderElement(e.child))
	case "padding":
		return widget.Padding(e.insets, renderElement(e.child))
	case "container":
		return widget.Container(e.style, renderElement(e.child))
	case "container-decoration":
		return widget.ContainerDecoration(e.decoration, renderElement(e.child), e.decoOpts...)
	default:
		return nil
	}
}

func (e *layoutElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *layoutElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	count := len(e.children)
	if count == 0 && e.child != nil {
		count = 1
	}
	return ElementIdentity{Kind: e.kind, ChildCount: count}
}

func (e *layoutElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	switch e.kind {
	case "column", "row", "stack":
		return append([]Element(nil), e.children...)
	case "center", "padding", "container", "container-decoration":
		return []Element{e.child}
	default:
		return nil
	}
}

func (e *layoutElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil {
		return nil
	}
	switch e.kind {
	case "column":
		return widget.Column(children...)
	case "row":
		return widget.Row(children...)
	case "stack":
		return widget.Stack(children...)
	case "center":
		return widget.Center(firstWidget(children))
	case "padding":
		return widget.Padding(e.insets, firstWidget(children))
	case "container":
		return widget.Container(e.style, firstWidget(children))
	case "container-decoration":
		return widget.ContainerDecoration(e.decoration, firstWidget(children), e.decoOpts...)
	default:
		return nil
	}
}

func renderElements(children []Element) []Widget {
	widgets := make([]Widget, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		if w := renderElement(child); w != nil {
			widgets = append(widgets, w)
		}
	}
	return widgets
}

func renderElementsWithContext(ctx *Context, children []Element) []Widget {
	widgets := make([]Widget, 0, len(children))
	for idx, child := range children {
		if child == nil {
			continue
		}
		if w := renderElementWithContextAt(ctx, child, idx, ""); w != nil {
			widgets = append(widgets, w)
		}
	}
	return widgets
}

func firstWidget(children []Widget) Widget {
	if len(children) == 0 {
		return nil
	}
	return children[0]
}

func (e *singleChildElement) render() widget.Widget {
	if e == nil || e.renderFn == nil {
		return nil
	}
	return e.renderFn(renderElement(e.child))
}

func (e *singleChildElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: e.kind, ChildCount: 1}
}

func (e *singleChildElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return []Element{e.child}
}

func (e *singleChildElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil || e.renderFn == nil {
		return nil
	}
	return e.renderFn(firstWidget(children))
}

func (e *multiChildElement) render() widget.Widget {
	if e == nil || e.renderFn == nil {
		return nil
	}
	return e.renderFn(renderElements(e.children))
}

func (e *multiChildElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: e.kind, ChildCount: len(e.children)}
}

func (e *multiChildElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return append([]Element(nil), e.children...)
}

func (e *multiChildElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil || e.renderFn == nil {
		return nil
	}
	return e.renderFn(children)
}

func (e *withFontElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	return widget.WithFont(e.font, renderElement(e.child))
}

func (e *withFontElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *withFontElement) identity() ElementIdentity {
	return ElementIdentity{Kind: "with-font", ChildCount: 1}
}

func (e *withFontElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return []Element{e.child}
}

func (e *withFontElement) ChildContext(ctx *Context) *Context {
	if ctx == nil || e == nil {
		return ctx
	}
	return ctx.WithFont(e.font)
}

func (e *withFontElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil {
		return nil
	}
	return widget.WithFont(e.font, firstWidget(children))
}

func (e *flexElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	if e.kind == "expanded" {
		return widget.Expanded(renderElement(e.child))
	}
	return widget.Flexed(e.weight, renderElement(e.child))
}

func (e *flexElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *flexElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: e.kind, ChildCount: 1}
}

func (e *flexElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return []Element{e.child}
}

func (e *flexElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil {
		return nil
	}
	if e.kind == "expanded" {
		return widget.Expanded(firstWidget(children))
	}
	return widget.Flexed(e.weight, firstWidget(children))
}

func (e *listViewElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	return widget.ListView(e.count, func(ctx *internal.Context, index int) widget.Widget {
		if e.builder == nil {
			return nil
		}
		return renderElement(e.builder(ctx, index))
	}, e.opts...)
}

func (e *listViewElement) renderWithContext(ctx *Context) widget.Widget {
	if e == nil {
		return nil
	}
	return widget.ListView(e.count, func(itemCtx *internal.Context, index int) widget.Widget {
		if e.builder == nil {
			return nil
		}
		return renderElementWithContext(itemCtx, e.builder(itemCtx, index))
	}, e.opts...)
}

func (e *listViewElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: "list-view", ChildCount: e.count}
}

func (e *gridElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	return widget.Grid(e.columns, renderElements(e.children)...)
}

func (e *gridElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *gridElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: "grid", ChildCount: len(e.children)}
}

func (e *gridElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	return append([]Element(nil), e.children...)
}

func (e *gridElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil {
		return nil
	}
	return widget.Grid(e.columns, children...)
}

func (e *gridViewElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	return widget.GridView(e.count, e.columns, func(ctx *internal.Context, index int) widget.Widget {
		if e.builder == nil {
			return nil
		}
		return renderElement(e.builder(ctx, index))
	}, e.opts...)
}

func (e *gridViewElement) renderWithContext(ctx *Context) widget.Widget {
	if e == nil {
		return nil
	}
	return widget.GridView(e.count, e.columns, func(itemCtx *internal.Context, index int) widget.Widget {
		if e.builder == nil {
			return nil
		}
		return renderElementWithContext(itemCtx, e.builder(itemCtx, index))
	}, e.opts...)
}

func (e *gridViewElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: "grid-view", ChildCount: e.count}
}

func (e *appBarElement) render() widget.Widget {
	if e == nil {
		return nil
	}
	opts := append([]AppBarOption(nil), e.opts...)
	if e.leading != nil {
		opts = append(opts, widget.AppBarLeading(renderElement(e.leading)))
	}
	if len(e.actions) > 0 {
		opts = append(opts, widget.AppBarActions(renderElements(e.actions)...))
	}
	return widget.AppBar(renderElement(e.title), opts...)
}

func (e *appBarElement) renderWithContext(ctx *Context) widget.Widget {
	return renderCompositeElementWithContext(ctx, e, 0, "")
}

func (e *appBarElement) identity() ElementIdentity {
	if e == nil {
		return ElementIdentity{}
	}
	return ElementIdentity{Kind: "app-bar", ChildCount: 1 + len(e.actions)}
}

func (e *appBarElement) HostChildren() []Element {
	if e == nil {
		return nil
	}
	children := make([]Element, 0, 2+len(e.actions))
	children = append(children, e.title, e.leading)
	children = append(children, e.actions...)
	return children
}

func (e *appBarElement) RenderWithChildren(ctx *Context, children []Widget) Widget {
	if e == nil {
		return nil
	}
	title := firstWidget(children)
	opts := append([]AppBarOption(nil), e.opts...)
	if len(children) > 1 && children[1] != nil {
		opts = append(opts, widget.AppBarLeading(children[1]))
	}
	if len(children) > 2 {
		opts = append(opts, widget.AppBarActions(children[2:]...))
	}
	return widget.AppBar(title, opts...)
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

func ImageDecoration(d Decoration) ImageOption {
	return widget.ImageDecoration(d)
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

func FilledCard(child Widget, opts ...CardOption) Widget {
	return widget.FilledCard(child, opts...)
}

func ElevatedCard(child Widget, opts ...CardOption) Widget {
	return widget.ElevatedCard(child, opts...)
}

func OutlinedCard(child Widget, opts ...CardOption) Widget {
	return widget.OutlinedCard(child, opts...)
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

func CardDecoration(d Decoration) CardOption {
	return widget.CardDecoration(d)
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

func RadioGroupDecoration(d Decoration) RadioGroupOption {
	return widget.RadioGroupDecoration(d)
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

func SelectDecoration[T comparable](d Decoration) SelectOption[T] {
	return widget.SelectDecoration[T](d)
}

func Menu(items []MenuItem, opts ...MenuOption) Widget {
	return widget.Menu(items, opts...)
}

func MenuSelectedKey(key string) MenuOption {
	return widget.MenuSelectedKey(key)
}

func MenuOnSelect(fn func(ctx *Context, key string)) MenuOption {
	return widget.MenuOnSelect(fn)
}

func MenuWidth(width float32) MenuOption {
	return widget.MenuWidth(width)
}

func MenuMaxHeight(height float32) MenuOption {
	return widget.MenuMaxHeight(height)
}

func MenuDecoration(d Decoration) MenuOption {
	return widget.MenuDecoration(d)
}

func DropdownMenu(open bool, trigger Widget, items []MenuItem, opts ...DropdownMenuOption) Widget {
	return widget.DropdownMenu(open, trigger, items, opts...)
}

func DropdownMenuSelectedKey(key string) DropdownMenuOption {
	return widget.DropdownMenuSelectedKey(key)
}

func DropdownMenuOnSelect(fn func(ctx *Context, key string)) DropdownMenuOption {
	return widget.DropdownMenuOnSelect(fn)
}

func DropdownMenuOnOpenChange(fn func(ctx *Context, open bool)) DropdownMenuOption {
	return widget.DropdownMenuOnOpenChange(fn)
}

func DropdownMenuWidth(width float32) DropdownMenuOption {
	return widget.DropdownMenuWidth(width)
}

func DropdownMenuMaxHeight(height float32) DropdownMenuOption {
	return widget.DropdownMenuMaxHeight(height)
}

func DropdownMenuDecoration(d Decoration) DropdownMenuOption {
	return widget.DropdownMenuDecoration(d)
}

func ListItem(headline string, opts ...ListItemOption) Widget {
	return widget.ListItem(headline, opts...)
}

func ListItemWithSlots(headline, supporting, leading, trailing Widget, opts ...ListItemOption) Widget {
	return widget.ListItemWithSlots(headline, supporting, leading, trailing, opts...)
}

func ListItemSelected(selected bool) ListItemOption {
	return widget.ListItemSelected(selected)
}

func ListItemDisabled(disabled bool) ListItemOption {
	return widget.ListItemDisabled(disabled)
}

func ListItemOnClick(fn func(ctx *Context)) ListItemOption {
	return widget.ListItemOnClick(fn)
}

func ListItemMinHeight(height float32) ListItemOption {
	return widget.ListItemMinHeight(height)
}

func ListItemDecoration(d Decoration) ListItemOption {
	return widget.ListItemDecoration(d)
}

func IconButton(child Widget, opts ...IconButtonOption) Widget {
	return widget.IconButton(child, opts...)
}

func FilledIconButton(child Widget, opts ...IconButtonOption) Widget {
	return widget.FilledIconButton(child, opts...)
}

func FilledTonalIconButton(child Widget, opts ...IconButtonOption) Widget {
	return widget.FilledTonalIconButton(child, opts...)
}

func OutlinedIconButton(child Widget, opts ...IconButtonOption) Widget {
	return widget.OutlinedIconButton(child, opts...)
}

func IconButtonOnClick(fn func(ctx *Context)) IconButtonOption {
	return widget.IconButtonOnClick(fn)
}

func IconButtonDisabled(disabled bool) IconButtonOption {
	return widget.IconButtonDisabled(disabled)
}

func IconButtonSelected(selected bool) IconButtonOption {
	return widget.IconButtonSelected(selected)
}

func IconButtonSize(size float32) IconButtonOption {
	return widget.IconButtonSize(size)
}

func IconButtonBackground(col color.NRGBA) IconButtonOption {
	return widget.IconButtonBackground(col)
}

func IconButtonForeground(col color.NRGBA) IconButtonOption {
	return widget.IconButtonForeground(col)
}

func IconButtonDecoration(d Decoration) IconButtonOption {
	return widget.IconButtonDecoration(d)
}

func FloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return widget.FloatingActionButton(icon, opts...)
}

func SmallFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return widget.SmallFloatingActionButton(icon, opts...)
}

func LargeFloatingActionButton(icon Widget, opts ...FloatingActionButtonOption) Widget {
	return widget.LargeFloatingActionButton(icon, opts...)
}

func ExtendedFloatingActionButton(icon, label Widget, opts ...FloatingActionButtonOption) Widget {
	return widget.ExtendedFloatingActionButton(icon, label, opts...)
}

func FloatingActionButtonOnClick(fn func(ctx *Context)) FloatingActionButtonOption {
	return widget.FloatingActionButtonOnClick(fn)
}

func FloatingActionButtonDisabled(disabled bool) FloatingActionButtonOption {
	return widget.FloatingActionButtonDisabled(disabled)
}

func FloatingActionButtonBackground(col color.NRGBA) FloatingActionButtonOption {
	return widget.FloatingActionButtonBackground(col)
}

func FloatingActionButtonForeground(col color.NRGBA) FloatingActionButtonOption {
	return widget.FloatingActionButtonForeground(col)
}

func FloatingActionButtonDecoration(d Decoration) FloatingActionButtonOption {
	return widget.FloatingActionButtonDecoration(d)
}

func ProgressBar(value float32, opts ...ProgressOption) Widget {
	return widget.ProgressBar(value, opts...)
}

func LinearProgressIndicator(value float32, opts ...ProgressOption) Widget {
	return widget.LinearProgressIndicator(value, opts...)
}

func CircularProgress(value float32, opts ...ProgressOption) Widget {
	return widget.CircularProgress(value, opts...)
}

func CircularProgressIndicator(value float32, opts ...ProgressOption) Widget {
	return widget.CircularProgressIndicator(value, opts...)
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

func ProgressLabelVisible(visible bool) ProgressOption {
	return widget.ProgressLabelVisible(visible)
}

func ProgressDecoration(d Decoration) ProgressOption {
	return widget.ProgressDecoration(d)
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

func TabsDecoration(d Decoration) TabsOption {
	return widget.TabsDecoration(d)
}

func TabsTabDecoration(d Decoration) TabsOption {
	return widget.TabsTabDecoration(d)
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

func DialogDecoration(d Decoration) DialogOption {
	return widget.DialogDecoration(d)
}

func DialogMaskColor(col color.NRGBA) DialogOption {
	return widget.DialogMaskColor(col)
}

func DialogMaskAlpha(alpha uint8) DialogOption {
	return widget.DialogMaskAlpha(alpha)
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

func NewPopupRef() *PopupRef {
	return widget.NewPopupRef()
}

func PopupAttachRef(ref *PopupRef) PopupOption {
	return widget.PopupAttachRef(ref)
}

func PopupDecoration(d Decoration) PopupOption {
	return widget.PopupDecoration(d)
}

func PopupMaskColor(col color.NRGBA) PopupOption {
	return widget.PopupMaskColor(col)
}

func PopupMaskAlpha(alpha uint8) PopupOption {
	return widget.PopupMaskAlpha(alpha)
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

func ToastDecoration(d Decoration) ToastOption {
	return widget.ToastDecoration(d)
}

func ToastTextColor(col color.NRGBA) ToastOption {
	return widget.ToastTextColor(col)
}

func ToastAction(label string, fn func(ctx *Context)) ToastOption {
	return widget.ToastAction(label, fn)
}

func Snackbar(message string, opts ...SnackbarOption) Widget {
	return widget.Snackbar(message, opts...)
}

func SnackbarAction(label string, fn func(ctx *Context)) SnackbarOption {
	return widget.SnackbarAction(label, fn)
}

func Tooltip(label string, child Widget, opts ...TooltipOption) Widget {
	return widget.Tooltip(label, child, opts...)
}

func TooltipDisabled(disabled bool) TooltipOption {
	return widget.TooltipDisabled(disabled)
}

func TooltipOffset(offset float32) TooltipOption {
	return widget.TooltipOffset(offset)
}

func TooltipDecoration(d Decoration) TooltipOption {
	return widget.TooltipDecoration(d)
}

func TooltipTextColor(col color.NRGBA) TooltipOption {
	return widget.TooltipTextColor(col)
}

func Badge(child Widget, label string, opts ...BadgeOption) Widget {
	return widget.Badge(child, label, opts...)
}

func BadgeVisible(visible bool) BadgeOption {
	return widget.BadgeVisible(visible)
}

func BadgeBackground(col color.NRGBA) BadgeOption {
	return widget.BadgeBackground(col)
}

func BadgeForeground(col color.NRGBA) BadgeOption {
	return widget.BadgeForeground(col)
}

func BadgeDecoration(d Decoration) BadgeOption {
	return widget.BadgeDecoration(d)
}

func BadgeOffset(x, y int) BadgeOption {
	return widget.BadgeOffset(x, y)
}

func AssistChip(label string, opts ...ChipOption) Widget {
	return widget.AssistChip(label, opts...)
}

func FilterChip(label string, opts ...ChipOption) Widget {
	return widget.FilterChip(label, opts...)
}

func InputChip(label string, opts ...ChipOption) Widget {
	return widget.InputChip(label, opts...)
}

func SuggestionChip(label string, opts ...ChipOption) Widget {
	return widget.SuggestionChip(label, opts...)
}

func ChipWithSlots(label Widget, opts ...ChipOption) Widget {
	return widget.ChipWithSlots(label, opts...)
}

func ChipSelected(selected bool) ChipOption {
	return widget.ChipSelected(selected)
}

func ChipDisabled(disabled bool) ChipOption {
	return widget.ChipDisabled(disabled)
}

func ChipOnClick(fn func(ctx *Context)) ChipOption {
	return widget.ChipOnClick(fn)
}

func ChipLeading(leading Widget) ChipOption {
	return widget.ChipLeading(leading)
}

func ChipTrailing(trailing Widget) ChipOption {
	return widget.ChipTrailing(trailing)
}

func ChipBackground(col color.NRGBA) ChipOption {
	return widget.ChipBackground(col)
}

func ChipForeground(col color.NRGBA) ChipOption {
	return widget.ChipForeground(col)
}

func ChipDecoration(d Decoration) ChipOption {
	return widget.ChipDecoration(d)
}

func SearchBar(value string, opts ...SearchBarOption) Widget {
	return widget.SearchBar(value, opts...)
}

func SearchBarPlaceholder(text string) SearchBarOption {
	return widget.SearchBarPlaceholder(text)
}

func SearchBarDisabled(disabled bool) SearchBarOption {
	return widget.SearchBarDisabled(disabled)
}

func SearchBarOnChange(fn func(ctx *Context, value string)) SearchBarOption {
	return widget.SearchBarOnChange(fn)
}

func SearchBarLeading(leading Widget) SearchBarOption {
	return widget.SearchBarLeading(leading)
}

func SearchBarTrailing(trailing Widget) SearchBarOption {
	return widget.SearchBarTrailing(trailing)
}

func SearchBarDecoration(d Decoration) SearchBarOption {
	return widget.SearchBarDecoration(d)
}

func SearchBarInputOptions(opts ...InputOption) SearchBarOption {
	return widget.SearchBarInputOptions(opts...)
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
	if itemBuilder == nil {
		return widget.ListView(count, nil, opts...)
	}
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

func ListDecoration(d Decoration) ListOption {
	return widget.ListDecoration(d)
}

func ListOnReachEnd(fn func(ctx *Context)) ListOption {
	return widget.ListOnReachEnd(fn)
}

func Grid(columns int, children ...Widget) Widget {
	return widget.Grid(columns, children...)
}

func GridView(count int, columns int, itemBuilder func(ctx *Context, index int) Widget, opts ...GridOption) Widget {
	if itemBuilder == nil {
		return widget.GridView(count, columns, nil, opts...)
	}
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

func GridDecoration(d Decoration) GridOption {
	return widget.GridDecoration(d)
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

func AppBarDecoration(d Decoration) AppBarOption {
	return widget.AppBarDecoration(d)
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

func BottomNavDecoration(d Decoration) BottomNavOption {
	return widget.BottomNavDecoration(d)
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

func NavigationRail(active string, items []NavItem, opts ...NavigationRailOption) Widget {
	return widget.NavigationRail(active, items, opts...)
}

func NavigationRailOnChange(fn func(ctx *Context, key string)) NavigationRailOption {
	return widget.NavigationRailOnChange(fn)
}

func NavigationRailWidth(width float32) NavigationRailOption {
	return widget.NavigationRailWidth(width)
}

func NavigationRailHeader(header Widget) NavigationRailOption {
	return widget.NavigationRailHeader(header)
}

func NavigationRailFooter(footer Widget) NavigationRailOption {
	return widget.NavigationRailFooter(footer)
}

func NavigationRailActiveColor(col color.NRGBA) NavigationRailOption {
	return widget.NavigationRailActiveColor(col)
}

func NavigationRailInactiveColor(col color.NRGBA) NavigationRailOption {
	return widget.NavigationRailInactiveColor(col)
}

func NavigationRailDecoration(d Decoration) NavigationRailOption {
	return widget.NavigationRailDecoration(d)
}

func NavigationDrawer(active string, items []NavItem, opts ...NavigationDrawerOption) Widget {
	return widget.NavigationDrawer(active, items, opts...)
}

func NavigationDrawerOnChange(fn func(ctx *Context, key string)) NavigationDrawerOption {
	return widget.NavigationDrawerOnChange(fn)
}

func NavigationDrawerWidth(width float32) NavigationDrawerOption {
	return widget.NavigationDrawerWidth(width)
}

func NavigationDrawerHeader(header Widget) NavigationDrawerOption {
	return widget.NavigationDrawerHeader(header)
}

func NavigationDrawerFooter(footer Widget) NavigationDrawerOption {
	return widget.NavigationDrawerFooter(footer)
}

func NavigationDrawerActiveColor(col color.NRGBA) NavigationDrawerOption {
	return widget.NavigationDrawerActiveColor(col)
}

func NavigationDrawerInactiveColor(col color.NRGBA) NavigationDrawerOption {
	return widget.NavigationDrawerInactiveColor(col)
}

func NavigationDrawerDecoration(d Decoration) NavigationDrawerOption {
	return widget.NavigationDrawerDecoration(d)
}
