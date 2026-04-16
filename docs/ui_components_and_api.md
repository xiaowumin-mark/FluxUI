# FluxUI 组件与 API 说明文档（第一步：文档基线）

> 本文档目标：在项目尚未完工阶段，先建立一份可落地、可扩展、可对齐实现的组件与接口说明。  
> 约定：本文档分为“已实现（可用）”与“规划中（建议补全）”两部分，避免把未实现能力误导为已完成。

---

## 1. 框架定位与设计原则

FluxUI 是基于 Gio 的声明式 UI 框架，设计原则如下：

1. UI 由函数树声明（类似 `build`/`render`）。
2. 每一帧都会重建 UI 树（immediate mode 思路）。
3. 状态由 `state` 统一托管，Widget 只消费状态与事件。
4. 动画由 frame tick 驱动，不依赖 goroutine 轮询。
5. 对外只暴露 `ui` 包，`internal`/底层 Gio 细节不直接暴露给业务层。

---

## 2. 对外入口（ui 包）

### 2.1 应用生命周期

```go
func App(root func(ctx *Context) Widget, opts ...AppOption) *app.Application
func Run(root func(ctx *Context) Widget, opts ...AppOption) error
```

- `root`：每帧调用的根组件构建函数。
- 返回值：
  - `App`：返回应用实例，可自行控制启动时机。
  - `Run`：直接运行窗口循环。

### 2.2 应用配置

```go
func Title(value string) AppOption
func Size(width, height int) AppOption
func WithTheme(th *Theme) AppOption
func UseTheme(ctx *Context) *Theme
```

- `Title`：窗口标题。
- `Size`：窗口初始尺寸（逻辑像素）。
- `WithTheme`：设置主题。
- `UseTheme`：在当前 frame 读取主题。

---

## 3. 已实现组件（可用）

### 3.1 布局组件

#### Column / Row

```go
func Column(children ...Widget) Widget
func Row(children ...Widget) Widget
```

- 用途：
  - `Column`：垂直排列子组件。
  - `Row`：水平排列子组件。
- 特性：按声明顺序布局；可嵌套。

#### Container

```go
func Container(st Style, child Widget) Widget
```

- 用途：背景、内边距、圆角、外边距容器。
- 关键参数（`Style`）：
  - `Background color.NRGBA`
  - `Padding Insets`
  - `Margin Insets`
  - `Radius float32`

#### Padding

```go
func Padding(insets Insets, child Widget) Widget
```

- 用途：纯内边距包装。
- 常用辅助：
  - `All(value float32) Insets`
  - `Symmetric(vertical, horizontal float32) Insets`

---

### 3.2 文本组件

#### Text

```go
func Text(content string, opts ...TextOption) Widget
```

##### TextOption

```go
func TextSize(size float32) TextOption
func TextColor(value color.NRGBA) TextOption
func TextAlign(alignment TextAlignment) TextOption
```

- `TextSize`：字号。
- `TextColor`：文本颜色。
- `TextAlign`：对齐方式。

##### TextAlignment 枚举

```go
const (
    AlignStart
    AlignCenter
    AlignEnd
)
```

---

### 3.3 按钮组件

#### Button

```go
func Button(child Widget, opts ...ButtonOption) Widget
```

##### ButtonOption

```go
func OnClick(fn func(ctx *Context)) ButtonOption
func OnHover(fn func(ctx *Context, hovering bool)) ButtonOption
func Disabled(disabled bool) ButtonOption
func ButtonPadding(insets Insets) ButtonOption
func ButtonRadius(radius float32) ButtonOption
func ButtonBackground(value color.NRGBA) ButtonOption
func ButtonForeground(value color.NRGBA) ButtonOption
```

- `OnClick`：点击事件。
- `OnHover`：悬浮事件（进入/离开）。
- `Disabled`：禁用态。
- 其余用于控制外观。

---

### 3.4 输入组件

#### TextField

```go
func TextField(value string, opts ...InputOption) Widget
```

- 受控输入：`value` 由外部状态提供，变化通过 `InputOnChange` 回传。

##### InputOption

```go
func InputValue(value string) InputOption
func InputPlaceholder(text string) InputOption
func InputPadding(insets Insets) InputOption
func InputRadius(radius float32) InputOption
func InputBorder(color color.NRGBA) InputOption
func InputBorderFocus(color color.NRGBA) InputOption
func InputBackground(color color.NRGBA) InputOption
func InputForeground(color color.NRGBA) InputOption
func InputTextSize(size float32) InputOption
func InputMaxLen(maxLen int) InputOption
func InputPassword(password bool) InputOption
func InputSingleLine(singleLine bool) InputOption
func InputDisabled(disabled bool) InputOption
func InputOnChange(fn func(ctx *Context, value string)) InputOption
func InputOnFocus(fn func(ctx *Context, focused bool)) InputOption
```

