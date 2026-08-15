package app

import (
	"testing"
	"time"
)

func TestRelativeAgeTiers(t *testing.T) {
	now := time.Now()
	casos := []struct {
		atras time.Duration
		quer  string
	}{
		{30 * time.Minute, "30m"},
		{3 * time.Hour, "3h"},
		{2 * 24 * time.Hour, "2d"},
		{21 * 24 * time.Hour, "3sem"},
		{90 * 24 * time.Hour, "3mes"},
		{800 * 24 * time.Hour, "2a"},
	}
	for _, c := range casos {
		iso := now.Add(-c.atras).Format(time.RFC3339)
		if got := relativeAge(iso); got != c.quer {
			t.Errorf("relativeAge(%v atras) = %q, quer %q", c.atras, got, c.quer)
		}
		if len([]rune(relativeAge(iso))) > 5 {
			t.Errorf("idade %q passou de 5 caracteres (trunca na coluna QUANDO)", relativeAge(iso))
		}
	}
}

func TestItoa(t *testing.T) {
	casos := map[int]string{0: "0", 7: "7", 42: "42", -5: "-5", -100: "-100"}
	for n, quer := range casos {
		if got := itoa(n); got != quer {
			t.Errorf("itoa(%d)=%q, quer %q", n, got, quer)
		}
	}
}
