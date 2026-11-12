// Copyright 2026 The gogpu Authors
// SPDX-License-Identifier: MIT

package g3d

import (
	"math"
	"strings"
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	"github.com/gogpu/wgpu/wgputest"
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

func newGeometryDrawFixture(
	t *testing.T,
) (*Renderer, *wgpu.TextureView, *wgputest.Recorder) {
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
	return renderer, view, recorder
}

// bufferCreates counts the public buffer creations that carry the given label.
func bufferCreates(trace *wgputest.Trace, label string) int {
	n := 0
	for _, b := range trace.Buffers {
		if b.Label == label {
			n++
		}
	}
	return n
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

			trace := recorder.Trace()
			if len(trace.Draws) != 2 {
				t.Fatalf("non-indexed draws = %+v, want two", trace.Draws)
			}
			for i, draw := range trace.Draws {
				if draw.VertexCount != 3 || draw.InstanceCount != 1 ||
					draw.FirstVertex != 0 || draw.FirstInstance != 0 {
					t.Fatalf("draw %d = %+v, want vertexCount 3, instanceCount 1, firstVertex/firstInstance 0", i, draw)
				}
			}
			if len(trace.DrawsIndexed) != 0 {
				t.Fatalf("indexed draws = %+v, want none", trace.DrawsIndexed)
			}
			if got := len(trace.VertexBuffers); got != 2 {
				t.Errorf("vertex buffer binds = %d, want 2", got)
			}
			if got := len(trace.IndexBuffers); got != 0 {
				t.Errorf("index buffer binds = %d, want 0", got)
			}
			if got := bufferCreates(&trace, "g3d_vertex"); got != 1 {
				t.Errorf("vertex buffer creations = %d, want 1", got)
			}
			if got := bufferCreates(&trace, "g3d_index"); got != 0 {
				t.Errorf("index buffer creations = %d, want 0", got)
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
	trace := recorder.Trace()
	if len(trace.Draws) != 0 || len(trace.DrawsIndexed) != 0 {
		t.Fatalf("draws recorded for empty geometry: non-indexed=%+v indexed=%+v",
			trace.Draws, trace.DrawsIndexed)
	}
	if len(trace.VertexBuffers) != 0 || len(trace.IndexBuffers) != 0 {
		t.Fatalf("buffers bound for empty geometry: vertex=%d index=%d",
			len(trace.VertexBuffers), len(trace.IndexBuffers))
	}
	if v, i := bufferCreates(&trace, "g3d_vertex"), bufferCreates(&trace, "g3d_index"); v != 0 || i != 0 {
		t.Fatalf("buffers created for empty geometry: vertex=%d index=%d", v, i)
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
			trace := recorder.Trace()
			if len(trace.Draws) != 0 || len(trace.DrawsIndexed) != 0 {
				t.Fatalf("draws recorded for invalid count: non-indexed=%+v indexed=%+v",
					trace.Draws, trace.DrawsIndexed)
			}
			if len(trace.VertexBuffers) != 0 || len(trace.IndexBuffers) != 0 {
				t.Fatalf("buffers bound for invalid count: vertex=%d index=%d",
					len(trace.VertexBuffers), len(trace.IndexBuffers))
			}
			if v, i := bufferCreates(&trace, "g3d_vertex"), bufferCreates(&trace, "g3d_index"); v != 1 || i != 0 {
				t.Fatalf("buffer creations = vertex %d, index %d; want 1, 0", v, i)
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

	trace := recorder.Trace()
	if len(trace.DrawsIndexed) != 2 {
		t.Fatalf("indexed draws = %+v, want two", trace.DrawsIndexed)
	}
	for i, draw := range trace.DrawsIndexed {
		if draw.IndexCount != 3 || draw.InstanceCount != 1 ||
			draw.FirstIndex != 0 || draw.BaseVertex != 0 || draw.FirstInstance != 0 {
			t.Fatalf("indexed draw %d = %+v, want indexCount 3, instanceCount 1, firstIndex/baseVertex/firstInstance 0",
				i, draw)
		}
	}
	if len(trace.Draws) != 0 {
		t.Fatalf("non-indexed draws = %+v, want none", trace.Draws)
	}
	if len(trace.VertexBuffers) != 2 || len(trace.IndexBuffers) != 2 {
		t.Errorf("buffer binds = vertex %d, index %d; want 2 each",
			len(trace.VertexBuffers), len(trace.IndexBuffers))
	}
	if v, i := bufferCreates(&trace, "g3d_vertex"), bufferCreates(&trace, "g3d_index"); v != 1 || i != 1 {
		t.Errorf("buffer creations = vertex %d, index %d; want 1 each", v, i)
	}
	if len(renderer.geomIdxBufs) != 1 || len(renderer.geomIdxCounts) != 1 {
		t.Errorf("indexed geometry was not cached")
	}
}
