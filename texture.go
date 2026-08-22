package g3d

import (
	"fmt"
	"image"
	"os"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gputypes"
)

// LoadTexture loads a texture from a file path. Supports PNG and JPEG formats.
func (r *Renderer) LoadTexture(ctx *gogpu.Context, path, id string, smooth bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	return r.AddTextureAsImage(ctx, img, id, smooth)
}

// AddTextureAsImage adds texture as image.
func (r *Renderer) AddTextureAsImage(ctx *gogpu.Context, img image.Image, id string, smooth bool) error {
	if _, has := r.textures[id]; has {
		return fmt.Errorf("texture with id: %s is already loaded", id)
	}

	options := gogpu.DefaultTextureOptions()
	if !smooth {
		options.MagFilter = gputypes.FilterModeNearest
		options.MinFilter = gputypes.FilterModeNearest
	}

	texture, err := ctx.Renderer().NewTextureFromImageWithOptions(img, options)
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
		texture.Handle().Release()
		texture.View().Release()
	}
	delete(r.textures, id)
}

// UnloadAllTextures unloads all textures.
func (r *Renderer) UnloadAllTextures() {
	for _, tex := range r.textures {
		tex.Handle().Release()
		tex.View().Release()
	}
	r.textures = make(map[string]*gogpu.Texture)
}

func (r *Renderer) textureById(id string) (*gogpu.Texture, error) {
	texture, ok := r.textures[id]
	if !ok {
		return nil, fmt.Errorf("texture with id: %s does not exist", id)
	}
	return texture, nil
}
