package app

import "testing"

func TestItoa(t *testing.T) {
	casos := map[int]string{0: "0", 7: "7", 42: "42", -5: "-5", -100: "-100"}
	for n, quer := range casos {
		if got := itoa(n); got != quer {
			t.Errorf("itoa(%d)=%q, quer %q", n, got, quer)
		}
	}
}
