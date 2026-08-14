package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/vynazevedo/gh-triage/internal/app"
	"github.com/vynazevedo/gh-triage/internal/cache"
	"github.com/vynazevedo/gh-triage/internal/gh"
	"github.com/vynazevedo/gh-triage/internal/output"
	"github.com/vynazevedo/gh-triage/internal/ui"
)

var version = "dev"

func main() {
	if err := fang.Execute(context.Background(), rootCmd(), fang.WithVersion(version)); err != nil {
		os.Exit(1)
	}
}

type options struct {
	repo     string
	state    string
	kind     string
	assignee string
	format   string
	radar    bool
	nerd     bool
	watch    bool
	refresh  bool
	maxAge   int
	term     string
}

func rootCmd() *cobra.Command {
	var o options

	cmd := &cobra.Command{
		Use:   "gh-triage [consulta]",
		Short: "Sua caixa de entrada do GitHub no terminal",
		Long: "Abre no Radar: tudo que espera por você em todos os repositórios,\n" +
			"ordenado por urgência. Com -R, vira uma mesa de triagem do repositório\n" +
			"ou organização, cruzando issues, PRs e commits.",
		Example: "  gh triage\n" +
			"  gh triage -R cli/cli \"avaliação\"\n" +
			"  gh triage -R cli --state all\n" +
			"  gh triage --radar -o json\n" +
			"  gh triage -o count      # resumo compacto para statusline\n" +
			"  gh triage --watch       # Radar ao vivo",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				o.term = args[0]
			}
			return run(o)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.repo, "repo", "R", "", "repositório owner/repo ou org (padrão: repo atual via gh)")
	f.StringVar(&o.state, "state", "open", "open | closed | all")
	f.StringVar(&o.kind, "type", "", "issue | pr (padrão: ambos)")
	f.StringVar(&o.assignee, "assignee", "", "assignee (aceita @me)")
	f.StringVarP(&o.format, "output", "o", "", "json | tsv | plain | count | count-json (força modo não-interativo)")
	f.BoolVar(&o.radar, "radar", false, "abre no Radar: tudo que espera por você")
	f.BoolVar(&o.nerd, "nerd", true, "ícones Nerd Font (desative se não tiver a fonte)")
	f.BoolVar(&o.watch, "watch", false, "atualiza o Radar automaticamente enquanto a TUI está aberta")
	f.BoolVar(&o.refresh, "refresh", false, "ignora o cache e busca agora (para -o count)")
	f.IntVar(&o.maxAge, "max-age", 60, "idade máxima do cache em segundos para -o count")

	cmd.AddCommand(standupCmd())
	cmd.AddCommand(blockingCmd())
	return cmd
}

func blockingCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "blocking",
		Short: "Quem espera pelo seu review e por quem você espera",
		Long: "Mostra os dois lados do bloqueio de review: PRs em que pediram seu\n" +
			"review (pessoas esperando por você) e seus PRs ainda sem review\n" +
			"(você esperando por elas), ordenados pelo que espera há mais tempo.",
		Example: "  gh triage blocking\n" +
			"  gh triage blocking -o json",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gh.NewClient()
			if err != nil {
				return err
			}
			b, err := client.FetchBlocking()
			if err != nil {
				return err
			}
			f := format
			if f == "" {
				f = "plain"
			}
			return output.WriteBlocking(os.Stdout, b, f)
		},
	}
	c.Flags().StringVarP(&format, "output", "o", "", "plain | json")
	return c
}

func standupCmd() *cobra.Command {
	var sinceDays int
	var format string
	c := &cobra.Command{
		Use:   "standup",
		Short: "Resumo da sua atividade: entregue, em andamento, revisando, bloqueado",
		Long: "Monta um resumo pronto para colar na daily a partir da sua atividade\n" +
			"no GitHub: PRs merged na janela, seus PRs abertos, reviews pedidos a\n" +
			"você e o que está bloqueado (mudanças pedidas ou CI vermelho).",
		Example: "  gh triage standup\n" +
			"  gh triage standup --since 3\n" +
			"  gh triage standup -o json",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gh.NewClient()
			if err != nil {
				return err
			}
			su, err := client.FetchStandup(sinceDays)
			if err != nil {
				return err
			}
			f := format
			if f == "" {
				f = "plain"
			}
			return output.WriteStandup(os.Stdout, su, f)
		},
	}
	c.Flags().IntVar(&sinceDays, "since", 1, "janela em dias para 'entregue'")
	c.Flags().StringVarP(&format, "output", "o", "", "plain | json")
	return c
}

