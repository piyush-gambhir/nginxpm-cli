# Nginx Proxy Manager CLI - Agent Guide

> **Contributors:** When you add or change commands, flags, auth, or output, update this file in the same PR as the code. Keep `SKILL.md` aligned if the CLI's described scope changes.

## Quick Reference

- **Binary:** `nginxpm`
- **Config file:** `~/.config/nginxpm-cli/config.yaml`
- **Env vars:** `NGINXPM_URL`, `NGINXPM_EMAIL`, `NGINXPM_PASSWORD`, `NGINXPM_INSECURE`
- **Auth method:** Email/password -> JWT token
- **Config priority:** CLI flags > environment variables > profile config > defaults

## Setup

```bash
# Interactive login (prompts for URL, email, password, TLS, profile name)
nginxpm login

# Or set environment variables for non-interactive use
export NGINXPM_URL=https://npm.example.com
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=changeme
```

## Output Formats

All list/get commands support three output formats via `-o`:

- `-o table` (default) -- human-readable tabular output
- `-o json` -- JSON, ideal for programmatic parsing with jq
- `-o yaml` -- YAML, useful for config management

**For agents:** Always use `-o json` when you need to parse or process output programmatically.

## Common Workflows

### Check server status

```bash
# Check server status (no auth required)
nginxpm status -o json
```

### Manage proxy hosts

```bash
# List all proxy hosts
nginxpm proxy list -o json

# Get a proxy host by ID
nginxpm proxy get 1 -o json

# Create a proxy host from a JSON file
nginxpm proxy create -f proxy.json

# Create from stdin
echo '{"domain_names":["app.example.com"],"forward_scheme":"http","forward_host":"192.168.1.10","forward_port":3000}' | nginxpm proxy create -f -

# Update a proxy host
nginxpm proxy update 1 -f proxy.json

# Delete a proxy host (requires confirmation)
nginxpm proxy delete 1

# Delete without confirmation
nginxpm proxy delete 1 --confirm

# Enable/disable a proxy host
nginxpm proxy enable 1
nginxpm proxy disable 1
```

### Manage redirection hosts

```bash
# List all redirection hosts
nginxpm redirect list -o json

# Get a redirection host
nginxpm redirect get 1 -o json

# Create a redirection host
nginxpm redirect create -f redirect.json

# Update a redirection host
nginxpm redirect update 1 -f redirect.json

# Delete a redirection host
nginxpm redirect delete 1 --confirm

# Enable/disable
nginxpm redirect enable 1
nginxpm redirect disable 1
```

### Manage TCP/UDP streams

```bash
# List all streams
nginxpm stream list -o json

# Get a stream
nginxpm stream get 5 -o json

# Create a stream
nginxpm stream create -f stream.json

# Update a stream
nginxpm stream update 5 -f stream.json

# Delete a stream
nginxpm stream delete 5 --confirm

# Enable/disable
nginxpm stream enable 5
nginxpm stream disable 5
```

### Manage 404 dead hosts

```bash
# List all dead hosts
nginxpm dead list -o json

# Get a dead host
nginxpm dead get 3 -o json

# Create a dead host
nginxpm dead create -f dead.json

# Update a dead host
nginxpm dead update 3 -f dead.json

# Delete a dead host
nginxpm dead delete 3 --confirm

# Enable/disable
nginxpm dead enable 3
nginxpm dead disable 3
```

### Manage SSL certificates

```bash
# List all certificates
nginxpm cert list -o json

# Get a certificate
nginxpm cert get 1 -o json

# Create a Let's Encrypt certificate
nginxpm cert create -f cert.json

# Renew a Let's Encrypt certificate
nginxpm cert renew 1

# Delete a certificate
nginxpm cert delete 1 --confirm

# List DNS challenge providers
nginxpm cert dns-providers

# Test HTTP reachability for domains
nginxpm cert test-http example.com www.example.com
```

### Manage access lists

```bash
# List all access lists
nginxpm access list -o json

# Get an access list
nginxpm access get 1 -o json

# Create an access list
nginxpm access create -f access.json

# Update an access list
nginxpm access update 1 -f access.json

# Delete an access list
nginxpm access delete 1 --confirm
```

### Manage users

