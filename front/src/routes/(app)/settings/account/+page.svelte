<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const title = $derived.by(() => { i18n.locale; return m.settings_account_title(); });
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-lg font-semibold">{title}</h2>
		<hr class="mt-3 border-border" />
	</div>

	<div class="flex flex-col items-center justify-center gap-3 py-16 text-center text-muted-foreground">
		<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24"
			fill="none" stroke="currentColor" stroke-width="1.5"
			stroke-linecap="round" stroke-linejoin="round">
			<circle cx="12" cy="8" r="5"/>
			<path d="M20 21a8 8 0 1 0-16 0"/>
		</svg>
		<p class="text-sm">Coming soon</p>
	</div>
</div>
