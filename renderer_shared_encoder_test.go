// Copyright 2026 The gogpu Authors
// SPDX-License-Identifier: MIT

package g3d

import (
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	"github.com/gogpu/wgpu/wgputest"
)

func newRenderToFixture(
	t *testing.T,
) (*Renderer, *wgpu.Device, *wgpu.TextureView, *wgputest.Recorder) {
	t.Helper()

	recorder := wgputest.NewRecorder()
	device, queue, cleanup, err := wgputest.NewNoopDevice(wgputest.WithRecorder(recorder))
	if err != nil {
		t.Fatalf("wgputest.NewNoopDevice: %v", err)
	}

	renderer, err := NewRendererFromDevice(device, queue, gputypes.TextureFormatBGRA8Unorm)
	if err != nil {
		cleanup()
		t.Fatalf("NewRendererFromDevice: %v", err)
	}
	renderer.SetSize(16, 16)

	texture, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "g3d-render-to-target",
		Size:          wgpu.Extent3D{Width: 16, Height: 16, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        gputypes.TextureFormatBGRA8Unorm,
		Usage:         gputypes.TextureUsageRenderAttachment,
	})
	if err != nil {
		renderer.Release()
		cleanup()
		t.Fatalf("CreateTexture: %v", err)
	}
	view, err := device.CreateTextureView(texture, nil)
	if err != nil {
		texture.Release()
		renderer.Release()
		cleanup()
		t.Fatalf("CreateTextureView: %v", err)
	}

	t.Cleanup(func() {
		view.Release()
		texture.Release()
		renderer.Release()
		cleanup()
	})
	return renderer, device, view, recorder
}

// submits counts the public Queue.Submit calls recorded so far.
func submits(recorder *wgputest.Recorder) int {
	return len(recorder.Trace().Submits)
}

func TestRenderToRejectsNilEncoder(t *testing.T) {
	renderer, _, view, _ := newRenderToFixture(t)
	err := renderer.RenderTo(nil, NewScene(), NewPerspectiveCamera(60, 1, 0.1, 100), view)
	if err == nil {
		t.Fatal("RenderTo(nil encoder) succeeded, want error")
	}
}

func TestRenderToRecordsWithoutSubmitting(t *testing.T) {
	renderer, device, view, recorder := newRenderToFixture(t)
	encoder, err := device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "shared-frame"})
	if err != nil {
		t.Fatalf("CreateCommandEncoder: %v", err)
	}

	scene := NewScene()
	camera := NewPerspectiveCamera(60, 1, 0.1, 100)
	if err := renderer.RenderTo(encoder, scene, camera, view); err != nil {
		encoder.DiscardEncoding()
		t.Fatalf("RenderTo: %v", err)
	}
	if got := submits(recorder); got != 0 {
		t.Fatalf("submissions after RenderTo = %d, want 0", got)
	}

	commands, err := encoder.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if _, err := device.Queue().Submit(commands); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if got := submits(recorder); got != 1 {
		t.Fatalf("submissions after caller submit = %d, want 1", got)
	}
}

func TestRenderStillOwnsOneSubmission(t *testing.T) {
	renderer, _, view, recorder := newRenderToFixture(t)
	if err := renderer.Render(
		NewScene(), NewPerspectiveCamera(60, 1, 0.1, 100), view,
	); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got := submits(recorder); got != 1 {
		t.Fatalf("submissions after Render = %d, want 1", got)
	}
}
