<!-- fluxui-doc-meta
{
  "id": "getting_started",
  "title": "快速开始",
  "category": "使用指南",
  "order": 700,
  "summary": "从最小 RunElement 应用开始，理解 Element、Widget、状态和事件的基本用法。",
  "example": { "id": "getting_started_basic" },
  "apis": [
    "RunElement(root Component, opts ...AppOption) error",
    "type Component func(ctx *Context) Element",
    "Title(title string) AppOption",
    "Size(width, height int) AppOption",
    "UseState[T](ctx *Context, initial T) *State[T]",
    "TextElement(content string, opts ...TextOption) Element",
    "ButtonElement(child Element, opts ...ButtonOption) Element",
    "ColumnElement(children ...Element) Element",
    "FromWidget(w Widget) Element"
  ]
}
-->

# 快速开始

## 最小应用
FluxUI 推荐使用 React-style Element API。应用入口是 `RunElement`，根组件是 `func(ctx *ui.Context) ui.Element`。

```go
package main

import (
    "fmt"

    ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func App(ctx *ui.Context) ui.Element {
    count := ui.UseState(ctx, 0)

    return ui.ColumnElement(
        ui.TextElement("Hello FluxUI"),
        ui.ButtonElement(
            ui.TextElement("count +1"),
            ui.OnClick(func(ctx *ui.Context) {
                count.Set(count.Value() + 1)
            }),
        ),
        ui.TextElement(fmt.Sprintf("count = %d", count.Value())),
    )
}

func main() {
    ui.RunElement(App, ui.Title("FluxUI App"), ui.Size(900, 640))
}
```

## Element 与 Widget
- `Element` 是新 API，支持组件身份、HookSlot、生命周期和可组合渲染。
- `Widget` 是兼容 API，仍可直接使用。
- 新代码优先使用 `TextElement`、`ButtonElement`、`ColumnElement` 等 Element 构造器。
- 需要接入旧 Widget 时，使用 `FromWidget(w)` 作为逃生口。

## 常见写法
```go
func Panel(ctx *ui.Context) ui.Element {
    th := ui.UseTheme(ctx)

    return ui.ContainerDecorationElement(
        ui.Bg(th.Surface).WithPad(ui.All(16)).WithRad(12),
        ui.ColumnElement(
            ui.TextElement("标题", ui.TextSize(20)),
            ui.TextElement("内容", ui.TextColor(th.TextColor)),
        ),
    )
}
```

## 下一步
- 组件文档：`docs/widgets`。
- 样式系统：`docs/style/decoration.md`。
- 主题系统：`docs/theme/theme.md`。
- Hooks 与生命周期：`docs/widgets/hooks_lifecycle.md`。
