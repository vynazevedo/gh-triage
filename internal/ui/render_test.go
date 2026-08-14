package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestGradientEndpoints(t *testing.T) {
	g := Gradient("#000000", "#808080", "#ffffff", 5)
	if len(g) != 5 {
		t.Fatalf("len=%d, quer 5", len(g))
	}
	if strings.ToLower(g[0]) != "#000000" {
		t.Errorf("primeira cor=%q, quer #000000", g[0])
	}
	if strings.ToLower(g[4]) != "#ffffff" {
		t.Errorf("última cor=%q, quer #ffffff", g[4])
	}
}

func TestGradientSingle(t *testing.T) {
	g := Gradient("#112233", "#445566", "#778899", 1)
	if len(g) != 1 || strings.ToLower(g[0]) != "#112233" {
		t.Fatalf("n=1 deveria dar a cor inicial, veio %v", g)
	}
}

func TestSparklineWidthAndLevels(t *testing.T) {
	out := stripANSIu(Sparkline([]float64{0, 1, 2, 4, 8}, "#3FB950", "#F85149"))
	r := []rune(out)
	if len(r) != 5 {
		t.Fatalf("largura=%d, quer 5", len(r))
	}
	// menor valor -> bloco baixo (espaço), maior -> bloco cheio
	if r[0] != ' ' {
		t.Errorf("valor 0 deveria ser espaço, veio %q", string(r[0]))
	}
	if r[4] != '█' {
		t.Errorf("valor máx deveria ser █, veio %q", string(r[4]))
	}
}

func TestSparklineEmpty(t *testing.T) {
	if Sparkline(nil, "#000", "#fff") != "" {
		t.Error("sparkline vazio deveria ser string vazia")
	}
}

func TestMeterFill(t *testing.T) {
	cheio := stripANSIu(Meter(1.0, 10, "#4C86F0"))
	if strings.Count(cheio, "█") != 10 {
		t.Errorf("frac 1.0 deveria encher 10 blocos, veio %d", strings.Count(cheio, "█"))
	}
	vazio := stripANSIu(Meter(0.0, 10, "#4C86F0"))
	if strings.Count(vazio, "░") != 10 {
		t.Errorf("frac 0.0 deveria ser 10 de trilho, veio %d", strings.Count(vazio, "░"))
	}
	meio := stripANSIu(Meter(0.5, 10, "#4C86F0"))
	if r := []rune(meio); len(r) != 10 {
		t.Errorf("largura do meter=%d, quer 10", len(r))
	}
}

func TestSegmentedMeterWidth(t *testing.T) {
	out := stripANSIu(SegmentedMeter([]float64{2, 1, 1}, []string{"#F85149", "#D29922", "#8B949E"}, 12))
	if r := []rune(out); len(r) != 12 {
		t.Errorf("largura=%d, quer 12", len(r))
	}
}

func stripANSIu(s string) string { return ansi.Strip(s) }
