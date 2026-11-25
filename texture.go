package g3d

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// NewTexture creates new Texture with the given options.
// Defaults: magFilter=gputypes.FilterModeLinear, minFilter=gputypes.FilterModeLinear,
// addressModeU=gputypes.AddressModeClampToEdge, addressModeV=gputypes.AddressModeClampToEdge.
func NewTexture(img image.Image, options ...TextureOption) *Texture {
	var (
		bounds = img.Bounds()
		rgba   *image.NRGBA
		ok     bool
	)
	if rgba, ok = img.(*image.NRGBA); !ok {
		rgba = image.NewNRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	}

	tex := &Texture{
		data:         rgba.Pix,
		width:        bounds.Dy(),
		height:       bounds.Dx(),
		magFilter:    gputypes.FilterModeLinear,
		minFilter:    gputypes.FilterModeLinear,
		addressModeU: gputypes.AddressModeClampToEdge,
		addressModeV: gputypes.AddressModeClampToEdge,
	}

	for _, opt := range options {
		opt.apply(tex)
	}

	return tex
}

// Texture represents texture that can be passed as parameter for a material.
type Texture struct {
	data          []byte
	width, height int

	magFilter    gputypes.FilterMode
	minFilter    gputypes.FilterMode
	addressModeU gputypes.AddressMode
	addressModeV gputypes.AddressMode
}

// Dimensions returns dimensions of the texture.
func (t *Texture) Dimensions() (width, height int) {
	return t.width, t.height
}

// AddressMode returns address mode of the texture.
func (t *Texture) AddressMode() (u, v gputypes.AddressMode) {
	return t.addressModeU, t.addressModeV
}

// textureDescriptor holds information used by gpu to display texture.
type textureDescriptor struct {
	texture *wgpu.Texture
	view    *wgpu.TextureView
	sampler *wgpu.Sampler
	width   int
	height  int
	format  gputypes.TextureFormat
}

// Release ...
func (t *textureDescriptor) Release() {
	t.texture.Release()
	t.view.Release()
	t.sampler.Release()
}

// getTextureDescriptor returns existing texture descriptor or creates new.
func (r *Renderer) getTextureDescriptor(device *wgpu.Device, texture *Texture) (*textureDescriptor, error) {
	if texture == nil {
		return r.emptyTexture, nil
	}

	var (
		descriptor *textureDescriptor
		err        error
		ok         bool
	)

	descriptor, ok = r.textures[texture]
	if !ok {
		descriptor, err = r.createTextureDescriptor(device, texture)
		if err != nil {
			return nil, err
		}
		r.textures[texture] = descriptor
	}
	return descriptor, err
}

// createTextureDescriptor creates new texture descriptor.
func (r *Renderer) createTextureDescriptor(device *wgpu.Device, tex *Texture) (*textureDescriptor, error) {
	if len(tex.data) == 0 {
		return nil, fmt.Errorf("g3d: invalid data size: expected at least %d bytes, got %d", 4, len(tex.data))
	}

	expectedSize := tex.width * tex.height * 4
	if len(tex.data) != expectedSize {
		return nil, fmt.Errorf("g3d: invalid data size: expected %d bytes, got %d", expectedSize, len(tex.data))
	}

	texture, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Size: wgpu.Extent3D{
			Width:              uint32(tex.width),  //nolint:gosec // G115: width validated positive above
			Height:             uint32(tex.height), //nolint:gosec // G115: height validated positive above
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
		tex.data,
		&wgpu.ImageDataLayout{
			Offset:       0,
			BytesPerRow:  uint32(tex.width * 4), //nolint:gosec // G115: width validated positive above
			RowsPerImage: uint32(tex.height),    //nolint:gosec // G115: height validated positive above
		},
		&wgpu.Extent3D{
			Width:              uint32(tex.width),  //nolint:gosec // G115: width validated positive above
			Height:             uint32(tex.height), //nolint:gosec // G115: height validated positive above
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
		AddressModeU: tex.addressModeU,
		AddressModeV: tex.addressModeV,
		AddressModeW: gputypes.AddressModeClampToEdge,
		MagFilter:    tex.magFilter,
		MinFilter:    tex.minFilter,
		MipmapFilter: gputypes.FilterModeNearest,
		LodMinClamp:  0,
		LodMaxClamp:  32,
	})
	if err != nil {
		view.Release()
		texture.Release()
		return nil, fmt.Errorf("g3d: failed to create sampler: %w", err)
	}

	return &textureDescriptor{
		texture: texture,
		view:    view,
		sampler: sampler,
		width:   tex.width,
		height:  tex.height,
		format:  gputypes.TextureFormatRGBA8UnormSrgb,
	}, nil
}
