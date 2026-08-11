// Copyright 2026 The gogpu Authors
// SPDX-License-Identifier: MIT

package g3d

import (
	"math"
	"strings"
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	"github.com/gogpu/wgpu/hal"
	"github.com/gogpu/wgpu/hal/noop"
)

type testGeometry struct {
	vertices    []float32
	indices     []uint32
	vertexCount int
}

func (g *testGeometry) Vertices() []float32 { return g.vertices }

func (g *testGeometry) Indices() []uint32 { return g.indices }

func (g *testGeometry) VertexCount() int { return g.vertexCount }

func (g *testGeometry) BoundingBox() AABB { return computeAABB(g.vertices) }

type geometryDrawCall struct {
	vertexCount   uint32
	instanceCount uint32
	firstVertex   uint32
	firstInstance uint32
}

type geometryIndexedDrawCall struct {
	indexCount    uint32
	instanceCount uint32
	firstIndex    uint32
	baseVertex    int32
	firstInstance uint32
}

type geometryDrawRecorder struct {
	vertexBufferBindings int
	indexBufferBindings  int
	vertexBufferCreates  int
	indexBufferCreates   int
	draws                []geometryDrawCall
	indexedDraws         []geometryIndexedDrawCall
}

type geometryRecordingDevice struct {
	noop.Device
	recorder *geometryDrawRecorder
}

func (d *geometryRecordingDevice) CreateBuffer(desc *hal.BufferDescriptor) (hal.Buffer, error) {
	switch desc.Label {
	case "g3d_vertex":
		d.recorder.vertexBufferCreates++
	case "g3d_index":
		d.recorder.indexBufferCreates++
	}
	return d.Device.CreateBuffer(desc)
}

func (d *geometryRecordingDevice) CreateCommandEncoder(
	_ *hal.CommandEncoderDescriptor,
) (hal.CommandEncoder, error) {
	return &geometryRecordingCommandEncoder{recorder: d.recorder}, nil
}

type geometryRecordingCommandEncoder struct {
	noop.CommandEncoder
	recorder *geometryDrawRecorder
}

func (e *geometryRecordingCommandEncoder) BeginRenderPass(
	_ *hal.RenderPassDescriptor,
) hal.RenderPassEncoder {
	return &geometryRecordingRenderPass{recorder: e.recorder}
}

type geometryRecordingRenderPass struct {
	noop.RenderPassEncoder
	recorder *geometryDrawRecorder
}

func (p *geometryRecordingRenderPass) SetVertexBuffer(_ uint32, _ hal.Buffer, _ uint64) {
	p.recorder.vertexBufferBindings++
}

func (p *geometryRecordingRenderPass) SetIndexBuffer(
	_ hal.Buffer,
	_ gputypes.IndexFormat,
	_ uint64,
) {
	p.recorder.indexBufferBindings++
}

func (p *geometryRecordingRenderPass) Draw(
	vertexCount, instanceCount, firstVertex, firstInstance uint32,
) {
	p.recorder.draws = append(p.recorder.draws, geometryDrawCall{
		vertexCount:   vertexCount,
		instanceCount: instanceCount,
		firstVertex:   firstVertex,
		firstInstance: firstInstance,
	})
}

func (p *geometryRecordingRenderPass) DrawIndexed(
	indexCount, instanceCount, firstIndex uint32,
	baseVertex int32,
	firstInstance uint32,
) {
	p.recorder.indexedDraws = append(p.recorder.indexedDraws, geometryIndexedDrawCall{
		indexCount:    indexCount,
		instanceCount: instanceCount,
		firstIndex:    firstIndex,
		baseVertex:    baseVertex,
		firstInstance: firstInstance,
	})
}

func newGeometryDrawFixture(
	t *testing.T,
) (*Renderer, *wgpu.TextureView, *geometryDrawRecorder) {
	t.Helper()

	recorder := &geometryDrawRecorder{}
	rawDevice := &geometryRecordingDevice{recorder: recorder}
	device, err := wgpu.NewDeviceFromHAL(
		rawDevice, &noop.Queue{}, 0, wgpu.DefaultLimits(), "g3d-geometry-draw-test",
	)
	if err != nil {
		t.Fatalf("NewDeviceFromHAL: %v", err)
	}

	renderer, err := NewRendererFromDevice(
		device, device.Queue(), gputypes.TextureFormatBGRA8Unorm,
	)
	if err != nil {
		device.Release()
		t.Fatalf("NewRendererFromDevice: %v", err)
	}
	renderer.SetSize(16, 16)

	texture, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "g3d-geometry-draw-target",
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
	return renderer, view, recorder
}

