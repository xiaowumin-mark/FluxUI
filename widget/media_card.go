package widget

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaowumin-mark/FluxUI/internal"
	"github.com/xiaowumin-mark/FluxUI/layout"
	"github.com/xiaowumin-mark/FluxUI/style"

	gioLayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	gioWidget "gioui.org/widget"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// ImageSource 描述图片来源。
type ImageSource struct {
	Path  string
	Label string
}

// ImageFit 描述图片缩放模式。
type ImageFit int

const (
	ImageFitContain ImageFit = iota
	ImageFitCover
	ImageFitFill
	ImageFitNone
)

// ImageOption 定义图片配置。
type ImageOption func(*imageConfig)

type imageConfig struct {
	width         float32
	height        float32
	fit           ImageFit
	radius        float32
	background    color.NRGBA
	hasBackground bool
	onClick       func(ctx *internal.Context)
	ref           *ButtonRef
}

type imageWidget struct {
	src    ImageSource
	config imageConfig
}

type imageRenderWidget struct {
	owner *imageWidget
}

type decodedImageState struct {
	loadedPath string
	loaded     bool
	img        image.Image
	op         paint.ImageOp
	err        error
}

// Image 创建图片组件。
func Image(src ImageSource, opts ...ImageOption) Widget {
	cfg := imageConfig{
		width:  120,
		height: 80,
		fit:    ImageFitContain,
		radius: 8,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &imageWidget{
		src:    src,
		config: cfg,
	}
}

// ImageWidth 设置宽度。
func ImageWidth(width float32) ImageOption {
	return func(cfg *imageConfig) {
		cfg.width = width
	}
}

// ImageHeight 设置高度。
func ImageHeight(height float32) ImageOption {
	return func(cfg *imageConfig) {
		cfg.height = height
	}
}

// ImageFitMode 设置缩放模式。
func ImageFitMode(fit ImageFit) ImageOption {
	return func(cfg *imageConfig) {
		cfg.fit = fit
	}
}

// ImageRadius 设置圆角。
func ImageRadius(radius float32) ImageOption {
	return func(cfg *imageConfig) {
		cfg.radius = radius
	}
}

// ImageBackground 设置背景色。
func ImageBackground(col color.NRGBA) ImageOption {
	return func(cfg *imageConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

// ImageOnClick 设置点击回调。
func ImageOnClick(fn func(ctx *internal.Context)) ImageOption {
	return func(cfg *imageConfig) {
		cfg.onClick = fn
	}
}

// ImageAttachRef 绑定命令型引用，用于外部主动触发点击。
func ImageAttachRef(ref *ButtonRef) ImageOption {
	return func(cfg *imageConfig) {
		cfg.ref = ref
	}
}

func (i *imageWidget) Layout(ctx *internal.Context) layout.Dimensions {
	var root Widget = &fixedSizeWidget{
		width:  i.config.width,
		height: i.config.height,
		child:  &imageRenderWidget{owner: i},
	}
	if i.config.onClick != nil || i.config.ref != nil {
		opts := []ButtonOption{
			ButtonBackground(color.NRGBA{}),
			ButtonForeground(ctx.Theme().TextColor),
			ButtonPadding(style.All(0)),
			ButtonRadius(i.config.radius),
		}
		if i.config.onClick != nil {
			opts = append(opts, OnClick(i.config.onClick))
		}
		if i.config.ref != nil {
			opts = append(opts, ButtonAttachRef(i.config.ref))
		}
		root = Button(
			root,
			opts...,
		)
	}
	return root.Layout(ctx.Child(0))
}

func (r *imageRenderWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if r.owner == nil {
		return layout.Dimensions{}
	}

	gtx := ctx.Gtx
	size := clampPointToConstraints(gtx.Constraints.Min, gtx.Constraints.Min, gtx.Constraints.Max)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{}
	}

	bg := ctx.Theme().SurfaceMuted
	if r.owner.config.hasBackground {
		bg = r.owner.config.background
	}

	radius := gtx.Dp(safeDp(r.owner.config.radius))
	clipArea := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, bg)

	state := imageStateFor(ctx)
	r.owner.refreshImageState(state)
	if state.img == nil {
		r.owner.layoutFallback(ctx, size, state)
		clipArea.Pop()
		return layout.Dimensions{Size: size}
	}

	img := gioWidget.Image{
		Src:      state.op,
		Fit:      toGioImageFit(r.owner.config.fit),
		Position: gioLayout.Center,
	}
	if gtx.Metric.PxPerDp > 0 {
		img.Scale = 1 / gtx.Metric.PxPerDp
	}

	imgCtx := gtx
	imgCtx.Constraints = gioLayout.Exact(size)
	_ = img.Layout(imgCtx)
	clipArea.Pop()
	return layout.Dimensions{Size: size}
}

func imageStateFor(ctx *internal.Context) *decodedImageState {
	value := ctx.Memo("image", func() any {
		return &decodedImageState{}
	})
	state, ok := value.(*decodedImageState)
	if !ok {
		panic("github.com/xiaowumin-mark/FluxUIwidget: image state type mismatch")
	}
	return state
}

func (i *imageWidget) refreshImageState(state *decodedImageState) {
	if state == nil {
		return
	}

	path := strings.TrimSpace(i.src.Path)
	resolvedPath := resolveImagePath(path)
	if state.loaded && state.loadedPath == resolvedPath {
		return
	}

	state.loaded = true
	state.loadedPath = resolvedPath
	state.img = nil
	state.err = nil
	state.op = paint.ImageOp{}

	if resolvedPath == "" {
		return
	}

	file, err := os.Open(resolvedPath)
	if err != nil {
		state.err = err
		return
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		state.err = err
		return
	}

	state.img = decoded
	state.op = paint.NewImageOp(decoded)
}

func resolveImagePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	if filepath.IsAbs(path) {
		if fileExists(path) {
			return path
		}
		return filepath.Clean(path)
	}

	candidates := make([]string, 0, 10)
	candidates = append(candidates, path)

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, path))
		candidates = append(candidates, filepath.Join(cwd, "..", path))
		candidates = append(candidates, filepath.Join(cwd, "..", "..", path))
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, path),
			filepath.Join(exeDir, "..", path),
			filepath.Join(exeDir, "..", "..", path),
		)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		if fileExists(cleaned) {
			return cleaned
		}
	}
	return filepath.Clean(path)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (i *imageWidget) layoutFallback(ctx *internal.Context, size image.Point, state *decodedImageState) {
	label := i.src.Label
	if label == "" && i.src.Path != "" {
		label = filepath.Base(i.src.Path)
	}
	if state != nil && state.err != nil {
		label = "图片加载失败"
	}
	if label == "" {
		label = "Image"
	}

	fallback := Text(label, TextSize(12), TextColor(ctx.Theme().TextColor))

	gtx := ctx.Gtx
	gtx.Constraints = gioLayout.Exact(size)
	_ = gioLayout.Center.Layout(gtx, func(gtx gioLayout.Context) gioLayout.Dimensions {
		next := *ctx
		next.Gtx = gtx
		dims := fallback.Layout(&next)
		return gioLayout.Dimensions{Size: dims.Size}
	})
}

