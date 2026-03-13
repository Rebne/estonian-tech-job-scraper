# scrapy_project_v2
Golang rewrite of scrapy_project_v1 (Python. This project scrapes Estonian IT-sector job offers and posts them.

## Dependencies
- go
- sqlc
- golang-migrate
- just

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
