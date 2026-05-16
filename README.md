# gop1x

Developer workstation provisioner & preset toolkit. Set up terminal, shell, tools, and dev environment in one command.

## Install

```bash
go install github.com/ohp1x/gop1x@latest
```

## Usage

```bash
gop1x init                  # Initialize ~/.ohp1x/
gop1x sync                  # Converge state to ohp1x.yaml
gop1x add tools/fzf         # Install a preset
gop1x rm tools/fzf          # Remove a preset
gop1x status                # Show installed presets & drift
gop1x doctor                # Validate environment
```

Global flags: `--dry-run` (`-n`), `--yes` (`-y`)

## Development

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run go vet
make clean      # Remove binary
```

Requires Go 1.22+.

## License

MIT
