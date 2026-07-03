# AGENTS.md

## Commands

```bash
go run main.go          # runs the bot (needs .env)
go build -o bot main.go # single-binary build
go vet ./...            # static analysis
go test ./...           # run tests (none exist yet)
```

No Makefile, Dockerfile, or CI exists.

## Environment

- `.env` is loaded via `godotenv`. Required vars: `DISCORD_BOT_TOKEN`, `APP_ID`
- `DISCORD_GUILD_ID`: set for dev (instant command updates), leave empty for prod (global, 1hr propagation)
- `DATABASE_PATH`: defaults to `bot.db`, loaded from env
- `APP_PUBLIC_KEY` is defined in `.env.example` but **never used** by the code (bot uses gateway, not HTTP interactions)

## Architecture

```
main.go           # entrypoint, wires everything
internal/
  commands/       # slash command + component handlers
  config/         # env-based config (no struct tags)
  database/       # SQLite (WAL, foreign_keys ON, busy_timeout=5000)
  logger/         # slog + discordgo log capture
  middleware/     # Recover (panic) + GuildOnly (DM guard)
  router/         # dispatches interactions by name/custom_id
```

Single binary at root. No `cmd/` directory.

## Middleware

`middleware.Chain(h, mw...)` applies middleware **in order** — first in the list is outermost:

```go
middleware.Chain(handler, middleware.Recover, middleware.GuildOnly)
// exec: Recover -> GuildOnly -> handler
```

Slash commands and component handlers are wrapped **separately** in main.go. Component handlers are NOT wrapped with `GuildOnly` (only `Recover`).

## SQLite / Database

- Migrations run automatically on startup via `database.Migrate()`
- Add new migrations by appending to the `migrations` slice in `internal/database/migrate.go`
- Repository functions are write-only (`INSERT OR REPLACE` upserts); no read queries exist
- The `messages` table exists in schema but has no repository function and is never written to

## Data Flow

1. `godotenv.Load()` → `logger.Setup()` → `config.Load()` → `database.Open()` → `database.Migrate()`
2. Router registers slash commands + component handlers
3. If `APP_ID` is set, `r.Sync()` bulk-overwrites commands (guild-scoped or global based on `DISCORD_GUILD_ID`)
4. `dg.AddHandler(r.Handle)` handles all `InteractionCreate` events
5. On shutdown (SIGINT/SIGTERM): deferred `db.Close()` runs

## Style / Conventions

- Zero test files. New tests go in same package as source (`_test.go` alongside)
- Logging uses `log/slog` with `source=discordgo` tag for library logs
- Component `custom_id` convention: `commandname_action` (e.g., `embed_demo_confirm`)
- All commands are guild-only; `GuildOnly` middleware returns ephemeral error for DMs