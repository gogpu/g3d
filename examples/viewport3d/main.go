// Command viewport3d embeds a g3d 3D viewport inside a gogpu/ui application.
//
// This example demonstrates the integration between g3d (3D rendering) and
// gogpu/ui (GUI toolkit) using the Viewport3D widget. A rotating PBR-lit cube
// is rendered into an offscreen GPU texture, which the ui compositor then
// blits into the widget tree alongside standard 2D UI elements.
//
// Architecture:
//
//	gogpu.App (window + GPU device)
//	  ├── ui.App (widget tree, event dispatch)
//	  │     └── Column layout
//	  │           ├── Title text
//	  │           ├── Viewport3D widget ← g3d renders here
//	  │           └── Control buttons
//	  └── g3d.Renderer (forward pipeline, PBR, depth)
//
// The g3d renderer is created lazily on the first OnRender callback because
// the GPU device is not available until gogpu.App.Run() initializes it.
// Scene graph objects (scene, camera, lights, mesh) are pure CPU data and
// can be created before Run().
//
// Run with different backends:
//
//	GOGPU_GRAPHICS_API=vulkan   go run .
//	GOGPU_GRAPHICS_API=dx12     go run .
//	GOGPU_GRAPHICS_API=gles     go run .
//	GOGPU_GRAPHICS_API=software go run .
package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration for ui text rendering

	"github.com/gogpu/g3d"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/viewport3d"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
	"github.com/gogpu/wgpu"
)

// viewportWidth and viewportHeight define the offscreen texture dimensions
// for the 3D viewport in logical pixels.
const (
	viewportWidth  = 600
	viewportHeight = 400
)

func main() {
	// --- Application setup ---

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("g3d + ui — Viewport3D Demo").
		WithSize(750, 620))

	m3 := material3.New(widget.Hex(0x6750A4)) // Material 3 purple
	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)

	// --- Scene graph setup (CPU only, no GPU needed) ---

	scene, camera, cube := buildScene()

	// --- Renderer state (lazy init inside OnRender) ---

	var renderer *g3d.Renderer
	var lastTime time.Time
	var rotationY, rotationX float32
	paused := false

	// --- Viewport3D widget ---

	vp := viewport3d.New(
		viewport3d.Size(viewportWidth, viewportHeight),
		viewport3d.Continuous(true), // real-time 3D rendering every frame
		viewport3d.OnRender(func(view gpucontext.TextureView) {
			if view.IsNil() {
				return
			}

			// Lazy renderer init: GPU device is available only after Run().
			if renderer == nil {
				provider := gogpuApp.GPUContextProvider()
				if provider == nil {
					return
				}
				var err error
				renderer, err = g3d.NewRenderer(provider)
				if err != nil {
					log.Printf("g3d: create renderer: %v", err)
					return
				}
				lastTime = time.Now()
			}

			// Update animation: rotate the cube based on elapsed time.
			now := time.Now()
			dt := float32(now.Sub(lastTime).Seconds())
			lastTime = now

			if !paused {
				rotationY += dt * 0.8
				rotationX += dt * 0.3
				cube.MeshNode().SetRotation(g3d.Euler{X: rotationX, Y: rotationY})
			}

			// Resize depth buffer to match viewport dimensions.
			w, h := viewportWidth, viewportHeight
			renderer.SetSize(uint32(w), uint32(h))

			// Convert gpucontext.TextureView to *wgpu.TextureView.
			// gpucontext.TextureView is an opaque handle wrapping an
			// unsafe.Pointer to the concrete GPU type. This is the
			// standard pattern for cross-package GPU resource sharing
			// in the gogpu ecosystem (see gpucontext/handle.go).
			wgpuView := (*wgpu.TextureView)(view.Pointer())

			if err := renderer.Render(scene, camera, wgpuView); err != nil {
				log.Printf("g3d: render: %v", err)
			}
		}),
	)

	// --- UI layout ---

	uiApp.SetRoot(buildUI(vp, &paused, &rotationX, &rotationY, cube))

	// --- Cleanup and run ---

	gogpuApp.OnClose(func() {
		if renderer != nil {
			renderer.Release()
		}
	})

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

