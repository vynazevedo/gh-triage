<h1 align="center">gh-triage</h1>

<p align="center">
  Sua caixa de entrada do GitHub no terminal, ordenada por urgência, sem configurar nada.
</p>

<p align="center">
  <a href="https://github.com/vynazevedo/gh-triage/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/vynazevedo/gh-triage/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/vynazevedo/gh-triage/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/vynazevedo/gh-triage?sort=semver&style=flat-square&color=4C86F0"></a>
  <a href="LICENSE"><img alt="Licença MIT" src="https://img.shields.io/badge/licen%C3%A7a-MIT-3BE8B0?style=flat-square"></a>
  <img alt="Go 1.25" src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white">
  <a href="#instalação"><img alt="gh extension" src="https://img.shields.io/badge/gh-extension-2088FF?style=flat-square&logo=github&logoColor=white"></a>
  <img alt="TUI Bubble Tea" src="https://img.shields.io/badge/TUI-Bubble_Tea-FF75B7?style=flat-square">
</p>

---

Digite `gh triage` e veja, em um segundo, tudo do GitHub que está esperando
por você: PRs bloqueando colegas, mudanças pedidas no seu código, CI quebrado,
issues atribuídas e menções, em todos os seus repositórios e organizações,
ordenado por urgência. Sem configurar nada.

O Radar responde a pergunta que a página de notificações do GitHub não
responde: "o que eu preciso fazer agora?"

```
review pedido a você     pr    #384   acme/api          Anexo de documentos
review pedido a você     pr    #480   acme/service-web  Remoção de logs
atribuída a você         issue #109   acme/service-crm  Estrutura de tabelas
você foi mencionado      issue #436   acme/api          Refatoração da estrutura
```

E quando você mergulha num repositório, vira uma mesa de triagem completa:
issues, PRs e commits cruzados entre si, view profunda com comentários e
reviews, mutações em lote, tudo sem sair do terminal.

## Índice

