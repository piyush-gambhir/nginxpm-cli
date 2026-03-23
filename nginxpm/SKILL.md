---
name: nginxpm
description: CLI for managing Nginx Proxy Manager -- proxy hosts, redirections, streams, SSL certificates, access lists, users, and settings.
---

# Nginx Proxy Manager CLI (`nginxpm`) -- agent skill

## Maintainer note

**When you change the CLI** (commands, flags, auth, config, output): update **`CLAUDE.md`** and any affected **`README.md` / `docs/`** in the same PR. If the change affects what this skill claims (scope, install, major workflows), update the YAML **`description`** above or the sections below. See **`CONTRIBUTING.md` -> Documentation and agent materials**.

## Canonical guide

The full command reference, workflows, and examples live in **`CLAUDE.md`** in this repository. Read that file for complete coverage.

## Quick orientation

| Item | Value |
|------|--------|
| Binary | `nginxpm` |
| Config | `~/.config/nginxpm-cli/config.yaml` |
| Machine-readable output | `-o json` (prefer for agents) |
| Non-interactive | set env vars (`NGINXPM_URL`, `NGINXPM_EMAIL`, `NGINXPM_PASSWORD`) and/or `--no-input` |

## Discovering commands

Cobra adds **`-h` / `--help`** on the root and every subcommand:

```bash
nginxpm --help
nginxpm proxy --help
nginxpm proxy list --help
```

Use this when the repo is unavailable or to confirm flags after upgrades.

## Install (reference)

```bash
go install github.com/piyush-gambhir/nginxpm-cli@latest
# or: clone repo && make install -- see README.md
```

## What it does

The `nginxpm` CLI provides complete management of Nginx Proxy Manager instances from the terminal. It covers:

- **Proxy hosts** -- create, list, get, update, delete, enable, disable reverse proxy entries
- **Redirection hosts** -- manage URL redirections (301/302)
- **Streams** -- manage TCP/UDP stream forwarding
- **Dead hosts** -- manage 404 catch-all hosts
- **SSL certificates** -- create, list, delete, renew Let's Encrypt certs; test HTTP reachability; list DNS providers
- **Access lists** -- manage HTTP basic auth and IP-based access control
- **Users** -- create, list, update, delete users; manage permissions and passwords
- **Audit log** -- view the action audit trail
- **Settings** -- view and update server settings (e.g., default site behavior)
- **Server status** -- check if the NPM instance is reachable (no auth required)

## Key commands

```bash
# Authenticate
nginxpm login

# Server status (no auth)
nginxpm status

# Proxy hosts
nginxpm proxy list -o json
nginxpm proxy get 1 -o json
nginxpm proxy create -f proxy.json
nginxpm proxy delete 1 --confirm

# Certificates
nginxpm cert list -o json
nginxpm cert create -f cert.json
nginxpm cert renew 1

# Users
nginxpm user list -o json
nginxpm user current -o json

# Configuration
nginxpm config list-profiles
nginxpm config use-profile prod
```

## Example use cases

1. **Inventory all proxy hosts as JSON for processing:**
   ```bash
   nginxpm proxy list -o json | jq '.[] | {id, domains: .domain_names}'
   ```

2. **Create a new proxy host from a template:**
   ```bash
   nginxpm proxy create -f proxy-template.json
   ```

3. **Renew all Let's Encrypt certificates:**
   ```bash
   nginxpm cert list -o json | jq '.[].id' | xargs -I{} nginxpm cert renew {}
   ```

4. **Disable a proxy host for maintenance:**
   ```bash
   nginxpm proxy disable 5
   ```

5. **Check server reachability in CI:**
   ```bash
   export NGINXPM_URL=http://npm.example.com:81
   nginxpm status -o json
   ```

6. **Audit recent actions:**
   ```bash
   nginxpm audit list -o json
   ```
