# run scraper sync
run:
    go run cmd/main.go

# run scraper async
async:
    go run cmd/main.go --async

# create migration
create-migration name:
    migrate create -seq -dir migrations -ext sql {{name}}

# run migrations for development postgres
migrate:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5455/postgres?sslmode=disable" up

# force migration for development migration
force-migration count:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5455/postgres?sslmode=disable" force {{count}}

# run development postgres on port 5455
run-dev-postgres:
    docker run --name postgres-dev -e POSTGRES_PASSWORD=postgres -p 5455:5432 -d postgres

# build binary for mac arm64 architecture
build-mac:
    mkdir -p bin
    GOOS=darwin GOARCH=arm64 go build -o bin/scraper_mac cmd/main.go

# build binary for linux amd64 architecture
build:
    mkdir -p bin
    GOOS=linux GOARCH=amd64 go build -o bin/scraper cmd/main.go
