<!-- fluxui-doc-meta
{
  "id": "image",
  "title": "Image 图片",
  "category": "基础显示",
  "order": 120,
  "summary": "Image 支持本地图片显示与缩放控制。",
  "example": { "id": "image_basic" },
  "apis": [
    "Image(src ImageSource, opts ...ImageOption) Widget",
    "ImageElement(src ImageSource, opts ...ImageOption) Element",
    "ImageWidth(width float32) ImageOption",
    "ImageHeight(height float32) ImageOption",
    "ImageFitMode(fit ImageFit) ImageOption",
    "ImageRadius(radius float32) ImageOption",
    "ImageBackground(col color.NRGBA) ImageOption",
    "ImageDecoration(d Decoration) ImageOption",
    "ImageOnClick(fn func(ctx *Context)) ImageOption",
    "ImageAttachRef(ref *ButtonRef) ImageOption"
  ]
}
-->

# Image 图片

## 组件说明
Image 用于显示本地图片资源，支持 `Contain/Cover/Fill` 等适配策略。

## 使用方法
- 建议设置明确宽高，避免布局抖动。
- 常见模式：封面图用 `Cover`，素材图用 `Contain`。
- 图片作为可点击区域时，可用 `ImageAttachRef` 进行外部触发。

## 使用示例

### React-style Element

`ImageElement` 已可在 `RunElement` root 下直接使用。图片加载缓存、尺寸约束、点击事件和 `ImageAttachRef` 仍由底层 image widget host 管理。

```go
func Cover(ctx *ui.Context) ui.Element {
    return ui.ImageElement(
        ui.ImageSource{Path: "examples/assets/sample.png", Label: "示例图"},
        ui.ImageWidth(160),
        ui.ImageHeight(96),
        ui.ImageFitMode(ui.ImageFitCover),
    )
}
```

### Legacy Widget
旧 `ui.Image` / `Widget` 写法继续可用：

```go
ui.Image(
    ui.ImageSource{Path: "examples/assets/sample.png", Label: "示例图"},
    ui.ImageWidth(160),
    ui.ImageHeight(96),
    ui.ImageFitMode(ui.ImageFitContain),
    ui.ImageRadius(8),
)
```
