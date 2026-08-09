// Command fullscreen-overlay demonstrates g3d + gg compositing in a single
// window: a lit, rotating PBR cube fills the entire viewport while gg draws
// a 2D HUD (heads-up display) on top.
//
// This is the canonical pattern for any application that mixes 3D content
// with 2D overlays: games (health bar, minimap, crosshair), CAD tools
// (status bar, toolbars), scientific visualizers (legends, annotations).
//
// Architecture (two-pass, same surface):
//
//	Pass 1 — g3d forward renderer (LoadOp::Clear + depth)
//	       ↓ dc.MarkPreserveContent() + g3dSource.ReportDamageWithReason(...)
//	Pass 2 — gg canvas overlay    (LoadOp::Load, no depth, alpha blend)
//
// MarkPreserveContent signals gogpu that a prior render pass has already
// written to the surface, so gg's pass uses LoadOp::Load instead of
// LoadOp::Clear — preserving the 3D content underneath.
//
// RegisterDamageSource + ReportDamageWithReason (ADR-065) reports per-frame
// damage so the compositor can union damage from all renderers and use
// VkPresentRegionsKHR on Wayland for power-efficient partial presents.
//
// Enterprise references: Chromium cc DamageTracker (union pattern),
// Flutter InlinePassContext, Qt6 QRhi beginExternal/endExternal.
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
	"image"
	"log"
	"math"
	"os"
	"time"

	"github.com/gogpu/g3d"
	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu" // Register GPU accelerator (SDF + MSDF text)
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
)

func main() {
	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("g3d + gg — Fullscreen 3D + 2D Overlay").
		WithSize(960, 640).
		WithContinuousRender(true))

	// --- g3d scene setup (CPU only, no GPU device needed) ---

	scene, camera, cube := buildScene()

	// --- Font loading for HUD text (before Run, no GPU needed) ---

	fontSource := loadFontSource()
	var faceTitle, faceFPS, faceStatus text.Face
	if fontSource != nil {
		faceTitle = fontSource.Face(22)
		faceFPS = fontSource.Face(16)
		faceStatus = fontSource.Face(14)
	}

	// --- Rendering state (lazy-initialized in OnDraw) ---

	var renderer3d *g3d.Renderer
	var canvas *ggcanvas.Canvas
	var g3dSource *gogpu.DamageSource // ADR-065: damage reporting for 3D content
	var backendName string
	var lastDrawTime time.Time
	var animTime float64
	var fpsFrames int
	var lastFPSTime time.Time
	var currentFPS float64
	paused := false

	// OnUpdate: rotate the cube at a constant angular velocity. Rotation is
	// accumulated into animTime which pauses cleanly. SetRotation triggers
	// the dirty flag that propagates through the scene graph transform cache.
	app.OnUpdate(func(dt float64) {
		if paused {
			return
		}
		animTime += dt
		cube.MeshNode().SetRotation(g3d.Euler{
			Y: float32(animTime) * 0.8,
			X: float32(animTime) * 0.3,
		})
	})

	// OnResize: update the camera aspect ratio when the window changes size.
	// The renderer's SetSize is called in OnDraw with physical pixel dimensions.
	app.OnResize(func(width, height int) {
		if height > 0 {
			camera.SetAspect(float32(width) / float32(height))
		}
	})

	// OnDraw: the core two-pass rendering loop.
	app.OnDraw(func(dc *gogpu.Context) {
		// ---- Lazy initialization (first frame only) ----

		provider := app.GPUContextProvider()
		if provider == nil {
			return
		}

		// Create g3d renderer on first frame when GPU is ready.
		if renderer3d == nil {
			var err error
			renderer3d, err = g3d.NewRenderer(provider)
			if err != nil {
				log.Fatalf("g3d: create renderer: %v", err)
			}
			// Register g3d as a damage source (ADR-065). The compositor
			// unions damage from all sources (gg, g3d) at present time.
			g3dSource = dc.RegisterDamageSource("g3d")
			backendName = dc.Backend()
			log.Printf("Backend: %s", backendName)
		}

		// Create gg canvas on first frame.
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}

		if canvas == nil {
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Fatalf("gg: create canvas: %v", err)
			}
			lastDrawTime = time.Now()
			lastFPSTime = time.Now()
			log.Printf("Canvas created: %dx%d", w, h)
		}

		// Handle window resize for the gg canvas.
		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("gg: canvas resize error: %v", err)
			}
		}

		// ---- FPS counter (wall-clock, independent of animation) ----

		now := time.Now()
		if !lastDrawTime.IsZero() {
			// Clamp dt to avoid FPS spike after resume.
			dt := now.Sub(lastDrawTime).Seconds()
			if dt > 0.1 {
				dt = 1.0 / 60.0
			}
			_ = dt // dt already accumulated in OnUpdate
		}
		lastDrawTime = now

		fpsFrames++
		if time.Since(lastFPSTime) >= time.Second {
			currentFPS = float64(fpsFrames) / time.Since(lastFPSTime).Seconds()
			fpsFrames = 0
			lastFPSTime = now
		}

		// ---- Pass 1: g3d renders the 3D scene (LoadOp::Clear + depth) ----

		fbW, fbH := dc.FramebufferSize()
		renderer3d.SetSize(uint32(fbW), uint32(fbH))

		view := dc.SurfaceView()
		if view == nil {
			return // Surface not available (minimized window).
		}

		// Use the framework's shared encoder so g3d and gg record into
		// one command buffer — single queue.Submit at frame end.
		// This avoids a Vulkan present-semaphore race on TBDR GPUs
		// (e.g. Apple Silicon via Mesa Asahi). See g3d#22.
		encoder := dc.CommandEncoder()
		if encoder == nil {
			return
		}

		if err := renderer3d.RenderTo(encoder, scene, camera, view); err != nil {
			log.Printf("g3d: render error: %v", err)
			return
		}

		// ---- Signal: 3D content already on the surface ----
		// ADR-065: report full-viewport damage for the 3D scene so the
		// compositor knows this area changed. MarkPreserveContent tells
		// subsequent render passes (gg) to use LoadOp::Load.
		g3dSource.ReportDamageWithReason(
			gpucontext.DamageReason{
				Category: gpucontext.DamageCategoryAnimation,
				Detail:   "3D scene",
			},
			image.Rect(0, 0, fbW, fbH),
		)
		dc.MarkPreserveContent()

		// ---- Pass 2: gg renders the 2D HUD overlay (LoadOp::Load) ----

		if err := canvas.Draw(func(cc *gg.Context) {
			drawHUD(cc, w, h, currentFPS, animTime, paused, backendName,
				faceTitle, faceFPS, faceStatus)
		}); err != nil {
			log.Printf("gg: draw error: %v", err)
		}

		if err := canvas.Render(dc.RenderTarget()); err != nil {
			log.Printf("gg: render error: %v", err)
		}
	})

	// Space toggles animation pause/resume.
	app.EventSource().OnKeyPress(func(key gpucontext.Key, _ gpucontext.Modifiers) {
		if key == gpucontext.KeySpace {
			paused = !paused
			if paused {
				log.Println("Animation paused")
			} else {
				log.Println("Animation resumed")
			}
		}
	})

	// Clean up GPU resources when the application exits.
	app.OnClose(func() {
		if renderer3d != nil {
			renderer3d.Release()
		}
		gg.CloseAccelerator()
	})

	if err := app.Run(); err != nil {
		log.Fatalf("gogpu: %v", err)
	}
}