func toGioImageFit(fit ImageFit) gioWidget.Fit {
	switch fit {
	case ImageFitCover:
		return gioWidget.Cover
	case ImageFitFill:
		return gioWidget.Fill
	case ImageFitNone:
		return gioWidget.Unscaled
	default:
		return gioWidget.Contain
	}
}

// IconOption 定义图标配置。
type IconOption func(*iconConfig)

type iconConfig struct {
	size     float32
	color    color.NRGBA
	hasColor bool
	onClick  func(ctx *internal.Context)
	ref      *ButtonRef
}

type iconWidget struct {
	name   string
	config iconConfig
}

// Icon 创建图标组件（当前为文本图标占位）。
func Icon(name string, opts ...IconOption) Widget {
	cfg := iconConfig{
		size: 16,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &iconWidget{
		name:   name,
		config: cfg,
	}
}

// IconSize 设置尺寸。
func IconSize(size float32) IconOption {
	return func(cfg *iconConfig) {
		cfg.size = size
	}
}

// IconColor 设置颜色。
func IconColor(col color.NRGBA) IconOption {
	return func(cfg *iconConfig) {
		cfg.color = col
		cfg.hasColor = true
	}
}

// IconOnClick 设置点击回调。
func IconOnClick(fn func(ctx *internal.Context)) IconOption {
	return func(cfg *iconConfig) {
		cfg.onClick = fn
	}
}

// IconAttachRef 绑定命令型引用，用于外部主动触发点击。
func IconAttachRef(ref *ButtonRef) IconOption {
	return func(cfg *iconConfig) {
		cfg.ref = ref
	}
}

func (i *iconWidget) Layout(ctx *internal.Context) layout.Dimensions {
	col := ctx.Theme().TextColor
	if i.config.hasColor {
		col = i.config.color
	}

	name := i.name
	if name == "" {
		name = "icon"
	}

	label := Text(name, TextSize(i.config.size), TextColor(col))
	root := Widget(label)
	if i.config.onClick != nil || i.config.ref != nil {
		opts := []ButtonOption{
			ButtonBackground(color.NRGBA{}),
			ButtonForeground(col),
			ButtonPadding(style.All(0)),
		}
		if i.config.onClick != nil {
			opts = append(opts, OnClick(i.config.onClick))
		}
		if i.config.ref != nil {
			opts = append(opts, ButtonAttachRef(i.config.ref))
		}
		root = Button(
			label,
			opts...,
		)
	}
	return root.Layout(ctx.Child(0))
}

// CardOption 定义卡片配置。
type CardOption func(*cardConfig)

type cardConfig struct {
	padding        style.Insets
	radius         float32
	background     color.NRGBA
	hasBackground  bool
	borderColor    color.NRGBA
	hasBorderColor bool
	borderWidth    float32
	shadowLevel    int
	onClick        func(ctx *internal.Context)
	ref            *ButtonRef
}

type cardWidget struct {
	child  Widget
	config cardConfig
}

// Card 创建卡片组件。
func Card(child Widget, opts ...CardOption) Widget {
	cfg := cardConfig{
		padding: style.All(12),
		radius:  12,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &cardWidget{
		child:  child,
		config: cfg,
	}
}

// CardPadding 设置内边距。
func CardPadding(insets style.Insets) CardOption {
	return func(cfg *cardConfig) {
		cfg.padding = insets
	}
}

// CardRadius 设置圆角。
func CardRadius(radius float32) CardOption {
	return func(cfg *cardConfig) {
		cfg.radius = radius
	}
}

// CardBackground 设置背景色。
func CardBackground(col color.NRGBA) CardOption {
	return func(cfg *cardConfig) {
		cfg.background = col
		cfg.hasBackground = true
	}
}

// CardBorder 设置边框颜色和宽度。
func CardBorder(col color.NRGBA, width float32) CardOption {
	return func(cfg *cardConfig) {
		cfg.borderColor = col
		cfg.hasBorderColor = true
		cfg.borderWidth = width
	}
}

// CardShadow 设置阴影等级（当前为预留参数）。
func CardShadow(level int) CardOption {
	return func(cfg *cardConfig) {
		cfg.shadowLevel = level
	}
}

// CardOnClick 设置点击回调。
func CardOnClick(fn func(ctx *internal.Context)) CardOption {
	return func(cfg *cardConfig) {
		cfg.onClick = fn
	}
}

// CardAttachRef 绑定命令型引用，用于外部主动触发点击。
func CardAttachRef(ref *ButtonRef) CardOption {
	return func(cfg *cardConfig) {
		cfg.ref = ref
	}
}

func (c *cardWidget) Layout(ctx *internal.Context) layout.Dimensions {
	bg := ctx.Theme().Surface
	if c.config.hasBackground {
		bg = c.config.background
	}

	inner := Container(
		style.Style{
			Background: bg,
			Padding:    c.config.padding,
			Radius:     c.config.radius,
		},
		c.child,
	)

	var root Widget = inner
	if c.config.hasBorderColor && c.config.borderWidth > 0 {
		innerRadius := c.config.radius - c.config.borderWidth
		if innerRadius < 0 {
			innerRadius = 0
		}
		root = Container(
			style.Style{
				Background: c.config.borderColor,
				Padding:    style.All(c.config.borderWidth),
				Radius:     c.config.radius,
			},
			Container(
				style.Style{
					Background: bg,
					Padding:    c.config.padding,
					Radius:     innerRadius,
				},
				c.child,
			),
		)
	}

	if c.config.onClick != nil || c.config.ref != nil {
		opts := []ButtonOption{
			ButtonBackground(color.NRGBA{}),
			ButtonForeground(ctx.Theme().TextColor),
			ButtonPadding(style.All(0)),
			ButtonRadius(c.config.radius),
		}
		if c.config.onClick != nil {
			opts = append(opts, OnClick(c.config.onClick))
		}
		if c.config.ref != nil {
			opts = append(opts, ButtonAttachRef(c.config.ref))
		}
		root = Button(
			root,
			opts...,
		)
	}

	return root.Layout(ctx.Child(0))
}

type fixedSizeWidget struct {
	width  float32
	height float32
	child  Widget
}

func (f *fixedSizeWidget) Layout(ctx *internal.Context) layout.Dimensions {
	if f.child == nil {
		return layout.Dimensions{}
	}

	gtx := ctx.Gtx
	cgtx := gtx

	hasWidth := f.width > 0
	hasHeight := f.height > 0
	width := gtx.Dp(safeDp(f.width))
	height := gtx.Dp(safeDp(f.height))
	if hasWidth && width <= 0 {
		width = 1
	}
	if hasHeight && height <= 0 {
		height = 1
	}

	if hasWidth {
		if width > cgtx.Constraints.Max.X {
			width = cgtx.Constraints.Max.X
		}
		if width < cgtx.Constraints.Min.X {
			width = cgtx.Constraints.Min.X
		}
		cgtx.Constraints.Min.X = width
		cgtx.Constraints.Max.X = width
	}
	if hasHeight {
		if height > cgtx.Constraints.Max.Y {
			height = cgtx.Constraints.Max.Y
		}
		if height < cgtx.Constraints.Min.Y {
			height = cgtx.Constraints.Min.Y
		}
		cgtx.Constraints.Min.Y = height
		cgtx.Constraints.Max.Y = height
	}
	if cgtx.Constraints.Min.X > cgtx.Constraints.Max.X {
		cgtx.Constraints.Min.X = cgtx.Constraints.Max.X
	}
	if cgtx.Constraints.Min.Y > cgtx.Constraints.Max.Y {
		cgtx.Constraints.Min.Y = cgtx.Constraints.Max.Y
	}

	next := *ctx
	next.Gtx = cgtx
	dims := f.child.Layout(next.Child(0))

	size := dims.Size
	if hasWidth {
		size.X = width
	}
	if hasHeight {
		size.Y = height
	}
	dims.Size = clampPointToConstraints(size, cgtx.Constraints.Min, cgtx.Constraints.Max)
	return dims
}
