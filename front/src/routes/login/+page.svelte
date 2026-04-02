<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import type { PageData } from './$types';
	import LoginForm from "$lib/components/login-form.svelte";
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const resetSuccess = $derived(page.url.searchParams.get('reset') === 'success');
	const oidcError = $derived(page.url.searchParams.get('error'));

	const t = $derived.by(() => {
		i18n.locale;
		return {
			resetSuccess: m.login_reset_success(),
			oidcDenied: m.login_oidc_denied(),
			oidcNoAccount: m.login_oidc_no_account(),
			oidcArchived: m.login_oidc_account_archived(),
			oidcLoadError: m.login_oidc_load_error()
		};
	});

	const errorMessage = $derived.by(() => {
		if (!oidcError) return '';
		switch (oidcError) {
			case 'oidc_denied': return t.oidcDenied;
			case 'oidc_no_account': return t.oidcNoAccount;
			case 'account_archived': return t.oidcArchived;
			default: return '';
		}
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
		{#if resetSuccess}
			<div class="rounded-md bg-green-50 p-3 text-center text-sm text-green-700 border border-green-200">
				{t.resetSuccess}
			</div>
		{/if}
		{#if errorMessage}
			<div class="rounded-md bg-destructive/10 p-3 text-center text-sm text-destructive border border-destructive/20">
				{errorMessage}
			</div>
		{/if}
		{#if data.oidcLoadFailed}
			<div class="rounded-md bg-destructive/10 p-3 text-center text-sm text-destructive border border-destructive/20">
				{t.oidcLoadError}
			</div>
		{/if}
		<LoginForm providers={data.providers} />
	</div>
</div>
