# migrate-kit

Sibling migration program of the two demo services. It applies hand-written
go-migrate scripts against the same postgres database the services share, so the
schema comes from scripted migration and not from gorm AutoMigrate.

## Config

`config.yaml` holds the postgres source (same as the demo services) and the
scripts path. Pass `--conf=/path/to/other.yaml` to point at a different config.

## Commands

- `migrate-kit migrate batch` — run each pending step forward (called `all` before go-migrate v0.0.6)
- `migrate-kit migrate inc` — step one forward
- `migrate-kit migrate dec` — step one back
- `migrate-kit migrate` — show the current schema version

## Run the demo

Use Go 1.27.0. From this directory, with PostgreSQL running and an empty demo
database matching `config.yaml`:

```bash
go run . migrate batch
go run . migrate
```

The first command applies `scripts/000001_create_tables.up.sql`: it creates
`students`, `articles`, and `idx_articles_student_id`. The version command
then prints `1`; running `batch` again has no pending steps. `make all` runs
the same batch command.

Start `demo1kratos` and `demo2kratos` from their module directories with
`go run ./cmd/demo1kratos -conf ./configs` and
`go run ./cmd/demo2kratos -conf ./configs`. Both services must use the same
database as this tool; set `KRATOS_DATABASE_SOURCE` for a custom database and
pass a matching YAML file to this tool with `--conf` (use an absolute scripts
path in that file).

The services use `checkmigration.CheckMigrate` instead of `AutoMigrate`.
Before migration they reject the missing schema without creating tables;
after migration they start on HTTP ports 8001 and 8002. Probe with
`curl 'http://localhost:8001/students?page=1&page_size=1'` and
`curl 'http://localhost:8002/articles?page=1&page_size=1'`.

Use an empty database for the initial script: it creates tables rather than
adopting tables previously created by another demo. `migrate dec` reverses
that script and drops both tables, so use it only on disposable demo data.
