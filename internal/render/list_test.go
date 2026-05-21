package render

import (
	"testing"
)

// --- RenderBucket tests ---

func TestRenderBucketString(t *testing.T) {
	tests := []struct {
		bucket RenderBucket
		want   string
	}{
		{BucketBackground, "Background"},
		{BucketOpaque, "Opaque"},
		{BucketTransmissive, "Transmissive"},
		{BucketTransparent, "Transparent"},
		{RenderBucket(99), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.bucket.String()
		if got != tt.want {
			t.Errorf("RenderBucket(%d).String() = %q, want %q", tt.bucket, got, tt.want)
		}
	}
}

func TestRenderBucketOrder(t *testing.T) {
	// Verify the enum values match expected draw order.
	if BucketBackground != 0 {
		t.Errorf("BucketBackground = %d, want 0", BucketBackground)
	}
	if BucketOpaque != 1 {
		t.Errorf("BucketOpaque = %d, want 1", BucketOpaque)
	}
	if BucketTransmissive != 2 {
		t.Errorf("BucketTransmissive = %d, want 2", BucketTransmissive)
	}
	if BucketTransparent != 3 {
		t.Errorf("BucketTransparent = %d, want 3", BucketTransparent)
	}
}

// --- NewRenderList tests ---

func TestNewRenderList(t *testing.T) {
	rl := NewRenderList()
	if rl == nil {
		t.Fatal("NewRenderList() returned nil")
	}
	if rl.TotalCount() != 0 {
		t.Errorf("new list TotalCount() = %d, want 0", rl.TotalCount())
	}
}

// --- Add routing tests ---

func TestAddRoutesToCorrectBucket(t *testing.T) {
	rl := NewRenderList()

	rl.Add(DrawCall{Bucket: BucketBackground, MeshIndex: 0})
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 2})
	rl.Add(DrawCall{Bucket: BucketTransmissive, MeshIndex: 3})
	rl.Add(DrawCall{Bucket: BucketTransparent, MeshIndex: 4})
	rl.Add(DrawCall{Bucket: BucketTransparent, MeshIndex: 5})

	if got := len(rl.Background()); got != 1 {
		t.Errorf("Background len = %d, want 1", got)
	}
	if got := len(rl.Opaque()); got != 2 {
		t.Errorf("Opaque len = %d, want 2", got)
	}
	if got := len(rl.Transmissive()); got != 1 {
		t.Errorf("Transmissive len = %d, want 1", got)
	}
	if got := len(rl.Transparent()); got != 2 {
		t.Errorf("Transparent len = %d, want 2", got)
	}
	if got := rl.TotalCount(); got != 6 {
		t.Errorf("TotalCount() = %d, want 6", got)
	}
}

// --- Clear tests ---

func TestClearPreservesCapacity(t *testing.T) {
	rl := NewRenderList()

	// Add enough to force allocation.
	for i := 0; i < 100; i++ {
		rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: i})
		rl.Add(DrawCall{Bucket: BucketTransparent, MeshIndex: i})
	}

	opaqueCap := cap(rl.opaque)
	transparentCap := cap(rl.transparent)

	rl.Clear()

	if rl.TotalCount() != 0 {
		t.Errorf("after Clear(), TotalCount() = %d, want 0", rl.TotalCount())
	}
	if cap(rl.opaque) != opaqueCap {
		t.Errorf("opaque capacity changed: before=%d, after=%d", opaqueCap, cap(rl.opaque))
	}
	if cap(rl.transparent) != transparentCap {
		t.Errorf("transparent capacity changed: before=%d, after=%d", transparentCap, cap(rl.transparent))
	}
}

func TestClearThenAdd(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 1})
	rl.Clear()
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 2})

	opaque := rl.Opaque()
	if len(opaque) != 1 {
		t.Fatalf("after Clear+Add, Opaque len = %d, want 1", len(opaque))
	}
	if opaque[0].MeshIndex != 2 {
		t.Errorf("after Clear+Add, MeshIndex = %d, want 2", opaque[0].MeshIndex)
	}
}

// --- Empty list safety ---

