# migrate-kit

Sibling migration program of the two demo services. It applies hand-written
go-migrate scripts against the same postgres database the services share, so the
schema comes from scripted migration and not from gorm AutoMigrate.

## Config

`config.yaml` holds the postgres source (same as the demo services) and the
scripts path. Pass `--conf=/path/to/other.yaml` to point at a different config.

## Commands

- `migrate-kit migrate all` — run each pending step forward
- `migrate-kit migrate inc` — step one forward
- `migrate-kit migrate dec` — step one back
- `migrate-kit migrate` — show the current schema version
