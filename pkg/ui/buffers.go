package ui

import (
	"fmt"
	"reflect"
	"unsafe"

	gl "github.com/go-gl/gl/v3.1/gles2"
)

// IBuffer -
type IBuffer interface {
	Bind(uint32)
	Unbind(uint32)
	Type() uint32
	Len() int
	Size() int
}

// TBuffer -
type TBuffer struct {
	id        uint32
	typ       uint32
	len       int
	size      int
	usageHint uint32
}

// TArrayBuffer -
type TArrayBuffer struct {
	TBuffer
}

// TElementArrayBuffer -
type TElementArrayBuffer struct {
	TBuffer
}

// NewBuffer -
func NewBuffer() *TBuffer {
	ret := &TBuffer{usageHint: gl.STATIC_DRAW}
	gl.GenBuffers(1, &ret.id)
	return ret
}

// Bind -
func (o *TBuffer) Bind(target uint32) {
	gl.BindBuffer(target, o.id)
}

// Unbind -
func (o *TBuffer) Unbind(target uint32) {
	gl.BindBuffer(target, 0)
}

// Type -
func (o *TBuffer) Type() uint32 { return o.typ }

// Len -
func (o *TBuffer) Len() int { return o.len }

// Size -
func (o *TBuffer) Size() int { return o.size }

// Data -
func (o *TBuffer) Data(target uint32, data interface{}, usageHint ...uint32) {
	o.typ, o.size = loadData2(o, target, data, usageHint...)
}

// NewArrayBuffer -
func NewArrayBuffer() *TArrayBuffer {
	ret := &TArrayBuffer{TBuffer: *NewBuffer()}
	gl.GenBuffers(1, &ret.id)
	return ret
}

// Bind -
func (o *TArrayBuffer) Bind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, o.id)
}

// Unbind -
func (o *TArrayBuffer) Unbind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// Data -
func (o *TArrayBuffer) Data(data interface{}, usageHint ...uint32) {
	o.typ, o.size = loadData2(&o.TBuffer, gl.ARRAY_BUFFER, data, usageHint...)
}

// NewElementArrayBuffer -
func NewElementArrayBuffer() *TElementArrayBuffer {
	ret := &TElementArrayBuffer{TBuffer: *NewBuffer()}
	gl.GenBuffers(1, &ret.id)
	return ret
}

// Bind -
func (o *TElementArrayBuffer) Bind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, o.id)
}

// Unbind -
func (o *TElementArrayBuffer) Unbind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

// Data -
func (o *TElementArrayBuffer) Data(data interface{}, usageHint ...uint32) {
	o.typ, o.size = loadData2(&o.TBuffer, gl.ELEMENT_ARRAY_BUFFER, data, usageHint...)
}

// Draw -
func (o TElementArrayBuffer) Draw(mode uint32) {
	gl.DrawElements(mode, int32(o.size), o.typ, gl.PtrOffset(0)) //gl.Ptr(&ctx.elements[0]))
}

// func loadData(o *TBuffer, target uint32, data interface{}, usageHint ...uint32) int {
// 	if len(usageHint) > 0 {
// 		o.usageHint = usageHint[0]
// 	}

// 	val := reflect.ValueOf(data)
// 	if val.Len() == 0 {
// 		// ???
// 		return 0
// 	}
// 	typ := reflect.TypeOf(data)
// 	typeSize := typ.Elem().Size()
// 	sliceLen := val.Len() // for slice, arrays or chan only
// 	ptr := unsafe.Pointer(val.Pointer())
// 	size := int(sliceLen) * int(typeSize)
// 	// fmt.Println(sliceLen, " ", typeSize, " ", ptr)
// 	gl.BufferData(
// 		target,
// 		size,
// 		ptr,
// 		o.usageHint)
// 	// fmt.Printf("-> %v\n", *(*uint32)(ptr))
// 	return size
// }

func loadData2(o *TBuffer, target uint32, data interface{}, usageHint ...uint32) (uint32, int) {
	if len(usageHint) > 0 {
		o.usageHint = usageHint[0]
	}
	var elemTyp reflect.Type
	var glType uint32

	val := reflect.ValueOf(data)
	typ := val.Type()
	switch typ.Kind() {
	default:
		panic(fmt.Errorf("unsupported type %s; must be a slice or pointer to a scalar value", typ))
	case reflect.Ptr:
		panic(fmt.Errorf("not yet supported type %s", val.Type()))
	case reflect.Slice:
		elemTyp = typ.Elem()
		o.len = val.Len() // for slice, arrays or chan only
		switch elemTyp.Kind() {
		default:
			panic(fmt.Errorf("unsupported type of element %s; must be a scalar type", elemTyp))
		case reflect.Int8:
			glType = gl.BYTE
		case reflect.Uint8:
			glType = gl.UNSIGNED_BYTE
		case reflect.Int16:
			glType = gl.SHORT
		case reflect.Uint16:
			glType = gl.UNSIGNED_SHORT
		case reflect.Int32:
			glType = gl.INT
			// switch elemVal.Interface().(type) {
			// case fixed.Int16_16:
			// elemType = gl.FIXED
			// }
		case reflect.Uint32:
			glType = gl.UNSIGNED_INT
		// case reflect.Float16:
		// 	elemType = gl.HALF_FLOAT
		case reflect.Float32:
			glType = gl.FLOAT
			// case reflect.Float64:
			// 	elemType = gl.DOUBLE
		}
	}
	ptr := unsafe.Pointer(val.Pointer())
	typeSize := elemTyp.Size()
	size := int(o.len) * int(typeSize)
	// if val.Len() == 0 {
	// 	// ???
	// 	return glType, 0
	// }
	// fmt.Println(sliceLen, " ", typeSize, " ", ptr)
	gl.BufferData(
		target,
		size,
		ptr,
		o.usageHint)
	// fmt.Printf("-> %v\n", *(*uint32)(ptr))
	return glType, size
}
