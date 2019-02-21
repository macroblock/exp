package main

import (
	"fmt"

	"github.com/macroblock/exp/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

func logf(format string, args ...interface{}) {
	for _, arg := range args {
		fmt.Printf(format, arg)
	}
}

func main() {
	ctx, err := ui.NewContext("test", 800, 600)
	defer ctx.Close()
	if err != nil {
		logf("%v", err)
		return
	}
	logf("%v\n", ctx.Info())

	text := ui.NewText()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				// onStop(ctx)
				running = false
			case *sdl.MouseMotionEvent:
				// xpos = float32(t.X)
				// ypos = float32(t.Y)
			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYDOWN {
					switch t.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					}
				}
			case *sdl.WindowEvent:
				// w, h := ctx.window.GetSize()
				// onDraw(ctx, w, h)
				// sdl.PushEvent(&sdl.WindowEvent{})
			}
		}
		ctx.Draw()
		rect := ctx.Rect()
		text.SetTextColor(0.0, 0.0, 1.0, 1.0)
		text.RenderText("¾Ẁ\x00ABCDEFGHIabcdefgyhi", 0, 0, rect.Dx(), rect.Dy())
		text.SetTextColor(1.0, 0.0, 0.0, 1.0)
		text.RenderText("¾Ẁ\x00ABCDEFGHIabcdefgyhi", 50, 50, rect.Dx(), rect.Dy())
		ctx.Flush() // swap
	}
}
