package ui

import (
	"bytes"
	"fmt"
	"unsafe"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/macroblock/exp/pkg/ui/fontface"

	gl "github.com/go-gl/gl/v3.1/gles2"
)

type (
	// IMesh -
	// IMesh interface {
	// 	Data() []TBuffer
	// }

	// TTexture -
	TTexture struct {
		id  uint32
		typ string
	}

	// TText -
	TText struct {
		vertices  []float32
		indices   []uint32
		textures  []TTexture
		vao       *TVertexArrayObject
		stride    int32
		vbo       *TArrayBuffer
		ebo       *TElementArrayBuffer
		prog      *TProgram
		font      *fontface.TFontFace
		texHandle uint32

		winW int
		winH int

		orthoMatrix mgl32.Mat4
	}
)

// NewText -
func NewText() *TText {
	// vShader := `#version 300 es
	//     #extension GL_ARB_explicit_uniform_location : enable
	//     layout(location=0) in vec2 aPosition;
	//     out vec4 vColor;
	//     out vec2 uv;
	//     void main() {
	//         float k = 0.5;
	//         uv.x = aPosition.x;
	//         uv.y = aPosition.y;
	//         gl_Position = vec4(k*aPosition.x, k*aPosition.y,1.0,1.0);
	//         uv.x = (uv.x + 1.0)*1.0;
	//         uv.y = (-uv.y + 0.0)*1.0;
	//         vColor = vec4(0.7,0.7,0.0,1.0);
	//     }
	// ` + "\x00"
	// fShader := `#version 300 es
	//     precision mediump float;
	//     in vec4 vColor;
	//     in vec2 uv;
	//     out vec4 outColor;
	//     uniform sampler2D texSampler;
	//     void main() {
	//         outColor = vColor;
	//         outColor =  vec4(texture( texSampler, uv ).rrr, 1.0);
	//     }
	// ` + "\x00"
	vShader := `#version 300 es
        #extension GL_ARB_explicit_uniform_location : enable
        layout(location=0) in vec4 Vertex; // [xy, st]

        layout(location=3) uniform mat4 Ortho;
        layout(location=4) uniform vec3 inColor;

        out vec2 TexCoords;
        out vec3 Color;

        void main() {
            gl_Position = Ortho * vec4(Vertex.xy,0.0,1.0);
            // gl_Position = vec4(Vertex.xy,0.0,1.0);
            TexCoords = Vertex.zw;
            Color = inColor;
        }
    ` + "\x00"
	fShader := `#version 300 es
        precision mediump float;
        in vec3 Color;
        in vec2 TexCoords;
        out vec4 outColor;
        uniform sampler2D texSampler;
        void main() {
            vec4 sampled = vec4(Color.rgb, texture(texSampler,TexCoords).r);
            outColor = sampled ;
        }
    ` + "\x00"
	program, err := NewProgram(vShader, fShader)
	if err != nil {
		logPanicf("%v", err)
	}
	ret := &TText{}
	// l := float32(1.0)
	// z := float32(0.5)
	// u0 := float32(0.)
	// v0 := float32(0.0)
	// u1 := float32(3.0)
	// v1 := float32(3.0)
	// ret.vertices = []float32{-l, -l, z, -l, l, z, l, -l, z,
	// 	z, z, z,
	// 	z, z, z, z, z, z, z, z, z, z, z, z,
	// }
	// ret.vertices = make([]float32, 4*6, 4*6)
	// ret.indices = []uint32{0, 1, 2}
	// ret.stride = 3
	ret.prog = program

	ret.font, err = fontface.NewFromReader(bytes.NewReader(goregular.TTF), 14, '{', 'z')
	// ret.font, err = loadTTF(bytes.NewReader(gomono.TTF), 24, '{', 'z')
	if err != nil {
		fmt.Printf("\n Font: %v", err)
	}

	ret.Setup()
	return ret
}