- [Destaques](#destaques)
- [Instalação](#instalação)
- [Antes e depois](#antes-e-depois)
- [Uso](#uso)
- [Statusline](#statusline)
- [Relatórios](#relatórios)
- [Telas](#telas)
- [Teclas](#teclas)
- [Desenvolvimento](#desenvolvimento)
- [Estado atual](#estado-atual)
- [Licença](#licença)

## Destaques

- Radar de urgência entre todos os repos e organizações, sem YAML nem configuração.
- Início instantâneo: o radar é cacheado em disco e atualizado em segundo plano.
- `--watch` transforma o Radar num painel ao vivo, que se atualiza sozinho.
- Statusline compacta (`-o count`) para tmux, starship, sketchybar ou o prompt do shell.
- Relatórios `standup` e `blocking` prontos para colar na daily.
- Ranking ciente do tempo: o que espera há mais tempo sobe, e a idade ganha cor ao envelhecer.
- Modo pipe com registro JSON completo, para compor com `jq` e scripts.

## Instalação

```
gh extension install vynazevedo/gh-triage
```

Requer o `gh` autenticado (`gh auth login`). Os ícones usam Nerd Fonts;
sem a fonte, rode com `--nerd=false`.

## Antes e depois

O que antes exigia isto:

```
gh search issues --repo owner/repo --include-prs "avaliação" \
    --json number,title,isPullRequest,state \
    --jq '.[] | "\(if .isPullRequest then "PR " else "issue" end) #\(.number) [\(.state)] \(.title)"'
```

agora é isto:

```
gh triage -R owner/repo "avaliação" -o plain
```

E sem `-o`, abre a mesma busca numa TUI navegável.

## Uso

```
gh triage                                  # abre no Radar: o que espera por você
gh triage --watch                          # Radar ao vivo, atualiza sozinho
gh triage -R cli/cli                       # triagem de um repositório
gh triage -R cli                           # organização inteira (todos os repos)
gh triage -R cli/cli --state all "termo"   # filtros iniciais e termo de busca
gh triage --radar -o json                  # seu radar como JSON, para scripts
gh triage -R cli/cli -o plain              # linhas legíveis, sem TUI nem jq
gh triage -o count                         # resumo compacto para statusline
gh triage -o count-json                    # os números do radar como JSON
gh triage standup                          # entregue / em andamento / revisando / bloqueado
gh triage blocking                         # quem espera pelo seu review e por quem você espera
```

Flags em qualquer ordem, `--help` completo e autocompletion de shell
(`gh-triage completion bash|zsh|fish`).

Dentro da TUI, `.` troca de repositório ou organização a qualquer momento
(digite `owner/repo` para um repo ou só `owner` para a org inteira).

## Statusline

`gh triage -o count` imprime uma linha compacta do seu radar, feita para viver
no tmux, no starship, no sketchybar ou no prompt do zsh:

```
2 review · 1 ci · 9 atribuídas · 1 aguardando · 16 total
```

A saída é texto puro, sem cor, para você compor onde quiser. Para colorir por
conta própria, use `-o count-json`, que traz os números já separados por classe:

```json
{"review":2,"changes":0,"ci":1,"merge":1,"assigned":9,"mentions":2,
 "waiting":1,"drafts":0,"total":16,"urgent":3,"age_seconds":1,"stale":false}
```

`urgent` soma review, mudanças pedidas e CI falhando: o número que você quer num
único indicador. `stale` fica `true` quando a rede falhou e a linha veio de um
cache antigo (a linha compacta marca isso com um `*` no fim).

O radar é cacheado em disco (`os.UserCacheDir()/gh-triage/radar.json`), então
chamadas repetidas respondem na hora, sem tocar a rede. Por padrão o cache vale
60 segundos; ajuste com `--max-age <segundos>` ou force uma busca com `--refresh`.
Abrir a TUI ou rodar `--radar -o json` também aquece o mesmo cache.

Exemplo no tmux (`~/.tmux.conf`):

```
set -g status-interval 15
set -g status-right "#(gh triage -o count) "
```

## Relatórios

`gh triage standup` monta um resumo pronto para colar na daily a partir da sua
atividade: PRs merged na janela (`--since <dias>`, padrão 1), seus PRs abertos,
reviews pedidos a você e o que está bloqueado (mudanças pedidas ou CI vermelho).

`gh triage blocking` mostra os dois lados do bloqueio de review: quem pediu seu
review (pessoas esperando por você) e seus PRs ainda sem review (você esperando
por elas), ordenados pelo que espera há mais tempo.

Ambos aceitam `-o json`. A saída JSON de todas as superfícies (radar, triagem,
standup, blocking) traz o registro completo do item: `author`, `draft`, `ci`,
`review_decision`, `comments`, `linked`, `created_at` além do básico.

## Telas

Quatro abas, trocadas com `Tab` / `Shift+Tab` ou `1` `2` `3` `4`:

- Radar: tudo que espera por você, entre todos os repos, por urgência.
- Triagem: issues e PRs de um repo ou org na mesma lista.
- PRs: só pull requests, com coluna de review.
- Commits: histórico do branch padrão, com os PRs e issues que cada commit referencia.

A ordem do Radar: review pedido a você, mudanças pedidas nos seus PRs,
CI falhou, pronto para merge, issues atribuídas, menções, aguardando review.
Dentro de cada classe, o que espera há mais tempo sobe primeiro, e a idade
ganha cor conforme envelhece (amarelo após 3 dias, vermelho após 7). O preview
explica o porquê da urgência: o motivo, a idade e o sinal (CI vermelho, sem
review ainda, aprovado e pronto).

## Teclas

| Tecla | Ação |
|---|---|
| `Tab` / `Shift+Tab`, `1` `2` `3` | Trocar de aba |
| `j` `k` / setas, `g` `G` | Navegar / topo / fim |
| `Enter` | View completa (corpo, comentários, reviews) |
| `espaço`, `A` | Marcar / marcar todos |
| `/` | Busca fuzzy com ranking no carregado |
| `f` | Painel de filtros (tipo, estado, labels, milestone, assignee) |
| `.` | Picker de org/repo (lista das suas orgs para repos) |
| `>` | Trocar de escopo digitando (owner ou owner/repo) |
| `S` / `w` | Buscas salvas / salvar a busca atual |
| `L` `a` | Label / assignee na seleção |
| `x` `X` | Fechar / reabrir |
| `n` `c` `e` | Comentar / criar / editar (via `$EDITOR`) |
| `v` `M` `C` | Review / merge / checkout de PR |
| `o` `r` `?` | Browser / recarregar / ajuda |
| `q` `Esc` | Voltar ou sair |

## Desenvolvimento

Go 1.25. Stack: Bubble Tea (TEA), Lip Gloss (estilo), Bubbles (componentes),
Glamour (markdown), go-gh (auth e API oficiais do GitHub), Cobra + Fang (CLI),
sahilm/fuzzy (busca com ranking).

```
go build -o gh-triage .
go test ./...
go vet ./...
gofmt -l .
```

Estrutura:

```
main.go              detecção de TTY, dispatch pipe vs TUI, subcomandos
internal/gh/         go-gh: query, model, radar, urgency, standup, blocking, mutate
internal/app/        TEA: Model, Update, View, superfícies, modais
internal/ui/         lipgloss theme, ícones, markdown, render
internal/output/     pipe json/tsv/plain, statusline e relatórios
internal/cache/      cache do radar em disco (início instantâneo, statusline)
reference/rust/      implementação Rust original, mantida como especificação
```

## Estado atual

Implementado: Radar por urgência entre repos com cache em disco (início
instantâneo), ranking ciente do tempo e `--watch` ao vivo, statusline compacta
(`-o count` / `count-json`), relatórios `standup` e `blocking`, quatro abas com
cruzamento issue/PR/commit, view completa com comentários e reviews, painel de
filtros visual, paginação automática, mutações em lote com confirmação (label,
assignee, fechar, reabrir), comentar/criar/editar via `$EDITOR`,
review/merge/checkout de PR, buscas salvas persistidas, busca incremental, ajuda
e modo pipe com registro JSON completo. Overlays compostos de forma ANSI-aware.

Fora de escopo permanente: review de diff linha a linha e GitHub Enterprise Server.

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para o texto completo.