---

### 3.5 选择组件

#### Checkbox

```go
func Checkbox(label string, checked bool, opts ...CheckboxOption) Widget
```

##### CheckboxOption

```go
func CheckboxOnChange(fn func(ctx *Context, checked bool)) CheckboxOption
func CheckboxDisabled(disabled bool) CheckboxOption
func CheckboxSize(size float32) CheckboxOption
func CheckboxColor(color color.NRGBA) CheckboxOption
```

#### Switch

```go
func Switch(checked bool, opts ...SwitchOption) Widget
```

##### SwitchOption

```go
func SwitchOnChange(fn func(ctx *Context, checked bool)) SwitchOption
func SwitchDisabled(disabled bool) SwitchOption
func SwitchWidth(width float32) SwitchOption
func SwitchHeight(height float32) SwitchOption
func SwitchColor(color color.NRGBA) SwitchOption
func SwitchTrackColor(color color.NRGBA) SwitchOption
func SwitchThumbColor(color color.NRGBA) SwitchOption
```

#### Slider

```go
func Slider(value float32, opts ...SliderOption) Widget
```

##### SliderOption

```go
func SliderOnChange(fn func(ctx *Context, value float32)) SliderOption
func SliderDisabled(disabled bool) SliderOption
func SliderMin(min float32) SliderOption
func SliderMax(max float32) SliderOption
func SliderStep(step float32) SliderOption
func SliderWidth(width float32) SliderOption
func SliderTrackColor(color color.NRGBA) SliderOption
func SliderThumbColor(color color.NRGBA) SliderOption
func SliderProgressColor(color color.NRGBA) SliderOption
```

---

## 4. 状态与动画 API（已实现）

### 4.1 状态

```go
func State[T any](ctx *Context) *state.State[T]
```

`state.State[T]` 方法：

```go
func (s *State[T]) Key() string
func (s *State[T]) Value() T
func (s *State[T]) Set(v T)
```

- `Value`：读取当前值。
- `Set`：更新值并触发重绘。

### 4.2 动画

```go
func Animate(opts ...anim.Option) *anim.Animation
func Duration(duration time.Duration) anim.Option
func From(value float32) anim.Option
func To(value float32) anim.Option
func Ease(easing anim.Easing) anim.Option
```

缓动函数常量：

```go
var (
    Linear    anim.Easing
    EaseOut   anim.Easing
    EaseInOut anim.Easing
)
```

动画值读取：

```go
func (a *Animation) Value(ctx *internal.Context) float32
```

---

## 5. 常见用法示例（基于当前 API）

### 5.1 计数按钮

```go
count := ui.State[int](ctx)

ui.Button(
    ui.Text("+1"),
    ui.OnClick(func(ctx *ui.Context) {
        count.Set(count.Value() + 1)
    }),
)
```

### 5.2 受控输入框

```go
name := ui.State[string](ctx)

ui.TextField(
    name.Value(),
    ui.InputPlaceholder("请输入名称"),
    ui.InputOnChange(func(ctx *ui.Context, value string) {
        name.Set(value)
    }),
)
```

### 5.3 动画插值

```go
progress := ui.Animate(
    ui.From(0),
    ui.To(1),
    ui.Duration(500*time.Millisecond),
    ui.Ease(ui.EaseInOut),
).Value(ctx)
```

---

## 6. 规划中的常用控件清单（建议补全）

> 下列控件是一个完整 UI 框架常见能力。当前版本文档先定义“目标组件集合”，后续分步落地实现。

### 6.1 基础显示类

1. `Image`：本地/内存图像渲染、缩放模式、圆角裁剪。
2. `Icon`：矢量或字体图标渲染。
3. `Divider`：水平/垂直分割线。
4. `Spacer`：显式空白占位。

### 6.2 输入交互类

1. `Radio` / `RadioGroup`
2. `Select` / `Dropdown`
3. `DatePicker` / `TimePicker`
4. `Stepper`（数字步进器）
5. `SearchField`
6. `Textarea`（多行输入）

### 6.3 反馈与提示类

1. `ProgressBar`（线性进度）
2. `CircularProgress`（环形进度）
3. `Toast`
4. `Snackbar`
5. `Tooltip`
6. `Badge`

### 6.4 容器与结构类

1. `Card`
2. `Panel`
3. `List`（虚拟化可选）
4. `ScrollView`
5. `Grid`
6. `Tabs`
7. `Accordion`
8. `Modal/Dialog`
9. `Drawer`

