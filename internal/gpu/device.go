package gpu

import (
	"fmt"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

// GPUState manages the wgpu device and associated GPU caches (shaders, pipelines).
// It supports two initialization modes:
//
//  1. External device — injected via SetDeviceProvider (used when g3d runs inside gogpu).
//  2. Standalone — created lazily via EnsureDevice (used for headless/test scenarios).
//
// The design follows the gg/internal/gpu/gpu_shared.go pattern exactly.
type GPUState struct {
	device   *wgpu.Device
	queue    *wgpu.Queue
	instance *wgpu.Instance // nil when using external device
	external bool           // true = device owned externally, don't release on Close

	shaders   *ShaderCache
	pipelines *PipelineCache

	ready bool
}

// NewGPUState creates an uninitialized GPUState. Call SetDeviceProvider or
// EnsureDevice before using GPU resources.
func NewGPUState() *GPUState {
	return &GPUState{}
}

// SetDeviceProvider configures the GPUState to use a shared GPU device from
// an external provider (e.g., gogpu.App). The provider's Device() must return
// a *wgpu.Device.
//
// This follows the gg/internal/gpu/gpu_shared.go:124-189 pattern exactly:
//  1. Check for software adapter -> return error
//  2. Type-assert provider.Device() to *wgpu.Device
//  3. Get queue
//  4. Destroy own resources if previously initialized
//  5. Store external device, create shader + pipeline caches
//  6. Set ready = true
func (g *GPUState) SetDeviceProvider(provider gpucontext.DeviceProvider) error {
	// Check if adapter is software/CPU — GPU shaders require a hardware adapter.
	if wgpuAdapter := wgpu.AdapterFromHandle(provider.Adapter()); wgpuAdapter != nil {
		if wgpuAdapter.Info().DeviceType == gputypes.DeviceTypeCPU {
			return fmt.Errorf("g3d: software adapter detected, GPU rendering requires a hardware adapter")
		}
	}

	wgpuDev := wgpu.DeviceFromHandle(provider.Device())
	if wgpuDev == nil {
		return fmt.Errorf("g3d: provider Device is nil")
	}

	wgpuQueue := wgpuDev.Queue()
	if wgpuQueue == nil {
		return fmt.Errorf("g3d: provider Queue is nil")
	}

	// Destroy own resources if we had a standalone device.
	g.destroyResources()

	// Use provided resources.
	g.device = wgpuDev
	g.queue = wgpuQueue
	g.external = true

	// Create shader and pipeline caches.
	if err := g.initCaches(); err != nil {
		g.device = nil
		g.queue = nil
		g.external = false
		return err
	}

	g.ready = true
	return nil
}

// SetDeviceDirect configures the GPUState with an explicit device and queue.
// Used by NewRendererFromDevice for standalone mode without a full DeviceProvider.
func (g *GPUState) SetDeviceDirect(device *wgpu.Device, queue *wgpu.Queue) error {
	if device == nil {
		return fmt.Errorf("g3d: device is nil")
	}
	if queue == nil {
		return fmt.Errorf("g3d: queue is nil")
	}

	g.destroyResources()

	g.device = device
	g.queue = queue
	g.external = true // caller owns the device

	if err := g.initCaches(); err != nil {
		g.device = nil
		g.queue = nil
		g.external = false
		return err
	}

	g.ready = true
	return nil
}

// EnsureDevice lazily creates a standalone wgpu device if none has been set.
// This is used for headless testing or standalone rendering without gogpu.
func (g *GPUState) EnsureDevice() error {
	if g.ready {
		return nil
	}

	inst, err := wgpu.CreateInstance(nil)
	if err != nil {
		return fmt.Errorf("g3d: create wgpu instance: %w", err)
	}

	adapter, err := inst.RequestAdapter(&wgpu.RequestAdapterOptions{
		PowerPreference: gputypes.PowerPreferenceHighPerformance,
	})
	if err != nil {
		inst.Release()
		return fmt.Errorf("g3d: request adapter: %w", err)
	}

	device, err := adapter.RequestDevice(nil)
	if err != nil {
		adapter.Release()
		inst.Release()
		return fmt.Errorf("g3d: request device: %w", err)
	}

	g.instance = inst
	g.device = device
	g.queue = device.Queue()
	g.external = false

	if err := g.initCaches(); err != nil {
		device.Release()
		adapter.Release()
		inst.Release()
		g.instance = nil
		g.device = nil
		g.queue = nil
		return err
	}

	g.ready = true
	return nil
}

// Device returns the wgpu device, or nil if not initialized.
func (g *GPUState) Device() *wgpu.Device { return g.device }

// Queue returns the wgpu queue, or nil if not initialized.
func (g *GPUState) Queue() *wgpu.Queue { return g.queue }

// IsReady reports whether the GPU is initialized and ready for rendering.
func (g *GPUState) IsReady() bool { return g.ready }

// Shaders returns the shader cache.
func (g *GPUState) Shaders() *ShaderCache { return g.shaders }

// Pipelines returns the pipeline cache.
func (g *GPUState) Pipelines() *PipelineCache { return g.pipelines }

// Close releases all GPU resources in the correct order:
//  1. Pipeline cache (pipelines + layouts + bind group layouts)
//  2. Shader cache (shader modules)
//  3. Device/Instance (only if standalone, external=false)
//
// Follows gg/internal/gpu/gpu_shared.go:216-257 cleanup order.
func (g *GPUState) Close() {
	g.destroyResources()
	g.ready = false
}

// initCaches creates shader and pipeline caches for the current device.
func (g *GPUState) initCaches() error {
	shaders, err := NewShaderCache(g.device)
	if err != nil {
		return fmt.Errorf("g3d: create shader cache: %w", err)
	}
	g.shaders = shaders
	g.pipelines = NewPipelineCache(g.device, shaders)
	return nil
}

// destroyResources releases all owned GPU resources in reverse creation order.
func (g *GPUState) destroyResources() {
	if g.pipelines != nil {
		g.pipelines.Close()
		g.pipelines = nil
	}
	if g.shaders != nil {
		g.shaders.Close()
		g.shaders = nil
	}
	if !g.external {
		if g.device != nil {
			g.device.Release()
		}
		if g.instance != nil {
			g.instance.Release()
		}
	}
	g.device = nil
	g.queue = nil
	g.instance = nil
	g.external = false
}
