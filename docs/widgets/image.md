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
    "ImageWidth(width float32) ImageOption",
    "ImageHeight(height float32) ImageOption",
    "ImageFitMode(fit ImageFit) ImageOption",
    "ImageRadius(radius float32) ImageOption",
    "ImageBackground(col color.NRGBA) ImageOption",
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
```go
ui.Image(
    ui.ImageSource{Path: "examples/assets/sample.png", Label: "示例图"},
    ui.ImageWidth(160),
    ui.ImageHeight(96),
    ui.ImageFitMode(ui.ImageFitContain),
    ui.ImageRadius(8),
)
```
