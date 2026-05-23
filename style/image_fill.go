package style

import (
	"image"
	"io"
	"net/http"
	"os"
	"strings"
)

// ImageFillFit 定义图片背景的缩放模式。
type ImageFillFit int

const (
	ImageFillContain ImageFillFit = iota // 1
	ImageFillCover                       // 2
	ImageFillFill                        // 3
	ImageFillNone                        // 4
)

// ImageFill 描述作为背景使用的图片及其缩放模式。
// Src 为已解码的 image.Image；Fit 决定图片如何填满容器区域。
type ImageFill struct {
	Src image.Image
	Fit ImageFillFit
}

// DecodeImageURL 从 HTTP(S) URL 下载并解码图片。
func DecodeImageURL(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeImageFromReader(resp.Body)
}

// DecodeImageFile 从本地文件路径解码图片。
func DecodeImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return decodeImageFromReader(f)
}

// LoadImage 自动检测 URL 或文件路径，解码图片。
// 以 "http://" 或 "https://" 开头视为 URL，否则视为本地文件。
func LoadImage(src string) (image.Image, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return DecodeImageURL(src)
	}
	return DecodeImageFile(src)
}

func decodeImageFromReader(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}
