// basic.wgsl — Unlit shader for BasicMaterial.
//
// Renders geometry with a solid color, no lighting calculations.
// Used by BasicMaterial (ShaderID = "basic").
//
// Bind groups:
//   Group 0, Binding 0: FrameUniforms
//   Group 1, Binding 0: ObjectUniforms
//   Group 1, Binding 1: MaterialUniforms
//   Group 1, Binding 2: Base Color Texture
//   Group 1, Binding 3: Base Color Sampler

// --- Uniform structs (must match Go layout byte-for-byte) ---

struct LightData {
    direction: vec3<f32>,    // 12 bytes (offset 0)
    light_type: u32,         //  4 bytes (offset 12) — row total: 16
    color: vec3<f32>,        // 12 bytes (offset 16)
    intensity: f32,          //  4 bytes (offset 28) — row total: 16
};                           // Total: 32 bytes, 16-byte aligned

struct FrameUniforms {
    view_projection: mat4x4<f32>,    //  64 bytes (offset   0)
    camera_position: vec4<f32>,      //  16 bytes (offset  64) — vec4, NOT vec3!
    lights: array<LightData, 4>,     // 128 bytes (offset  80) — 4 x 32
    light_count: u32,                //   4 bytes (offset 208)
    _pad0: u32,                      //   4 bytes (offset 212)
    _pad1: u32,                      //   4 bytes (offset 216)
    _pad2: u32,                      //   4 bytes (offset 220)
};                                   // Total: 224 bytes, 16-byte aligned

struct ObjectUniforms {
    model: mat4x4<f32>,              // 64 bytes (offset  0)
    normal_model: mat4x4<f32>,       // 64 bytes (offset 64)
};                                   // Total: 128 bytes

struct MaterialUniforms {
    color: vec4<f32>,                // 16 bytes (offset 0) — RGBA
};                                   // Total: 16 bytes

// --- Bind groups ---

@group(0) @binding(0) var<uniform> frame: FrameUniforms;
@group(1) @binding(0) var<uniform> object: ObjectUniforms;
@group(1) @binding(1) var<uniform> material: MaterialUniforms;
@group(1) @binding(2) var base_color_texture: texture_2d<f32>;
@group(1) @binding(3) var base_color_sampler: sampler;

// --- Vertex I/O ---

struct VertexInput {
    @location(0) position: vec3<f32>,  // 12 bytes (offset  0)
    @location(1) normal: vec3<f32>,    // 12 bytes (offset 12)
    @location(2) uv: vec2<f32>,        //  8 bytes (offset 24)
};                                     // Stride: 32 bytes

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) uv: vec2<f32>,
};

// --- Vertex shader ---

@vertex
fn vs_main(in: VertexInput) -> VertexOutput {
    var out: VertexOutput;

    let world_pos = object.model * vec4<f32>(in.position, 1.0);
    out.clip_position = frame.view_projection * world_pos;
    out.uv = in.uv;

    return out;
}

// --- Fragment shader ---

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
    // Color from texture.
    let tex_color = textureSample(base_color_texture, base_color_sampler, in.uv);

    // Material color * texture Color.
    return material.color * tex_color;
}