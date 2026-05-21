package gpu

import (
	"testing"

	"github.com/gogpu/gputypes"
)

func TestPipelineKeyEquality(t *testing.T) {
	a := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
		DoubleSided:   false,
	}
	b := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
		DoubleSided:   false,
	}

	if a != b {
		t.Errorf("identical PipelineKeys are not equal")
	}
}

func TestPipelineKeyDifferentShader(t *testing.T) {
	a := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}
	b := PipelineKey{
		ShaderID:      "standard",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}

	if a == b {
		t.Errorf("different shader PipelineKeys should not be equal")
	}
}

func TestPipelineKeyDifferentFormat(t *testing.T) {
	a := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}
	b := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatRGBA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}

	if a == b {
		t.Errorf("different format PipelineKeys should not be equal")
	}
}

func TestPipelineKeyDoubleSided(t *testing.T) {
	a := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
		DoubleSided:   false,
	}
	b := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
		DoubleSided:   true,
	}

	if a == b {
		t.Errorf("different DoubleSided PipelineKeys should not be equal")
	}
}

func TestPipelineKeyAsMapKey(t *testing.T) {
	m := make(map[PipelineKey]int)

	k1 := PipelineKey{
		ShaderID:      "basic",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}
	k2 := PipelineKey{
		ShaderID:      "standard",
		SurfaceFormat: gputypes.TextureFormatBGRA8Unorm,
		DepthFormat:   gputypes.TextureFormatDepth24Plus,
	}

	m[k1] = 1
	m[k2] = 2

	if m[k1] != 1 {
		t.Errorf("map[k1] = %d, want 1", m[k1])
	}
	if m[k2] != 2 {
		t.Errorf("map[k2] = %d, want 2", m[k2])
	}
	if len(m) != 2 {
		t.Errorf("map len = %d, want 2", len(m))
	}

	// Same key should overwrite.
	m[k1] = 3
	if m[k1] != 3 {
		t.Errorf("map[k1] after overwrite = %d, want 3", m[k1])
	}
	if len(m) != 2 {
		t.Errorf("map len after overwrite = %d, want 2", len(m))
	}
}
