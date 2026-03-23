# Contributing to Nginx Proxy Manager CLI

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go version matching `go.mod` (see the `go` directive)
- Make
- Git

### Clone and Build

```bash
git clone https://github.com/piyush-gambhir/nginxpm-cli.git
cd nginxpm-cli
make build
```

### Run Locally

```bash
./nginxpm --help
./nginxpm version
```

### Run Tests

```bash
make test
```

### Lint

```bash
make lint    # requires golangci-lint
make vet     # go vet
make fmt     # gofmt
```

## Project Structure

```
.
├── main.go                 # Entry point
├── cmd/                    # Cobra command definitions
│   ├── root.go             # Root command, global flags, auth resolution
│   ├── login.go            # Interactive login flow
│   ├── status.go           # Server status (no auth required)
│   ├── version.go          # Version command
│   ├── update.go           # Self-update command
│   ├── completion.go       # Shell completion generation
│   ├── config/             # CLI config management (view, set, use-profile, list-profiles)
│   ├── proxy/              # Proxy host commands (list, get, create, update, delete, enable, disable)
│   │   ├── proxy.go        # Parent command registration
│   │   ├── list.go
│   │   ├── get.go
│   │   ├── create.go
│   │   ├── update.go
│   │   ├── delete.go
│   │   ├── enable.go
│   │   └── disable.go
│   ├── redirect/           # Redirection host commands (list, get, create, update, delete, enable, disable)
│   ├── stream/             # TCP/UDP stream commands (list, get, create, update, delete, enable, disable)
│   ├── dead/               # Dead host commands (list, get, create, update, delete, enable, disable)
│   ├── cert/               # Certificate commands (list, get, create, delete, renew, dns-providers, test-http)
│   ├── access/             # Access list commands (list, get, create, update, delete)
│   ├── user/               # User commands (list, get, create, update, delete, current, permissions, password)
│   ├── audit/              # Audit log commands (list, get)
│   └── setting/            # Setting commands (list, get, update)
├── internal/
│   ├── client/             # HTTP API client
│   │   ├── client.go       # Base client (JWT auth, HTTP methods)
│   │   ├── transport.go    # Custom HTTP transport (TLS, verbose logging)
│   │   ├── response.go     # Response handling
│   │   ├── errors.go       # Error types and helpers
│   │   ├── proxy.go        # Proxy host API methods
│   │   ├── redirect.go     # Redirection host API methods
│   │   ├── stream.go       # Stream API methods
│   │   ├── dead.go         # Dead host API methods
│   │   ├── cert.go         # Certificate API methods
│   │   ├── access.go       # Access list API methods
│   │   ├── user.go         # User API methods
│   │   ├── audit.go        # Audit log API methods
│   │   ├── setting.go      # Setting API methods
│   │   ├── status.go       # Server status API methods
│   │   └── report.go       # Report API methods
│   ├── cmdutil/            # Shared command utilities
│   │   ├── factory.go      # Factory struct (client, config, IO streams)
│   │   ├── flags.go        # Shared flag helpers (AddFileFlag, AddConfirmFlag, etc.)
│   │   ├── file.go         # File input reading/unmarshaling
│   │   └── errors.go       # Error types
│   ├── config/             # Config file and auth resolution
│   │   ├── config.go       # Config struct, Load, Save, profile management
│   │   ├── auth.go         # ResolvedConfig, Resolve (flags > env > profile)
│   │   └── paths.go        # Config directory and file paths
│   ├── output/             # Output formatting
│   │   ├── formatter.go    # Print dispatcher (table/json/yaml)
│   │   ├── table.go        # Table formatter
│   │   ├── json.go         # JSON formatter
│   │   ├── yaml.go         # YAML formatter
│   │   └── errors.go       # Output error handling
│   ├── build/              # Build version info (Version, Commit, Date)
│   └── update/             # Self-update check logic
├── Makefile
├── .goreleaser.yaml
└── go.mod
```

## Adding a New Command

1. **Add the API method** in `internal/client/<resource>.go`:
   ```go
   func (c *Client) ListWidgets(ctx context.Context) ([]Widget, error) {
       // HTTP call to the NPM API
   }
   ```

