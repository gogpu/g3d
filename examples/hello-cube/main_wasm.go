//go:build js && wasm

package main

import (
	"log"
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
	const width, height = 800, 600
	canvas.Set("width", width)
	canvas.Set("height", height)
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
	camera.SetAspect(float32(width) / float32(height))
	renderer.SetSize(width, height)

	last := time.Now()
	var frame js.Func
	frame = js.FuncOf(func(js.Value, []js.Value) any {
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now
		if dt > 0.1 {
			dt = 1.0 / 60.0
		}

		rotation := cube.MeshNode().Rotation
		rotation.Y += float32(dt)
		rotation.X += float32(dt) * 0.3
		cube.MeshNode().SetRotation(rotation)

		texture, _, err := surface.GetCurrentTexture()
		if err == nil {
			view, viewErr := texture.CreateView(nil)
			if viewErr == nil {
				if renderErr := renderer.Render(scene, camera, view); renderErr != nil {
					log.Printf("hello-cube: render error: %v", renderErr)
				}
			} else {
				log.Printf("hello-cube: create view: %v", viewErr)
			}
		} else {
			log.Printf("hello-cube: acquire surface texture: %v", err)
		}

		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
	select {}
}
