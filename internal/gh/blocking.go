package gh

import (
	"sort"
	"sync"
)

type Blocking struct {
	OnYou    []Item
	OnOthers []Item
}

func oldestFirst(items []Item) {
	sort.SliceStable(items, func(a, b int) bool {
		return items[a].UpdatedAt < items[b].UpdatedAt
	})
}

func (c *Client) FetchBlocking() (Blocking, error) {
	consultas := []string{
		"is:pr is:open review-requested:@me archived:false",
		"is:pr is:open author:@me review:required draft:false archived:false",
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

	oldestFirst(res[0])
	oldestFirst(res[1])
	b := Blocking{OnYou: res[0], OnOthers: res[1]}

	var primeiro error
	for _, e := range erros {
		if e != nil {
			primeiro = e
			break
		}
	}
	if primeiro != nil && len(b.OnYou)+len(b.OnOthers) == 0 {
		return b, primeiro
	}
	return b, nil
}
