run:
    go run cmd/main.go

create-migration name:
    migrate create -seq -dir migrations -ext sql {{name}}

migrate:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

force-migration count:
    migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" force {{count}}