// buildScene creates the g3d scene graph: scene, lights, camera, and a PBR cube.
// All objects are CPU data structures — no GPU device needed at creation time.
func buildScene() (*g3d.Scene, *g3d.PerspectiveCamera, *g3d.Mesh) {
	scene := g3d.NewScene()
	scene.SetBackground(g3d.RGB(0.08, 0.08, 0.12))

	// Ambient light: base illumination so no face is pure black.
	ambient := g3d.NewAmbientLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(0.25),
	)
	ambientNode := g3d.NewNode()
	ambientNode.SetName("AmbientLight")
	ambientNode.SetUserData(ambient)
	scene.Add(ambientNode)

	// Directional light: sunlight from the upper-right for visible face shading.
	sun := g3d.NewDirectionalLight(
		g3d.WithLightColor(g3d.White),
		g3d.WithLightIntensity(1.0),
	)
	sun.LightNode().SetRotation(g3d.Euler{
		X: g3d.Radians(-45),
		Y: g3d.Radians(30),
	})
	scene.Add(sun.LightNode())

	// PBR cube with metallic-roughness material.
	cube := g3d.NewMesh(
		g3d.NewBoxGeometry(1, 1, 1),
		g3d.NewStandardMaterial(
			g3d.WithColor(g3d.RGB(0.4, 0.7, 1.0)),
			g3d.WithMetallic(0.3),
			g3d.WithRoughness(0.6),
		),
	)
	scene.Add(cube.MeshNode())

	// Camera: 75-degree vertical FOV, 3 units back with slight elevation.
	camera := g3d.NewPerspectiveCamera(75, 960.0/640.0, 0.1, 1000)
	camera.CameraNode().SetPosition(g3d.Vec3{X: 0, Y: 0.5, Z: 3})

	return scene, camera, cube
}

