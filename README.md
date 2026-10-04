# Nginx Proxy Manager CLI

A command-line interface for managing Nginx Proxy Manager instances -- proxy hosts, redirection hosts, streams, dead hosts, SSL certificates, access lists, users, and more.

Designed for both human operators and coding agents (LLMs). All commands support `--help` for detailed usage, and `-o json` / `-o yaml` for machine-readable output.

[![Go Version](https://img.shields.io/github/go-mod/go-version/piyush-gambhir/nginxpm-cli)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/piyush-gambhir/nginxpm-cli)](https://github.com/piyush-gambhir/nginxpm-cli/releases)
[![License](https://img.shields.io/github/license/piyush-gambhir/nginxpm-cli)](LICENSE)

## Features

- Full API coverage -- proxy hosts, redirections, streams, dead hosts, certificates, access lists, users, audit log, settings
- Multiple output formats -- table, JSON, YAML (`-o json`)
- Profile management -- multiple NPM instances with `--profile`
- Update notices -- in an interactive terminal, a once-a-day check says when a new release is out; `nginxpm update` installs it on macOS, Linux, and Windows
- Agent-friendly -- comprehensive help text, structured output for LLM coding agents (`CLAUDE.md` guide, `SKILL.md` for Cursor-style skills)
- Cross-platform -- macOS, Linux, Windows (amd64 and arm64)

## Installation

Install the latest release with the standalone installer (no Go required):

```bash
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/nginxpm-cli/main/install.sh | sh
```

Install a specific version:

```bash
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/nginxpm-cli/main/install.sh | VERSION=0.1.7 sh
```

The installer detects your OS/arch (macOS and Linux, amd64 and arm64) and installs the `nginxpm` binary to `/usr/local/bin` (override with `INSTALL_DIR`).

### Direct download

Download a prebuilt binary from the [releases page](https://github.com/piyush-gambhir/nginxpm-cli/releases), then extract and install it. For example, macOS arm64:

```bash
curl -sSfL -o nginxpm.tar.gz https://github.com/piyush-gambhir/nginxpm-cli/releases/download/v0.1.7/nginxpm-cli_darwin_arm64.tar.gz
tar -xzf nginxpm.tar.gz
sudo mv nginxpm /usr/local/bin/
```

### Alternative (build from source, requires Go)

```bash
git clone https://github.com/piyush-gambhir/nginxpm-cli.git
cd nginxpm-cli/cli-go && make install
```

### Updating

```bash
nginxpm update --check   # compare your version with the latest release (add -o json for scripts)
nginxpm update           # download, verify, and install it (asks first; --yes skips the prompt)
```

`nginxpm update` works on macOS, Linux, and Windows. It downloads the release archive for your platform, verifies its SHA-256 against the release's `checksums.txt`, and replaces the binary. On Windows the running `nginxpm.exe` is moved aside to `nginxpm.exe.old`, which the next run deletes. If the binary's directory is not writable, re-run with `sudo` or reinstall with the install script into a writable `INSTALL_DIR`. A source build in a Go bin directory (`$GOBIN`, `$GOPATH/bin`, `~/go/bin`) is not replaced: run `git pull && make install` in your clone instead. `--read-only` blocks installing (not `--check`).

In an interactive terminal, nginxpm checks GitHub for a new release at most once a day and, after the command's own output, prints a notice on stderr (once per release per day):

```text
A new version of nginxpm is available: v0.1.9 -> v0.1.10
Update with: nginxpm update
Release notes: https://github.com/piyush-gambhir/nginxpm-cli/releases/tag/v0.1.10
```

The check is skipped when stderr is not a terminal (pipes, scripts, coding agents), when `CI` is set, with `--quiet`, and for `update`, `version`, `completion`, and `help`. Turn it off with `NGINXPM_NO_UPDATE_NOTIFIER=1` or `NO_UPDATE_NOTIFIER=1`.

## Quick Start

```bash
# Authenticate
nginxpm login

# Check server status (no auth required)
nginxpm status

# List proxy hosts
nginxpm proxy list

# Get proxy host details as JSON
nginxpm proxy get 1 -o json

# Create a proxy host from a JSON file
nginxpm proxy create -f proxy.json
```

## Authentication

```bash
# Interactive login (saves profile to ~/.config/nginxpm-cli/config.yaml)
nginxpm login

# Environment variables
export NGINXPM_URL=https://npm.example.com
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=changeme
```

NPM uses email/password authentication. The CLI exchanges credentials for a JWT token, which is used for all subsequent API calls.

### Auth Priority

Configuration is resolved in this order (first match wins):

1. CLI flags (`--url`, `--email`, `--password`)
2. Environment variables (`NGINXPM_URL`, `NGINXPM_EMAIL`, `NGINXPM_PASSWORD`)
3. Config file profile (`~/.config/nginxpm-cli/config.yaml`)

## Configuration

### Profiles

```bash
# Interactive login creates a profile
nginxpm login

# List profiles
nginxpm config list-profiles

# Switch profiles
nginxpm config use-profile prod

# Use a profile for a single command
nginxpm proxy list --profile staging
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NGINXPM_URL` | Nginx Proxy Manager URL |
| `NGINXPM_EMAIL` | Email for authentication |
| `NGINXPM_PASSWORD` | Password for authentication |
| `NGINXPM_INSECURE` | Skip TLS verification (`true`/`1`) |
| `NGINXPM_NO_INPUT` | Disable interactive prompts |
| `NGINXPM_QUIET` | Suppress informational output |
| `NGINXPM_VERBOSE` | Enable verbose HTTP logging |
| `NGINXPM_NO_UPDATE_NOTIFIER` | Turn off the new-version notice (any value; `NO_UPDATE_NOTIFIER` also works) |

## Commands

| Group | Description | Aliases |
|-------|-------------|---------|
| `nginxpm proxy` | Manage proxy hosts | `proxy-host`, `ph` |
| `nginxpm redirect` | Manage redirection hosts | `redirection`, `redir`, `rh` |
| `nginxpm stream` | Manage TCP/UDP streams | `streams`, `st` |
| `nginxpm dead` | Manage 404 dead hosts | `dead-host`, `dh`, `404` |
| `nginxpm cert` | Manage SSL certificates | `certificate`, `certificates`, `ssl` |
| `nginxpm access` | Manage access lists | `access-list`, `acl` |
| `nginxpm user` | Manage users | `users` |
| `nginxpm audit` | View audit log | `audit-log`, `log` |
| `nginxpm setting` | Manage server settings | `settings` |
| `nginxpm config` | View/set configuration, manage profiles | |
| `nginxpm login` | Interactive authentication setup | |
| `nginxpm status` | Show server status (no auth required) | |
| `nginxpm version` | Print CLI version | |
| `nginxpm update` | Update to the latest release (`--check` only checks) | |
| `nginxpm completion` | Generate shell completions | |

## Proxy Hosts

```bash
# List all proxy hosts
nginxpm proxy list

# Get a proxy host by ID
nginxpm proxy get 1 -o json

# Create a proxy host from a file
nginxpm proxy create -f proxy.json

# Update a proxy host
nginxpm proxy update 1 -f proxy.json

# Delete a proxy host
nginxpm proxy delete 1 --confirm

# Enable / disable
nginxpm proxy enable 1
nginxpm proxy disable 1
```

## Redirection Hosts

```bash
nginxpm redirect list
nginxpm redirect get 1 -o json
nginxpm redirect create -f redirect.json
nginxpm redirect update 1 -f redirect.json
nginxpm redirect delete 1 --confirm
nginxpm redirect enable 1
nginxpm redirect disable 1
```

## Streams

```bash
nginxpm stream list
nginxpm stream get 1 -o json
nginxpm stream create -f stream.json
nginxpm stream update 1 -f stream.json
nginxpm stream delete 1 --confirm
nginxpm stream enable 1
nginxpm stream disable 1
```

## Dead Hosts

```bash
nginxpm dead list
nginxpm dead get 1 -o json
nginxpm dead create -f dead.json
nginxpm dead update 1 -f dead.json
nginxpm dead delete 1 --confirm
nginxpm dead enable 1
nginxpm dead disable 1
```

## SSL Certificates

```bash
# List all certificates
nginxpm cert list

# Get certificate details
nginxpm cert get 1 -o json

# Create a certificate
nginxpm cert create -f cert.json

# Delete a certificate
nginxpm cert delete 1 --confirm

# Renew a Let's Encrypt certificate
nginxpm cert renew 1

# List available DNS challenge providers
nginxpm cert dns-providers

# Test HTTP reachability for domains
nginxpm cert test-http example.com www.example.com
```

## Access Lists

```bash
nginxpm access list
nginxpm access get 1 -o json
nginxpm access create -f access.json
nginxpm access update 1 -f access.json
nginxpm access delete 1 --confirm
```

## Users

```bash
# List all users
nginxpm user list

# Get user details
nginxpm user get 1 -o json

# Show current authenticated user
nginxpm user current

# Create a user
nginxpm user create -f user.json

# Update a user
nginxpm user update 1 -f user.json

# Delete a user
nginxpm user delete 1 --confirm

# Set user permissions
nginxpm user permissions 1 -f permissions.json

# Change user password
nginxpm user password 1 -f password.json
```

## Audit Log

```bash
nginxpm audit list
nginxpm audit get 1 -o json
```

## Settings

```bash
nginxpm setting list
nginxpm setting get default-site -o json
nginxpm setting update default-site -f setting.json
```

## Output Formats

All list and get commands support three output formats:

```bash
nginxpm proxy list                  # table (default, human-readable)
nginxpm proxy list -o json          # JSON (machine-readable)
nginxpm proxy list -o yaml          # YAML
```

## Global Flags

These flags are available on every command:

| Flag | Description |
|------|-------------|
| `--output`, `-o` | Output format: `table`, `json`, `yaml` |
| `--profile` | Configuration profile to use |
| `--url` | Nginx Proxy Manager URL |
| `--email` | Email for authentication |
| `--password`, `-p` | Password for authentication |
| `--insecure`, `-k` | Skip TLS certificate verification |
| `--read-only` | Block all server-mutating commands (agent safety mode) |
| `--no-input` | Disable all interactive prompts (for CI/agent use) |
| `--quiet`, `-q` | Suppress informational output |
| `--verbose`, `-v` | Enable verbose HTTP logging |

## File Input Format

Commands that accept `--file/-f` support:
- JSON files (`.json`)
- YAML files (`.yaml`, `.yml`)
- Stdin (use `-f -` and pipe input)

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT

## Compatibility and maintenance

See [build requirements and upstream API compatibility](docs/COMPATIBILITY.md) for the checked API baseline and dependency update workflow.
