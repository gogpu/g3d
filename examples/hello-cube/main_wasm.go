//go:build js && wasm

package main

import (
	"log"
	"math"
	"syscall/js"
	"time"
	"unsafe"

	"github.com/gogpu/g3d"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

type browserProvider struct {
	device  *wgpu.Device
	adapter *wgpu.Adapter
	format  gputypes.TextureFormat
}

func (p *browserProvider) Device() gpucontext.Device {
	return gpucontext.NewDevice(unsafe.Pointer(p.device))
}

func (p *browserProvider) Queue() gpucontext.Queue {
	return gpucontext.NewQueue(unsafe.Pointer(p.device.Queue()))
}

func (p *browserProvider) SurfaceFormat() gputypes.TextureFormat { return p.format }

func (p *browserProvider) Adapter() gpucontext.Adapter {
	return gpucontext.NewAdapter(unsafe.Pointer(p.adapter))
}

func (p *browserProvider) AdapterInfo() gpucontext.AdapterInfo {
	return gpucontext.AdapterInfo{Type: gpucontext.AdapterTypeUnknown}
}

func getDimensions() (width, height uint32) {
	win := js.Global().Get("window")
	dpr := win.Get("devicePixelRatio").Float()
	if dpr <= 0 {
		dpr = 1
	}

	cssWidth := win.Get("innerWidth").Float()
	cssHeight := win.Get("innerHeight").Float()
	if cssWidth <= 0 || cssHeight <= 0 {
		cssWidth = 800
		cssHeight = 600
	}

	pixelWidth := uint32(math.Round(cssWidth * dpr))
	pixelHeight := uint32(math.Round(cssHeight * dpr))
	if pixelWidth <= 0 || pixelHeight <= 0 {
		pixelWidth = 800
		pixelHeight = 600
	}
	return pixelWidth, pixelHeight
}

func updateCanvasSize(canvas js.Value, renderer *g3d.Renderer, camera *g3d.PerspectiveCamera) {
	pixelWidth, pixelHeight := getDimensions()

	canvas.Set("width", pixelWidth)
	canvas.Set("height", pixelHeight)
	camera.SetAspect(float32(pixelWidth) / float32(pixelHeight))
	renderer.SetSize(pixelWidth, pixelHeight)
}

func renderFrame(surface *wgpu.Surface, renderer *g3d.Renderer, scene *g3d.Scene, camera *g3d.PerspectiveCamera) {
	texture, _, err := surface.GetCurrentTexture()
	if err != nil {
		log.Printf("hello-cube: acquire surface texture: %v", err)
		return
	}

	view, err := texture.CreateView(nil)
	if err != nil {
		log.Printf("hello-cube: create view: %v", err)
		return
	}

	if err := renderer.Render(scene, camera, view); err != nil {
		log.Printf("hello-cube: render error: %v", err)
	}
}

func main() {
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	if canvas.IsUndefined() || canvas.IsNull() {
		log.Fatal("hello-cube: canvas element not found")
	}

	instance, err := wgpu.CreateInstance(nil)
	if err != nil {
		log.Fatal(err)
	}
	surface, err := instance.CreateSurfaceFromCanvas(canvas)
	if err != nil {
		log.Fatal(err)
	}
	adapter, err := instance.RequestAdapter(&wgpu.RequestAdapterOptions{CompatibleSurface: surface})
	if err != nil {
		log.Fatal(err)
	}
	device, err := adapter.RequestDevice(nil)
	if err != nil {
		log.Fatal(err)
	}

	caps := adapter.GetSurfaceCapabilities(surface)
	if len(caps.Formats) == 0 || len(caps.AlphaModes) == 0 {
		log.Fatal("hello-cube: WebGPU surface has no supported configuration")
	}
	format := caps.Formats[0]
	width, height := getDimensions()
	if err := surface.Configure(device, &wgpu.SurfaceConfiguration{
		Width: width, Height: height, Format: format,
		Usage:       wgpu.TextureUsageRenderAttachment,
		PresentMode: wgpu.PresentModeFifo,
		AlphaMode:   caps.AlphaModes[0],
	}); err != nil {
		log.Fatal(err)
	}

	provider := &browserProvider{device: device, adapter: adapter, format: format}
	renderer, err := g3d.NewRenderer(provider)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Release()

	scene, camera, cube := buildScene()
	updateCanvasSize(canvas, renderer, camera)

	resize := js.FuncOf(func(this js.Value, args []js.Value) any {
		updateCanvasSize(canvas, renderer, camera)
		return nil
	})
	defer resize.Release()
	js.Global().Call("addEventListener", "resize", resize)

	last := time.Now()
	var frame js.Func
	frame = js.FuncOf(func(js.Value, []js.Value) any {
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now
		if dt > 0.1 {
			dt = 1.0 / 60.0
		}

		updateScene(cube, dt)
		renderFrame(surface, renderer, scene, camera)

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
	select {}
}
