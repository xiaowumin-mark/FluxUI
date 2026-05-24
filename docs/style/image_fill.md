<!-- fluxui-doc-meta
{
  "id": "image_fill",
  "title": "ImageFill 背景图",
  "category": "样式系统",
  "order": 560,
  "summary": "ImageFill 让 Decoration 可以为容器绘制背景图片，并支持多种填充模式。",
  "example": { "id": "image_fill_basic" },
  "apis": [
    "type ImageFill = style.ImageFill",
    "type ImageFillFit = style.ImageFillFit",
    "const ImageFillContain, ImageFillCover, ImageFillFill, ImageFillNone",
    "ImageBg(src image.Image, fit ImageFillFit) Decoration",
    "LoadImage(src string) (image.Image, error)",
    "DecodeImageURL(url string) (image.Image, error)",
    "DecodeImageFile(path string) (image.Image, error)",
    "(Decoration).WithImage(f ImageFill) Decoration",
    "(Decoration).ResolveImage() *ImageFill"
  ]
}
-->

# ImageFill 背景图

## API 说明
`ImageFill` 用于把 `image.Image` 作为 Decoration 的背景图片。

```go
type ImageFill struct {
    Src image.Image
    Fit ImageFillFit
}
```

填充模式：

- `ImageFillContain`：完整显示图片，可能留空。
- `ImageFillCover`：覆盖整个区域，可能裁切。
- `ImageFillFill`：拉伸填满。
- `ImageFillNone`：按原始大小绘制。

## 使用示例
```go
img, err := ui.LoadImage("assets/banner.png")
if err != nil {
    return ui.Text("图片加载失败")
}

deco := ui.ImageBg(img, ui.ImageFillCover).
    WithPad(ui.All(20)).
    WithRad(16)

return ui.ContainerDecorationElement(
    deco,
    ui.TextElement("背景图片容器", ui.TextColor(ui.NRGBA(255, 255, 255, 255))),
)
```

## 使用建议
- 背景图叠文字时，建议同时添加半透明背景层或使用高对比文字。
- 网络图片加载应放在 effect / async 流程中，不要在每帧布局中重复下载。
- 小图标适合 `ImageFillNone`，横幅适合 `Cover`。
