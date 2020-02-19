package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	vertices = []float32{
		// positions      // colors         // texture coords
		+0.5, +0.5, +0.0, +1.0, +0.0, +0.0, +1.0, +1.0, // top right
		+0.5, -0.5, +0.0, +0.0, +1.0, +0.0, +1.0, +0.0, // bottom right
		-0.5, -0.5, +0.0, +0.0, +0.0, +1.0, +0.0, +0.0, // bottom left
		-0.5, +0.5, +0.0, +1.0, +1.0, +0.0, +0.0, +1.0, // top left
	}
	indices = []uint16{
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
	debug     = true
	wireFrame = false
)

const floatSize = 4
const intSize = 2

func main() {
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	imgfile, err := os.Open("tron-grid.jpg")
	if err != nil {
		log.Fatalf("Failed to open %s with error: %v", "container.jpg", err)
	}
	defer imgfile.Close()

	img, _, err := image.Decode(imgfile)
	i := image.NewRGBA(img.Bounds())
	draw.Draw(i, img.Bounds(), img, img.Bounds().Min, draw.Src)

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
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(8*floatSize), nil)
	gl.EnableVertexAttribArray(0)
	// color data
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(8*floatSize), gl.PtrOffset(3*floatSize))
	gl.EnableVertexAttribArray(1)
	// texture coords
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(8*floatSize), gl.PtrOffset(6*floatSize))
	gl.EnableVertexAttribArray(2)

	// texture stuff
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// 	float borderColor[] = { 1.0f, 1.0f, 0.0f, 1.0f };
	//  glTexParameterfv(GL_TEXTURE_2D, GL_TEXTURE_BORDER_COLOR, borderColor);
	// texture texl interpolation
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// load texture data
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(i.Bounds().Max.X), int32(i.Bounds().Max.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(i.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// unbind the buffers here
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	for !window.ShouldClose() {
		drawScene(window, program, vertexArrayObject, texture)
		processInput(window)
	}
	gl.DeleteVertexArrays(1, &vertexArrayObject)
	gl.DeleteBuffers(1, &vertexBufferObject)
	gl.DeleteBuffers(1, &elementBufferObject)
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
	vertexShader, err := compileShader(readFile("vertex-tex.glsl"), gl.VERTEX_SHADER)
	if err != nil {
		log.Fatal(err)
	}
	fragmentShader, err := compileShader(readFile("fragment-tex.glsl"), gl.FRAGMENT_SHADER)
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

func drawScene(window *glfw.Window, program uint32, vao uint32, texture uint32) {
	gl.ClearColor(0.2, 0.3, 0.3, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// timeValue := glfw.GetTime()
	// green := (math.Sin(timeValue) / 2.0) + 0.5
	// vertexColourLocation := gl.GetUniformLocation(program, gl.Str("ourColor\x00"))
	// gl.Uniform4f(vertexColourLocation, 0.0, float32(green), 0.0, 1.0)
	gl.UseProgram(program)
	gl.BindTexture(gl.TEXTURE_2D, texture)
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
	shader := gl.CreateShader(shaderType)
	shaderSource, free := gl.Strs(code)
	gl.ShaderSource(shader, 1, shaderSource, nil)
	free()
	gl.CompileShader(shader)
	if debug {
		var status int32
		gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
			var logLength int32
			gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

			log := strings.Repeat("\x00", int(logLength))
			gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
			return 0, fmt.Errorf("Failed to compile shader %q with error: %q", code, log)
		}
	}

	return shader, nil
}
