package ui

import (
	gl "github.com/go-gl/gl/v3.1/gles2"
)

// TVertexArrayObject -
type TVertexArrayObject struct {
	id uint32
}

// NewVao -
func NewVao() *TVertexArrayObject {
	ret := &TVertexArrayObject{}
	gl.GenVertexArrays(1, &ret.id)
	return ret
}

// Bind -
func (o *TVertexArrayObject) Bind() {
	// fmt.Println("vao: ", o)
	gl.BindVertexArray(o.id)
}

// Unbind -
func (o *TVertexArrayObject) Unbind() {
	gl.BindVertexArray(0)
}

// Draw -
func (o *TVertexArrayObject) Draw(mode uint32, elems *TElementArrayBuffer) {
	o.Bind()
	elems.Bind()
	elems.Draw(mode)
	elems.Unbind()
	o.Unbind()
}

// AddAttribute -
// func (o *TVertexArrayObject) AddAttribute(name string, buffer *TArrayBuffer, stride int32) {
// 	index, err := globActiveProgram.AttribLocation(name)
// 	_ = index
// 	if err != nil {
// 		panic(err)
// 	}
// 	buffer.Bind()
// 	// _, typ, len := glDescribeType(
// 	attribs := globActiveProgram.AttribParams()
// 	desc := (*TAttribParams)(nil)
// 	for i := range attribs {
// 		if name == attribs[i].name {
// 			desc = &attribs[i]
// 			break
// 		}
// 	}
// 	if desc == nil {
// 		panic(fmt.Errorf("could not find attribute %q in program %v", name, globActiveProgram.id))
// 	}
// 	nm, typ, len := glDescribeType(desc.typ)
// 	if len <= 0 {
// 		panic(fmt.Errorf("bad type description %q %v %v of attribute %q", nm, typ, len, name))
// 	}
// 	len = len * desc.num
// 	if len < 1 || 4 < len || typ != buffer.Type() {
// 		nm, _, _ := glDescribeType(desc.typ)
// 		panic(fmt.Errorf("want 1..4 * %q but got %v * %q", buffer.Type(), len, nm))
// 	}
// 	stride = len
// 	gl.VertexAttribPointer(desc.index, stride, typ, false, 0, gl.PtrOffset(0))
// 	gl.EnableVertexAttribArray(index)

// 	nm, _, _ = glDescribeType(typ)
// 	fmt.Printf("new index: %v name: %q type: %v len: %v\n", desc.index, desc.name, nm, len)
// 	nm, _, _ = glDescribeType(buffer.Type())
// 	fmt.Printf("old index: %v name: %q type: %v len: %v\n", index, name, nm, stride)
// }
