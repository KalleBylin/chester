# chester

`chester` is named for Chesterton's Fence: before an agent deletes, rewrites, or "simplifies" a piece of code, it should first understand why that code exists. `chester` is a lightweight, stateless CLI for deterministic repository archaeology. It shells out to local `git` first, then layers in GitHub data through the authenticated `gh` CLI when it is available. The result is compressed Markdown for humans and deterministic JSON for agents.

## Install

### Go Install

```bash
go install github.com/KalleBylin/chester@latest
chester --help
```

### From Source

```bash
git clone git@github.com:KalleBylin/chester.git
cd chester
make test
make build
./bin/chester --help
```

## Shell Completions

`chester` ships Cobra-generated completions for `bash`, `zsh`, `fish`, and `powershell`.

```bash
chester completion bash > /etc/bash_completion.d/chester
chester completion zsh > "${fpath[1]}/_chester"
chester completion fish > ~/.config/fish/completions/chester.fish
```

For a one-off shell session, you can also source them directly:

```bash
source <(chester completion bash)
source <(chester completion zsh)
chester completion fish | source
```

## Agent Onboarding

Agents are the primary users. Run `chester onboard` and paste the emitted snippet into `AGENTS.md` (or `.github/copilot-instructions.md`) so coding agents know the anti-magic rule and the five core primitives.

## Principles

- No LLMs under the hood
- No heuristic file guessing
- No local cache, daemon, or database
- No repository mutation

## Commands

### `chester onboard`

Prints a minimal snippet for `AGENTS.md` that teaches agents when and how to use `chester` without bloating the repository's agent instructions.

### `chester completion <bash|zsh|fish|powershell>`

Generates shell completion scripts using Cobra's built-in completion support.

### `chester read-thread <id>`

Fetches a remote issue or pull request thread and prints the body, comments, reviews, and inline review comments. If the ID is not a pull request, `chester` falls back to `gh issue view`.

### `chester why-file <path>`

Walks local history for one exact file, collapsing adjacent commits from the same PR and enriching with GitHub PR context when available.

### `chester why-lines <file>:<start>:<end>`

Blames an exact line range, resolves the introducing commits to PRs, and keeps file-scoped review comments separate from the span summaries.

### `chester why-range <from_ref>..<to_ref>`

Walks a git revision range and renders a chronological summary with explicit summary provenance.

### `chester text-history <literal>`

Walks history for one exact string literal, optionally scoped to one path with `--path`.

## Output Modes

All archaeology commands support `--json`. Markdown is the default for direct reading. JSON is for wrappers and agents that need to chain `chester` without scraping prose.

## Requirements

- Go 1.26+ for `go install` or source builds
- `git` must be installed
- `gh` is optional for local archaeology commands and required for `read-thread`
- `gh auth login` must be completed before using GitHub-backed enrichment
- for GitHub-backed enrichment, either:
  - the current repository has a GitHub `origin` remote
  - or `--repo owner/name` is provided

## Development

Use the provided `Makefile` so Go build cache writes stay inside the workspace:

```bash
make test
make build
./bin/chester --help
```
