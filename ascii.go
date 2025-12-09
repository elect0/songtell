package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	targetCols = 60
	widthRatio = 2.2
	charRamp = "@&&%#$OEHMBDR8GAS0oVvC()s~-+=/|li.,:;' qzjkx"
)

func Convert(artURI string) ([]string, error) {
	if strings.TrimSpace(artURI) == "" {
		return nil, fmt.Errorf("Empty URI")
	}

	img, err := loadImage(artURI)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	newHeight := int(math.Max(1, (float64(h)/float64(w))*float64(targetCols)/widthRatio))

	resizedImg := resizeNearest(img, newHeight, targetCols)
	rampRunes := []rune(charRamp)

	var asciiArtLines []string
	for y := range newHeight {
		var lineBuilder strings.Builder
		lineBuilder.WriteString("  ") // left

		for x := range targetCols {
			c := resizedImg.At(x, y)
			r, g, b, _ := c.RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			brightness := (float64(r8) + float64(g8) + float64(b8)) / 3.0
			idx := int((brightness / 255.0) * float64(len(rampRunes)-1))

			if idx >= len(rampRunes)-1 {
				idx = len(rampRunes) - 1
			}

			char := rampRunes[idx]

			// colors
			rBoost := math.Min(255, float64(r8)*1.1)
			gBoost := math.Min(255, float64(g8)*1.1)
			bBoost := math.Min(255, float64(b8)*1.1)
			ansiCode := getNearestColor(uint8(rBoost), uint8(gBoost), uint8(bBoost))
			lineBuilder.WriteString(fmt.Sprintf("\x1b[%dm%c\x1b[0m", ansiCode, char))
		}
		lineBuilder.WriteString(" ")
		asciiArtLines = append(asciiArtLines, lineBuilder.String())
	}
	asciiArtLines = append(asciiArtLines, "")
	return asciiArtLines, nil
}

func getNearestColor(r, g, b uint8) int {
	palette := []struct {
		r, g, b float64
		code    int
	}{
		// --- Normale (Darker) ---
		{0, 0, 0, 30},       // Black
		{170, 0, 0, 31},     // Red
		{0, 170, 0, 32},     // Green
		{170, 85, 0, 33},    // Yellow
		{0, 0, 170, 34},     // Blue
		{170, 0, 170, 35},   // Magenta
		{0, 170, 170, 36},   // Cyan
		{170, 170, 170, 37}, // White (Light Gray)

		// --- Aprinse (Bright) ---
		{85, 85, 85, 90},    // Bright Black (Gray)
		{255, 85, 85, 91},   // Bright Red
		{85, 255, 85, 92},   // Bright Green
		{255, 255, 85, 93},  // Bright Yellow
		{85, 85, 255, 94},   // Bright Blue
		{255, 85, 255, 95},  // Bright Magenta
		{85, 255, 255, 96},  // Bright Cyan
		{255, 255, 255, 97}, // Bright White
	}

	minDist := math.MaxFloat64
	closestCode := 37 // white

	for _, p := range palette {
		dist := math.Sqrt(math.Pow(float64(r)-p.r, 2) + math.Pow(float64(g)-p.g, 2) + math.Pow(float64(b)-p.b, 2))
		if dist < minDist {
			minDist = dist
			closestCode = p.code
		}
	}

	return closestCode
}

func loadImage(uri string) (image.Image, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse uri.")
	}

	switch u.Scheme {
	case "http", "https":
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		return img, err
	default:
		// local file
		path := uri
		if u.Scheme == "file" {
			path = u.Path
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		return img, err
	}
}

func resizeNearest(img image.Image, newH, newW int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()

	for y := range newH {
		for x := range newW {
			srcX := bounds.Min.X + (x * dx / newW)
			srcY := bounds.Min.Y + (y * dy / newH)
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}

	return dst
}
