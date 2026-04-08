# scrapy_project_v2
Golang rewrite of scrapy_project_v1 (Python. This project scrapes Estonian IT-sector job offers and posts them.

## Dependencies
- go
- sqlc
- golang-migrate
- just
- chrome (for chromedp)

## Setting up and running
```shell
go mod install

# Installs playwright dependecies (browsers mostly)
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps

just run
```

## Environment variables
- `DATABASE_URL`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `MODE`
  - dev: only print to console, no database
  - test: only print to console, database (this value can actually be anything)
  - prod: send to telegram, database
- `PROXY_URL` (optional)
  - http and playwright fetchers will make requests through this proxy url
- `CHROME_EXECUTABLE_PATH` (optional)
  - explicit Chrome/Chromium binary path for chromedp

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
