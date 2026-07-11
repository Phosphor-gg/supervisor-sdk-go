package supervisor

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"strings"

	// Register the stdlib decoders so image.Decode can sniff these formats.
	_ "image/gif"
	_ "image/png"
)

// maxImageEdge is the longest edge, in pixels, PrepareImage will send.
const maxImageEdge = 1280

// PrepareImage normalizes a base64-encoded image before upload: it strips an
// optional "data:...;base64," prefix, downscales the image so its longest
// edge is at most 1280 pixels (aspect ratio preserved, never upscaled),
// flattens any transparency onto a white background, and re-encodes the
// result as JPEG. The returned string is always raw standard base64.
//
// Inputs that cannot be decoded — invalid base64, or formats the Go standard
// library cannot read (e.g. WebP) — are returned unchanged. If no resize was
// needed and the JPEG re-encode would not shrink the payload, the original
// image bytes are returned instead.
//
// The client's Moderate and ModerateBatch methods call PrepareImage
// automatically; it is exported for callers who want to preprocess images
// themselves (e.g. to cache the prepared form).
func PrepareImage(imageB64 string) string {
	raw := imageB64
	if strings.HasPrefix(raw, "data:") {
		idx := strings.Index(raw, ";base64,")
		if idx < 0 {
			return imageB64
		}
		raw = raw[idx+len(";base64,"):]
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(raw)
		if err != nil {
			return imageB64
		}
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return imageB64
	}

	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if sw <= 0 || sh <= 0 {
		return imageB64
	}

	dw, dh := sw, sh
	resized := false
	if sw > maxImageEdge || sh > maxImageEdge {
		resized = true
		if sw >= sh {
			dw = maxImageEdge
			dh = (sh*maxImageEdge + sw/2) / sw
		} else {
			dh = maxImageEdge
			dw = (sw*maxImageEdge + sh/2) / sh
		}
		if dw < 1 {
			dw = 1
		}
		if dh < 1 {
			dh = 1
		}
	}

	flat := flattenAndScale(src, dw, dh)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, flat, &jpeg.Options{Quality: 85}); err != nil {
		return imageB64
	}

	if !resized && buf.Len() >= len(data) {
		return base64.StdEncoding.EncodeToString(data)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// flattenAndScale converts src to an opaque RGBA image of dw x dh pixels,
// compositing transparency onto a white background. Downscaling uses an
// area-average (box) filter: every source pixel is accumulated into the
// destination bucket its coordinates map to, then each bucket is averaged.
// dw and dh must not exceed the source dimensions.
func flattenAndScale(src image.Image, dw, dh int) *image.RGBA {
	bounds := src.Bounds()
	sw, sh := bounds.Dx(), bounds.Dy()

	sumR := make([]float64, dw*dh)
	sumG := make([]float64, dw*dh)
	sumB := make([]float64, dw*dh)
	count := make([]float64, dw*dh)

	for sy := 0; sy < sh; sy++ {
		dy := sy * dh / sh
		for sx := 0; sx < sw; sx++ {
			dx := sx * dw / sw
			// RGBA() returns alpha-premultiplied 16-bit channels, so
			// compositing over white is c/257 + (255 - a/257).
			r, g, b, a := src.At(bounds.Min.X+sx, bounds.Min.Y+sy).RGBA()
			white := 255.0 - float64(a)/257.0
			i := dy*dw + dx
			sumR[i] += float64(r)/257.0 + white
			sumG[i] += float64(g)/257.0 + white
			sumB[i] += float64(b)/257.0 + white
			count[i]++
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for i := 0; i < dw*dh; i++ {
		n := count[i]
		if n == 0 {
			n = 1
		}
		p := i * 4
		dst.Pix[p] = clampByte(sumR[i] / n)
		dst.Pix[p+1] = clampByte(sumG[i] / n)
		dst.Pix[p+2] = clampByte(sumB[i] / n)
		dst.Pix[p+3] = 0xFF
	}
	return dst
}

func clampByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v + 0.5)
}
