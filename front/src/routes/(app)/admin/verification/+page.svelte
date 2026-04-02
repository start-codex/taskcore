<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import { instance } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.admin_verification_title(),
			toggle: m.admin_verification_toggle(),
			enabled: m.admin_verification_enabled(),
			disabled: m.admin_verification_disabled()
		};
	});

	let required = $state(false);
	let saving = $state(false);
	let saved = $state(false);

	$effect(() => { required = data.verificationRequired; });

	async function handleToggle() {
		saving = true;
		saved = false;
		try {
			required = !required;
			await instance.verification.save({ required });
			saved = true;
			setTimeout(() => { saved = false; }, 3000);
		} catch {
			required = !required; // revert
		} finally {
			saving = false;
		}
	}
</script>

<div class="space-y-6">
	<h2 class="text-lg font-semibold">{t.title}</h2>

	<Card.Root>
		<Card.Content class="flex items-center justify-between pt-6">
			<div>
				<p class="text-sm font-medium">{t.toggle}</p>
				<p class="text-xs text-muted-foreground">
					{required ? t.enabled : t.disabled}
				</p>
			</div>
			<div class="flex items-center gap-3">
				<Button
					variant={required ? 'default' : 'outline'}
					size="sm"
					onclick={handleToggle}
					disabled={saving}
				>
					{required ? 'On' : 'Off'}
				</Button>
				{#if saved}
					<span class="text-xs text-green-600">Saved</span>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>
</div>
