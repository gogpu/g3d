// Copyright 2026 The gogpu Authors
// SPDX-License-Identifier: MIT

package g3d

import (
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	"github.com/gogpu/wgpu/hal"
	"github.com/gogpu/wgpu/hal/noop"
)

type countingQueue struct {
	noop.Queue
	submits int
}

func (q *countingQueue) Submit(commands []hal.CommandBuffer) (uint64, error) {
	q.submits++
	return q.Queue.Submit(commands)
}

func newRenderToFixture(t *testing.T) (*Renderer, *wgpu.Device, *wgpu.TextureView, *countingQueue) {
	t.Helper()

	rawQueue := &countingQueue{}
	device, err := wgpu.NewDeviceFromHAL(
		&noop.Device{}, rawQueue, 0, wgpu.DefaultLimits(), "g3d-render-to-test",
	)
	if err != nil {
		t.Fatalf("NewDeviceFromHAL: %v", err)
	}

	renderer, err := NewRendererFromDevice(device, device.Queue(), gputypes.TextureFormatBGRA8Unorm)
	if err != nil {
		device.Release()
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
		device.Release()
		t.Fatalf("CreateTexture: %v", err)
	}
	view, err := device.CreateTextureView(texture, nil)
	if err != nil {
		texture.Release()
		renderer.Release()
		device.Release()
		t.Fatalf("CreateTextureView: %v", err)
	}

	t.Cleanup(func() {
		view.Release()
		texture.Release()
		renderer.Release()
		device.Release()
	})
	return renderer, device, view, rawQueue
}

func TestRenderToRejectsNilEncoder(t *testing.T) {
	renderer, _, view, _ := newRenderToFixture(t)
	err := renderer.RenderTo(nil, NewScene(), NewPerspectiveCamera(60, 1, 0.1, 100), view)
	if err == nil {
		t.Fatal("RenderTo(nil encoder) succeeded, want error")
	}
}

func TestRenderToRecordsWithoutSubmitting(t *testing.T) {
	renderer, device, view, queue := newRenderToFixture(t)
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
	if queue.submits != 0 {
		t.Fatalf("submissions after RenderTo = %d, want 0", queue.submits)
	}

	commands, err := encoder.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if _, err := device.Queue().Submit(commands); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if queue.submits != 1 {
		t.Fatalf("submissions after caller submit = %d, want 1", queue.submits)
	}
}

func TestRenderStillOwnsOneSubmission(t *testing.T) {
	renderer, _, view, queue := newRenderToFixture(t)
	if err := renderer.Render(
		NewScene(), NewPerspectiveCamera(60, 1, 0.1, 100), view,
	); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if queue.submits != 1 {
		t.Fatalf("submissions after Render = %d, want 1", queue.submits)
	}
}