// buildScene creates the g3d scene graph: scene, lights, camera, and a PBR cube.
// All objects are CPU data structures — no GPU device required.
func buildScene() (*g3d.Scene, g3d.Camera, *g3d.Mesh) {
	scene := g3d.NewScene()
	scene.SetBackground(g3d.RGB(0.12, 0.12, 0.18))

	// Ambient light: base illumination so no face is pure black.
	ambient := g3d.NewAmbientLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(0.25),
	)
	ambientNode := g3d.NewNode()
	ambientNode.SetName("AmbientLight")
	ambientNode.SetUserData(ambient)
	scene.Add(ambientNode)

	// Directional light: simulates sunlight from the upper-right.
	sun := g3d.NewDirectionalLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(1.0),
	)
	sun.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-45),
		Y: g3d.Radians(30),
	})
	scene.Add(sun.LightNode())

	// PBR cube: metallic-roughness material.
	cube := g3d.NewMesh(
		g3d.NewBoxGeometry(1, 1, 1),
		g3d.NewStandardMaterial(
			g3d.WithColor(g3d.RGB(0.4, 0.7, 1.0)),
			g3d.WithMetallic(0.3),
			g3d.WithRoughness(0.6),
		),
	)
	scene.Add(cube.MeshNode())

	// Camera: 60-degree vertical FOV, positioned to frame the cube.
	camera := g3d.NewPerspectiveCamera(
		60,
		float32(viewportWidth)/float32(viewportHeight),
		0.1, 100,
	)
	camera.CameraNode().SetPosition(g3d.Vec3{X: 0, Y: 0.8, Z: 3})

	return scene, camera, cube
}

// buildUI creates the widget tree: a card with the 3D viewport and controls.
func buildUI(
	vp *viewport3d.Widget,
	paused *bool,
	rotX, rotY *float32,
	cube *g3d.Mesh,
) *primitives.BoxWidget {
	card := primitives.Box(
		// Title.
		primitives.Text("g3d Viewport3D").
			FontSize(22).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		// Subtitle.
		primitives.Text("Rotating PBR cube rendered by g3d inside a ui widget").
			FontSize(13).
			Color(widget.RGBA8(100, 100, 100, 255)),

		// 3D viewport.
		vp,

		// Control row.
		primitives.Box(
			button.New(
				button.TextOpt("Pause / Resume"),
				button.OnClick(func() {
					*paused = !*paused
					if *paused {
						fmt.Println("Animation paused")
					} else {
						fmt.Println("Animation resumed")
					}
				}),
			),
			button.New(
				button.TextOpt("Reset"),
				button.OnClick(func() {
					*rotX = 0
					*rotY = 0
					cube.MeshNode().SetRotation(g3d.Euler{})
					fmt.Println("Rotation reset")
				}),
			),
			button.New(
				button.TextOpt("Toggle Continuous"),
				button.OnClick(func() {
					vp.SetContinuous(!vp.IsContinuous())
					fmt.Printf("Continuous rendering: %v\n", vp.IsContinuous())
				}),
			),
		).Gap(8),

		// Description.
		primitives.Text(
			"The 3D viewport uses g3d's forward renderer with PBR materials, "+
				"Blinn-Phong lighting, and depth testing. All 5 GPU backends supported.",
		).
			FontSize(11).
			Color(widget.RGBA8(140, 140, 140, 255)),
	).
		Padding(28).
		Gap(14).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)

	return primitives.Box(card).
		Padding(20).
		Background(widget.RGBA8(245, 245, 250, 255))
}

// Ensure unsafe import is used (required for the gpucontext.TextureView conversion).
var _ = unsafe.Pointer(nil)
