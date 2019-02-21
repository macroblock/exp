package fontface

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"unicode"

	"golang.org/x/image/font"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"
	"github.com/macroblock/imed/pkg/misc"
)

// TChar -
type TChar struct {
	Rect    image.Rectangle
	Center  image.Point
	Offset  image.Point
	Advance image.Point
}

// TFontFace -
type TFontFace struct {
	// texture    uint32
	// listbase   uint32
	// maxW, maxH int
	Tex     *image.Gray
	CharMap map[rune]*TChar
	Fixed   bool
}

type tMask struct {
	r         rune
	destRect  image.Rectangle
	maskPoint image.Point
	advance   fixed.Int26_6
}

type tRange struct {
	begin, end rune
}

func emptyCol(img *image.Gray, rect image.Rectangle, x int) bool {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		if img.GrayAt(x, y).Y > 0 {
			return false
		}
	}
	return true
}

func emptyRow(img *image.Gray, rect image.Rectangle, y int) bool {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		if img.GrayAt(x, y).Y > 0 {
			return false
		}
	}
	return true
}

func tightBounds(m *image.Gray) (r image.Rectangle) {
	r = m.Bounds()
	for ; r.Min.Y < r.Max.Y && emptyRow(m, r, r.Min.Y+0); r.Min.Y++ {
	}
	for ; r.Min.Y < r.Max.Y && emptyRow(m, r, r.Max.Y-1); r.Max.Y-- {
	}
	for ; r.Min.X < r.Max.X && emptyCol(m, r, r.Min.X+0); r.Min.X++ {
	}
	for ; r.Min.X < r.Max.X && emptyCol(m, r, r.Max.X-1); r.Max.X-- {
	}
	return r
}

func maxPow2(n int) int {
	x := 1
	for 0 < x && x < n {
		x = x << 1
	}
	if x < 0 {
		return -1
	}
	return x
}

func drawHLine(m *image.Gray, y int) {
	c := color.Gray{127}
	draw.Draw(m, image.Rect(0, y, m.Bounds().Dx(), y+1), &image.Uniform{c}, image.ZP, draw.Src)
}

func drawVLine(m *image.Gray, x int) {
	c := color.Gray{127}
	draw.Draw(m, image.Rect(x, 0, x+1, m.Bounds().Dy()), &image.Uniform{c}, image.ZP, draw.Src)
}
func drawBorder(m *image.Gray, bounds image.Rectangle) {
	c := color.Gray{127}
	minX := bounds.Min.X
	minY := bounds.Min.Y
	maxX := bounds.Max.X
	maxY := bounds.Max.Y
	draw.Draw(m, image.Rect(minX, minY, minX+1, maxY), &image.Uniform{c}, image.ZP, draw.Src)
	draw.Draw(m, image.Rect(maxX-1, minY, maxX, maxY), &image.Uniform{c}, image.ZP, draw.Src)
	draw.Draw(m, image.Rect(minX, minY, maxX, minY+1), &image.Uniform{c}, image.ZP, draw.Src)
	draw.Draw(m, image.Rect(minX, maxY-1, maxX, maxY), &image.Uniform{c}, image.ZP, draw.Src)
}

func drawRect(m *image.Gray, x, y, w, h int) {
	c := color.Gray{191}
	draw.Draw(m, image.Rect(x, y, x+w, y+h), &image.Uniform{c}, image.ZP, draw.Src)
	c = color.Gray{0}
	draw.Draw(m, image.Rect(x+1, y+1, x+w-2, y+h-2), &image.Uniform{c}, image.ZP, draw.Src)
}

