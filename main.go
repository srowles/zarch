package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	triangle = []float32{
		// first triangle
		0.5, 0.5, 0.0, // top right
		0.5, -0.5, 0.0, // bottom right
		-0.5, 0.5, 0.0, // top left
		// second triangle
		// 0.5, -0.5, 0.0, // bottom right
		// -0.5, -0.5, 0.0, // bottom left
		// -0.5, 0.5, 0.0, // top left
	}
	vertices = []float32{
		0.5, 0.5, 0.0, 0.0, 0.0, 1.0, // top right
		0.5, -0.5, 0.0, 1.0, 0.0, 0.0, // bottom right
		-0.5, -0.5, 0.0, 1.0, 1.0, 0.0, // bottom left
		-0.5, 0.5, 0.0, 1.0, 0.0, 1.0, // top left
	}
	indices = []uint16{ // note that we start from 0!
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
	debug     = true
	wireFrame = false
)

const floatSize = 4
const intSize = 2

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	var vertexBufferObject, vertexArrayObject, elementBufferObject uint32
	gl.GenVertexArrays(1, &vertexArrayObject)
	gl.GenBuffers(1, &vertexBufferObject)
	gl.GenBuffers(1, &elementBufferObject)

	gl.BindVertexArray(vertexArrayObject)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBufferObject)
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, elementBufferObject)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, intSize*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)

	// position data
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(6*floatSize), nil)
	gl.EnableVertexAttribArray(0)
	// color data
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(6*floatSize), gl.PtrOffset(3*floatSize))
	gl.EnableVertexAttribArray(1)

	// unbind the buffers here
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	for !window.ShouldClose() {
		draw(window, program, vertexArrayObject)
		processInput(window)
	}
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	// these versions match gl core imported above
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(800, 600, "Zarch (WIP)", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file %s with error %v", filename, err)
	}
	return string(data)
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	prog := gl.CreateProgram()

	// build shaders
	vertexShader, err := compileShader(readFile("vertex.glsl"), gl.VERTEX_SHADER)
	if err != nil {
		log.Fatal(err)
	}
	fragmentShader, err := compileShader(readFile("fragment.glsl"), gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatal(err)
	}
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)

	gl.LinkProgram(prog)

	if debug {
		var success int32
		gl.GetProgramiv(prog, gl.LINK_STATUS, &success)
		if success == gl.FALSE {
			var logLength int32
			gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
			logLines := strings.Repeat("\x00", int(logLength))
			gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(logLines))
			log.Fatalf("Failed to link program with error: %q", logLines)
		}
	}
	// once linked we can delete the shader objects
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return prog
}

func draw(window *glfw.Window, program uint32, vao uint32) {
	gl.ClearColor(0.2, 0.3, 0.3, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// timeValue := glfw.GetTime()
	// green := (math.Sin(timeValue) / 2.0) + 0.5
	// vertexColourLocation := gl.GetUniformLocation(program, gl.Str("ourColor\x00"))
	// gl.Uniform4f(vertexColourLocation, 0.0, float32(green), 0.0, 1.0)
	gl.UseProgram(program)
	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_SHORT, nil)

	glfw.PollEvents()
	window.SwapBuffers()
}

func processInput(window *glfw.Window) {
	if window.GetKey(glfw.KeyEscape) == glfw.Press {
		window.SetShouldClose(true)
	}
	if window.GetKey(glfw.KeyW) == glfw.Press {
		wireFrame = !wireFrame
		mode := uint32(gl.FILL)
		if wireFrame {
			mode = gl.LINE
		}
		gl.PolygonMode(gl.FRONT_AND_BACK, mode)
	}
}

func compileShader(code string, shaderType uint32) (uint32, error) {
	vertexShader := gl.CreateShader(shaderType)
	shaderSource, free := gl.Strs(code)
	gl.ShaderSource(vertexShader, 1, shaderSource, nil)
	free()
	gl.CompileShader(vertexShader)
	if debug {
		var status int32
		gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
			var logLength int32
			gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)

			log := strings.Repeat("\x00", int(logLength))
			gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
			return 0, fmt.Errorf("Failed to compile shader %q with error: %q", code, log)
		}
	}

	return vertexShader, nil
}