func TestEmptyListIsSafe(t *testing.T) {
	rl := NewRenderList()

	// Sort, access, and count on an empty list must not panic.
	rl.Sort()

	if len(rl.Background()) != 0 {
		t.Error("empty Background not empty")
	}
	if len(rl.Opaque()) != 0 {
		t.Error("empty Opaque not empty")
	}
	if len(rl.Transmissive()) != 0 {
		t.Error("empty Transmissive not empty")
	}
	if len(rl.Transparent()) != 0 {
		t.Error("empty Transparent not empty")
	}
	if rl.TotalCount() != 0 {
		t.Errorf("empty TotalCount() = %d, want 0", rl.TotalCount())
	}
}

// --- Opaque 3-key sort tests ---

func TestOpaqueSortByPipelineKey(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 5, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 5, Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 3, Bucket: BucketOpaque, MeshIndex: 2})
	rl.Sort()

	opaque := rl.Opaque()
	if len(opaque) != 3 {
		t.Fatalf("Opaque len = %d, want 3", len(opaque))
	}
	// "basic" < "standard" — basic must come first.
	if opaque[0].PipelineKey != "basic" {
		t.Errorf("opaque[0].PipelineKey = %q, want 'basic'", opaque[0].PipelineKey)
	}
	if opaque[1].PipelineKey != "standard" {
		t.Errorf("opaque[1].PipelineKey = %q, want 'standard'", opaque[1].PipelineKey)
	}
}

func TestOpaqueSortByMaterialID(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 42, Distance: 5, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 10, Distance: 5, Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 30, Distance: 5, Bucket: BucketOpaque, MeshIndex: 2})
	rl.Sort()

	opaque := rl.Opaque()
	if opaque[0].MaterialID != 10 {
		t.Errorf("opaque[0].MaterialID = %d, want 10", opaque[0].MaterialID)
	}
	if opaque[1].MaterialID != 30 {
		t.Errorf("opaque[1].MaterialID = %d, want 30", opaque[1].MaterialID)
	}
	if opaque[2].MaterialID != 42 {
		t.Errorf("opaque[2].MaterialID = %d, want 42", opaque[2].MaterialID)
	}
}

func TestOpaqueSortByDistanceFrontToBack(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 10.0, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 2.5, Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 2})
	rl.Sort()

	opaque := rl.Opaque()
	if opaque[0].Distance != 2.5 {
		t.Errorf("opaque[0].Distance = %f, want 2.5", opaque[0].Distance)
	}
	if opaque[1].Distance != 5.0 {
		t.Errorf("opaque[1].Distance = %f, want 5.0", opaque[1].Distance)
	}
	if opaque[2].Distance != 10.0 {
		t.Errorf("opaque[2].Distance = %f, want 10.0", opaque[2].Distance)
	}
}

func TestOpaque3KeySortOrder(t *testing.T) {
	// Full integration: pipeline > material > distance.
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 2, Distance: 1.0, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 3.0, Bucket: BucketOpaque, MeshIndex: 2})
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 2.0, Bucket: BucketOpaque, MeshIndex: 3})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 8.0, Bucket: BucketOpaque, MeshIndex: 4})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 2, Distance: 6.0, Bucket: BucketOpaque, MeshIndex: 5})
	rl.Sort()

	opaque := rl.Opaque()
	if len(opaque) != 6 {
		t.Fatalf("Opaque len = %d, want 6", len(opaque))
	}

	// Expected order:
	// 0: basic/1/2.0  (MeshIndex 3)
	// 1: basic/1/5.0  (MeshIndex 1)
	// 2: standard/1/3.0 (MeshIndex 2)
	// 3: standard/1/8.0 (MeshIndex 4)
	// 4: standard/2/1.0 (MeshIndex 0)
	// 5: standard/2/6.0 (MeshIndex 5)
	expected := []struct {
		pipelineKey string
		materialID  uint64
		distance    float32
		meshIndex   int
	}{
		{"basic", 1, 2.0, 3},
		{"basic", 1, 5.0, 1},
		{"standard", 1, 3.0, 2},
		{"standard", 1, 8.0, 4},
		{"standard", 2, 1.0, 0},
		{"standard", 2, 6.0, 5},
	}

	for i, e := range expected {
		dc := opaque[i]
		if dc.PipelineKey != e.pipelineKey || dc.MaterialID != e.materialID ||
			dc.Distance != e.distance || dc.MeshIndex != e.meshIndex {
			t.Errorf("opaque[%d] = {%q, %d, %.1f, %d}, want {%q, %d, %.1f, %d}",
				i, dc.PipelineKey, dc.MaterialID, dc.Distance, dc.MeshIndex,
				e.pipelineKey, e.materialID, e.distance, e.meshIndex)
		}
	}
}