### 6.5 导航类

1. `AppBar`
2. `BottomNavigation`
3. `Breadcrumb`
4. `Pagination`

### 6.6 规划控件接口草案（详细）

> 说明：以下接口均为**规划草案**，用于后续实现对齐，不代表当前可直接调用。

#### A. Spacer / Divider

```go
func Spacer(width, height float32) Widget
func HSpacer(width float32) Widget
func VSpacer(height float32) Widget

func Divider(opts ...DividerOption) Widget
type DividerOption func(*dividerConfig)

func DividerVertical(vertical bool) DividerOption
func DividerThickness(thickness float32) DividerOption
func DividerColor(col color.NRGBA) DividerOption
func DividerLength(length float32) DividerOption
func DividerMargin(insets Insets) DividerOption
```

#### B. Image / Icon

```go
func Image(src ImageSource, opts ...ImageOption) Widget
type ImageOption func(*imageConfig)

func ImageWidth(width float32) ImageOption
func ImageHeight(height float32) ImageOption
func ImageFit(fit ImageFit) ImageOption
func ImageRadius(radius float32) ImageOption
func ImageBackground(col color.NRGBA) ImageOption
func ImageOnClick(fn func(ctx *Context)) ImageOption

type ImageFit int
const (
    ImageFitContain ImageFit = iota
    ImageFitCover
    ImageFitFill
    ImageFitNone
)
```

```go
func Icon(name string, opts ...IconOption) Widget
type IconOption func(*iconConfig)

func IconSize(size float32) IconOption
func IconColor(col color.NRGBA) IconOption
func IconOnClick(fn func(ctx *Context)) IconOption
```

#### C. Card

```go
func Card(child Widget, opts ...CardOption) Widget
type CardOption func(*cardConfig)

func CardPadding(insets Insets) CardOption
func CardRadius(radius float32) CardOption
func CardBackground(col color.NRGBA) CardOption
func CardBorder(col color.NRGBA, width float32) CardOption
func CardShadow(level int) CardOption
func CardOnClick(fn func(ctx *Context)) CardOption
```

#### D. Radio / RadioGroup

```go
func RadioGroup(value string, items []RadioItem, opts ...RadioGroupOption) Widget
type RadioItem struct {
    Label string
    Value string
}
type RadioGroupOption func(*radioGroupConfig)

func RadioGroupDirection(axis Axis) RadioGroupOption
func RadioGroupDisabled(disabled bool) RadioGroupOption
func RadioGroupOnChange(fn func(ctx *Context, value string)) RadioGroupOption
func RadioGroupSize(size float32) RadioGroupOption
func RadioGroupColor(col color.NRGBA) RadioGroupOption
```

#### E. Select / Dropdown

```go
func Select[T comparable](value T, options []SelectOptionItem[T], opts ...SelectOption[T]) Widget
type SelectOptionItem[T comparable] struct {
    Label string
    Value T
}
type SelectOption[T comparable] func(*selectConfig[T])

func SelectPlaceholder(text string) SelectOption[T]
func SelectDisabled(disabled bool) SelectOption[T]
func SelectSearchable(searchable bool) SelectOption[T]
func SelectMaxHeight(height float32) SelectOption[T]
func SelectOnChange(fn func(ctx *Context, value T)) SelectOption[T]
func SelectOnOpenChange(fn func(ctx *Context, opened bool)) SelectOption[T]
```

#### F. ProgressBar / CircularProgress

```go
func ProgressBar(value float32, opts ...ProgressOption) Widget
func CircularProgress(value float32, opts ...ProgressOption) Widget
type ProgressOption func(*progressConfig)

func ProgressMin(min float32) ProgressOption
func ProgressMax(max float32) ProgressOption
func ProgressIndeterminate(indeterminate bool) ProgressOption
func ProgressThickness(thickness float32) ProgressOption
func ProgressTrackColor(col color.NRGBA) ProgressOption
func ProgressFillColor(col color.NRGBA) ProgressOption
func ProgressSize(size float32) ProgressOption // 对环形进度有效
```

#### G. Tabs

```go
func Tabs(active string, items []TabItem, opts ...TabsOption) Widget
type TabItem struct {
    Key   string
    Label string
}
type TabsOption func(*tabsConfig)

func TabsOnChange(fn func(ctx *Context, key string)) TabsOption
func TabsScrollable(scrollable bool) TabsOption
func TabsIndicatorColor(col color.NRGBA) TabsOption
func TabsTextColor(col color.NRGBA) TabsOption
func TabsActiveTextColor(col color.NRGBA) TabsOption
```

#### H. Dialog / Modal

