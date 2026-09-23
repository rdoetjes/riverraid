#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform float time;
uniform vec2 resolution;

// 2D Hash function
vec2 hash2(vec2 p) {
    p = vec2(dot(p, vec2(127.1, 311.7)), dot(p, vec2(269.5, 183.3)));
    return fract(sin(p) * 43758.5453);
}

// 2D Noise for distortion
float noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    float a = hash2(i).x;
    float b = hash2(i + vec2(1.0, 0.0)).x;
    float c = hash2(i + vec2(0.0, 1.0)).x;
    float d = hash2(i + vec2(1.0, 1.0)).x;
    return mix(a, b, u.x) + (c - a) * u.y * (1.0 - u.x) + (d - b) * u.x * u.y;
}

// Voronoi with coordinate warping support
float voronoi(vec2 x) {
    vec2 n = floor(x);
    vec2 f = fract(x);
    float m = 8.0;

    for(int j = -1; j <= 1; j++) {
        for(int i = -1; i <= 1; i++) {
            vec2 g = vec2(float(i), float(j));
            vec2 o = hash2(n + g);

            // Oscillate the points in the grid
            o = 0.8 + 0.8 * sin(time * 0.4 + 6.2831 * o);

            vec2 r = 0.86*g + o - f;
            float d = dot(r, r);

            if (d < m) {
                m = d;
            }
        }
    }
    return sqrt(m);
}

void main()
{
    // Screen coordinates
    vec2 uv = gl_FragCoord.xy / resolution.xy;
    float aspect = resolution.x / resolution.y;

    // Animate the water
    float speed = 0.8;
    float scale = 8.0;

    // Wave layers for complexity and "bobbing"
    float wave1 = sin(uv.x * scale + time * speed) * 0.02;
    float wave2 = sin(uv.y * scale * 1.5 + time * speed * 1.2) * 0.02;
    float wave3 = sin((uv.x + uv.y) * scale * 0.5 - time * speed * 0.8) * 0.01;

    // Distort UVs for the motion
    vec2 distortedUv = uv + vec2(wave1 + wave3, wave2 + wave3);

    // Background blue colors
    vec3 waterDeep = vec3(0.05, 0.25, 0.45);
    vec3 waterMid = vec3(0.1, 0.35, 0.6);

    // Mix colors based on waves for the "bobbing" look
    float bob = 0.5 + 0.5 * sin(time * 0.5 + distortedUv.y * 3.0);
    vec3 color = mix(waterDeep, waterMid, bob);

    // --- DISTORTED VORONOI WAVE HEADS ---
    vec2 voronoiUv = uv * vec2(aspect * 9.0, 11.0);
    voronoiUv.y -= time * 0.15;

    // ADD NOISE DISTORTION: Warp the Voronoi input to make it less circular
    float warpAmt = 0.65;
    vec2 warp = vec2(
        noise(voronoiUv * 1.85 + time * 0.5),
        noise(voronoiUv * 1.95 - time * 0.5)
    );

    float v = voronoi(voronoiUv + warp * warpAmt);

    // Revert to the soft pow() but with the warped input
    float waveHeads = pow(max(0.0, 1.0 - v), 3.5);

    // Light blue color for the wave heads
    vec3 headColor = vec3(0.55, 0.82, 1.0);

    // Pulse the intensity
    float pulse = 0.6 + 0.4 * sin(time * 1.2)*0.2;

    color = mix(color, headColor, waveHeads * 0.3 * pulse);

    // --- SUBTLE RIPPLES ---
    float r1 = sin(distortedUv.x * 4.5 + time * 0.6) * cos(distortedUv.y * 3.5 - time * 0.4);
    float rippleFactor = r1 * 0.5;
    color += vec3(0.08, 0.12, 0.18) * rippleFactor * 0.3;

    // Add soft gradient from top to bottom
    color *= (0.8 + 0.2 * uv.y);

    finalColor = vec4(color, 1.0);
}
