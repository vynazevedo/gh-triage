package gh

import (
	"sort"
	"sync"
)

type RadarSource int

const (
	SourceReviewRequested RadarSource = iota
	SourceMyPR
	SourceAssigned
	SourceMention
)

type RadarItem struct {
	Item
	Reason string
	Class  int
}

func ClassifyRadar(origem RadarSource, it Item) (int, string) {
	switch origem {
	case SourceReviewRequested:
		return 0, "review pedido a você"
	case SourceMyPR:
		switch {
		case it.Draft:
			return 7, "seu rascunho"
		case it.ReviewDecision == "CHANGES_REQUESTED":
			return 1, "mudanças pedidas"
		case it.CI == CIFail:
			return 2, "CI falhou"
		case it.ReviewDecision == "APPROVED":
			return 3, "pronto para merge"
		default:
			return 6, "aguardando review"
		}
	case SourceAssigned:
		return 4, "atribuída a você"
	default:
		return 5, "você foi mencionado"
	}
}

type SourceResult struct {
	Source RadarSource
	Items  []Item
}

func MergeRadar(results []SourceResult) []RadarItem {
	porChave := map[string]RadarItem{}
	for _, r := range results {
		for _, it := range r.Items {
			classe, motivo := ClassifyRadar(r.Source, it)
			ch := it.Key()
			current, existe := porChave[ch]
			if !existe || classe < current.Class {
				porChave[ch] = RadarItem{Item: it, Reason: motivo, Class: classe}
			}
		}
	}
	out := make([]RadarItem, 0, len(porChave))
	for _, v := range porChave {
		out = append(out, v)
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Class != out[b].Class {
			return out[a].Class < out[b].Class
		}
		if out[a].UpdatedAt != out[b].UpdatedAt {
			return out[a].UpdatedAt < out[b].UpdatedAt
		}
		return out[a].Key() < out[b].Key()
	})
	return out
}

var radarQueries = []struct {
	Source RadarSource
	Q      string
}{
	{SourceReviewRequested, "is:open is:pr review-requested:@me archived:false"},
	{SourceMyPR, "is:open is:pr author:@me archived:false"},
	{SourceAssigned, "is:open is:issue assignee:@me archived:false"},
	{SourceMention, "is:open mentions:@me -author:@me archived:false"},
}

func (c *Client) FetchRadar() ([]RadarItem, int, error) {
	results := make([]SourceResult, len(radarQueries))
	erros := make([]error, len(radarQueries))
	rates := make([]int, len(radarQueries))

	var wg sync.WaitGroup
	for i, cons := range radarQueries {
		wg.Add(1)
		go func(i int, origem RadarSource, q string) {
			defer wg.Done()
			pag, err := c.SearchAll(q, MaxSearchItems)
			results[i] = SourceResult{Source: origem, Items: pag.Items}
			erros[i] = err
			rates[i] = pag.RateLimitRemaining
		}(i, cons.Source, cons.Q)
	}
	wg.Wait()

	items := MergeRadar(results)
	rate := 0
	for _, r := range rates {
		if r > 0 && (rate == 0 || r < rate) {
			rate = r
		}
	}
	var primeiroErro error
	for _, e := range erros {
		if e != nil {
			primeiroErro = e
			break
		}
	}
	if len(items) == 0 && primeiroErro != nil {
		return nil, rate, primeiroErro
	}
	return items, rate, primeiroErro
}