```go
func Dialog(open bool, child Widget, opts ...DialogOption) Widget
type DialogOption func(*dialogConfig)

func DialogTitle(title string) DialogOption
func DialogWidth(width float32) DialogOption
func DialogRadius(radius float32) DialogOption
func DialogMaskClosable(maskClosable bool) DialogOption
func DialogOnOpenChange(fn func(ctx *Context, open bool)) DialogOption
func DialogOnConfirm(fn func(ctx *Context)) DialogOption
func DialogOnCancel(fn func(ctx *Context)) DialogOption
```

#### I. Toast / Snackbar

```go
func Toast(message string, opts ...ToastOption) Widget
type ToastOption func(*toastConfig)

func ToastType(t ToastType) ToastOption
func ToastDuration(duration time.Duration) ToastOption
func ToastPosition(p ToastPosition) ToastOption
func ToastOnClose(fn func(ctx *Context)) ToastOption

type ToastType int
const (
    ToastInfo ToastType = iota
    ToastSuccess
    ToastWarning
    ToastError
)
```

#### J. ScrollView

```go
func ScrollView(child Widget, opts ...ScrollOption) Widget
type ScrollOption func(*scrollConfig)

func ScrollVertical(vertical bool) ScrollOption
func ScrollHorizontal(horizontal bool) ScrollOption
func ScrollBarVisible(visible bool) ScrollOption
func ScrollOnChange(fn func(ctx *Context, x, y float32)) ScrollOption
```

#### K. ListView（大数据列表）

```go
func ListView(count int, itemBuilder func(ctx *Context, index int) Widget, opts ...ListOption) Widget
type ListOption func(*listConfig)

func ListAxis(axis Axis) ListOption
func ListVirtualized(virtualized bool) ListOption
func ListItemSpacing(spacing float32) ListOption
func ListPadding(insets Insets) ListOption
func ListOnReachEnd(fn func(ctx *Context)) ListOption
```

#### L. Grid

```go
func Grid(columns int, children ...Widget) Widget
func GridView(count int, columns int, itemBuilder func(ctx *Context, index int) Widget, opts ...GridOption) Widget
type GridOption func(*gridConfig)

func GridGap(rowGap, colGap float32) GridOption
func GridPadding(insets Insets) GridOption
func GridMinItemWidth(width float32) GridOption
```

#### M. AppBar / BottomNavigation

```go
func AppBar(title Widget, opts ...AppBarOption) Widget
type AppBarOption func(*appBarConfig)

func AppBarLeading(leading Widget) AppBarOption
func AppBarActions(actions ...Widget) AppBarOption
func AppBarHeight(height float32) AppBarOption
func AppBarBackground(col color.NRGBA) AppBarOption
```

```go
func BottomNavigation(active string, items []NavItem, opts ...BottomNavOption) Widget
type NavItem struct {
    Key   string
    Label string
    Icon  Widget
}
type BottomNavOption func(*bottomNavConfig)

func BottomNavOnChange(fn func(ctx *Context, key string)) BottomNavOption
func BottomNavBackground(col color.NRGBA) BottomNavOption
func BottomNavActiveColor(col color.NRGBA) BottomNavOption
func BottomNavInactiveColor(col color.NRGBA) BottomNavOption
```

---

## 7. 规划组件的接口规范模板（建议）

为保证后续扩展一致性，建议所有新控件遵循下面模式：

### 7.1 构造函数 + Option

```go
func Component(requiredArg ..., opts ...ComponentOption) ui.Widget
type ComponentOption func(*componentConfig)
```

### 7.2 事件回调命名

1. 值变化：`ComponentOnChange`
2. 点击：`ComponentOnClick`
3. 聚焦：`ComponentOnFocus`
4. 展开收起：`ComponentOnToggle`
5. 提交：`ComponentOnSubmit`

### 7.3 常用样式项命名

1. `ComponentDisabled`
2. `ComponentPadding`
3. `ComponentRadius`
4. `ComponentBackground`
5. `ComponentForeground`
6. `ComponentBorder`

---

## 8. 建议优先级（第二步可执行顺序）

1. `ScrollView`、`List`、`Image`、`Icon`
2. `RadioGroup`、`Select`、`ProgressBar`
3. `Tabs`、`Dialog`、`Toast`
4. `Grid`、`Drawer`、`Pagination`

> 原因：上述能力优先补齐后，可覆盖大多数业务界面搭建需求。

---

## 9. 文档维护规则

1. `ui` 每新增一个导出 API，必须同步更新本文档。
2. 每个组件需明确：
   - 构造函数
   - Option 列表
   - 事件回调
   - 最小示例
3. “规划中”组件实现后，需迁移到“已实现”章节并补齐示例。
