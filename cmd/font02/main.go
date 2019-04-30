package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"unicode"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type (
	// Range -
	Range struct{ min, max rune }
)

func ttfStats(ttf *truetype.Font) []*Range {
	slice := []*Range{}
	hasReplacementRune := ttf.Index(0xfffd) != 0

	rng := (*Range)(nil)
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if 0xe000 <= r && r <= 0xf8ff ||
			0xf0000 <= r && r <= 0xffffd ||
			0x100000 <= r && r <= 0x10fffd {
			rng = nil
			continue
		}
		if ttf.Index(r) == 0 {
			rng = nil
			continue
		}
		if rng == nil {
			rng = &Range{min: r}
			slice = append(slice, rng)
		}
		rng.max = r + 1
	}
	fmt.Printf("replacement rune: %v\n", hasReplacementRune)
	return slice
}

func faceStats(face font.Face) []*Range {
	slice := []*Range{}
	hasReplacementRune := false
	if _, ok := face.GlyphAdvance(0xfffd); ok {
		hasReplacementRune = true
	}

	rng := (*Range)(nil)
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if 0xe000 <= r && r <= 0xf8ff ||
			0xf0000 <= r && r <= 0xffffd ||
			0x100000 <= r && r <= 0x10fffd {
			rng = nil
			continue
		}
		// if _, ok := face.GlyphAdvance(0xfffd); !ok {
		// if _, _, ok := face.GlyphBounds(r); !ok {
		if _, _, _, _, ok := face.Glyph(fixed.Point26_6{}, r); !ok {
			rng = nil
			continue
		}
		if rng == nil {
			rng = &Range{min: r}
			slice = append(slice, rng)
		}
		rng.max = r + 1
	}
	fmt.Printf("replacement rune: %v\n", hasReplacementRune)
	return slice
}

func main() {

	data, err := ioutil.ReadAll(bytes.NewReader(goregular.TTF))
	if err != nil {
		fmt.Println("read all: ", err)
		return
	}

	ttf, err := truetype.Parse(data)
	if err != nil {
		fmt.Println("truetype parse: ", err)
		return
	}

	ttfSlice := ttfStats(ttf)
	_ = ttfSlice

	face := truetype.NewFace(ttf, &truetype.Options{
		Size:       float64(14),
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	})
	defer face.Close()

	faceSlice := faceStats(face)
	_ = faceSlice
	ttfSlice = faceSlice

	total := 0
	for i, v := range ttfSlice {
		fmt.Printf("  %3v %4x-%4x: %4v\n", i, v.min, v.max-1, v.max-v.min)
		total += int(v.max - v.min)
	}
	fmt.Printf("total %4x-%4x: %4v\n", ttfSlice[0].min, ttfSlice[len(ttfSlice)-1].max-1, total)

}
