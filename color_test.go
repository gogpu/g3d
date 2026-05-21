package g3d

import (
	"testing"
)

func TestRGB(t *testing.T) {
	c := RGB(0.5, 0.6, 0.7)
	if c.R != 0.5 || c.G != 0.6 || c.B != 0.7 || c.A != 1.0 {
		t.Errorf("RGB(0.5, 0.6, 0.7) = %+v, want R=0.5 G=0.6 B=0.7 A=1.0", c)
	}
}

func TestRGBA(t *testing.T) {
	c := RGBA(0.1, 0.2, 0.3, 0.4)
	if c.R != 0.1 || c.G != 0.2 || c.B != 0.3 || c.A != 0.4 {
		t.Errorf("RGBA(0.1, 0.2, 0.3, 0.4) = %+v, want R=0.1 G=0.2 B=0.3 A=0.4", c)
	}
}

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name       string
		color      Color
		r, g, b, a float32
	}{
		{"White", White, 1, 1, 1, 1},
		{"Black", Black, 0, 0, 0, 1},
		{"Red", Red, 1, 0, 0, 1},
		{"Green", Green, 0, 1, 0, 1},
		{"Blue", Blue, 0, 0, 1, 1},
		{"Yellow", Yellow, 1, 1, 0, 1},
		{"Cyan", Cyan, 0, 1, 1, 1},
		{"Magenta", Magenta, 1, 0, 1, 1},
		{"Gray", Gray, 0.5, 0.5, 0.5, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color.R != tt.r || tt.color.G != tt.g || tt.color.B != tt.b || tt.color.A != tt.a {
				t.Errorf("%s = %+v, want {R:%v G:%v B:%v A:%v}", tt.name, tt.color, tt.r, tt.g, tt.b, tt.a)
			}
		})
	}
}

func TestColorLerp(t *testing.T) {
	tests := []struct {
		name string
		a, b Color
		t    float32
		want Color
	}{
		{"t=0", Black, White, 0, Black},
		{"t=1", Black, White, 1, White},
		{"t=0.5", Black, White, 0.5, Color{0.5, 0.5, 0.5, 1}},
		{"alpha lerp", RGBA(0, 0, 0, 0), RGBA(1, 1, 1, 1), 0.25, RGBA(0.25, 0.25, 0.25, 0.25)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Lerp(tt.b, tt.t)
			if !approxEqualColor(got, tt.want, 1e-6) {
				t.Errorf("Color.Lerp() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func approxEqualColor(a, b Color, eps float32) bool {
	return approxEqual(a.R, b.R, eps) && approxEqual(a.G, b.G, eps) &&
		approxEqual(a.B, b.B, eps) && approxEqual(a.A, b.A, eps)
}
