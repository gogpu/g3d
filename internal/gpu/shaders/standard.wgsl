// standard.wgsl — Blinn-Phong lit shader for StandardMaterial.
//
// Phase 1: Blinn-Phong approximation (ambient + diffuse + specular).
// Phase 2 will upgrade to Cook-Torrance BRDF.
// Used by StandardMaterial (ShaderID = "standard").
//
// Bind groups:
//   Group 0, Binding 0: FrameUniforms  (per-frame, set once)
//   Group 1, Binding 0: ObjectUniforms (per-object)
//   Group 1, Binding 1: MaterialUniforms (per-material, 32 bytes)

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
    color: vec4<f32>,                // 16 bytes (offset  0) — RGBA
    metalness: f32,                  //  4 bytes (offset 16)
    roughness: f32,                  //  4 bytes (offset 20)
    alpha_cutoff: f32,               //  4 bytes (offset 24)
    _pad: f32,                       //  4 bytes (offset 28)
};                                   // Total: 32 bytes, 16-byte aligned

// --- Bind groups ---

@group(0) @binding(0) var<uniform> frame: FrameUniforms;
@group(1) @binding(0) var<uniform> object: ObjectUniforms;
@group(1) @binding(1) var<uniform> material: MaterialUniforms;

// --- Vertex I/O ---

struct VertexInput {
    @location(0) position: vec3<f32>,  // 12 bytes (offset  0)
    @location(1) normal: vec3<f32>,    // 12 bytes (offset 12)
    @location(2) uv: vec2<f32>,        //  8 bytes (offset 24)
};                                     // Stride: 32 bytes

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) world_position: vec3<f32>,
    @location(1) world_normal: vec3<f32>,
    @location(2) uv: vec2<f32>,
};

// --- Vertex shader ---

@vertex
fn vs_main(in: VertexInput) -> VertexOutput {
    var out: VertexOutput;

    let world_pos = object.model * vec4<f32>(in.position, 1.0);
    out.clip_position = frame.view_projection * world_pos;
    out.world_position = world_pos.xyz;

    // Normal matrix = transpose(inverse(model)).
    // Passed as mat4x4 to avoid mat3x3 16-byte stride issues.
    // Only the upper-left 3x3 matters; w component discarded.
    let world_normal = (object.normal_model * vec4<f32>(in.normal, 0.0)).xyz;
    out.world_normal = normalize(world_normal);

    out.uv = in.uv;

    return out;
}

// --- Fragment shader ---

// Light type constants (must match Go LightKind values).
const LIGHT_AMBIENT: u32 = 0u;
const LIGHT_DIRECTIONAL: u32 = 1u;

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
    let base_color = material.color.rgb;
    let alpha = material.color.a;

    // Alpha test (AlphaModeMask): discard fragments below cutoff.
    if alpha < material.alpha_cutoff {
        discard;
    }

    // Re-normalize after interpolation.
    let N = normalize(in.world_normal);
    // View direction: camera to fragment.
    let V = normalize(frame.camera_position.xyz - in.world_position);

    var result = vec3<f32>(0.0, 0.0, 0.0);

    // Shininess: high when smooth (roughness=0), low when rough (roughness=1).
    let shininess = mix(256.0, 2.0, material.roughness);
    // Specular intensity: no specular when fully rough.
    let specular_strength = (1.0 - material.roughness) * 0.5;

    let count = min(frame.light_count, 4u);
    for (var i = 0u; i < count; i = i + 1u) {
        let light = frame.lights[i];
        let light_color = light.color * light.intensity;

        if light.light_type == LIGHT_AMBIENT {
            // Ambient: uniform illumination, no direction.
            result = result + base_color * light_color;
        } else if light.light_type == LIGHT_DIRECTIONAL {
            // Blinn-Phong directional light.
            let L = normalize(-light.direction);
            let H = normalize(L + V);

            let NdotL = max(dot(N, L), 0.0);
            let NdotH = max(dot(N, H), 0.0);

            let diffuse = base_color * NdotL;
            let specular = vec3<f32>(pow(NdotH, shininess)) * specular_strength;

            result = result + (diffuse + specular) * light_color;
        }
    }

    return vec4<f32>(result, alpha);
}
