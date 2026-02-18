# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`fornost` is a Go CLI tool for inspecting and managing Hetzner Cloud infrastructure. It runs standalone or as an internal library imported by the `arnor` suite. Named after the ancient capital of Arnor. Status: **Planned** (no implementation code yet).

## Tech Stack

- **Language:** Go
- **CLI framework:** Cobra (`github.com/spf13/cobra`)
- **Env loading:** `github.com/joho/godotenv`
- **Config:** Credentials from `~/.dotfiles/.env` with fallback to `.env`; requires `HETZNER_API_TOKEN`

## Build & Development Commands

```bash
go build -o fornost .          # Build binary
go run .                       # Run without building
go test ./...                  # Run all tests
go test ./internal/hetzner/    # Run tests for a specific package
go vet ./...                   # Static analysis
gofmt -w .                     # Format code
```

## Architecture

```
main.go              # Entry point, calls cmd.Execute()
cmd/
  root.go            # Env loading, cobra setup, Hetzner client init
  server.go          # server list, server get
  ssh.go             # ssh list, ssh add, ssh delete
  firewall.go        # firewall list, firewall get
internal/
  hetzner/
    client.go        # HTTP client wrapping Hetzner Cloud API v1
```

- **cmd/** — Each file registers subcommands on the root cobra command. `root.go` handles env loading (godotenv) and initializes the shared Hetzner API client.
- **internal/hetzner/** — API client used by commands and importable by `arnor`. All list operations must handle pagination via `meta.pagination.next_page`.

## Key Design Decisions

- Multi-project support: `--project prod` resolves to `HETZNER_API_TOKEN_PROD` env var
- API base: `https://api.hetzner.cloud/v1` with Bearer token auth
- Mirror structure and conventions of sibling project `shadowfax`
- Server status values: `running`, `initializing`, `starting`, `stopping`, `off`, `deleting`, `migrating`, `rebuilding`, `unknown`
- `ping` command validates token via `GET /actions` with low `per_page`
