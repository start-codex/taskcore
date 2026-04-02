<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import ChangePasswordForm from '$lib/components/change-password-form.svelte';
	import { currentUser } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const title = $derived.by(() => { i18n.locale; return m.settings_account_title(); });
	const t = $derived.by(() => {
		i18n.locale;
		return {
			noPassword: m.settings_no_password(),
			setPassword: m.settings_set_password()
		};
	});
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-lg font-semibold">{title}</h2>
		<hr class="mt-3 border-border" />
	</div>

	{#if $currentUser?.has_password === false}
		<div class="rounded-md bg-muted p-4 text-sm text-muted-foreground">
			<p>{t.noPassword}</p>
			<a href="/forgot-password" class="mt-2 inline-block text-sm text-primary underline-offset-4 hover:underline">
				{t.setPassword}
			</a>
		</div>
	{:else}
		<ChangePasswordForm />
	{/if}
</div>
