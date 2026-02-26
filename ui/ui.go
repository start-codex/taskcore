package ui

import "embed"

// FS holds the built SvelteKit SPA. Populated by running `pnpm build` in front/.
//
//go:embed all:dist
var FS embed.FS
