package ui

import (
	"C"
	"fmt"
	"strings"

	gl "github.com/go-gl/gl/v3.1/gles2"
	"github.com/macroblock/imed/pkg/misc"
)

var globActiveProgram *TProgram

type (
	// TShader -
	TShader struct {
		id uint32
	}
	// TVertexShader -
	TVertexShader TShader
	// TFragmentShader -
	TFragmentShader TShader
	// TProgram -
	TProgram struct {
		id       uint32
		attribs  []TAttribParams
		uniforms []TAttribParams
	}

	// TAttribParams -
	TAttribParams struct {
		name  string
		index uint32
		typ   uint32
		num   int32
	}
)

// NewShader -
func NewShader(shaderType uint32, src string) (TShader, error) {
	id, err := createShader(shaderType, src)
	return TShader{id: id}, err
}

// NewVertexShader -
func NewVertexShader(src string) (TVertexShader, error) {
	shader, err := NewShader(gl.VERTEX_SHADER, src)
	return TVertexShader(shader), err
}

// NewFragmentShader -
func NewFragmentShader(src string) (TFragmentShader, error) {
	shader, err := NewShader(gl.FRAGMENT_SHADER, src)
	return TFragmentShader(shader), err
}

// NewProgram -
func NewProgram(vshader interface{}, fshader interface{}) (*TProgram, error) {
	var err error
	var vs TVertexShader
	var fs TFragmentShader

	switch shader := vshader.(type) {
	default:
		return nil, fmt.Errorf("unsupported type of a vertex shader: %T", vshader)
	case string:
		vs, err = NewVertexShader(shader)
		if err != nil {
			return nil, err
		}
	case TVertexShader:
		vs = shader
	case *TVertexShader:
		vs = *shader
	}
	switch shader := fshader.(type) {
	default:
		return nil, fmt.Errorf("unsupported type of a fragment shader: %T", vshader)
	case string:
		fs, err = NewFragmentShader(shader)
		if err != nil {
			return nil, err
		}
	case TFragmentShader:
		fs = shader
	case *TFragmentShader:
		fs = *shader
	}

	ret := &TProgram{}
	ret.id, err = createProgram(vs.id, fs.id)
	if err != nil {
		return nil, err
	}

	ret.attribs = ret.AttribParams()
	ret.uniforms = ret.UniformParams()
	return ret, err
}

// Use -
func (o *TProgram) Use() {
	globActiveProgram = o
	gl.UseProgram(o.id)
}

// AttribLocation -
func (o *TProgram) AttribLocation(name string) (uint32, error) {
	index := gl.GetAttribLocation(o.id, gl.Str(name+"\x00"))
	if index == -1 {
		return uint32(index), fmt.Errorf("attribute index %q at program %v is not found", name, o.id)
	}
	return uint32(index), nil
}

// UniformLocation -
func (o *TProgram) UniformLocation(name string) (uint32, error) {
	index := gl.GetUniformLocation(o.id, gl.Str(name+"\x00"))
	if index == -1 {
		return uint32(index), fmt.Errorf("uniform index %q at program %v is not found", name, o.id)
	}
	return uint32(index), nil
}

// AttribParams -
func (o *TProgram) AttribParams() []TAttribParams {
	var count, bufSize, nameLength, maxNameLength int32
	if o.attribs != nil {
		return o.attribs
	}
	gl.GetProgramiv(o.id, gl.ACTIVE_ATTRIBUTES, &count)
	count = int32(misc.MaxInt(0, int(count)))
	gl.GetProgramiv(o.id, gl.ACTIVE_ATTRIBUTE_MAX_LENGTH, &maxNameLength)
	buf := make([]byte, maxNameLength)
	bufSize = int32(len(buf))
	ret := []TAttribParams{}
	for i := uint32(0); i < uint32(count); i++ {
		params := TAttribParams{}
		gl.GetActiveAttrib(o.id, uint32(i), bufSize, &nameLength, &params.num, &params.typ, &buf[0])
		params.name = string(buf[:nameLength])
		index, err := o.AttribLocation(params.name)
		if err != nil {
			panic(err)
		}
		params.index = index
		ret = append(ret, params)
	}
	return ret
}

// UniformParams -
func (o *TProgram) UniformParams() []TAttribParams {
	var count, bufSize, nameLength, maxNameLength int32
	if o.uniforms != nil {
		return o.uniforms
	}
	gl.GetProgramiv(o.id, gl.ACTIVE_UNIFORMS, &count)
	count = int32(misc.MaxInt(0, int(count)))
	gl.GetProgramiv(o.id, gl.ACTIVE_UNIFORM_MAX_LENGTH, &maxNameLength)
	buf := make([]byte, maxNameLength)
	bufSize = int32(len(buf))
	ret := []TAttribParams{}
	for i := uint32(0); i < uint32(count); i++ {
		params := TAttribParams{}
		gl.GetActiveUniform(o.id, uint32(i), bufSize, &nameLength, &params.num, &params.typ, &buf[0])
		params.name = string(buf[:nameLength])
		index, err := o.UniformLocation(params.name)
		if err != nil {
			panic(err)
		}
		params.index = index
		ret = append(ret, params)
	}
	return ret
}

func createShader(shaderType uint32, src string) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	if shader == 0 {
		// ???
		return 0, fmt.Errorf("unable to create shader: %v", gl.GetError())
	}

	csrc, free := gl.Strs(src)
	gl.ShaderSource(shader, 1, csrc, nil)
	free()
	gl.CompileShader(shader)

	status := int32(0)
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		logLen := int32(0)
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLen)

		log := strings.Repeat("\x00", int(logLen+1))
		gl.GetShaderInfoLog(shader, logLen, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", src, log)
	}
	return shader, nil
}

func createProgram(vShader, fShader uint32) (uint32, error) {
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vShader)
	gl.AttachShader(prog, fShader)
	gl.LinkProgram(prog)

	status := int32(0)
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		logLen := int32(0)
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)

		log := strings.Repeat("\x00", int(logLen+1))
		gl.GetProgramInfoLog(prog, logLen, nil, gl.Str(log))

		gl.DeleteProgram(prog) // ???
		return 0, fmt.Errorf("failed to link program %v: %v", prog, log)
	}
	gl.DeleteShader(vShader)
	gl.DeleteShader(fShader)

	return prog, nil
}
