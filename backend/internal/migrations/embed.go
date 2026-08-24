package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

// Dir is the embedded directory that roots the golang-migrate iofs source.
const Dir = "."