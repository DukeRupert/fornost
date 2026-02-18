# fornost — Project Scope

## Overview

`fornost` is a CLI tool for inspecting and managing Hetzner Cloud infrastructure. It is designed to be used standalone or as a component of the `arnor` infrastructure management suite.

Named after the ancient capital of Arnor — the administrative seat from which the northern kingdom was managed.

## Repository

`github.com/fireflysoftware/fornost`

## Status

**Planned**

## Technology

- **Language:** Go
- **CLI framework:** Cobra
- **Key dependencies:** `github.com/spf13/cobra`, `github.com/joho/godotenv`

## Configuration

Credentials are loaded from `~/.dotfiles/.env` with fallback to `.env` in the current directory.

Required environment variables:

```env
HETZNER_API_TOKEN=your_api_token
```

> API tokens are scoped per Hetzner project. To manage multiple projects, use named tokens:

```env
HETZNER_API_TOKEN_PROD=your_prod_token
HETZNER_API_TOKEN_DEV=your_dev_token
```

> The `--project` flag (or `HETZNER_PROJECT` env var) selects which token to use.

## Project Structure

```
fornost/
├── main.go
├── cmd/
│   ├── root.go         # env loading, cobra setup, client init
│   ├── server.go       # server subcommands
│   ├── ssh.go          # ssh key subcommands
│   └── firewall.go     # firewall subcommands
├── internal/
│   └── hetzner/
│       └── client.go   # Hetzner API client
├── .env.example
├── ROADMAP.md
└── README.md
```

## Commands

### v1.0.0

#### Server

| Command | Description |
|---|---|
| `server list` | List all servers in the project |
| `server get` | Get details for a specific server by name or ID |

#### SSH Keys

| Command | Description |
|---|---|
| `ssh list` | List all SSH keys in the project |
| `ssh add` | Upload a new SSH key |
| `ssh delete` | Delete an SSH key by name or ID |

#### Firewall

| Command | Description |
|---|---|
| `firewall list` | List all firewalls in the project |
| `firewall get` | Get rules for a specific firewall by name or ID |

#### General

| Command | Description |
|---|---|
| `ping` | Verify credentials and return account info |

### v1.1.0 — Extended Reading

| Command | Description |
|---|---|
| `server list --project all` | List servers across all projects |
| `network list` | List networks in the project |
| `location list` | List available Hetzner datacenters |
| `image list` | List available server images and snapshots |

### v2.0.0 — Quality of Life

- `--output json` flag for machine-readable output
- `--quiet` flag for scripting
- `--project` flag on all commands to switch between Hetzner projects
- Shell autocompletion (bash, zsh, fish)
- Config file support via Viper (`~/.config/fornost/config.yaml`)

## Flags — `server list`

| Flag | Required | Default | Description |
|---|---|---|---|
| `--project` | no | `$HETZNER_PROJECT` | Project token alias to use |
| `--output` | no | table | Output format (`table`, `json`) |

## Flags — `ssh add`

| Flag | Required | Default | Description |
|---|---|---|---|
| `--name` | yes | — | Name for the key in Hetzner |
| `--key` | yes | — | Path to public key file |

## Server Output

`server list` should display:

| Field | Description |
|---|---|
| ID | Hetzner server ID |
| Name | Server name |
| Status | `running`, `off`, etc. |
| IP | Public IPv4 address |
| Type | Server type (e.g. `cx22`) |
| Location | Datacenter location |
| Created | Creation date |

## API Reference

Base URL: `https://api.hetzner.cloud/v1`

Authentication: `Authorization: Bearer {token}` header

Key endpoints:

```
GET /servers                  # list servers
GET /servers/{id}             # get server details
GET /ssh_keys                 # list SSH keys
POST /ssh_keys                # add SSH key
DELETE /ssh_keys/{id}         # delete SSH key
GET /firewalls                # list firewalls
GET /firewalls/{id}           # get firewall rules
GET /datacenters              # list datacenters
```

## Multi-Project Support

Hetzner tokens are scoped per project. `fornost` handles multiple projects by mapping named aliases to tokens in `~/.dotfiles/.env`:

```env
HETZNER_API_TOKEN_PROD=token_for_prod
HETZNER_API_TOKEN_DEV=token_for_dev
```

The `--project` flag accepts the alias suffix (e.g. `--project prod` resolves to `HETZNER_API_TOKEN_PROD`). If no `--project` flag is provided, `HETZNER_API_TOKEN` is used as the default.

## Integration

When used as part of `arnor`, `fornost`'s client is imported as an internal library. A key use case in `arnor project create` is looking up the VPS IP by server name rather than requiring the user to know it:

```bash
# Instead of prompting for VPS IP, arnor can look it up
arnor project create --server my-vps --domain myapp.example.com --port 8080
```

## Notes for Implementing Agents

- Hetzner API responses are paginated — implement pagination handling in all list operations using the `meta.pagination` field in responses
- The API returns `next_page` in the pagination meta — follow it until `next_page` is null
- Server status values include: `running`, `initializing`, `starting`, `stopping`, `off`, `deleting`, `migrating`, `rebuilding`, `unknown`
- SSH key fingerprints are returned by the API — display them in `ssh list` output for verification
- Mirror the structure and conventions of `shadowfax` as closely as possible for consistency across the suite
- The `ping` command should call `GET /actions` with a low `per_page` value — a successful response confirms the token is valid
