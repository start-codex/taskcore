<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import type { LayoutData } from './$types';
	import { currentUser, logout } from '$lib/stores/auth';
	import { auth } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { children, data }: { children: any; data: LayoutData } = $props();

	const isWorkspaceRoute = $derived(!!page.params.workspace);

	const showVerifyBanner = $derived(
		!!$currentUser &&
		!$currentUser.email_verified_at &&
		data.emailVerificationRequired
	);

	const t = $derived.by(() => {
		i18n.locale;
		return {
			verifyBanner: m.verify_banner(),
			resend: m.verify_resend(),
			resent: m.verify_resent(),
			resendError: m.verify_resend_error()
		};
	});

	let resending = $state(false);
	let resendSuccess = $state(false);
	let resendError = $state('');

	async function handleResend() {
		resending = true;
		resendSuccess = false;
		resendError = '';
		try {
			await auth.resendVerification();
			resendSuccess = true;
			setTimeout(() => { resendSuccess = false; }, 5000);
		} catch {
			resendError = t.resendError;
		} finally {
			resending = false;
		}
	}
</script>

<!-- Verification banner (above everything) -->
{#if showVerifyBanner}
	<div class="bg-yellow-50 border-b border-yellow-200 px-4 py-2 text-center text-sm text-yellow-800">
		{t.verifyBanner}
		<button
			class="ml-2 underline underline-offset-4 hover:text-yellow-900"
			onclick={handleResend}
			disabled={resending}
		>
			{resending ? '...' : t.resend}
		</button>
		{#if resendSuccess}
			<span class="ml-2 text-green-700">{t.resent}</span>
		{/if}
		{#if resendError}
			<span class="ml-2 text-red-700">{resendError}</span>
		{/if}
	</div>
{/if}

{#if isWorkspaceRoute}
	{@render children()}
{:else}
	<div class="bg-background min-h-screen">
		<header class="border-b">
			<div class="mx-auto flex h-14 max-w-screen-xl items-center justify-between px-6">
				<a href="/" class="flex items-center gap-2 font-semibold">
					<div class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md text-xs">
						T
					</div>
					Tookly
				</a>
				{#if $currentUser}
					<div class="flex items-center gap-3">
						<span class="text-muted-foreground text-sm">{$currentUser.email}</span>
						<Button variant="ghost" size="sm" onclick={() => logout()}>
							<LogOutIcon class="size-4" />
						</Button>
					</div>
				{/if}
			</div>
		</header>
		{@render children()}
	</div>
{/if}