// --- Transparent sort tests ---

func TestTransparentSortBackToFront(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{Distance: 2.0, Bucket: BucketTransparent, MeshIndex: 0})
	rl.Add(DrawCall{Distance: 10.0, Bucket: BucketTransparent, MeshIndex: 1})
	rl.Add(DrawCall{Distance: 5.0, Bucket: BucketTransparent, MeshIndex: 2})
	rl.Sort()

	tr := rl.Transparent()
	if len(tr) != 3 {
		t.Fatalf("Transparent len = %d, want 3", len(tr))
	}
	// Back-to-front: 10.0, 5.0, 2.0
	if tr[0].Distance != 10.0 {
		t.Errorf("transparent[0].Distance = %f, want 10.0", tr[0].Distance)
	}
	if tr[1].Distance != 5.0 {
		t.Errorf("transparent[1].Distance = %f, want 5.0", tr[1].Distance)
	}
	if tr[2].Distance != 2.0 {
		t.Errorf("transparent[2].Distance = %f, want 2.0", tr[2].Distance)
	}
}

// --- Mixed bucket independence ---

func TestMixedBucketsDontInterfere(t *testing.T) {
	rl := NewRenderList()

	// Add opaque and transparent interleaved.
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 10.0, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{Distance: 2.0, Bucket: BucketTransparent, MeshIndex: 1})
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 2})
	rl.Add(DrawCall{Distance: 8.0, Bucket: BucketTransparent, MeshIndex: 3})
	rl.Add(DrawCall{Bucket: BucketBackground, MeshIndex: 4})

	rl.Sort()

	// Opaque: basic/5.0 before standard/10.0 (pipeline key sort).
	opaque := rl.Opaque()
	if len(opaque) != 2 {
		t.Fatalf("Opaque len = %d, want 2", len(opaque))
	}
	if opaque[0].PipelineKey != "basic" {
		t.Errorf("opaque[0].PipelineKey = %q, want 'basic'", opaque[0].PipelineKey)
	}

	// Transparent: 8.0 before 2.0 (back-to-front).
	tr := rl.Transparent()
	if len(tr) != 2 {
		t.Fatalf("Transparent len = %d, want 2", len(tr))
	}
	if tr[0].Distance != 8.0 {
		t.Errorf("transparent[0].Distance = %f, want 8.0", tr[0].Distance)
	}

	// Background: exactly 1, unchanged.
	bg := rl.Background()
	if len(bg) != 1 {
		t.Fatalf("Background len = %d, want 1", len(bg))
	}
	if bg[0].MeshIndex != 4 {
		t.Errorf("background[0].MeshIndex = %d, want 4", bg[0].MeshIndex)
	}

	if rl.TotalCount() != 5 {
		t.Errorf("TotalCount() = %d, want 5", rl.TotalCount())
	}
}

// --- Sort stability ---

func TestSortStabilityWithinSameKey(t *testing.T) {
	rl := NewRenderList()

	// Three opaque calls with identical sort keys — insertion order must be preserved.
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 10})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 20})
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 30})
	rl.Sort()

	opaque := rl.Opaque()
	if opaque[0].MeshIndex != 10 || opaque[1].MeshIndex != 20 || opaque[2].MeshIndex != 30 {
		t.Errorf("stable sort broken: got [%d, %d, %d], want [10, 20, 30]",
			opaque[0].MeshIndex, opaque[1].MeshIndex, opaque[2].MeshIndex)
	}
}

func TestTransparentSortStability(t *testing.T) {
	rl := NewRenderList()

	// Three transparent calls with same distance — insertion order preserved.
	rl.Add(DrawCall{Distance: 5.0, Bucket: BucketTransparent, MeshIndex: 10})
	rl.Add(DrawCall{Distance: 5.0, Bucket: BucketTransparent, MeshIndex: 20})
	rl.Add(DrawCall{Distance: 5.0, Bucket: BucketTransparent, MeshIndex: 30})
	rl.Sort()

	tr := rl.Transparent()
	if tr[0].MeshIndex != 10 || tr[1].MeshIndex != 20 || tr[2].MeshIndex != 30 {
		t.Errorf("transparent stable sort broken: got [%d, %d, %d], want [10, 20, 30]",
			tr[0].MeshIndex, tr[1].MeshIndex, tr[2].MeshIndex)
	}
}

