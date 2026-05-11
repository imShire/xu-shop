// Package poster 实现海报图片渲染。
package poster

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png" // 注册 PNG 解码器（小程序码为 PNG 格式）
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Renderer 海报渲染接口。
type Renderer interface {
	// Render 生成海报图片，返回 JPEG bytes。
	Render(ctx context.Context, req RenderReq) ([]byte, error)
}

// RenderReq 渲染请求参数。
type RenderReq struct {
	ProductImageURL string // 商品主图 URL
	Title           string // 商品标题
	PriceCents      int64  // 价格（分）
	QRCodePNG       []byte // 小程序码 PNG bytes
}

const (
	canvasW  = 600
	canvasH  = 800
	prodImgH = 500
	qrSize   = 150
	jpegQ    = 85
)

// ImageRenderer 使用标准库 image/draw 渲染海报。
type ImageRenderer struct {
	httpCli *http.Client
}

// NewImageRenderer 构造 ImageRenderer，HTTP 超时 5s。
func NewImageRenderer() *ImageRenderer {
	return &ImageRenderer{
		httpCli: &http.Client{Timeout: 5 * time.Second},
	}
}

// Render 生成海报 JPEG bytes。
// 布局：顶部 600×500 商品主图，下方标题（最多 2 行），价格（红色），右下角 150×150 小程序码。
func (r *ImageRenderer) Render(ctx context.Context, req RenderReq) ([]byte, error) {
	// 1. 创建白色画布 600×800
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	draw.Draw(canvas, canvas.Bounds(), image.White, image.Point{}, draw.Src)

	// 2. 商品主图（顶部 600×500）
	if req.ProductImageURL != "" {
		if img, err := fetchImage(ctx, r.httpCli, req.ProductImageURL); err == nil {
			drawScaled(canvas, image.Rect(0, 0, canvasW, prodImgH), img)
		}
	}

	// 3. 小程序码（右下角 150×150，距边缘 10px）
	if len(req.QRCodePNG) > 0 {
		if qr, _, err := image.Decode(bytes.NewReader(req.QRCodePNG)); err == nil {
			qrX := canvasW - qrSize - 10
			qrY := canvasH - qrSize - 10
			drawScaled(canvas, image.Rect(qrX, qrY, qrX+qrSize, qrY+qrSize), qr)
		}
	}

	// 4. 标题（最多 2 行，每行 20 字；basicfont 仅支持 ASCII，中文显示为缺字符）
	runes := []rune(req.Title)
	textY := prodImgH + 28
	for line := 0; line < 2 && line*20 < len(runes); line++ {
		start := line * 20
		end := start + 20
		if end > len(runes) {
			end = len(runes)
		}
		drawText(canvas, 16, textY, string(runes[start:end]), color.Black)
		textY += 22
	}

	// 5. 价格（红色；basicfont 不支持 ¥，以 Y 替代）
	yuan := req.PriceCents / 100
	fen := req.PriceCents % 100
	priceStr := fmt.Sprintf("Y%d.%02d", yuan, fen)
	drawText(canvas, 16, textY+12, priceStr, color.RGBA{R: 200, A: 255})

	// 6. 编码为 JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, canvas, &jpeg.Options{Quality: jpegQ}); err != nil {
		return nil, fmt.Errorf("poster: encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// fetchImage 通过 HTTP 下载并解码图片。
func fetchImage(ctx context.Context, cli *http.Client, url string) (image.Image, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("poster: image URL must use https scheme")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("poster: build image request: %w", err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poster: fetch image: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("poster: read image body: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("poster: decode image: %w", err)
	}
	return img, nil
}

// drawScaled 将 src 按最近邻算法缩放绘制到 dst 的 dstRect 区域。
func drawScaled(dst draw.Image, dstRect image.Rectangle, src image.Image) {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dw, dh := dstRect.Dx(), dstRect.Dy()
	if sw == 0 || sh == 0 || dw == 0 || dh == 0 {
		return
	}
	for dy := 0; dy < dh; dy++ {
		for dx := 0; dx < dw; dx++ {
			sx := dx * sw / dw
			sy := dy * sh / dh
			dst.Set(dstRect.Min.X+dx, dstRect.Min.Y+dy, src.At(sb.Min.X+sx, sb.Min.Y+sy))
		}
	}
}

// drawText 在画布 (x,y) 处绘制文本（使用 basicfont 7×13，仅支持 ASCII）。
func drawText(dst draw.Image, x, y int, text string, col color.Color) {
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}
