# chester

`chester` is named for Chesterton's Fence: before an agent deletes, rewrites, or "simplifies" a piece of code, it should first understand why that code exists. `chester` is a lightweight, stateless CLI for deterministic repository archaeology. It shells out to local `git` and the authenticated `gh` CLI, then emits compressed Markdown for LLM context windows so agents can pay down verification debt before they make changes.

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

Agents are the primary users. Run `chester onboard` and paste the emitted snippet into `AGENTS.md` (or `.github/copilot-instructions.md`) so coding agents know the anti-magic rule and the four core primitives.

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

Fetches a remote issue or pull request thread, strips GitHub UI noise, and prints:

- the original body
- a `Comments` section with chronological human comments
- a `Reviews` section for PR review summaries

If the ID is not a pull request, `chester` falls back to `gh issue view`.

### `chester file-history <path>`

Walks local history for one exact file with:

- `git log --follow --reverse`
- GitHub commit-to-PR association lookups
- deterministic fallback to direct commit messages when no PR exists

Adjacent commits that map to the same PR are collapsed into one timeline entry.

### `chester unearth-lines <file>:<start>:<end>`

Blames an exact line range with `git blame --line-porcelain`, resolves the introducing commits to PRs, and prints:

- line spans
- the blamed commit SHA
- the PR or direct-commit source
- top review comments for the exact file path only

If a blamed commit has no PR, `chester` falls back to the direct commit message.

### `chester unearth-range <from_ref>..<to_ref>`

Walks a git revision range with `git log --reverse`, resolves each commit to a PR, deduplicates PRs by first appearance, and renders a dense high-level list of architectural changes.

## Requirements

- Go 1.26+ for `go install` or source builds
- `git` must be installed
- `gh` must be installed and authenticated for GitHub
- `gh auth login` must have been completed before using remote-backed commands
- either:
  - the current repository has a GitHub `origin` remote
  - or `--repo owner/name` is provided

## Development

Use the provided `Makefile` so Go build cache writes stay inside the workspace:

```bash
make test
make build
./bin/chester --help
```