func run(o options) error {
	filters := gh.Filters{
		State:    o.state,
		Kind:     o.kind,
		Assignee: o.assignee,
		Term:     o.term,
	}

	if o.repo != "" {
		if err := gh.ValidateRepo(o.repo); err != nil {
			return err
		}
		filters.Repo = o.repo
	} else if r, err := gh.CurrentRepo(); err == nil {
		filters.Repo = r
	}

	client, err := gh.NewClient()
	if err != nil {
		return err
	}

	format := o.format
	if format == "" && !isatty.IsTerminal(os.Stdout.Fd()) {
		format = "json"
	}

	if format == "count" || format == "count-json" {
		return pipeCount(client, format, o.refresh, o.maxAge)
	}

	if format != "" {
		if o.radar || filters.Repo == "" {
			return pipeRadar(client, format)
		}
		return pipeMode(client, filters, format)
	}

	ui.UseNerdFonts(o.nerd)
	ui.SetMarkdownStyle(lipgloss.HasDarkBackground())
	browser.Stdout = io.Discard
	browser.Stderr = io.Discard

	abaInicial := app.TabRadar
	if !o.radar && (o.repo != "" || o.term != "" || o.kind != "" || o.assignee != "") {
		abaInicial = app.TabTriage
	}
	if filters.Repo == "" {
		abaInicial = app.TabRadar
	}

	modelo := app.New(client, filters.Repo, filters, abaInicial)
	if o.watch {
		modelo = modelo.WithWatch(30)
	}

	programa := tea.NewProgram(modelo, tea.WithAltScreen())
	_, err = programa.Run()
	return err
}

func pipeRadar(client *gh.Client, format string) error {
	items, rate, err := client.FetchRadar()
	if err != nil && len(items) == 0 {
		return err
	}
	if err == nil {
		_ = cache.SaveRadar(items, rate)
	} else {
		fmt.Fprintln(os.Stderr, "aviso: radar parcial ("+err.Error()+")")
	}
	return output.WriteRadar(os.Stdout, items, format)
}

func pipeCount(client *gh.Client, format string, refresh bool, maxAge int) error {
	var items []gh.RadarItem
	ageSeconds := 0
	stale := false

	if !refresh {
		if snap, ok := cache.LoadRadar(); ok && snap.Age() <= time.Duration(maxAge)*time.Second {
			items = snap.Items
			ageSeconds = int(snap.Age().Seconds())
			return emitCount(format, items, ageSeconds, stale)
		}
	}

	fresh, rate, err := client.FetchRadar()
	if err != nil && len(fresh) == 0 {
		snap, ok := cache.LoadRadar()
		if !ok {
			return err
		}
		items = snap.Items
		ageSeconds = int(snap.Age().Seconds())
		stale = true
		return emitCount(format, items, ageSeconds, stale)
	}

	items = fresh
	if err == nil {
		_ = cache.SaveRadar(fresh, rate)
	}
	return emitCount(format, items, ageSeconds, stale)
}

func emitCount(format string, items []gh.RadarItem, ageSeconds int, stale bool) error {
	if format == "count-json" {
		return output.WriteCountJSON(os.Stdout, items, ageSeconds, stale)
	}
	return output.WriteCount(os.Stdout, items, stale)
}

func pipeMode(client *gh.Client, filters gh.Filters, format string) error {
	page, err := client.SearchAll(gh.BuildQuery(filters), gh.MaxSearchItems)
	if err != nil && len(page.Items) == 0 {
		return err
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "aviso: resultado parcial ("+err.Error()+")")
	}
	return output.Write(os.Stdout, page.Items, format)
}
