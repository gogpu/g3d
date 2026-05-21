package render

import "sort"

// DrawCall represents a single object to be drawn. Sort keys are extracted by
// the Renderer from g3d public types; the render package never imports g3d.
type DrawCall struct {
	// PipelineKey identifies the render pipeline (derived from Material.ShaderID).
	// Pipeline state changes are the most expensive GPU operation, so this is
	// the primary sort key for opaque draw calls.
	PipelineKey string

	// MaterialID uniquely identifies a material instance. Bind group changes are
	// the second-most expensive GPU operation (after pipeline switches), so this
	// is the secondary sort key for opaque draw calls.
	MaterialID uint64

	// Distance is the squared distance from the camera to the object's center.
	// Used as the tertiary sort key: front-to-back for opaque (early Z rejection),
	// back-to-front for transparent (correct alpha compositing).
	Distance float32

	// Bucket determines which render pass draws this call.
	Bucket RenderBucket

	// MeshIndex is an opaque handle that the Renderer uses to look up the
	// associated mesh, geometry, material, and world transform. The render
	// package does not interpret this value.
	MeshIndex int
}

// RenderList collects and sorts visible meshes for rendering. It maintains
// separate slices per bucket so that sorting strategies can differ.
//
// Phase 1 uses Opaque and Transparent. Background and Transmissive slices
// are allocated lazily when the first draw call targets them.
//
// All slices are reused across frames via Clear() to avoid GC pressure.
type RenderList struct {
	background   []DrawCall
	opaque       []DrawCall
	transmissive []DrawCall
	transparent  []DrawCall
}

// NewRenderList creates an empty RenderList ready for draw call collection.
func NewRenderList() *RenderList {
	return &RenderList{}
}

// Clear resets all buckets for the next frame. Slice capacity is preserved
// to avoid allocation on subsequent frames (zero-alloc render path).
func (rl *RenderList) Clear() {
	rl.background = rl.background[:0]
	rl.opaque = rl.opaque[:0]
	rl.transmissive = rl.transmissive[:0]
	rl.transparent = rl.transparent[:0]
}

// Add inserts a draw call into the appropriate bucket based on dc.Bucket.
func (rl *RenderList) Add(dc DrawCall) {
	switch dc.Bucket {
	case BucketBackground:
		rl.background = append(rl.background, dc)
	case BucketOpaque:
		rl.opaque = append(rl.opaque, dc)
	case BucketTransmissive:
		rl.transmissive = append(rl.transmissive, dc)
	case BucketTransparent:
		rl.transparent = append(rl.transparent, dc)
	}
}

// Sort sorts each bucket by its strategy.
//
// Opaque uses a 3-key sort validated against Three.js PR #15484:
//  1. PipelineKey ascending — minimize pipeline state changes (most expensive)
//  2. MaterialID ascending — minimize bind group changes (second most expensive)
//  3. Distance ascending — front-to-back for early Z rejection (minimize overdraw)
//
// Transparent uses distance descending (back-to-front) for correct alpha blending.
//
// Background and Transmissive use insertion order (no sort) in Phase 1.
func (rl *RenderList) Sort() {
	sortOpaque(rl.opaque)
	sortTransparent(rl.transparent)
}

// Opaque returns the sorted opaque draw calls. The returned slice must not
// be modified by the caller; it is valid until the next Clear() or Add().
func (rl *RenderList) Opaque() []DrawCall {
	return rl.opaque
}

// Transparent returns the sorted transparent draw calls. The returned slice
// must not be modified by the caller; it is valid until the next Clear() or Add().
func (rl *RenderList) Transparent() []DrawCall {
	return rl.transparent
}

// Background returns the background draw calls. Valid until the next Clear().
func (rl *RenderList) Background() []DrawCall {
	return rl.background
}

// Transmissive returns the transmissive draw calls. Valid until the next Clear().
func (rl *RenderList) Transmissive() []DrawCall {
	return rl.transmissive
}

// TotalCount returns the total number of draw calls across all buckets.
func (rl *RenderList) TotalCount() int {
	return len(rl.background) + len(rl.opaque) + len(rl.transmissive) + len(rl.transparent)
}

// sortOpaque sorts opaque draw calls using the 3-key strategy:
// PipelineKey (asc) > MaterialID (asc) > Distance (asc, front-to-back).
//
// sort.SliceStable is used so that draw calls with identical keys maintain
// their insertion order (deterministic rendering).
func sortOpaque(calls []DrawCall) {
	sort.SliceStable(calls, func(i, j int) bool {
		a, b := &calls[i], &calls[j]

		// Key 1: pipeline (shader) — most expensive state change.
		if a.PipelineKey != b.PipelineKey {
			return a.PipelineKey < b.PipelineKey
		}

		// Key 2: material instance — bind group change.
		if a.MaterialID != b.MaterialID {
			return a.MaterialID < b.MaterialID
		}

		// Key 3: distance — front-to-back for early Z rejection.
		return a.Distance < b.Distance
	})
}

// sortTransparent sorts transparent draw calls by distance descending
// (back-to-front) for correct alpha compositing. Stable sort preserves
// insertion order for equal distances.
func sortTransparent(calls []DrawCall) {
	sort.SliceStable(calls, func(i, j int) bool {
		return calls[i].Distance > calls[j].Distance
	})
}
