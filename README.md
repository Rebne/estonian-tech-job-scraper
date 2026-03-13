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
docker run --name postgres-dev -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres
```

### Create a new migration

```shell
migrate create -seq -dir migrations -ext sql create_job_table
```

### Run migrations

```shell
migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
```
