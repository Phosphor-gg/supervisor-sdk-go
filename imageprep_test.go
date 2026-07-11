package supervisor

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// pngB64 encodes img as PNG and returns it base64-encoded.
func pngB64(t *testing.T, img image.Image) string {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// decodeB64Image decodes a base64 string produced by PrepareImage back into
// an image, returning the image and its sniffed format.
func decodeB64Image(t *testing.T, b64 string) (image.Image, string) {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("output is not valid standard base64: %v", err)
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode output image: %v", err)
	}
	return img, format
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func TestPrepareImageDownscalesLargeOpaqueImage(t *testing.T) {
	const sw, sh = 3000, 2000
	src := image.NewRGBA(image.Rect(0, 0, sw, sh))
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			src.SetRGBA(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 120, A: 255})
		}
	}

	out := PrepareImage(pngB64(t, src))
	img, format := decodeB64Image(t, out)
	if format != "jpeg" {
		t.Fatalf("expected jpeg output, got %q", format)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w != 1280 {
		t.Fatalf("expected longest edge exactly 1280, got %dx%d", w, h)
	}
	expectedH := sh * 1280 / sw // aspect-preserving height
	if absInt(h-expectedH) > 1 {
		t.Fatalf("aspect not preserved: got %dx%d, expected height ~%d", w, h, expectedH)
	}
}

func TestPrepareImageDownscalesTransparentImage(t *testing.T) {
	const sw, sh = 4000, 1000
	src := image.NewNRGBA(image.Rect(0, 0, sw, sh))
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			if x < sw/2 {
				// Left half: fully transparent.
				src.SetNRGBA(x, y, color.NRGBA{})
			} else {
				// Right half: half-transparent red.
				src.SetNRGBA(x, y, color.NRGBA{R: 255, A: 128})
			}
		}
	}

	out := PrepareImage(pngB64(t, src))
	img, format := decodeB64Image(t, out)
	if format != "jpeg" {
		t.Fatalf("expected jpeg output, got %q", format)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w != 1280 {
		t.Fatalf("expected longest edge exactly 1280, got %dx%d", w, h)
	}
	expectedH := sh * 1280 / sw
	if absInt(h-expectedH) > 1 {
		t.Fatalf("aspect not preserved: got %dx%d, expected height ~%d", w, h, expectedH)
	}

	// Fully transparent pixels must have been flattened onto white.
	r, g, b, _ := img.At(10, h/2).RGBA()
	if r/257 < 240 || g/257 < 240 || b/257 < 240 {
		t.Fatalf("transparent region not flattened to white: got rgb(%d, %d, %d)", r/257, g/257, b/257)
	}
}

func TestPrepareImageSmallImagePassesThroughResizeFree(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			src.SetRGBA(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 200, A: 255})
		}
	}

	in := pngB64(t, src)
	out := PrepareImage(in)
	if out == in {
		return // original returned unchanged: acceptable
	}
	img, _ := decodeB64Image(t, out)
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Fatalf("64x64 image was resized: got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestPrepareImageGarbageReturnsUnchanged(t *testing.T) {
	for _, in := range []string{
		"!!!", // not base64
		base64.StdEncoding.EncodeToString([]byte("hello")), // base64, not an image
	} {
		if out := PrepareImage(in); out != in {
			t.Errorf("PrepareImage(%q) = %q, expected input unchanged", in, out)
		}
	}
}

func TestPrepareImageStripsDataURLPrefix(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2000, 2000))
	for y := 0; y < 2000; y++ {
		for x := 0; x < 2000; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 30, G: 60, B: 90, A: 255})
		}
	}

	out := PrepareImage("data:image/png;base64," + pngB64(t, src))
	img, format := decodeB64Image(t, out) // fails if any prefix survived
	if format != "jpeg" {
		t.Fatalf("expected jpeg output, got %q", format)
	}
	if img.Bounds().Dx() != 1280 || img.Bounds().Dy() != 1280 {
		t.Fatalf("expected 1280x1280, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}
