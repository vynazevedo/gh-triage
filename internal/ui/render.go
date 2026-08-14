package ui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type rgb struct{ r, g, b float64 }

func hexRGB(h string) rgb {
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	if len(h) != 6 {
		return rgb{1, 1, 1}
	}
	return rgb{
		float64(hexPair(h[0:2])) / 255,
		float64(hexPair(h[2:4])) / 255,
		float64(hexPair(h[4:6])) / 255,
	}
}

func (c rgb) hex() string {
	conv := func(v float64) int {
		n := int(math.Round(v * 255))
		if n < 0 {
			return 0
		}
		if n > 255 {
			return 255
		}
		return n
	}
	return "#" + hx(conv(c.r)) + hx(conv(c.g)) + hx(conv(c.b))
}

func lerp(a, b rgb, t float64) rgb {
	lin := func(v float64) float64 {
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	enc := func(v float64) float64 {
		if v <= 0.0031308 {
			return v * 12.92
		}
		return 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	mix := func(x, y float64) float64 { return enc(lin(x) + (lin(y)-lin(x))*t) }
	return rgb{mix(a.r, b.r), mix(a.g, b.g), mix(a.b, b.b)}
}

func Gradient(start, mid, end string, n int) []string {
	if n <= 0 {
		return nil
	}
	s, m, e := hexRGB(start), hexRGB(mid), hexRGB(end)
	out := make([]string, n)
	for i := 0; i < n; i++ {
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		var c rgb
		if t < 0.5 {
			c = lerp(s, m, t*2)
		} else {
			c = lerp(m, e, (t-0.5)*2)
		}
		out[i] = c.hex()
	}
	return out
}

func GradientText(s string, cols []string) string {
	r := []rune(s)
	if len(r) == 0 || len(cols) == 0 {
		return s
	}
	var b strings.Builder
	for i, ch := range r {
		col := cols[i*len(cols)/len(r)]
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(string(ch)))
	}
	return b.String()
}

var vertBlocks = []rune(" ▁▂▃▄▅▆▇█")

func Sparkline(vals []float64, from, to string) string {
	if len(vals) == 0 {
		return ""
	}
	mx := 0.0
	for _, v := range vals {
		if v > mx {
			mx = v
		}
	}
	if mx == 0 {
		mx = 1
	}
	cols := Gradient(from, from, to, len(vals))
	var b strings.Builder
	for i, v := range vals {
		lvl := int(math.Round(v / mx * 8))
		if lvl < 0 {
			lvl = 0
		}
		if lvl > 8 {
			lvl = 8
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(cols[i])).Render(string(vertBlocks[lvl])))
	}
	return b.String()
}

var eighths = []rune("▏▎▍▌▋▊▉█")

func Meter(frac float64, width int, col string) string {
	if width <= 0 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	total := int(math.Round(frac * float64(width) * 8))
	full := total / 8
	rem := total % 8
	if full > width {
		full = width
		rem = 0
	}
	preenchido := strings.Repeat("█", full)
	if rem > 0 && full < width {
		preenchido += string(eighths[rem-1])
		full++
	}
	trilho := strings.Repeat("░", width-full)
	corTrilho := lipgloss.NewStyle().Foreground(Border)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(preenchido) + corTrilho.Render(trilho)
}

// MeterSegmentado desenha segmentos proporcionais, cada um com sua cor, num
// mesmo trilho — usado pelo medidor de urgência do Radar.
func SegmentedMeter(valores []float64, cores []string, width int) string {
	if width <= 0 || len(valores) == 0 {
		return strings.Repeat("░", max(0, width))
	}
	total := 0.0
	for _, v := range valores {
		total += v
	}
	if total == 0 {
		return lipgloss.NewStyle().Foreground(Border).Render(strings.Repeat("░", width))
	}
	var b strings.Builder
	usado := 0
	for i, v := range valores {
		n := int(math.Round(v / total * float64(width)))
		if usado+n > width {
			n = width - usado
		}
		if n <= 0 {
			continue
		}
		cor := "#888888"
		if i < len(cores) {
			cor = cores[i]
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(cor)).Render(strings.Repeat("█", n)))
		usado += n
	}
	if usado < width {
		b.WriteString(lipgloss.NewStyle().Foreground(Border).Render(strings.Repeat("░", width-usado)))
	}
	return b.String()
}
