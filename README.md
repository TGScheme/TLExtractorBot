# TLExtractorBot

Extracts the Telegram TL scheme from the Android apk, tdesktop and TDLib, merges
the three, and publishes the result to GitHub, Telegraph and Telegram.

## Running

```sh
cp .env.sample .env          # fill in tokens and ids
mkdir -p secrets             # drop the GitHub App key in secrets/github.pem
docker compose up -d --build
```

Development, with hot reload through air:

```sh
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

## Importing an existing storage.json

Older releases kept everything in a single JSON file. To carry that state over:

```sh
docker compose run --rm -v /path/to/storage.json:/tmp/storage.json bot \
    importjson /tmp/storage.json
```

## Database

Schema migrations live in `internal/db/schema` and are applied by goose at
startup. Queries live in `internal/db/query`; `cmd/sqlgen` turns them into typed
stores under `internal/db` — run it after touching either:

```sh
go run ./cmd/sqlgen
```

The generated `*_gen.go` files are not committed.

## Admin commands

Sent from the chat configured as `LOG_CHAT_ID`:

| command | effect |
|---|---|
| `/patch` | force an extraction even if the store version has not moved |
| `/model` | list Gemini models; `/model <id>` selects one |
| `/branch` | show the tdesktop branch; `/branch <sha>` pins another |
| `/backup` | dump the database and upload it right here |

## Backup and restore

Every six hours the bot dumps the database with `pg_dump -Fc` and uploads it to
`BACKUP_CHAT_ID`. `/backup` does the same on demand, into the log chat.

To restore, send one of those dumps back into the log chat as a document. Any
`.sql` or `.dump` file posted there is fed to `pg_restore --clean --if-exists`,
which **replaces the current database**. The restore is refused while an
extraction is running.
