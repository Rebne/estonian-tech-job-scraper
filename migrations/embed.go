package dbmigrations

import "embed"

// Files contains the embedded runtime migration files.
//
//go:embed *.up.sql
var Files embed.FS