func prepData(ttf *truetype.Font, face font.Face) ([]tMask, int, int, fixed.Int26_6, error) {
	slice := []tMask{}
	maxAdvance := fixed.Int26_6(-1)
	volume := -1
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if 0xe000 <= r && r <= 0xf8ff ||
			0xf0000 <= r && r <= 0xffffd ||
			0x100000 <= r && r <= 0x10fffd {
			continue
		}
		// if dr, mask, mpt, adv, ok := face.Glyph(fixed.Point26_6{}, r); ok {
		// 	_ = mask
		// 	_ = dr
		// 	_ = mpt
		// 	_ = adv
		// 	// fmt.Printf("valid skip %U %v %v %v\n", r, dr, mpt, adv)
		// 	// continue
		// }
		if ttf.Index(r) == 0 {
			// fmt.Printf("skiped %U\n", r)
			continue
		}
		dr, _, maskp, adv, ok := face.Glyph(fixed.Point26_6{}, r)
		if !ok {
			return nil, -1, -1, fixed.Int26_6(0), fmt.Errorf("could not load glyph %q %U", r, r)
		}
		maxAdvance = fixed.Int26_6(misc.MaxInt(int(maxAdvance), int(adv)))
		volume += dr.Dx() * dr.Dy()
		maskData := tMask{
			r:         r,
			destRect:  dr,
			maskPoint: maskp,
			advance:   adv,
		}
		slice = append(slice, maskData)
	}
	sort.SliceStable(slice, func(i, j int) bool { return slice[i].destRect.Dy() > slice[j].destRect.Dy() })

	w := int(math.Ceil(math.Sqrt(float64(volume))))
	w = maxPow2(w)
	h := int(math.Ceil(float64(volume) / float64(w)))
	h = maxPow2(h)
	fmt.Printf("volume: %v side: %v x %v\n", volume, w, h)
	fmt.Printf("needed: %v\n", w*h)
	return slice, w, h, maxAdvance, nil
}

// NewFromReader -
func NewFromReader(r io.Reader, size int32, lrune, hrune rune) (*TFontFace, error) {

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(ttf, &truetype.Options{
		Size:       float64(size),
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	})
	defer face.Close()

	fBounds := ttf.Bounds(fixed.Int26_6(size << 6))
	iBounds := image.Rect(
		+fBounds.Min.X.Floor(),
		-fBounds.Max.Y.Ceil(),
		+fBounds.Max.X.Ceil(),
		-fBounds.Min.Y.Floor(),
	)

	slice, texW, texH, maxAdvance, err := prepData(ttf, face)
	if err != nil {
		return nil, err
	}

	charMap := map[rune]*TChar{}
	posX := 0
	posY := 0
	maxH := 0
	adv := maxAdvance
	isFixed := true
	tex := image.NewGray(image.Rect(0, 0, texW, texH))
	for _, item := range slice {
		r := item.r
		tBounds := item.destRect

		if adv != item.advance {
			isFixed = false
		}

		if posX+tBounds.Dx() > tex.Bounds().Dx() {
			posX = 0
			posY += maxH
			maxH = 0
			if posY+tBounds.Dy() > tex.Bounds().Dy() {
				fmt.Println("not enough space in texture")
			}
		}

		if _, ok := charMap[r]; !ok {
			charMap[r] = &TChar{}
		}
		char := charMap[r]

		char.Rect.Min.X = posX
		char.Rect.Min.Y = posY
		char.Rect.Max.X = posX + tBounds.Dx()
		char.Rect.Max.Y = posY + tBounds.Dy()

		char.Advance.X = int(item.advance >> 6)
		char.Advance.Y = iBounds.Dy()

		char.Offset.X = tBounds.Min.X - iBounds.Min.X
		char.Offset.Y = tBounds.Min.Y - iBounds.Min.Y

		char.Center.X = -tBounds.Min.X
		char.Center.Y = -tBounds.Min.Y

		if char.Rect.Dx() == 0 || char.Rect.Dy() == 0 {
			continue
		}
		_, mask, _, _, ok := face.Glyph(fixed.Point26_6{}, r)
		if !ok {
			return nil, fmt.Errorf("unreachable %q, %U", r, r)
		}
		// rect := image.Rect(posX, posY, posX+char.Size.X, posY+char.Size.Y)
		draw.DrawMask(
			tex, char.Rect,
			image.White, image.Point{},
			mask, item.maskPoint,
			draw.Src)

		posX += char.Rect.Dx()
		maxH = misc.MaxInt(maxH, char.Rect.Dy())
	}

	f, err := os.Create("img.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, tex)

	return &TFontFace{
		CharMap: charMap,
		Tex:     tex,
		Fixed:   isFixed,
	}, nil
}

func printBounds(b fixed.Rectangle26_6) {
	fmt.Printf("Min.X:%d Min.Y:%d Max.X:%d Max.Y:%d\n", b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)
}
