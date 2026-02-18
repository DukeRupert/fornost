# fornost

A Go CLI tool for inspecting and managing Hetzner Cloud infrastructure. Named after the ancient capital of Arnor. Part of the `arnor` suite.

## Install

```bash
go install github.com/dukerupert/fornost@latest
```

Or build from source:

```bash
go build -o fornost .
```

## Configuration

Set your Hetzner Cloud API token in `~/.dotfiles/.env` or a local `.env` file:

```env
HETZNER_API_TOKEN=your_api_token_here
```

API tokens are created in the Hetzner Cloud Console under Project > Security > API Tokens.

## Usage

```bash
# Verify credentials
fornost ping

# Servers
fornost server list
fornost server get <name-or-id>

# SSH Keys
fornost ssh list
fornost ssh add --name my-key --key ~/.ssh/id_ed25519.pub
fornost ssh delete <name-or-id>

# Firewalls
fornost firewall list
fornost firewall get <name-or-id>
```

## Development

```bash
go run .              # Run without building
go test ./...         # Run all tests
go vet ./...          # Static analysis
gofmt -w .            # Format code
```
