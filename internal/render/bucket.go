// Package render manages the draw call list and bucket-based sorting for the
// g3d forward renderer. It operates on primitives only and has no dependency
// on the root g3d package.
package render

// RenderBucket determines render order. Four buckets follow the Three.js
// validated pattern (see PHASE1-ARCHITECTURE-VALIDATION.md, correction C3).
//
// Phase 1 uses Opaque and Transparent only. Background and Transmissive
// buckets exist for forward-compatibility but are not rendered until Phase 4.
type RenderBucket uint8

const (
	// BucketBackground is drawn first. Used for skyboxes and image-based lighting.
	BucketBackground RenderBucket = iota

	// BucketOpaque is drawn second with depth testing and no blending.
	// Sorted front-to-back within pipeline/material groups to minimize overdraw.
	BucketOpaque

	// BucketTransmissive is drawn third. Used for transparent objects that still
	// write to the depth buffer (e.g., refraction effects). Added in Phase 4.
	BucketTransmissive

	// BucketTransparent is drawn last with alpha blending enabled.
	// Sorted back-to-front for correct alpha compositing.
	BucketTransparent
)

// String returns a human-readable name for the bucket.
func (b RenderBucket) String() string {
	switch b {
	case BucketBackground:
		return "Background"
	case BucketOpaque:
		return "Opaque"
	case BucketTransmissive:
		return "Transmissive"
	case BucketTransparent:
		return "Transparent"
	default:
		return "Unknown"
	}
}
