# scrapy_project_v2
Golang rewrite of scrapy_project_v1 (Python. This project scrapes Estonian IT-sector job offers and posts them.

## Dependencies
- go
- sqlc
- golang-migrate
- just

## Environment variables
- `DATABASE_URL`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `MODE`
  - dev: only print to console, no database
  - test: only print to console, database (this value can actually be anything)
  - prod: send to telegram, database

**Important:** if MODE=dev other environment variables can be omitted

## For dev

### Run development postgres
```shell
just run-dev-postgres
```

### Create a new migration

```shell
just create-migration <name>
```

### Run migrations

```shell
just migrate
```