// Setup -
func (o *TText) Setup() {
	bmp := []uint8{
		// 0, 127, 255, 255, 0, 127,
		// 127, 255, 0, 255, 127, 0,
		63, 127, 191, 255,
	}
	_ = bmp
	o.prog.Use()
	vao := NewVao()
	vbo := NewArrayBuffer()
	ebo := NewElementArrayBuffer()
	o.vao = vao
	o.vbo = vbo
	o.ebo = ebo
	gl.GenTextures(1, &o.texHandle)

	vao.Bind()

	// p := o.font.glyphMap['▓']
	p := o.font.Tex
	b := p.Bounds()
	_, _ = p, b
	gl.BindTexture(gl.TEXTURE_2D, o.texHandle)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	gl.TexImage2D(gl.TEXTURE_2D, 0,
		gl.RGB, // gl.LUMINANCE, // gl.RGB,
		int32(b.Dx()), int32(b.Dy()),
		0, gl.LUMINANCE, gl.UNSIGNED_BYTE, unsafe.Pointer(&p.Pix[0]))
	// gl.TexImage2D(gl.TEXTURE_2D, 0,
	// 	gl.RGB,
	// 	int32(2), int32(2),
	// 	0, gl.RED, gl.UNSIGNED_BYTE, unsafe.Pointer(&bmp[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER_NV)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER_NV)
	// gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR_NV, &[]float32{1, 1, 0, 1}[0])

	vbo.Bind()
	vbo.Data(make([]float32, 4*6, 4*6), gl.DYNAMIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, nil)

	// ebo.Bind()
	// ebo.Data(o.indices, gl.STATIC_DRAW)

	// vao.AddAttribute("Vertex", vbo, o.stride)
	// vao.AddAttribute("aPosition", vbo, o.stride)
	vbo.Unbind()

	vao.Unbind()
}

// Draw -
func (o *TText) Draw() {
	// o.prog.Use()
	// o.vao.Bind()
	// o.vao.Draw(gl.TRIANGLES, o.ebo)
	// o.vao.Unbind()
}

// ResizeWindow -
func (o *TText) ResizeWindow(w int, h int) {
	o.winW = w
	o.winH = h
	o.orthoMatrix = mgl32.Ortho2D(
		-float32(o.winW)/2, float32(o.winW)/2,
		-float32(o.winH)/2, float32(o.winH)/2,
	)
}

// SetTextColor -
func (o *TText) SetTextColor(r, g, b, a float32) {
	o.prog.Use()
	gl.Uniform3f(4, r, g, b)
}

// RenderText -
func (o *TText) RenderText(s string, x0, y0 int, screenW, screenH int) {
	gl.Disable(gl.DEPTH_TEST)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	o.prog.Use()

	mtx := mgl32.Ortho(float32(0), float32(screenW), float32(screenH), float32(0), -1.0, 1.0)
	mtx = mgl32.Ortho2D(float32(0), float32(screenW), float32(screenH), float32(0))
	gl.UniformMatrix4fv(3, 1, false, &mtx[0])

	// cr := float32(0.5+255) / 256
	// cg := float32(0.5+128) / 256
	// cb := float32(0.5+255) / 256
	// cr = 1
	// cg = 1
	// cb = 1
	// gl.Uniform3f(4, cr, cg, cb)

	scale := float32(1.0 / 1)
	gl.BindBuffer(gl.TEXTURE_2D, o.texHandle)
	o.vao.Bind()
	x := float32(x0)
	y := float32(y0)
	for _, r := range s {
		ch, ok := o.font.CharMap[r]
		if !ok {
			continue
		}

		// fmt.Printf("%q xy %v %v wh %vx%v adv %v | ", r, ch.texX, ch.texY, ch.W, ch.H, ch.advanceX)
		xpos0 := x + float32(ch.Offset.X) //float32(ch.TTexX)
		ypos0 := y + float32(ch.Offset.Y) //float32(ch.TTexY)
		xpos1 := xpos0 + float32(ch.Rect.Dx())*scale
		ypos1 := ypos0 + float32(ch.Rect.Dy())*scale
		s0 := float32(ch.Rect.Min.X) / float32(o.font.Tex.Bounds().Dx())
		t0 := float32(ch.Rect.Max.Y) / float32(o.font.Tex.Bounds().Dy())
		s1 := float32(ch.Rect.Max.X) / float32(o.font.Tex.Bounds().Dx())
		t1 := float32(ch.Rect.Min.Y) / float32(o.font.Tex.Bounds().Dy())
		// xpos = 5
		// ypos = 5
		// w = float32(screenW - 10)
		// h = float32(screenH - 10)
		// s = 0
		// t = 0
		// texW = 1
		// texH = 1
		vertices := []float32{
			xpos0, ypos1, s0, t0,
			xpos0, ypos0, s0, t1,
			xpos1, ypos0, s1, t1,

			xpos0, ypos1, s0, t0,
			xpos1, ypos0, s1, t1,
			xpos1, ypos1, s1, t0,
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, o.vbo.id)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, unsafe.Pointer(&vertices[0]))
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		x += float32(ch.Advance.X) * scale
		// break
	}
	o.vao.Unbind()
}
