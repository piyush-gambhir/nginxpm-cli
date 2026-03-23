# Nginx Proxy Manager CLI - Authentication & Credentials Guide

This guide covers every authentication scenario supported by the Nginx Proxy Manager CLI (`nginxpm`). Whether you are running a local NPM instance, a Docker deployment, or managing multiple servers, this document walks you through obtaining credentials, configuring TLS, and connecting securely.

---

## Table of Contents

- [Quick Start](#quick-start)
- [How Authentication Works](#how-authentication-works)
- [Getting Your Credentials](#getting-your-credentials)
- [Configuration](#configuration)
  - [Config File](#config-file)
  - [Environment Variables](#environment-variables)
  - [CLI Flags](#cli-flags)
  - [Multiple Profiles](#multiple-profiles)
  - [Resolution Order](#resolution-order)
- [TLS / SSL Configuration](#tls--ssl-configuration)
- [Deployment Scenarios](#deployment-scenarios)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

There are three ways to authenticate. Pick whichever fits your workflow.

### 1. Interactive Login (recommended for first-time setup)

```bash
nginxpm login
```

The CLI will prompt you for:

1. **Nginx Proxy Manager URL** -- e.g. `http://localhost:81`
2. **Email** -- the admin email address
3. **Password** -- the admin password
4. **Skip TLS verification** -- answer `y` only in development
5. **Profile name** -- defaults to `default`

The CLI tests the connection by checking server status, then authenticates with the provided credentials. On success, the profile is written to `~/.config/nginxpm-cli/config.yaml` and set as the active profile.

### 2. Environment Variables (recommended for CI/CD)

```bash
export NGINXPM_URL=http://npm.example.com:81
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=changeme

# Then run any command -- no login needed
nginxpm proxy list
```

### 3. CLI Flags (one-off commands)

```bash
nginxpm proxy list \
  --url http://npm.example.com:81 \
  --email admin@example.com \
  --password changeme
```

Flags override environment variables, which override profile config. See [Resolution Order](#resolution-order) for full details.

---

## How Authentication Works

Nginx Proxy Manager uses **email/password authentication** that produces a **JWT token**.

1. The CLI sends a `POST /api/tokens` request with `{ "identity": "<email>", "secret": "<password>" }`.
2. NPM responds with a JWT token.
3. The CLI uses this token as a `Bearer` token in the `Authorization` header for all subsequent API requests.

This authentication happens automatically on every CLI invocation. The email and password are resolved from flags, environment variables, or the config file profile -- the CLI handles the token exchange transparently.

> **Note:** The CLI does not store JWT tokens. It stores the email and password in the profile config file and re-authenticates on each command invocation.

---

## Getting Your Credentials

### Default Admin Account

When Nginx Proxy Manager is first installed, it creates a default admin account:

- **Email:** `admin@example.com`
- **Password:** `changeme`

You should change these immediately after first login (either via the NPM web UI or via the CLI).

### Docker Deployment

Most NPM installations run via Docker:

```bash
docker run -d \
  --name npm \
  -p 80:80 \
  -p 81:81 \
  -p 443:443 \
  jc21/nginx-proxy-manager:latest
```

The admin panel is available at `http://localhost:81`. Use the default credentials above for first login. The API endpoint for the CLI is the same URL:

```bash
nginxpm login
# URL: http://localhost:81
# Email: admin@example.com
# Password: changeme
```

### Creating Additional Users

Via the CLI:

```bash
# Create a user from a JSON file
nginxpm user create -f user.json
```

Where `user.json` contains:

```json
{
  "name": "CLI User",
  "nickname": "cli",
  "email": "cli@example.com",
  "roles": ["admin"],
  "is_disabled": false,
  "auth": {
    "type": "password",
    "secret": "s3cure-pa55w0rd"
  }
}
```

---

## Configuration

### Config File

The config file lives at `~/.config/nginxpm-cli/config.yaml` (or `$XDG_CONFIG_HOME/nginxpm-cli/config.yaml` if `XDG_CONFIG_HOME` is set).

Example:

```yaml
current_profile: default
profiles:
  default:
    url: http://localhost:81
    email: admin@example.com
    password: changeme
    insecure: false
  production:
    url: https://npm.example.com
    email: admin@example.com
    password: pr0d-p4ssw0rd
    insecure: false
defaults:
  output: table
```

The file is created automatically by `nginxpm login`. Permissions are set to `0600` (owner read/write only).

### Environment Variables

| Variable | Description |
|----------|-------------|
| `NGINXPM_URL` | Nginx Proxy Manager URL (e.g. `http://localhost:81`) |
| `NGINXPM_EMAIL` | Email for authentication |
| `NGINXPM_PASSWORD` | Password for authentication |
| `NGINXPM_INSECURE` | Skip TLS verification (`true` or `1`) |
| `NGINXPM_NO_INPUT` | Disable interactive prompts |
| `NGINXPM_QUIET` | Suppress informational output |
| `NGINXPM_VERBOSE` | Enable verbose HTTP logging |

### CLI Flags

| Flag | Description |
|------|-------------|
| `--url` | Nginx Proxy Manager URL |
| `--email` | Email for authentication |
| `--password`, `-p` | Password for authentication |
| `--insecure`, `-k` | Skip TLS certificate verification |
| `--profile` | Use a specific named profile |

### Multiple Profiles

You can manage connections to multiple NPM instances using named profiles:

```bash
# Create profiles via interactive login
nginxpm login
# → saves as "default"

nginxpm login
# → saves as "production"

# List all profiles
nginxpm config list-profiles

# Switch the active profile
nginxpm config use-profile production

# Use a profile for a single command
nginxpm proxy list --profile staging
```

### Resolution Order

Configuration is resolved in this order (first match wins):

1. **CLI flags** (`--url`, `--email`, `--password`, `--insecure`)
2. **Environment variables** (`NGINXPM_URL`, `NGINXPM_EMAIL`, `NGINXPM_PASSWORD`, `NGINXPM_INSECURE`)
3. **Config file profile** (`~/.config/nginxpm-cli/config.yaml`, using the current profile or `--profile`)

---

## TLS / SSL Configuration

If your NPM instance uses HTTPS:

```bash
# Skip TLS verification (development only)
nginxpm proxy list --insecure

# Or via environment variable
export NGINXPM_INSECURE=true

# Or save it in a profile during login
nginxpm login
# → Skip TLS verification? (y/N): y
```

---

## Deployment Scenarios

### Local Docker (HTTP)

```bash
export NGINXPM_URL=http://localhost:81
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=changeme
nginxpm proxy list
```

### Remote Server (HTTPS with valid certificate)

```bash
export NGINXPM_URL=https://npm.example.com
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=s3cure-p4ss
nginxpm proxy list
```

### Remote Server (HTTPS with self-signed certificate)

```bash
export NGINXPM_URL=https://npm.internal:81
export NGINXPM_EMAIL=admin@example.com
export NGINXPM_PASSWORD=s3cure-p4ss
export NGINXPM_INSECURE=true
nginxpm proxy list
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
env:
  NGINXPM_URL: ${{ secrets.NGINXPM_URL }}
  NGINXPM_EMAIL: ${{ secrets.NGINXPM_EMAIL }}
  NGINXPM_PASSWORD: ${{ secrets.NGINXPM_PASSWORD }}
  NGINXPM_NO_INPUT: "true"

steps:
  - name: List proxy hosts
    run: nginxpm proxy list -o json
```

---

## Security Best Practices

1. **Change default credentials immediately** after installing NPM. The default `admin@example.com` / `changeme` is well-known.
2. **Use environment variables or a config file** instead of passing credentials as CLI flags -- flags may appear in shell history and process listings.
3. **Restrict config file permissions.** The CLI writes the config with `0600` permissions, but verify this if you create the file manually.
4. **Use HTTPS in production.** Run NPM behind a reverse proxy with a valid TLS certificate, or enable HTTPS directly. Avoid `--insecure` in production.
5. **Create dedicated users** with appropriate roles instead of sharing the admin account.
6. **Use `--no-input`** in automation to prevent the CLI from blocking on interactive prompts.
7. **Rotate passwords regularly** and update profiles with `nginxpm login` or by editing the config file.

---

## Troubleshooting

### "authentication failed" error

- Verify the email and password are correct by logging in to the NPM web UI.
- Check that the URL points to the correct NPM instance and port (default admin port is `81`).
- If using HTTPS with a self-signed certificate, add `--insecure` or set `NGINXPM_INSECURE=true`.

### "connection test failed" error

- Ensure the NPM server is running: `nginxpm status --url http://localhost:81`.
- Check firewall rules and network connectivity.
- Verify the URL includes the correct port.

### "URL is required" error

- Set the URL via `--url`, `NGINXPM_URL`, or configure a profile with `nginxpm login`.

### "empty token in response" error

- This may indicate the NPM version requires 2FA or the credentials are incorrect.
- Check the NPM logs for more details.
