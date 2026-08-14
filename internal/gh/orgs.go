package gh

import "fmt"

type Owner struct {
	Login string
}

type RepoRef struct {
	Owner       string
	Name        string
	Description string
	Private     bool
	PushedAt    string
}

func (r RepoRef) FullName() string { return r.Owner + "/" + r.Name }

func (c *Client) MyLogin() (string, error) {
	var resp struct {
		Login string `json:"login"`
	}
	if err := c.rest.Get("user", &resp); err != nil {
		return "", fmt.Errorf("falha ao obter o usuário atual: %w", err)
	}
	return resp.Login, nil
}

func (c *Client) MyOrgs() ([]Owner, error) {
	var resp []struct {
		Login string `json:"login"`
	}
	if err := c.rest.Get("user/orgs?per_page=100", &resp); err != nil {
		return nil, fmt.Errorf("falha ao listar organizações: %w", err)
	}
	out := make([]Owner, 0, len(resp))
	for _, o := range resp {
		out = append(out, Owner{Login: o.Login})
	}
	return out, nil
}

func (c *Client) OrgRepos(org string) ([]RepoRef, error) {
	return c.reposFrom("orgs/"+org+"/repos?sort=pushed&per_page=100", org)
}

func (c *Client) MyRepos() ([]RepoRef, error) {
	return c.reposFrom("user/repos?affiliation=owner,collaborator,organization_member&sort=pushed&per_page=100", "")
}

func (c *Client) reposFrom(path, fallbackOwner string) ([]RepoRef, error) {
	var resp []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		PushedAt    string `json:"pushed_at"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	if err := c.rest.Get(path, &resp); err != nil {
		return nil, fmt.Errorf("falha ao listar repositórios: %w", err)
	}
	out := make([]RepoRef, 0, len(resp))
	for _, r := range resp {
		owner := r.Owner.Login
		if owner == "" {
			owner = fallbackOwner
		}
		out = append(out, RepoRef{
			Owner:       owner,
			Name:        r.Name,
			Description: r.Description,
			Private:     r.Private,
			PushedAt:    r.PushedAt,
		})
	}
	return out, nil
}