func TestRendererDrawsNonIndexedGeometry(t *testing.T) {
	triangleVertices := []float32{
		-0.5, -0.5, 0, 0, 0, 1, 0, 0,
		0.5, -0.5, 0, 0, 0, 1, 1, 0,
		0, 0.5, 0, 0, 0, 1, 0.5, 1,
	}

	for _, tc := range []struct {
		name    string
		indices []uint32
	}{
		{name: "nil indices"},
		{name: "empty indices", indices: []uint32{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			renderer, target, recorder := newGeometryDrawFixture(t)
			geom := &testGeometry{vertices: triangleVertices, indices: tc.indices, vertexCount: 3}
			scene := NewScene()
			scene.Add(NewMesh(geom, NewBasicMaterial()).MeshNode())
			camera := NewPerspectiveCamera(60, 1, 0.1, 100)
			camera.CameraNode().SetPosition(Vec3{Z: 3})

			if err := renderer.Render(scene, camera, target); err != nil {
				t.Fatalf("Render: %v", err)
			}
			if err := renderer.Render(scene, camera, target); err != nil {
				t.Fatalf("second Render: %v", err)
			}

			wantDraw := geometryDrawCall{
				vertexCount: 3, instanceCount: 1, firstVertex: 0, firstInstance: 0,
			}
			if len(recorder.draws) != 2 || recorder.draws[0] != wantDraw || recorder.draws[1] != wantDraw {
				t.Fatalf("non-indexed draws = %+v, want two %+v calls", recorder.draws, wantDraw)
			}
			if len(recorder.indexedDraws) != 0 {
				t.Fatalf("indexed draws = %+v, want none", recorder.indexedDraws)
			}
			if recorder.vertexBufferBindings != 2 {
				t.Errorf("vertex buffer bindings = %d, want 2", recorder.vertexBufferBindings)
			}
			if recorder.indexBufferBindings != 0 {
				t.Errorf("index buffer bindings = %d, want 0", recorder.indexBufferBindings)
			}
			if recorder.vertexBufferCreates != 1 {
				t.Errorf("vertex buffer creations = %d, want 1", recorder.vertexBufferCreates)
			}
			if recorder.indexBufferCreates != 0 {
				t.Errorf("index buffer creations = %d, want 0", recorder.indexBufferCreates)
			}
			if len(renderer.geomIdxBufs) != 0 || len(renderer.geomIdxCounts) != 0 {
				t.Errorf("index caches populated for non-indexed geometry")
			}
		})
	}
}

func TestRendererRejectsNonIndexedGeometryWithoutVertices(t *testing.T) {
	renderer, target, recorder := newGeometryDrawFixture(t)
	scene := NewScene()
	scene.Add(NewMesh(&testGeometry{}, NewBasicMaterial()).MeshNode())
	camera := NewPerspectiveCamera(60, 1, 0.1, 100)
	camera.CameraNode().SetPosition(Vec3{Z: 3})

	err := renderer.Render(scene, camera, target)
	if err == nil || !strings.Contains(err.Error(), "geometry has no vertices") {
		t.Fatalf("Render error = %v, want geometry has no vertices", err)
	}
	if len(recorder.draws) != 0 || len(recorder.indexedDraws) != 0 {
		t.Fatalf("draws recorded for empty geometry: non-indexed=%+v indexed=%+v",
			recorder.draws, recorder.indexedDraws)
	}
	if recorder.vertexBufferBindings != 0 || recorder.indexBufferBindings != 0 {
		t.Fatalf("buffers bound for empty geometry: vertex=%d index=%d",
			recorder.vertexBufferBindings, recorder.indexBufferBindings)
	}
	if recorder.vertexBufferCreates != 0 || recorder.indexBufferCreates != 0 {
		t.Fatalf("buffers created for empty geometry: vertex=%d index=%d",
			recorder.vertexBufferCreates, recorder.indexBufferCreates)
	}
}

