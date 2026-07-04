# go-discord-bot

A Discord bot built with [discordgo](https://github.com/bwmarrin/discordgo) featuring slash commands, interactive components, and SQLite persistence.

## Features

- **`/ping`** — Health-check command that replies with "Pong!"
- **`/sync-server`** — Syncs all guild members, channels, and roles to the local SQLite database
- **`/embed-demo`** — Interactive embed with confirm/cancel buttons demonstrating Discord message components
- **`/send-message`** — Send rich messages (embeds, buttons, select menus) via a JSON payload, with optional reply targeting
- **`/get-message-as-json`** — Fetch a bot message by ID and download it as a JSON payload (compatible with `/send-message` and `/edit-message`)
- **`/edit-message`** — Edit an existing bot message by ID using the same JSON payload format

## Prerequisites

- [Go](https://go.dev/) 1.26+
- A Discord bot application (create one at [Discord Developer Portal](https://discord.com/developers/applications))

## Setup

```bash
git clone https://github.com/mridhop/go-discord-bot
cd go-discord-bot
cp .env.example .env
```

Edit `.env` and fill in your credentials:

```env
DISCORD_BOT_TOKEN=your_bot_token_here
APP_ID=your_application_id_here
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DISCORD_BOT_TOKEN` | Yes | — | Bot token from Discord Developer Portal |
| `APP_ID` | No | — | Application ID; required for command registration (`r.Sync()`) |
| `DISCORD_GUILD_ID` | No | — | Guild ID for instant command updates (dev). Leave empty for global commands (up to 1hr propagation) |
| `DATABASE_PATH` | No | `bot.db` | Path to the SQLite database file |

## Usage

```bash
go run main.go
```

Or build and run:

```bash
go build -o bot main.go
./bot
```

## Development

- **Dev mode:** Set `DISCORD_GUILD_ID` to register commands to a single guild (instant updates)
- **Prod mode:** Leave `DISCORD_GUILD_ID` empty for global commands

Available tooling:

```bash
go vet ./...        # static analysis
go test ./...       # run tests
```

## Project Structure

```
.
├── main.go                    # entrypoint — wires everything together
├── internal/
│   ├── commands/              # slash command + component handlers
│   │   ├── ping.go            # /ping
│   │   ├── sync_server.go     # /sync-server
│   │   ├── send_message.go    # /send-message
│   │   ├── embed_demo.go      # /embed-demo + button handlers
│   │   ├── get_message_json.go # /get-message-as-json
│   │   └── edit_message.go    # /edit-message
│   ├── config/                # env-based configuration
│   ├── database/              # SQLite (WAL, foreign keys, migrations)
│   ├── logger/                # slog + discordgo log bridge
│   ├── middleware/            # panic recovery + guild-only guard
│   └── router/                # interaction dispatch (commands + components)
├── .env.example               # environment template
└── .gitignore
```

## JSON Payload Format

The `/send-message` and `/edit-message` commands share the same JSON payload format. The `/get-message-as-json` command exports bot messages in this format. The top-level object has three optional fields (at least one required):

### Top-level

```json
{
  "content": "plain text content",
  "embeds": [{ ... }],
  "components": [ ... ]
}
```

### Embeds

```json
{
  "embeds": [{
    "title": "Title",
    "description": "Description text",
    "url": "https://example.com",
    "color": 16711680,
    "timestamp": "2025-01-01T00:00:00Z",
    "footer": { "text": "Footer text", "icon_url": "https://example.com/icon.png" },
    "thumbnail": { "url": "https://example.com/thumb.png" },
    "image": { "url": "https://example.com/img.png" },
    "author": { "name": "Author", "url": "https://example.com", "icon_url": "https://example.com/avatar.png" },
    "fields": [
      { "name": "Field 1", "value": "Value 1", "inline": true },
      { "name": "Field 2", "value": "Value 2", "inline": false }
    ]
  }]
}
```

### Components (Action Rows)

Each action row contains one or more components (buttons or select menus):

```json
{
  "components": [
    {
      "type": 1,
      "components": [
        { "type": 2, "style": 1, "label": "Primary", "custom_id": "btn_primary" },
        { "type": 2, "style": 3, "label": "Green", "emoji": { "name": "✅" } },
        { "type": 2, "style": 5, "label": "Link", "url": "https://example.com" }
      ]
    },
    {
      "type": 1,
      "components": [
        {
          "type": 3,
          "custom_id": "menu_select",
          "placeholder": "Choose an option",
          "min_values": 1,
          "max_values": 1,
          "options": [
            { "label": "Option A", "value": "a", "description": "First option", "emoji": { "name": "🅰️" } },
            { "label": "Option B", "value": "b", "default": true }
          ]
        }
      ]
    }
  ]
}
```

**Component types** — `1`: Action Row, `2`: Button, `3`: String Select, `5`: User Select, `6`: Role Select, `7`: Mentionable Select, `8`: Channel Select

**Button styles** — `1`: Primary (blurple), `2`: Secondary (grey), `3`: Success (green), `4`: Danger (red), `5`: Link (grey, requires `url`)

### Complete Example

```json
{
  "content": "Check this out!",
  "embeds": [{
    "title": "Announcement",
    "description": "Something important happened.",
    "color": 3447003,
    "footer": { "text": "Bot Announcement" },
    "fields": [
      { "name": "Time", "value": "Now", "inline": true },
      { "name": "Place", "value": "Here", "inline": true }
    ]
  }],
  "components": [
    {
      "type": 1,
      "components": [
        { "type": 2, "style": 3, "label": "Approve", "custom_id": "action_approve", "emoji": { "name": "✅" } },
        { "type": 2, "style": 4, "label": "Reject", "custom_id": "action_reject", "emoji": { "name": "❌" } }
      ]
    }
  ]
}
```

### Command Options

**`/send-message`**

| Option | Required | Description |
|---|---|---|
| `json` | Yes | JSON payload as described above |
| `message-id` | No | ID of a message to reply to (creates a reply thread) |

**`/edit-message`**

| Option | Required | Description |
|---|---|---|
| `json` | Yes | JSON payload with content, embeds, and components |
| `message-id` | Yes | ID of the bot message to edit |

**`/get-message-as-json`**

| Option | Required | Description |
|---|---|---|
| `message-id` | Yes | ID of the bot message to fetch |

## Architecture

- **Gateway-only bot** — Uses Discord WebSocket gateway; no HTTP interactions endpoint
- **Middleware chain** — `Recover` (panic recovery) and `GuildOnly` (blocks DMs) wrap every slash command; component handlers use only `Recover`
- **SQLite with WAL** — Write-Ahead Logging enabled, 5-second busy timeout, foreign key enforcement
- **Auto-migrations** — Schema migrations run at startup from a versioned migration slice
- **Write-only repository** — All DB operations use `INSERT OR REPLACE` upserts; no read queries exist outside of the migration tracker
- **Component convention** — Button `custom_id` format: `commandname_action` (e.g., `embed_demo_confirm`)
