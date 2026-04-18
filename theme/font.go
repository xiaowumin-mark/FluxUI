package theme

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	giofont "gioui.org/font"
	"gioui.org/font/opentype"
)

// FontStyle 表示字体样式。
type FontStyle int

const (
	FontStyleRegular FontStyle = iota
	FontStyleItalic
)

// FontWeight 表示字体字重。
type FontWeight int

const (
	FontWeightThin       FontWeight = -300
	FontWeightExtraLight FontWeight = -200
	FontWeightLight      FontWeight = -100
	FontWeightNormal     FontWeight = 0
	FontWeightMedium     FontWeight = 100
	FontWeightSemiBold   FontWeight = 200
	FontWeightBold       FontWeight = 300
	FontWeightExtraBold  FontWeight = 400
	FontWeightBlack      FontWeight = 500
)

// FontSpec 定义一组字体选择偏好。
type FontSpec struct {
	Family string
	Style  FontStyle
	Weight FontWeight
}

// FontFace 表示一个可供 Shaper 使用的字体面。
type FontFace struct {
	Spec      FontSpec
	source    *fontSource
	faceIndex int
}

// DefaultFontSpec 返回默认字体样式。
func DefaultFontSpec() FontSpec {
	return systemDefaultFontSpec()
}

// FontFamily 设置字体族。
func FontFamily(family string) FontSpec {
	return FontSpec{Family: strings.TrimSpace(family)}
}

// FontStyleOf 设置字体样式。
func FontStyleOf(style FontStyle) FontSpec {
	return FontSpec{Style: style}
}

// FontWeightOf 设置字体字重。
func FontWeightOf(weight FontWeight) FontSpec {
	return FontSpec{Weight: weight}
}

// WithStyle 复制并更新字体样式。
func (f FontSpec) WithStyle(style FontStyle) FontSpec {
	f.Style = style
	return f
}

// WithWeight 复制并更新字体字重。
func (f FontSpec) WithWeight(weight FontWeight) FontSpec {
	f.Weight = weight
	return f
}

// WithFamily 复制并更新字体族。
func (f FontSpec) WithFamily(family string) FontSpec {
	f.Family = strings.TrimSpace(family)
	return f
}

// Normalize 规范化字体参数。
func (f FontSpec) Normalize() FontSpec {
	if strings.TrimSpace(f.Family) == "" {
		f.Family = DefaultFontSpec().Family
	}
	return f
}

// ParseFontFile 解析单个字体文件，支持 ttf/otf/ttc/otc。
func ParseFontFile(path string) ([]FontFace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rawFaces, err := opentype.ParseCollection(data)
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(data))
	copy(copied, data)
	src := &fontSource{
		data: copied,
		path: path,
	}

	faces := make([]FontFace, 0, len(rawFaces))
	for idx, raw := range rawFaces {
		faces = append(faces, FontFace{
			Spec:      fromGioFont(raw.Font),
			source:    src,
			faceIndex: idx,
		})
	}
	return faces, nil
}

// LoadFontsFromPaths 加载多个字体文件。
func LoadFontsFromPaths(paths ...string) ([]FontFace, error) {
	out := make([]FontFace, 0)
	for _, path := range paths {
		p := strings.TrimSpace(path)
		if p == "" {
			continue
		}
		faces, err := ParseFontFile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, faces...)
	}
	return out, nil
}

// LoadFontsFromDir 递归加载目录下的字体文件。
func LoadFontsFromDir(dir string) ([]FontFace, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("theme: empty font directory")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("theme: font path is not a directory")
	}

	out := make([]FontFace, 0)
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".ttf" && ext != ".otf" && ext != ".ttc" && ext != ".otc" {
			return nil
		}
		faces, err := ParseFontFile(path)
		if err != nil {
			return err
		}
		out = append(out, faces...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AddFonts 追加自定义字体集合。
func (t *Theme) AddFonts(faces ...FontFace) {
	if t == nil || len(faces) == 0 {
		return
	}
	t.Fonts = append(t.Fonts, faces...)
}

// SetFonts 覆盖字体集合。
func (t *Theme) SetFonts(faces ...FontFace) {
	if t == nil {
		return
	}
	if len(faces) == 0 {
		t.Fonts = nil
		return
	}
	t.Fonts = append([]FontFace(nil), faces...)
}

// WithFonts 基于拷贝返回带字体集合的新主题。
func (t *Theme) WithFonts(faces ...FontFace) *Theme {
	if t == nil {
		return nil
	}
	cp := *t
	cp.SetFonts(faces...)
	return &cp
}

// SetDefaultFont 设置主题默认字体。
func (t *Theme) SetDefaultFont(spec FontSpec) {
	if t == nil {
		return
	}
	t.DefaultFont = spec.Normalize()
}

// WithDefaultFont 基于拷贝返回带默认字体的新主题。
func (t *Theme) WithDefaultFont(spec FontSpec) *Theme {
	if t == nil {
		return nil
	}
	cp := *t
	cp.SetDefaultFont(spec)
	return &cp
}

// SetUseSystemFonts 设置是否启用系统字体参与回退。
func (t *Theme) SetUseSystemFonts(enabled bool) {
	if t == nil {
		return
	}
	t.UseSystemFonts = enabled
}

// WithSystemFonts 基于拷贝返回系统字体开关已更新的新主题。
func (t *Theme) WithSystemFonts(enabled bool) *Theme {
	if t == nil {
		return nil
	}
	cp := *t
	cp.SetUseSystemFonts(enabled)
	return &cp
}

// SystemFontDirs 返回当前平台常见系统字体目录（用于扫描预览，不保证全部存在）。
func SystemFontDirs() []string {
	switch runtime.GOOS {
	case "windows":
		root := os.Getenv("WINDIR")
		if strings.TrimSpace(root) == "" {
			root = `C:\Windows`
		}
		return []string{filepath.Join(root, "Fonts")}
	case "darwin":
		home, _ := os.UserHomeDir()
		return []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			filepath.Join(home, "Library", "Fonts"),
		}
	default:
		home, _ := os.UserHomeDir()
		return []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(home, ".fonts"),
			filepath.Join(home, ".local", "share", "fonts"),
		}
	}
}

