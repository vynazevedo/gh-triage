package gh

import "time"

func AgeSince(iso string) (time.Duration, bool) {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return 0, false
	}
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	return d, true
}

func UrgencyHint(it Item) string {
	if it.Kind == KindPR {
		switch {
		case it.Draft:
			return "rascunho"
		case it.CI == CIFail:
			return "CI vermelho"
		case it.ReviewDecision == "CHANGES_REQUESTED":
			return "mudanças pedidas"
		case it.ReviewDecision == "APPROVED":
			return "aprovado, pronto para merge"
		case it.CI == CIPending:
			return "CI rodando"
		case it.ReviewDecision == "REVIEW_REQUIRED" || it.ReviewDecision == "":
			return "sem review ainda"
		}
		return ""
	}
	if len(it.Linked) > 0 {
		return "tem PR vinculado"
	}
	return ""
}
