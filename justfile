run:
    go run cmd/main.go

async:
    go run cmd/main.go --async

create-migration name:
    migrate create -seq -dir migrations -ext sql {{name}}

migrate:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5455/postgres?sslmode=disable" up

force-migration count:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5455/postgres?sslmode=disable" force {{count}}

run-dev-postgres:
    docker run --name postgres-dev -e POSTGRES_PASSWORD=postgres -p 5455:5432 -d postgres

build:
    mkdir -p bin
    go build -o bin/scraper cmd/main.go
