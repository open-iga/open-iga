package sql

import "embed"

//go:embed migration/*.sql
var MigrationFiles embed.FS
