# nginxpm -- Command Reference

Quick reference of all commands. For full details, see `CLAUDE.md` or run `nginxpm <command> --help`.

## Top-level commands

| Command | Description |
|---------|-------------|
| `nginxpm login` | Interactively log in and save a connection profile |
| `nginxpm status` | Show server status (no auth required) |
| `nginxpm version` | Print CLI version, commit, and build date |
| `nginxpm update` | Check for and install CLI updates (`--check` for check only) |
| `nginxpm completion` | Generate shell completion scripts |

## `nginxpm config` -- Manage CLI configuration

| Command | Description |
|---------|-------------|
| `nginxpm config view` | Display the current configuration |
| `nginxpm config set <key> <value>` | Set a config value (`defaults.output`, `current_profile`) |
| `nginxpm config use-profile <name>` | Switch to a different profile |
| `nginxpm config list-profiles` | List all configured profiles |

## `nginxpm proxy` (aliases: `proxy-host`, `ph`) -- Manage proxy hosts

| Command | Description |
|---------|-------------|
| `nginxpm proxy list` (alias: `ls`) | List all proxy hosts |
| `nginxpm proxy get <id>` | Get proxy host details |
| `nginxpm proxy create` | Create a proxy host (`-f` required) |
| `nginxpm proxy update <id>` | Update a proxy host (`-f` required) |
| `nginxpm proxy delete <id>` | Delete a proxy host (`--confirm`) |
| `nginxpm proxy enable <id>` | Enable a proxy host |
| `nginxpm proxy disable <id>` | Disable a proxy host |

## `nginxpm redirect` (aliases: `redirection`, `redir`, `rh`) -- Manage redirection hosts

| Command | Description |
|---------|-------------|
| `nginxpm redirect list` (alias: `ls`) | List all redirection hosts |
| `nginxpm redirect get <id>` | Get redirection host details |
| `nginxpm redirect create` | Create a redirection host (`-f` required) |
| `nginxpm redirect update <id>` | Update a redirection host (`-f` required) |
| `nginxpm redirect delete <id>` | Delete a redirection host (`--confirm`) |
| `nginxpm redirect enable <id>` | Enable a redirection host |
| `nginxpm redirect disable <id>` | Disable a redirection host |

## `nginxpm stream` (aliases: `streams`, `st`) -- Manage TCP/UDP streams

| Command | Description |
|---------|-------------|
| `nginxpm stream list` (alias: `ls`) | List all streams |
| `nginxpm stream get <id>` | Get stream details |
| `nginxpm stream create` | Create a stream (`-f` required) |
| `nginxpm stream update <id>` | Update a stream (`-f` required) |
| `nginxpm stream delete <id>` | Delete a stream (`--confirm`) |
| `nginxpm stream enable <id>` | Enable a stream |
| `nginxpm stream disable <id>` | Disable a stream |

## `nginxpm dead` (aliases: `dead-host`, `dh`, `404`) -- Manage dead hosts

| Command | Description |
|---------|-------------|
| `nginxpm dead list` (alias: `ls`) | List all dead hosts |
| `nginxpm dead get <id>` | Get dead host details |
| `nginxpm dead create` | Create a dead host (`-f` required) |
| `nginxpm dead update <id>` | Update a dead host (`-f` required) |
| `nginxpm dead delete <id>` | Delete a dead host (`--confirm`) |
| `nginxpm dead enable <id>` | Enable a dead host |
| `nginxpm dead disable <id>` | Disable a dead host |

## `nginxpm cert` (aliases: `certificate`, `certificates`, `ssl`) -- Manage SSL certificates

| Command | Description |
|---------|-------------|
| `nginxpm cert list` (alias: `ls`) | List all certificates |
| `nginxpm cert get <id>` | Get certificate details |
| `nginxpm cert create` | Create a certificate (`-f` required) |
| `nginxpm cert delete <id>` | Delete a certificate (`--confirm`) |
| `nginxpm cert renew <id>` | Renew a Let's Encrypt certificate |
| `nginxpm cert dns-providers` | List available DNS challenge providers |
| `nginxpm cert test-http <domain> [domain...]` | Test HTTP reachability for domains |

## `nginxpm access` (aliases: `access-list`, `acl`) -- Manage access lists

| Command | Description |
|---------|-------------|
| `nginxpm access list` (alias: `ls`) | List all access lists |
| `nginxpm access get <id>` | Get access list details |
| `nginxpm access create` | Create an access list (`-f` required) |
| `nginxpm access update <id>` | Update an access list (`-f` required) |
| `nginxpm access delete <id>` | Delete an access list (`--confirm`) |

## `nginxpm user` (alias: `users`) -- Manage users

| Command | Description |
|---------|-------------|
| `nginxpm user list` (alias: `ls`) | List all users |
| `nginxpm user get <id>` | Get user details |
| `nginxpm user create` | Create a user (`-f` required) |
| `nginxpm user update <id>` | Update a user (`-f` required) |
| `nginxpm user delete <id>` | Delete a user (`--confirm`) |
| `nginxpm user current` (aliases: `me`, `whoami`) | Show current authenticated user |
| `nginxpm user permissions <id>` | Set user permissions (`-f` required) |
| `nginxpm user password <id>` | Change user password (`-f` required) |

## `nginxpm audit` (aliases: `audit-log`, `log`) -- View audit log

| Command | Description |
|---------|-------------|
| `nginxpm audit list` (alias: `ls`) | List audit log entries |
| `nginxpm audit get <id>` | Get audit log entry details |

## `nginxpm setting` (alias: `settings`) -- Manage server settings

| Command | Description |
|---------|-------------|
| `nginxpm setting list` (alias: `ls`) | List all settings |
| `nginxpm setting get <id>` | Get a setting value |
| `nginxpm setting update <id>` | Update a setting (`-f` required) |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table` (default), `json`, `yaml` |
| `--profile <name>` | Configuration profile to use |
| `--url <url>` | Nginx Proxy Manager URL override |
| `--email <email>` | Email for authentication override |
| `-p, --password <pass>` | Password for authentication override |
| `-k, --insecure` | Skip TLS certificate verification |
| `--no-input` | Disable all interactive prompts (for CI/agent use) |
| `-q, --quiet` | Suppress informational output |
| `-v, --verbose` | Enable verbose HTTP logging |
