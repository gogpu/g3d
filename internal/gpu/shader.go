// Package gpu manages wgpu pipeline state for the g3d forward renderer.
//
// This package operates on wgpu types and raw byte slices. It does NOT import
// the g3d root package — all data arrives as [16]float32 (matrices), []byte
// (uniform data), or gputypes constants. This prevents circular dependencies.
package gpu

import (
	_ "embed"
	"fmt"

	"github.com/gogpu/wgpu"
)

//go:embed shaders/basic.wgsl
var basicShaderSource string

//go:embed shaders/standard.wgsl
var standardShaderSource string

// Shader ID constants used for pipeline keying.
const (
	ShaderBasic    = "basic"
	ShaderStandard = "standard"
)

// ShaderCache compiles and caches WGSL shader modules. Modules are compiled
// once at creation time and reused for all pipeline creation and rendering.
type ShaderCache struct {
	device  *wgpu.Device
	modules map[string]*wgpu.ShaderModule
}

// NewShaderCache compiles the built-in shaders (basic, standard) and returns
// a cache. Returns an error if any shader fails to compile.
func NewShaderCache(device *wgpu.Device) (*ShaderCache, error) {
	c := &ShaderCache{
		device:  device,
		modules: make(map[string]*wgpu.ShaderModule, 2),
	}

	shaders := map[string]string{
		ShaderBasic:    basicShaderSource,
		ShaderStandard: standardShaderSource,
	}

	for name, source := range shaders {
		module, err := device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label: "g3d_" + name,
			WGSL:  source,
		})
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("compile shader %q: %w", name, err)
		}
		c.modules[name] = module
	}

	return c, nil
}

// Get returns the cached ShaderModule for the given name, or nil if not found.
func (c *ShaderCache) Get(name string) *wgpu.ShaderModule {
	return c.modules[name]
}

// Close releases all cached shader modules.
func (c *ShaderCache) Close() {
	for name, m := range c.modules {
		m.Release()
		delete(c.modules, name)
	}
}
