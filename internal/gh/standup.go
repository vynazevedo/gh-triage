package gh

import (
	"sync"
	"time"
)

type Standup struct {
	SinceDays int
	Shipped   []Item
	Ongoing   []Item
	Reviewing []Item
	Blocked   []Item
}

func BlockedFrom(items []Item) []Item {
	var out []Item
	for _, it := range items {
		if it.ReviewDecision == "CHANGES_REQUESTED" || it.CI == CIFail {
			out = append(out, it)
		}
	}
	return out
}

func (c *Client) FetchStandup(sinceDays int) (Standup, error) {
	if sinceDays < 1 {
		sinceDays = 1
	}
	desde := time.Now().AddDate(0, 0, -sinceDays).Format("2006-01-02")
	consultas := []string{
		"is:pr author:@me is:merged merged:>=" + desde + " archived:false",
		"is:pr author:@me is:open archived:false",
		"is:pr review-requested:@me is:open archived:false",
	}
	res := make([][]Item, len(consultas))
	erros := make([]error, len(consultas))

	var wg sync.WaitGroup
	for i, q := range consultas {
		wg.Add(1)
		go func(i int, q string) {
			defer wg.Done()
			pag, err := c.Search(q, "")
			res[i] = pag.Items
			erros[i] = err
		}(i, q)
	}
	wg.Wait()

	su := Standup{SinceDays: sinceDays, Shipped: res[0], Ongoing: res[1], Reviewing: res[2]}
	su.Blocked = BlockedFrom(su.Ongoing)

	var primeiro error
	for _, e := range erros {
		if e != nil {
			primeiro = e
			break
		}
	}
	if primeiro != nil && len(su.Shipped)+len(su.Ongoing)+len(su.Reviewing) == 0 {
		return su, primeiro
	}
	return su, nil
}
