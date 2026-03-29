<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import * as Item from '$lib/components/ui/item/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n, switchLocale } from '$lib/i18n.svelte';
	import type { Locale } from '$lib/paraglide/runtime';

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.settings_preferences(),
			languageTitle: m.settings_language_title(),
			languageDesc: m.settings_language_desc()
		};
	});

	const languages: { value: Locale; label: string }[] = [
		{ value: 'en', label: 'English' },
		{ value: 'es', label: 'Español' }
	];
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-lg font-semibold">{t.title}</h2>
		<hr class="mt-3 border-border" />
	</div>

	<div class="mx-auto max-w-[760px]">
	<Item.Group>
		<Item.Root variant="outline">
			<Item.Content>
				<Item.Title>{t.languageTitle}</Item.Title>
				<Item.Description>{t.languageDesc}</Item.Description>
			</Item.Content>
			<Item.Actions>
				<select
					value={i18n.locale}
					onchange={(e) => switchLocale((e.target as HTMLSelectElement).value as Locale)}
					class="rounded-md border border-input bg-background px-3 py-1.5 text-sm shadow-xs focus:outline-none focus:ring-2 focus:ring-ring"
				>
					{#each languages as lang}
						<option value={lang.value}>{lang.label}</option>
					{/each}
				</select>
			</Item.Actions>
		</Item.Root>
	</Item.Group>
	</div>
</div>
