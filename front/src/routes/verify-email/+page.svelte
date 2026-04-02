<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			success: m.verify_success(),
			invalid: m.verify_invalid(),
			error: m.verify_error()
		};
	});
</script>

<div class="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="/" class="flex items-center gap-2 self-center font-medium">
			<div class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md">
				<svg xmlns="http://www.w3.org/2000/svg" class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<rect width="8" height="8" x="2" y="2" rx="2" /><rect width="8" height="8" x="14" y="2" rx="2" />
					<rect width="8" height="8" x="2" y="14" rx="2" /><rect width="8" height="8" x="14" y="14" rx="2" />
				</svg>
			</div>
			Tookly
		</a>
		<Card.Root>
			<Card.Content class="py-8 text-center">
				{#if data.status === 'success'}
					<p class="text-sm text-green-700">{t.success}</p>
					<a href="/" class="mt-4 inline-block text-sm text-primary underline-offset-4 hover:underline">
						Continue to Tookly
					</a>
				{:else if data.status === 'invalid'}
					<p class="text-sm text-destructive">{t.invalid}</p>
				{:else}
					<p class="text-sm text-destructive">{t.error}</p>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
</div>
