// Copyright 2026 The gogpu Authors
// SPDX-License-Identifier: MIT

//go:build !rust && !(js && wasm) && !android

package g3d

import (
	"context"
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	_ "github.com/gogpu/wgpu/hal/software"
)

// softwareProvider is a gpucontext.DeviceProvider over wgpu's software
// backend, the way gogpu hands its device to g3d. The adapter reports
// DeviceTypeCPU, like lavapipe or llvmpipe do (g3d#47).
type softwareProvider struct {
	adapter *wgpu.Adapter
	device  *wgpu.Device
	format  gputypes.TextureFormat
}

func (p *softwareProvider) Device() gpucontext.Device { return wgpu.DeviceToHandle(p.device) }
func (p *softwareProvider) Queue() gpucontext.Queue   { return wgpu.QueueToHandle(p.device.Queue()) }
func (p *softwareProvider) SurfaceFormat() gputypes.TextureFormat {
	return p.format
}
func (p *softwareProvider) Adapter() gpucontext.Adapter { return wgpu.AdapterToHandle(p.adapter) }
func (p *softwareProvider) AdapterInfo() gpucontext.AdapterInfo {
	return gpucontext.AdapterInfo{Name: p.adapter.Info().Name, Type: gpucontext.AdapterTypeSoftware}
}
func (p *softwareProvider) Features() gputypes.Features { return 0 }
func (p *softwareProvider) DownlevelCapabilities() gputypes.DownlevelCapabilities {
	return gputypes.DownlevelCapabilities{}
}

func newSoftwareProvider(t *testing.T, format gputypes.TextureFormat) *softwareProvider {
	t.Helper()

	instance, err := wgpu.CreateInstance(nil)
	if err != nil {
		t.Fatalf("CreateInstance: %v", err)
	}
	t.Cleanup(instance.Release)

	adapter, err := instance.RequestAdapter(&wgpu.RequestAdapterOptions{ForceFallbackAdapter: true})
	if err != nil {
		t.Fatalf("RequestAdapter(ForceFallbackAdapter): %v", err)
	}
	t.Cleanup(adapter.Release)
	if got := adapter.Info().DeviceType; got != gputypes.DeviceTypeCPU {
		t.Fatalf("adapter device type = %v, want DeviceTypeCPU", got)
	}

	device, err := adapter.RequestDevice(nil)
	if err != nil {
		t.Fatalf("RequestDevice: %v", err)
	}
	t.Cleanup(device.Release)

	return &softwareProvider{adapter: adapter, device: device, format: format}
}

// TestNewRendererAcceptsCPUAdapter: g3d renders on whatever device it is given;
// choosing a hardware adapter is the application's call (g3d#47).
func TestNewRendererAcceptsCPUAdapter(t *testing.T) {
	provider := newSoftwareProvider(t, gputypes.TextureFormatRGBA8Unorm)

	renderer, err := NewRenderer(provider)
	if err != nil {
		t.Fatalf("NewRenderer with a CPU adapter: %v", err)
	}
	renderer.Release()
}

// TestRenderSphereOnSoftwareBackend renders a lit sphere end to end on a CPU
// adapter and reads the pixels back (g3d#47). The light shines along the view
// direction, so the front of the sphere faces it and the center pixel is lit.
//
// On a GPU this also catches an inside-out sphere (g3d#46): back-face culling
// leaves the inner far wall, which faces away from the light, and the center
// pixel is black (checked on Vulkan, Intel Iris Xe). wgpu's software backend
// does not apply CullMode yet, so here the depth test keeps the near surface
// either way; TestGenerateSphere_WindingOrder covers the winding.
func TestRenderSphereOnSoftwareBackend(t *testing.T) {
	const size uint32 = 64 // 64 px × 4 B = 256 B per row, the copy alignment

	provider := newSoftwareProvider(t, gputypes.TextureFormatRGBA8Unorm)
	renderer, err := NewRenderer(provider)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	defer renderer.Release()
	renderer.SetSize(size, size)

	scene := NewScene()
	scene.SetBackground(Black)
	sun := NewDirectionalLight(WithLightColor(White), WithLightIntensity(1)) // default direction (0, 0, -1)
	scene.Add(sun.LightNode())
	sphere := NewMesh(
		NewSphereGeometry(1, WithSegments(32, 16)),
		NewStandardMaterial(WithColor(White), WithRoughness(0.8)),
	)
	scene.Add(sphere.MeshNode())

	camera := NewPerspectiveCamera(60, 1, 0.1, 100)
	camera.CameraNode().SetPosition(Vec3{X: 0, Y: 0, Z: 3})

	device := provider.device
	target, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "g3d-software-target",
		Size:          wgpu.Extent3D{Width: size, Height: size, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension2D,
		Format:        wgpu.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageCopySrc,
	})
	if err != nil {
		t.Fatalf("CreateTexture: %v", err)
	}
	defer target.Release()
	view, err := device.CreateTextureView(target, nil)
	if err != nil {
		t.Fatalf("CreateTextureView: %v", err)
	}
	defer view.Release()

	if err := renderer.Render(scene, camera, view); err != nil {
		t.Fatalf("Render: %v", err)
	}

	pixels := readbackRGBA8(t, device, target, size)
	at := func(x, y uint32) [4]byte {
		i := (y*size + x) * 4
		return [4]byte{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}
	}

	if corner := at(0, 0); corner[0] > 8 || corner[1] > 8 || corner[2] > 8 {
		t.Errorf("corner pixel = %v, want the black background", corner)
	}
	if center := at(size/2, size/2); center[0] < 128 || center[1] < 128 || center[2] < 128 {
		t.Errorf("center pixel = %v, want the lit front of the sphere (each channel >= 128)", center)
	}
}

func readbackRGBA8(t *testing.T, device *wgpu.Device, src *wgpu.Texture, size uint32) []byte {
	t.Helper()

	bytesPerRow := size * 4
	bufSize := uint64(bytesPerRow) * uint64(size)
	staging, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "g3d-software-readback",
		Size:  bufSize,
		Usage: wgpu.BufferUsageMapRead | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		t.Fatalf("CreateBuffer: %v", err)
	}
	defer staging.Release()

	enc, err := device.CreateCommandEncoder(nil)
	if err != nil {
		t.Fatalf("CreateCommandEncoder: %v", err)
	}
	enc.TransitionTextures([]wgpu.TextureBarrier{{
		Texture: src,
		Usage: wgpu.TextureUsageTransition{
			OldUsage: wgpu.TextureUsageRenderAttachment,
			NewUsage: wgpu.TextureUsageCopySrc,
		},
	}})
	enc.CopyTextureToBuffer(src, staging, []wgpu.BufferTextureCopy{{
		BufferLayout: wgpu.ImageDataLayout{BytesPerRow: bytesPerRow, RowsPerImage: size},
		TextureBase:  wgpu.ImageCopyTexture{Texture: src},
		Size:         wgpu.Extent3D{Width: size, Height: size, DepthOrArrayLayers: 1},
	}})
	cmd, err := enc.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if _, err := device.Queue().Submit(cmd); err != nil {
		t.Fatalf("Submit: %v", err)
	}

	if err := staging.Map(context.Background(), wgpu.MapModeRead, 0, bufSize); err != nil {
		t.Fatalf("Map: %v", err)
	}
	rng, err := staging.MappedRange(0, bufSize)
	if err != nil {
		t.Fatalf("MappedRange: %v", err)
	}
	pixels := append([]byte(nil), rng.Bytes()...)
	if err := staging.Unmap(); err != nil {
		t.Fatalf("Unmap: %v", err)
	}
	return pixels
}
