package theme

import (
	"image/color"
	"unsafe"
)

// TTheme -
type TTheme struct {
	DPI     float64
	FonFace unsafe.Pointer
	palette *TPalette
}

// TPalette -
type TPalette [paletteLen]color.Color

// TPaletteIndex -
type TPaletteIndex int

// color condtants
const (
	Light = TPaletteIndex(iota)
	Neutral
	Dark

	paletteLen
)

// DefaultDPI -
const DefaultDPI = 72.0

var (
	defaultPalette = TPalette{
		Light:   color.RGBA{0xf5, 0xf5, 0xf5, 0xff}, // Material Design "Grey 100"
		Neutral: color.RGBA{0xee, 0xee, 0xee, 0xff}, // Material Design "Grey 200"
		Dark:    color.RGBA{0x0, 0xb0, 0xe0, 0xff},  // Material Design "Grey 300"
	}

	// Default -
	Default *TTheme
)

// GetDPI -
func (o *TTheme) GetDPI() float64 {
	if o == nil || o.DPI == 0.0 {
		return DefaultDPI
	}
	return o.DPI
}

// GetPalette -
func (o *TTheme) GetPalette() *TPalette {
	if o == nil || o.palette == nil {
		return &defaultPalette
	}
	return o.palette
}
