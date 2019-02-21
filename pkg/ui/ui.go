package ui

import "C"
import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"unsafe"

	gl "github.com/go-gl/gl/v3.1/gles2"
	"github.com/macroblock/exp/pkg/ui/theme"
	"github.com/veandco/go-sdl2/sdl"
)

func logErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args)
}

func logPanicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args))
}

// TContext -
type TContext struct {
	valid    bool
	window   *sdl.Window
	context  sdl.GLContext
	userData interface{}
}

// NewContext -
func NewContext(title string, w, h int) (*TContext, error) {
	ctx := &TContext{}
	err := error(nil)
	runtime.LockOSThread()
	err = sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, logErrorf("sdl.Init: %v", err)
	}

	err = sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	if err != nil {
		return nil, logErrorf("sdl.GLSetAttribute: %v", err)
	}

	ctx.window, err = sdl.CreateWindow(title,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(w), int32(h),
		sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, logErrorf("sdl.CreateWindow: %v", err)
	}

	ctx.context, err = ctx.window.GLCreateContext()
	if err != nil {
		return nil, logErrorf("sdl.GLCreateContext: %v", err)
	}

	err = gl.Init()
	if err != nil {
		return nil, logErrorf("gles.Init: %v", err)
	}
	ctx.valid = true
	return ctx, nil
}

// Close -
func Close(ctx **TContext) {
	if ctx == nil || *ctx == nil {
		return
	}
	c := *ctx
	c.valid = false
	if c.context != nil {
		sdl.GLDeleteContext(c.context)
		c.context = nil
	}
	if c.window != nil {
		c.window.Destroy()
		c.window = nil
	}
	sdl.Quit()
	*ctx = nil
}

// Invalid -
func (o *TContext) Invalid() bool {
	if o == nil || !o.valid {
		return true
	}
	return false
}

// Close -
func (o *TContext) Close() {
	Close(&o)
}

func glGetString(val uint32) string {
	x := gl.GetString(val)
	p := (*C.char)(unsafe.Pointer(x))
	return C.GoString(p)
}

// SetUserData -
func (o *TContext) SetUserData(data interface{}) {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	o.userData = data
}

// UserData -
func (o *TContext) UserData() interface{} {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	return o.userData
}

// Info -
func (o *TContext) Info() string {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	ret := fmt.Sprintf("sdl\n")
	v := &sdl.Version{}
	sdl.GetVersion(v)
	ret += fmt.Sprintf("  version: %v\n", v)

	ret += fmt.Sprintf("opengl\n")
	ret += fmt.Sprintf("  version       : %v\n", glGetString(gl.VERSION))
	ret += fmt.Sprintf("  shader version: %v\n", glGetString(gl.SHADING_LANGUAGE_VERSION))
	ret += fmt.Sprintf("  vendor        : %v\n", glGetString(gl.VENDOR))
	ret += fmt.Sprintf("  renderer      : %v\n", glGetString(gl.RENDERER))
	return ret
}

// Rect -
func (o *TContext) Rect() image.Rectangle {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	x, y := o.window.GetPosition()
	w, h := o.window.GetSize()
	return image.Rectangle{
		image.Point{int(x), int(y)},
		image.Point{int(x) + int(w) + 1, int(y) + int(h) + 1},
	}
}

func colorToFloat32(color color.Color) (float32, float32, float32, float32) {
	r, g, b, a := color.RGBA()
	fr := float32(r) / float32(0xffff)
	fg := float32(g) / float32(0xffff)
	fb := float32(b) / float32(0xffff)
	fa := float32(a) / float32(0xffff)
	return fr, fg, fb, fa
}

// Draw -
func (o *TContext) Draw() {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	pal := theme.Default.GetPalette()
	r, g, b, a := colorToFloat32(pal[theme.Dark])
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// Flush -
func (o *TContext) Flush() {
	if o.Invalid() {
		logPanicf("invalid context\n")
	}
	o.window.GLSwap()
}
