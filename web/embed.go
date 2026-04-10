// Package web holds the embedded Svelte build output.
// Run `bun run build` inside web/ to populate dist/ before building the Go binary.
package web

import "embed"

//go:embed all:dist
var Assets embed.FS