2. **Create the command** in `cmd/<resource>/list.go`:
   ```go
   func newCmdWidgetList(f *cmdutil.Factory) *cobra.Command {
       cmd := &cobra.Command{
           Use:   "list",
           Short: "List widgets",
           Long:  `List all widgets...`,
           // Add Aliases for common abbreviations (e.g., "ls")
           Aliases: []string{"ls"},
           // Mark mutating commands with annotations
           // Annotations: map[string]string{"mutates": "true"},
           RunE: func(cmd *cobra.Command, args []string) error {
               c, err := f.Client()
               if err != nil {
                   return err
               }
               result, err := c.ListWidgets(context.Background())
               if err != nil {
                   return err
               }
               return output.Print(f.IOStreams.Out, f.Resolved.Output, result, &output.TableDef{...})
           },
       }
       return cmd
   }
   ```

3. **Register** the command in the parent command's `NewCmd*()` function (e.g., `cmd/<resource>/<resource>.go`).

4. **Mark mutating commands** with the `"mutates": "true"` annotation so safety checks can be applied.

5. **Use shared flag helpers** from `cmdutil`:
   - `cmdutil.AddFileFlag(cmd, &file)` -- adds `--file/-f`
   - `cmdutil.AddConfirmFlag(cmd, &confirm)` -- adds `--confirm`

6. **Add a test** in the corresponding `_test.go` file using `httptest.NewServer`.

7. **Update documentation and agent materials** (same PR as the code -- see [Documentation and agent materials](#documentation-and-agent-materials)):
   - Add a `Long` description with examples on the command
   - Update `README.md` if the feature is user-visible in overview / examples
   - Update `CLAUDE.md` with workflows, flags, and examples (especially `-o json` for agents)
   - Update `nginxpm/SKILL.md` if you change the CLI's scope or anything called out in the skill's description
   - Update `docs/` (e.g. `docs/CREDENTIALS.md`) if auth, env vars, or config paths change

## Documentation and agent materials

Any change to **commands, subcommands, flags, defaults, auth, config paths, or output shape** should ship with matching docs in the **same change** (or a follow-up PR immediately after):

| Audience | What to update |
|----------|----------------|
| End users | `README.md`, command `Long` / `--help` text |
| Agents / LLMs | `CLAUDE.md` (full guide), `nginxpm/SKILL.md` (triggers + short orientation; must stay consistent with `CLAUDE.md`) |
| Deep topics | `docs/*.md` where relevant |

**Discovery:** Cobra provides `-h` / `--help` on every command; keep `Short`, `Long`, and flag help strings accurate -- agents rely on them when the repo is unavailable.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Every command must have:
  - `Short` description (one line)
  - `Long` description with usage examples
  - Proper flag definitions with descriptions
- Use `-o json` output in all examples for agent-friendliness
- Table output should have meaningful column headers
- Destructive commands must:
  - Have `Annotations: map[string]string{"mutates": "true"}`
  - Use `cmdutil.ConfirmAction()` for confirmation prompts
  - Support `--confirm` to skip prompts

## Commit Messages

Follow conventional commits:
```
feat: add widget list command
fix: correct pagination in proxy list
docs: update README with new cert commands
test: add tests for proxy host creation
chore: update dependencies
```

## Pull Requests

1. Fork the repo and create a feature branch
2. Make your changes with tests
3. Update `README.md`, `CLAUDE.md`, `nginxpm/SKILL.md` (if applicable), and `docs/` per [Documentation and agent materials](#documentation-and-agent-materials)
4. Run `make test` and `make vet` to ensure everything passes
5. Commit with a clear message
6. Open a PR against `main`

## Releasing

Releases are automated via GoReleaser. To create a release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This triggers the release workflow to:
1. Build binaries for all platforms (macOS, Linux, Windows -- amd64 and arm64)
2. Create a GitHub Release with assets
3. Generate a changelog
4. Publish SHA256 checksums

## Reporting Issues

- Use GitHub Issues
- Include: CLI version (`nginxpm version`), OS/arch, command that failed, error output
- For feature requests, describe the use case

## License

This project is licensed under the MIT License -- see [LICENSE](LICENSE) for details.