```bash
# List all users
nginxpm user list -o json

# Get a user by ID
nginxpm user get 1 -o json

# Get current authenticated user
nginxpm user current -o json

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

### View audit log

```bash
# List audit log entries
nginxpm audit list -o json

# Get a specific audit log entry
nginxpm audit get 1 -o json
```

### Manage settings (default site)

```bash
# List all server settings
nginxpm setting list -o json

# Get a specific setting
nginxpm setting get default-site -o json

# Update a setting
nginxpm setting update default-site -f setting.json
```

### Configuration management

```bash
# View current configuration
nginxpm config view

# Set default output format
nginxpm config set defaults.output json

# List all profiles
nginxpm config list-profiles

# Switch to a different profile
nginxpm config use-profile staging
```

## Tips for Agents

- Always use `-o json` when you need to parse output programmatically.
- Use `--confirm` on destructive commands (delete) to skip interactive prompts.
- Use `--no-input` to disable all interactive prompts in CI/automation.
- For bulk operations: list with `-o json`, parse with jq, then loop over results.
- Many create/update commands require a `-f` flag pointing to a JSON or YAML file. Prepare the file first, then pass it.
- Use `-f -` to pipe content from stdin into any command that accepts a file.
- The `proxy` command has aliases: `proxy-host`, `ph`. The `redirect` command has aliases: `redirection`, `redir`, `rh`. The `stream` command has aliases: `streams`, `st`. The `dead` command has aliases: `dead-host`, `dh`, `404`. The `cert` command has aliases: `certificate`, `certificates`, `ssl`. The `access` command has aliases: `access-list`, `acl`. The `user` command has alias: `users`. The `audit` command has aliases: `audit-log`, `log`. The `setting` command has alias: `settings`.
- List subcommands often have an `ls` alias (e.g., `nginxpm proxy ls`, `nginxpm cert ls`).
- The `user current` command has aliases: `me`, `whoami`.
- The `status` command does not require authentication -- useful for checking server availability.
- Use `nginxpm config view` to check current connection settings and confirm which profile is active.
- The `cert dns-providers` output is always JSON regardless of `--output` setting.

## Complete Command Reference

### Top-level commands

| Command | Description |
|---------|-------------|
| `nginxpm login` | Interactively log in and save a connection profile |
| `nginxpm status` | Show server status, version, and setup state (no auth required) |
| `nginxpm version` | Print CLI version, commit, and build date |
| `nginxpm update` | Check for and install CLI updates (--check for check only) |
| `nginxpm completion` | Generate shell completion scripts (bash, zsh, fish, powershell) |

### `nginxpm config` -- Manage CLI configuration

| Command | Description |
|---------|-------------|
| `nginxpm config view` | Display the current configuration |
| `nginxpm config set <key> <value>` | Set a configuration value (defaults.output, current_profile) |
| `nginxpm config use-profile <name>` | Switch to a different profile |
| `nginxpm config list-profiles` | List all configured profiles |

### `nginxpm proxy` (aliases: `proxy-host`, `ph`) -- Manage proxy hosts

| Command | Description |
|---------|-------------|
| `nginxpm proxy list` (alias: `ls`) | List all proxy hosts |
| `nginxpm proxy get <id>` | Get a proxy host by ID |
| `nginxpm proxy create` | Create a proxy host (-f required) |
| `nginxpm proxy update <id>` | Update a proxy host (-f required) |
| `nginxpm proxy delete <id>` | Delete a proxy host (--confirm) |
| `nginxpm proxy enable <id>` | Enable a proxy host |
| `nginxpm proxy disable <id>` | Disable a proxy host |

### `nginxpm redirect` (aliases: `redirection`, `redir`, `rh`) -- Manage redirection hosts

| Command | Description |
|---------|-------------|
| `nginxpm redirect list` (alias: `ls`) | List all redirection hosts |
| `nginxpm redirect get <id>` | Get a redirection host by ID |
| `nginxpm redirect create` | Create a redirection host (-f required) |
| `nginxpm redirect update <id>` | Update a redirection host (-f required) |
| `nginxpm redirect delete <id>` | Delete a redirection host (--confirm) |
| `nginxpm redirect enable <id>` | Enable a redirection host |
| `nginxpm redirect disable <id>` | Disable a redirection host |

### `nginxpm stream` (aliases: `streams`, `st`) -- Manage TCP/UDP streams

| Command | Description |
|---------|-------------|
| `nginxpm stream list` (alias: `ls`) | List all TCP/UDP streams |
| `nginxpm stream get <id>` | Get a stream by ID |
| `nginxpm stream create` | Create a stream (-f required) |
| `nginxpm stream update <id>` | Update a stream (-f required) |
| `nginxpm stream delete <id>` | Delete a stream (--confirm) |
| `nginxpm stream enable <id>` | Enable a stream |
| `nginxpm stream disable <id>` | Disable a stream |

### `nginxpm dead` (aliases: `dead-host`, `dh`, `404`) -- Manage 404 dead hosts

| Command | Description |
|---------|-------------|
| `nginxpm dead list` (alias: `ls`) | List all 404 dead hosts |
| `nginxpm dead get <id>` | Get a dead host by ID |
| `nginxpm dead create` | Create a dead host (-f required) |
| `nginxpm dead update <id>` | Update a dead host (-f required) |
| `nginxpm dead delete <id>` | Delete a dead host (--confirm) |
| `nginxpm dead enable <id>` | Enable a dead host |
| `nginxpm dead disable <id>` | Disable a dead host |

### `nginxpm cert` (aliases: `certificate`, `certificates`, `ssl`) -- Manage SSL certificates

| Command | Description |
|---------|-------------|
| `nginxpm cert list` (alias: `ls`) | List all SSL certificates |
| `nginxpm cert get <id>` | Get a certificate by ID |
| `nginxpm cert create` | Create a certificate (-f required) |
| `nginxpm cert delete <id>` | Delete a certificate (--confirm) |
| `nginxpm cert renew <id>` | Renew a Let's Encrypt certificate |
| `nginxpm cert dns-providers` | List available DNS challenge providers |
| `nginxpm cert test-http <domain> [domain...]` | Test HTTP reachability for domains |

### `nginxpm access` (aliases: `access-list`, `acl`) -- Manage access lists

| Command | Description |
|---------|-------------|
| `nginxpm access list` (alias: `ls`) | List all access lists |
| `nginxpm access get <id>` | Get an access list by ID |
| `nginxpm access create` | Create an access list (-f required) |
| `nginxpm access update <id>` | Update an access list (-f required) |
| `nginxpm access delete <id>` | Delete an access list (--confirm) |

### `nginxpm user` (alias: `users`) -- Manage users

| Command | Description |
|---------|-------------|
| `nginxpm user list` (alias: `ls`) | List all users |
| `nginxpm user get <id>` | Get a user by ID (use "me" for current user) |
| `nginxpm user create` | Create a user (-f required) |
| `nginxpm user update <id>` | Update a user (-f required) |
| `nginxpm user delete <id>` | Delete a user (--confirm) |
| `nginxpm user current` (aliases: `me`, `whoami`) | Show current authenticated user |
| `nginxpm user permissions <id>` | Set user permissions (-f required) |
| `nginxpm user password <id>` | Change user password (-f required) |

### `nginxpm audit` (aliases: `audit-log`, `log`) -- View audit log

| Command | Description |
|---------|-------------|
| `nginxpm audit list` (alias: `ls`) | List all audit log entries |
| `nginxpm audit get <id>` | Get a specific audit log entry |

### `nginxpm setting` (alias: `settings`) -- Manage server settings

| Command | Description |
|---------|-------------|
| `nginxpm setting list` (alias: `ls`) | List all server settings |
| `nginxpm setting get <id>` | Get a specific setting (e.g., default-site) |
| `nginxpm setting update <id>` | Update a setting (-f required) |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: table (default), json, yaml |
| `--profile <name>` | Configuration profile to use |
| `--url <url>` | Nginx Proxy Manager URL override |
| `--email <email>` | Email for authentication override |
| `-p, --password <pass>` | Password for authentication override |
| `-k, --insecure` | Skip TLS certificate verification |
| `--no-input` | Disable all interactive prompts (for CI/agent use) |
| `-q, --quiet` | Suppress informational output |
| `-v, --verbose` | Enable verbose HTTP logging |