// drawHUD renders the 2D heads-up display elements on top of the 3D scene.
// All coordinates are in logical pixels (DIP); gg handles HiDPI internally.
func drawHUD(
	cc *gg.Context,
	width, height int,
	fps, elapsed float64,
	paused bool,
	backend string,
	faceTitle, faceFPS, faceStatus text.Face,
) {
	fw := float64(width)
	fh := float64(height)

	// ---- Title (top-left) ----

	if faceTitle != nil {
		drawLabelWithBackground(cc, "g3d + gg Overlay Demo", 16, 16, faceTitle,
			0.0, 0.0, 0.0, 0.55, // background: semi-transparent black
			1.0, 1.0, 1.0, 1.0, // text: white
			10, 6) // paddingH, paddingV
	}

	// ---- FPS counter (top-right) ----

	if faceFPS != nil {
		fpsText := fmt.Sprintf("%.0f FPS", fps)
		// Measure text width to right-align.
		cc.SetFont(faceFPS)
		tw, _ := cc.MeasureString(fpsText)
		x := fw - tw - 26
		drawLabelWithBackground(cc, fpsText, x, 16, faceFPS,
			0.0, 0.0, 0.0, 0.55,
			0.4, 1.0, 0.4, 1.0, // text: green
			10, 6)
	}

	// ---- Crosshair (center) ----

	drawCrosshair(cc, fw/2, fh/2, elapsed)

	// ---- Status bar (bottom) ----

	if faceStatus != nil {
		state := "Running"
		if paused {
			state = "Paused"
		}
		statusText := fmt.Sprintf(
			"Backend: %s  |  Objects: 1  |  Time: %.1fs  |  %s  |  Space to pause",
			backend, elapsed, state,
		)
		drawStatusBar(cc, statusText, fw, fh, faceStatus)
	}
}

// drawLabelWithBackground renders text with a rounded semi-transparent
// background rectangle for legibility over 3D content.
func drawLabelWithBackground(
	cc *gg.Context,
	label string,
	x, y float64,
	face text.Face,
	bgR, bgG, bgB, bgA float64,
	fgR, fgG, fgB, fgA float64,
	padH, padV float64,
) {
	cc.SetFont(face)
	tw, th := cc.MeasureString(label)

	// Background rectangle.
	cc.SetRGBA(bgR, bgG, bgB, bgA)
	cc.DrawRoundedRectangle(x, y, tw+padH*2, th+padV*2, 6)
	_ = cc.Fill()

	// Text, vertically centered in the rectangle.
	cc.SetRGBA(fgR, fgG, fgB, fgA)
	cc.DrawString(label, x+padH, y+padV+th*0.85)
}

// drawCrosshair draws an animated crosshair at the center of the viewport.
// The outer ring pulses gently; the inner cross is fixed-size.
func drawCrosshair(cc *gg.Context, cx, cy, elapsed float64) {
	// Pulsing outer ring.
	pulse := 1.0 + 0.15*math.Sin(elapsed*2.0)
	radius := 20.0 * pulse
	cc.SetRGBA(1, 1, 1, 0.35)
	cc.SetLineWidth(1.5)
	cc.DrawCircle(cx, cy, radius)
	_ = cc.Stroke()

	// Inner cross lines (fixed 10px arms).
	arm := 10.0
	cc.SetRGBA(1, 1, 1, 0.7)
	cc.SetLineWidth(1.5)

	// Horizontal arm.
	cc.MoveTo(cx-arm, cy)
	cc.LineTo(cx+arm, cy)
	_ = cc.Stroke()

	// Vertical arm.
	cc.MoveTo(cx, cy-arm)
	cc.LineTo(cx, cy+arm)
	_ = cc.Stroke()

	// Center dot.
	cc.SetRGBA(1, 1, 1, 0.9)
	cc.DrawCircle(cx, cy, 2)
	_ = cc.Fill()
}

// drawStatusBar renders a full-width status bar at the bottom of the viewport.
func drawStatusBar(cc *gg.Context, status string, width, height float64, face text.Face) {
	barH := 32.0
	barY := height - barH

	// Background: dark translucent strip.
	cc.SetRGBA(0, 0, 0, 0.6)
	cc.DrawRectangle(0, barY, width, barH)
	_ = cc.Fill()

	// Status text: centered vertically in the bar.
	cc.SetFont(face)
	_, th := cc.MeasureString(status)
	cc.SetRGBA(0.85, 0.85, 0.85, 1.0)
	cc.DrawString(status, 16, barY+(barH+th*0.85)/2)
}

// loadFontSource locates and loads a system TTF font for HUD text rendering.
// Returns nil if no suitable font is found (text rendering is skipped).
func loadFontSource() *text.FontSource {
	path := findSystemFont()
	if path == "" {
		log.Println("No system font found — HUD text disabled")
		return nil
	}
	source, err := text.NewFontSourceFromFile(path)
	if err != nil {
		log.Printf("Font load failed (%s): %v", path, err)
		return nil
	}
	log.Printf("Loaded font: %s", source.Name())
	return source
}

// findSystemFont probes well-known paths for a TTF font.
// TTC collections are not supported by gg's text engine.
func findSystemFont() string {
	candidates := []string{
		// Windows
		`C:\Windows\Fonts\segoeui.ttf`,
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\calibri.ttf`,
		// macOS
		"/Library/Fonts/Arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
		"/System/Library/Fonts/Monaco.ttf",
		// Linux
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
