# Discord Bot Control Panel

A self-hosted web dashboard for managing Discord bots and configuring automated messages—without editing bot code for every rule.

![Discord bot management dashboard](imgs/admin.png)

## What it does

- **Manage bots:** Add bot credentials, verify connections, rename bots, and start or stop workers from the dashboard.
- **Schedule messages:** Send a message repeatedly at an interval or once at a chosen date and time.
- **Configure message replies:** Match a prefix and command, then reply in a selected channel or across channels.
- **Administer the app:** Review logs, change the admin password, and download or restore the SQLite database.

## Requirements

- Go **1.25.8** or newer
- A C compiler/toolchain for local builds (the SQLite Go driver uses CGO)
- The `sqlite3` command-line executable available on `PATH` (the app checks for it at startup)
- A Discord application and bot token for Discord features
- Docker Compose, if using the containerized development setup

## Run locally

1. Clone the repository and enter its directory:

   ```bash
   git clone https://github.com/KrzysztofMarciniak/dc.git
   cd dc
   ```

2. Create the required environment file and set your own admin credentials:

   ```bash
   cp .env_example .env
   ```

   Edit `.env`:

   ```dotenv
   ADMIN_USERNAME=your-admin-name
   ADMIN_PASSWORD=use-a-strong-password
   ```

   The sample values in `.env_example` are for local setup only; replace them before running the app.

3. Start the server:

   ```bash
   go run .
   ```

4. Open [http://localhost:8080/login](http://localhost:8080/login) and sign in with the credentials from `.env`.

The app creates `database.db` in its working directory on first start. Keep this file safe: it contains application data, including configured bot credentials.

## Run with Docker Compose

The development Compose setup starts the app behind Nginx and serves it on port `80`:

```bash
cp .env_example .env
# Edit .env and replace the sample credentials before continuing.
docker compose -f compose.dev.yml up --build
```

Open [http://localhost](http://localhost). To stop the development stack, press `Ctrl+C`; remove its containers with:

```bash
docker compose -f compose.dev.yml down
```

For the production Compose setup, first create and edit `.env` as described above, then run:

```bash
docker compose -f compose.prod.yml up --build -d
```

Open [http://localhost](http://localhost). The production app reads `.env` from a read-only mount and stores `database.db` in the persistent `app-data` Docker volume. Stop the services with `docker compose -f compose.prod.yml down`; the database volume remains unless you explicitly remove it with `docker compose -f compose.prod.yml down --volumes`.

This stack serves plain HTTP on port `80` and does not configure TLS. Put it behind an HTTPS reverse proxy before exposing it publicly.

## Set up a Discord bot

1. Create an application and bot in the [Discord Developer Portal](https://discord.com/developers/applications).
2. Enable the **Message Content Intent** in the bot settings if you want message-triggered replies.
3. In the dashboard, add the bot's name, token, and application client ID, then verify its connection.
4. Invite it to your server and grant only the permissions it needs. The dashboard's generated invite link currently requests the `Administrator` permission; review that before using the link.
5. Start the bot worker, then configure scheduled messages or message-response rules.

Workers run in the application process. If the app restarts, start the bot workers again from the dashboard.

## Scheduling notes

- Repeating schedules use an interval in **seconds** (for example, `3600` for one hour).
- One-time schedules use a date and time entered in the dashboard.
- Scheduled messages are checked by a worker that runs while the corresponding bot worker is active.

## Project structure

The app is written in Go, uses Gin for HTTP routing and templates, GORM with SQLite for persistence, and DiscordGo to connect to Discord. The UI is rendered from Go templates; Nginx is included in the Docker Compose setups as a reverse proxy.

## Current limitations

- Bot workers are process-local and are not automatically restarted after an app restart.
- The generated bot invite link requests Administrator permissions; use a least-privilege invite instead where possible.
