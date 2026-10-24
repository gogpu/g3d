package g3d

import (
	"fmt"
	"image"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// TODO: docs.
type Texture struct {
	texture *wgpu.Texture
	view    *wgpu.TextureView
	sampler *wgpu.Sampler
	width   int
	height  int
	format  gputypes.TextureFormat
}

func (t *Texture) Release() {
	t.texture.Release()
	t.view.Release()
	t.sampler.Release()
}

// TextureOptions configures texture creation.
type TextureOptions struct {
	// Label for debugging (optional)
	Label string

	// Filter mode for magnification (default: Linear)
	MagFilter gputypes.FilterMode

	// Filter mode for minification (default: Linear)
	MinFilter gputypes.FilterMode

	// Address mode for U coordinate (default: ClampToEdge)
	AddressModeU gputypes.AddressMode

	// Address mode for V coordinate (default: ClampToEdge)
	AddressModeV gputypes.AddressMode
}

// DefaultTextureOptions returns sensible defaults for texture creation.
func DefaultTextureOptions() TextureOptions {
	return TextureOptions{
		MagFilter:    gputypes.FilterModeLinear,
		MinFilter:    gputypes.FilterModeLinear,
		AddressModeU: gputypes.AddressModeClampToEdge,
		AddressModeV: gputypes.AddressModeClampToEdge,
	}
}

func (r *Renderer) createTexture(device *wgpu.Device, data []byte, opts TextureOptions, width, height int) (*Texture, error) {
	expectedSize := width * height * 4
	if len(data) != expectedSize {
		return nil, fmt.Errorf("g3d: invalid data size: expected %d bytes, got %d", expectedSize, len(data))
	}

	texture, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label: opts.Label,
		Size: wgpu.Extent3D{
			Width:              uint32(width),  //nolint:gosec // G115: width validated positive above
			Height:             uint32(height), //nolint:gosec // G115: height validated positive above
			DepthOrArrayLayers: 1,
		},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        gputypes.TextureFormatRGBA8Unorm,
		Usage:         gputypes.TextureUsageTextureBinding | gputypes.TextureUsageCopyDst,
	})

	if err != nil {
		return nil, fmt.Errorf("g3d: failed to create texture: %w", err)
	}

	// Upload pixel data via wgpu Queue
	if err := device.Queue().WriteTexture(
		&wgpu.ImageCopyTexture{
			Texture:  texture,
			MipLevel: 0,
			Origin:   wgpu.Origin3D{X: 0, Y: 0, Z: 0},
			Aspect:   gputypes.TextureAspectAll,
		},
		data,
		&wgpu.ImageDataLayout{
			Offset:       0,
			BytesPerRow:  uint32(width * 4), //nolint:gosec // G115: width validated positive above
			RowsPerImage: uint32(height),    //nolint:gosec // G115: height validated positive above
		},
		&wgpu.Extent3D{
			Width:              uint32(width),  //nolint:gosec // G115: width validated positive above
			Height:             uint32(height), //nolint:gosec // G115: height validated positive above
			DepthOrArrayLayers: 1,
		},
	); err != nil {
		texture.Release()
		return nil, fmt.Errorf("g3d: failed to upload texture data: %w", err)
	}

	// Create texture view
	view, err := device.CreateTextureView(texture, nil)
	if err != nil {
		texture.Release()
		return nil, fmt.Errorf("g3d: failed to create texture view: %w", err)
	}

	// Create sampler
	sampler, err := device.CreateSampler(&wgpu.SamplerDescriptor{
		Label:        opts.Label,
		AddressModeU: opts.AddressModeU,
		AddressModeV: opts.AddressModeV,
		AddressModeW: gputypes.AddressModeClampToEdge,
		MagFilter:    opts.MagFilter,
		MinFilter:    opts.MinFilter,
		MipmapFilter: gputypes.FilterModeNearest,
		LodMinClamp:  0,
		LodMaxClamp:  32,
	})
	if err != nil {
		view.Release()
		texture.Release()
		return nil, fmt.Errorf("g3d: failed to create sampler: %w", err)
	}

	return &Texture{
		texture: texture,
		view:    view,
		sampler: sampler,
		width:   width,
		height:  height,
		format:  gputypes.TextureFormatRGBA8Unorm,
	}, nil
}

// AddTextureAsImage adds texture as image.
func (r *Renderer) AddTextureAsImage(img *image.RGBA, id string, smooth bool) error {
	if _, has := r.textures[id]; has {
		return fmt.Errorf("texture with id: %s is already loaded", id)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	opts := DefaultTextureOptions()
	if !smooth {
		opts.MagFilter = gputypes.FilterModeNearest
		opts.MinFilter = gputypes.FilterModeNearest
	}

	texture, err := r.createTexture(r.gpuState.Device(), img.Pix, opts, width, height)
	if err != nil {
		return err
	}

	r.textures[id] = texture

	return nil
}

// UnloadTexture unloads the texture by its id.
func (r *Renderer) UnloadTexture(id string) {
	texture, ok := r.textures[id]
	if ok {
		texture.Release()
	}
	delete(r.textures, id)
}

// UnloadAllTextures unloads all textures.
func (r *Renderer) UnloadAllTextures() {
	for _, tex := range r.textures {
		tex.Release()
	}
	r.textures = make(map[string]*Texture)
}

func (r *Renderer) textureById(id string) (*Texture, error) {
	texture, ok := r.textures[id]
	if !ok {
		return nil, fmt.Errorf("texture with id: %s does not exist", id)
	}
	return texture, nil
}
