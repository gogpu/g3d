package g3d

import (
	"testing"

	"github.com/gogpu/g3d/internal/render"
)

func TestMapBucketOpaque(t *testing.T) {
	got := mapBucket(RenderBucketOpaque)
	if got != render.BucketOpaque {
		t.Errorf("mapBucket(Opaque) = %v, want BucketOpaque", got)
	}
}

func TestMapBucketTransparent(t *testing.T) {
	got := mapBucket(RenderBucketTransparent)
	if got != render.BucketTransparent {
		t.Errorf("mapBucket(Transparent) = %v, want BucketTransparent", got)
	}
}

func TestMapBucketBackground(t *testing.T) {
	got := mapBucket(RenderBucketBackground)
	if got != render.BucketBackground {
		t.Errorf("mapBucket(Background) = %v, want BucketBackground", got)
	}
}

func TestMapBucketTransmissive(t *testing.T) {
	got := mapBucket(RenderBucketTransmissive)
	if got != render.BucketTransmissive {
		t.Errorf("mapBucket(Transmissive) = %v, want BucketTransmissive", got)
	}
}

func TestMapBucketDefault(t *testing.T) {
	// Unknown bucket should map to opaque.
	got := mapBucket(RenderBucket(99))
	if got != render.BucketOpaque {
		t.Errorf("mapBucket(99) = %v, want BucketOpaque", got)
	}
}

func TestMeshFromNodeWithMesh(t *testing.T) {
	mesh := NewMesh(NewBoxGeometry(1, 1, 1), NewBasicMaterial())
	got := meshFromNode(mesh.MeshNode())
	if got != mesh {
		t.Error("meshFromNode should return the mesh stored in UserData")
	}
}

func TestMeshFromNodeWithoutMesh(t *testing.T) {
	node := NewNode()
	got := meshFromNode(node)
	if got != nil {
		t.Error("meshFromNode should return nil for a plain node")
	}
}

func TestLightFromNodeWithDirectionalLight(t *testing.T) {
	light := NewDirectionalLight()
	got := lightFromNode(light.LightNode())
	if got == nil {
		t.Error("lightFromNode should return the light stored in UserData")
	}
}

func TestLightFromNodeWithoutLight(t *testing.T) {
	node := NewNode()
	got := lightFromNode(node)
	if got != nil {
		t.Error("lightFromNode should return nil for a plain node")
	}
}

func TestMaterialIDDifferent(t *testing.T) {
	m1 := NewBasicMaterial()
	m2 := NewBasicMaterial()

	id1 := materialID(m1)
	id2 := materialID(m2)

	if id1 == id2 {
		t.Error("different material instances should have different IDs")
	}
}

func TestMaterialIDStable(t *testing.T) {
	m := NewBasicMaterial()
	id1 := materialID(m)
	id2 := materialID(m)

	if id1 != id2 {
		t.Error("same material instance should have the same ID across calls")
	}
}

func TestNewRendererFromDeviceNilDevice(t *testing.T) {
	_, err := NewRendererFromDevice(nil, nil, 0)
	if err == nil {
		t.Error("expected error for nil device")
	}
}