// --- TotalCount accuracy ---

func TestTotalCountAccuracy(t *testing.T) {
	rl := NewRenderList()

	if rl.TotalCount() != 0 {
		t.Errorf("empty TotalCount() = %d, want 0", rl.TotalCount())
	}

	rl.Add(DrawCall{Bucket: BucketBackground, MeshIndex: 0})
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{Bucket: BucketTransmissive, MeshIndex: 2})
	rl.Add(DrawCall{Bucket: BucketTransparent, MeshIndex: 3})
	if rl.TotalCount() != 4 {
		t.Errorf("after 4 adds, TotalCount() = %d, want 4", rl.TotalCount())
	}

	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 4})
	if rl.TotalCount() != 5 {
		t.Errorf("after 5 adds, TotalCount() = %d, want 5", rl.TotalCount())
	}

	rl.Clear()
	if rl.TotalCount() != 0 {
		t.Errorf("after Clear(), TotalCount() = %d, want 0", rl.TotalCount())
	}
}

// --- Single element edge cases ---

func TestSingleElementOpaque(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 3.0, Bucket: BucketOpaque, MeshIndex: 7})
	rl.Sort()

	opaque := rl.Opaque()
	if len(opaque) != 1 {
		t.Fatalf("Opaque len = %d, want 1", len(opaque))
	}
	if opaque[0].MeshIndex != 7 {
		t.Errorf("opaque[0].MeshIndex = %d, want 7", opaque[0].MeshIndex)
	}
}

func TestSingleElementTransparent(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{Distance: 5.0, Bucket: BucketTransparent, MeshIndex: 9})
	rl.Sort()

	tr := rl.Transparent()
	if len(tr) != 1 {
		t.Fatalf("Transparent len = %d, want 1", len(tr))
	}
	if tr[0].MeshIndex != 9 {
		t.Errorf("transparent[0].MeshIndex = %d, want 9", tr[0].MeshIndex)
	}
}

// --- Multiple Sort calls (idempotent) ---

func TestSortIsIdempotent(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{PipelineKey: "standard", MaterialID: 2, Distance: 1.0, Bucket: BucketOpaque, MeshIndex: 0})
	rl.Add(DrawCall{PipelineKey: "basic", MaterialID: 1, Distance: 5.0, Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{Distance: 3.0, Bucket: BucketTransparent, MeshIndex: 2})
	rl.Add(DrawCall{Distance: 9.0, Bucket: BucketTransparent, MeshIndex: 3})

	rl.Sort()

	// Capture first sort result.
	firstOpaque0 := rl.Opaque()[0].MeshIndex
	firstTr0 := rl.Transparent()[0].MeshIndex

	// Sort again — same result.
	rl.Sort()
	if rl.Opaque()[0].MeshIndex != firstOpaque0 {
		t.Error("opaque sort is not idempotent")
	}
	if rl.Transparent()[0].MeshIndex != firstTr0 {
		t.Error("transparent sort is not idempotent")
	}
}

// --- Clear resets all four buckets ---

func TestClearResetsAllBuckets(t *testing.T) {
	rl := NewRenderList()
	rl.Add(DrawCall{Bucket: BucketBackground, MeshIndex: 0})
	rl.Add(DrawCall{Bucket: BucketOpaque, MeshIndex: 1})
	rl.Add(DrawCall{Bucket: BucketTransmissive, MeshIndex: 2})
	rl.Add(DrawCall{Bucket: BucketTransparent, MeshIndex: 3})

	rl.Clear()

	if len(rl.Background()) != 0 {
		t.Error("Background not cleared")
	}
	if len(rl.Opaque()) != 0 {
		t.Error("Opaque not cleared")
	}
	if len(rl.Transmissive()) != 0 {
		t.Error("Transmissive not cleared")
	}
	if len(rl.Transparent()) != 0 {
		t.Error("Transparent not cleared")
	}
}