// DiscoverSystemFonts 尝试读取系统字体文件并返回可用字体集合。
func DiscoverSystemFonts() ([]FontFace, error) {
	dirs := SystemFontDirs()
	out := make([]FontFace, 0)
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || d == nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext != ".ttf" && ext != ".otf" && ext != ".ttc" && ext != ".otc" {
				return nil
			}
			faces, err := ParseFontFile(path)
			if err != nil {
				// 单个字体文件损坏时跳过，保证扫描过程可继续。
				return nil
			}
			out = append(out, faces...)
			return nil
		})
	}
	if len(out) == 0 {
		return nil, errors.New("theme: no system fonts discovered")
	}
	return out, nil
}

func systemDefaultFontSpec() FontSpec {
	switch runtime.GOOS {
	case "windows":
		return FontSpec{
			Family: `"Segoe UI", "Microsoft YaHei UI", "Microsoft YaHei", sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	case "darwin":
		return FontSpec{
			Family: `"SF Pro Text", "Helvetica Neue", "PingFang SC", sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	case "linux":
		return FontSpec{
			Family: `"Noto Sans CJK SC", "Noto Sans", "Ubuntu", "Cantarell", "DejaVu Sans", sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	case "android":
		return FontSpec{
			Family: `"Roboto", "Noto Sans CJK SC", sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	case "ios":
		return FontSpec{
			Family: `"SF Pro Text", "PingFang SC", "Helvetica Neue", sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	default:
		return FontSpec{
			Family: `sans-serif, Go`,
			Style:  FontStyleRegular,
			Weight: FontWeightNormal,
		}
	}
}

// ListFontFamilies 返回去重后的字体族列表。
func ListFontFamilies(faces []FontFace) []string {
	if len(faces) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(faces))
	out := make([]string, 0, len(faces))
	for _, face := range faces {
		name := strings.TrimSpace(face.Spec.Family)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// DiscoverSystemFontFamilies 返回系统可识别字体族名称列表。
func DiscoverSystemFontFamilies() ([]string, error) {
	faces, err := DiscoverSystemFonts()
	if err != nil {
		return nil, err
	}
	return ListFontFamilies(faces), nil
}

// Data 返回字体二进制数据。
func (f FontFace) Data() []byte {
	if f.source == nil {
		return nil
	}
	return f.source.data
}

// FaceIndex 返回在字体集合文件中的索引。
func (f FontFace) FaceIndex() int {
	return f.faceIndex
}

// SourcePath 返回字体来源路径（若有）。
func (f FontFace) SourcePath() string {
	if f.source == nil {
		return ""
	}
	return f.source.path
}

// WithFamily 返回仅修改字体族的新副本。
func (f FontFace) WithFamily(family string) FontFace {
	f.Spec = f.Spec.WithFamily(family)
	return f
}

// WithStyle 返回仅修改字体样式的新副本。
func (f FontFace) WithStyle(style FontStyle) FontFace {
	f.Spec = f.Spec.WithStyle(style)
	return f
}

// WithWeight 返回仅修改字体字重的新副本。
func (f FontFace) WithWeight(weight FontWeight) FontFace {
	f.Spec = f.Spec.WithWeight(weight)
	return f
}

func toGioFont(spec FontSpec) giofont.Font {
	normalized := spec.Normalize()
	return giofont.Font{
		Typeface: giofont.Typeface(normalized.Family),
		Style:    toGioFontStyle(normalized.Style),
		Weight:   toGioFontWeight(normalized.Weight),
	}
}

func toGioFontStyle(style FontStyle) giofont.Style {
	switch style {
	case FontStyleItalic:
		return giofont.Italic
	default:
		return giofont.Regular
	}
}

func toGioFontWeight(weight FontWeight) giofont.Weight {
	return giofont.Weight(weight)
}

func fromGioFont(g giofont.Font) FontSpec {
	return FontSpec{
		Family: strings.TrimSpace(string(g.Typeface)),
		Style:  fromGioFontStyle(g.Style),
		Weight: FontWeight(g.Weight),
	}.Normalize()
}

func fromGioFontStyle(style giofont.Style) FontStyle {
	switch style {
	case giofont.Italic:
		return FontStyleItalic
	default:
		return FontStyleRegular
	}
}

type fontSource struct {
	data []byte
	path string
}
