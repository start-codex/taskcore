// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package ui

import "embed"

// FS holds the built SvelteKit SPA. Populated by running `pnpm build` in front/.
//
//go:embed all:dist
var FS embed.FS
