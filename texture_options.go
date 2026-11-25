package g3d

import "github.com/gogpu/gputypes"

// TextureOption configures texture properties via functional options.
type TextureOption interface {
	apply(t *Texture)
}

type magFilterOption struct{ filter gputypes.FilterMode }

func (m magFilterOption) apply(t *Texture) { t.magFilter = m.filter }

// WithMagFilter sets magnification filter.
func WithMagFilter(filter gputypes.FilterMode) TextureOption { return magFilterOption{filter} }

type minFilterOption struct{ filter gputypes.FilterMode }

func (m minFilterOption) apply(t *Texture) { t.magFilter = m.filter }

// WithMinFilter sets minification filter.
func WithMinFilter(filter gputypes.FilterMode) TextureOption { return magFilterOption{filter} }

type addressMode struct{ uMode, vMode gputypes.AddressMode }

func (a addressMode) apply(t *Texture) { t.addressModeU, t.addressModeV = a.uMode, a.vMode }

// WithUVAddressMode sets UV address mode.
func WithUVAddressMode(u, v gputypes.AddressMode) TextureOption {
	return addressMode{uMode: u, vMode: v}
}