func TestRendererRejectsInvalidNonIndexedVertexCount(t *testing.T) {
	triangleVertices := []float32{
		-0.5, -0.5, 0, 0, 0, 1, 0, 0,
		0.5, -0.5, 0, 0, 0, 1, 1, 0,
		0, 0.5, 0, 0, 0, 1, 0.5, 1,
	}
	invalidCounts := []struct {
		name  string
		count int
	}{
		{name: "negative", count: -1},
	}
	tooLarge := uint64(math.MaxUint32) + 1
	if uint64(math.MaxInt) > math.MaxUint32 {
		invalidCounts = append(invalidCounts, struct {
			name  string
			count int
		}{name: "overflow", count: int(tooLarge)})
	}

	for _, tc := range invalidCounts {
		t.Run(tc.name, func(t *testing.T) {
			renderer, target, recorder := newGeometryDrawFixture(t)
			geom := &testGeometry{vertices: triangleVertices, vertexCount: tc.count}
			scene := NewScene()
			scene.Add(NewMesh(geom, NewBasicMaterial()).MeshNode())
			camera := NewPerspectiveCamera(60, 1, 0.1, 100)
			camera.CameraNode().SetPosition(Vec3{Z: 3})

			err := renderer.Render(scene, camera, target)
			if err == nil || !strings.Contains(err.Error(), "invalid vertex count") {
				t.Fatalf("Render error = %v, want invalid vertex count", err)
			}
			if len(recorder.draws) != 0 || len(recorder.indexedDraws) != 0 {
				t.Fatalf("draws recorded for invalid count: non-indexed=%+v indexed=%+v",
					recorder.draws, recorder.indexedDraws)
			}
			if recorder.vertexBufferBindings != 0 || recorder.indexBufferBindings != 0 {
				t.Fatalf("buffers bound for invalid count: vertex=%d index=%d",
					recorder.vertexBufferBindings, recorder.indexBufferBindings)
			}
			if recorder.vertexBufferCreates != 1 || recorder.indexBufferCreates != 0 {
				t.Fatalf("buffer creations = vertex %d, index %d; want 1, 0",
					recorder.vertexBufferCreates, recorder.indexBufferCreates)
			}
		})
	}
}

func TestRendererPreservesIndexedGeometryDraw(t *testing.T) {
	triangleVertices := []float32{
		-0.5, -0.5, 0, 0, 0, 1, 0, 0,
		0.5, -0.5, 0, 0, 0, 1, 1, 0,
		0, 0.5, 0, 0, 0, 1, 0.5, 1,
	}
	renderer, target, recorder := newGeometryDrawFixture(t)
	geom := NewBufferGeometry(triangleVertices, []uint32{0, 1, 2})
	scene := NewScene()
	scene.Add(NewMesh(geom, NewBasicMaterial()).MeshNode())
	camera := NewPerspectiveCamera(60, 1, 0.1, 100)
	camera.CameraNode().SetPosition(Vec3{Z: 3})

	if err := renderer.Render(scene, camera, target); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if err := renderer.Render(scene, camera, target); err != nil {
		t.Fatalf("second Render: %v", err)
	}

	wantDraw := geometryIndexedDrawCall{
		indexCount: 3, instanceCount: 1, firstIndex: 0, baseVertex: 0, firstInstance: 0,
	}
	if len(recorder.indexedDraws) != 2 || recorder.indexedDraws[0] != wantDraw || recorder.indexedDraws[1] != wantDraw {
		t.Fatalf("indexed draws = %+v, want two %+v calls", recorder.indexedDraws, wantDraw)
	}
	if len(recorder.draws) != 0 {
		t.Fatalf("non-indexed draws = %+v, want none", recorder.draws)
	}
	if recorder.vertexBufferBindings != 2 || recorder.indexBufferBindings != 2 {
		t.Errorf("buffer bindings = vertex %d, index %d; want 2 each",
			recorder.vertexBufferBindings, recorder.indexBufferBindings)
	}
	if recorder.vertexBufferCreates != 1 || recorder.indexBufferCreates != 1 {
		t.Errorf("buffer creations = vertex %d, index %d; want 1 each",
			recorder.vertexBufferCreates, recorder.indexBufferCreates)
	}
	if len(renderer.geomIdxBufs) != 1 || len(renderer.geomIdxCounts) != 1 {
		t.Errorf("indexed geometry was not cached")
	}
}
