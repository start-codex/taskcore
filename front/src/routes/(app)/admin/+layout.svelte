<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import KeyIcon from '@lucide/svelte/icons/key';
	import MailIcon from '@lucide/svelte/icons/mail';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { children } = $props();

	const navItems = $derived.by(() => {
		i18n.locale;
		return [
			{ href: '/admin/smtp', label: m.admin_smtp_title(), icon: MailIcon },
			{ href: '/admin/verification', label: m.admin_verification_title(), icon: ShieldCheckIcon },
			{ href: '/admin/oidc', label: m.admin_oidc_title(), icon: KeyIcon }
		];
	});

	const title = $derived.by(() => { i18n.locale; return m.admin_title(); });
</script>

<div class="mx-auto max-w-screen-lg px-6 py-8">
	<div class="flex gap-0">
		<!-- Left nav -->
		<aside class="w-48 shrink-0 border-r pr-6">
			<p class="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
				{title}
			</p>
			<nav class="flex flex-col gap-0.5">
				{#each navItems as item}
					{@const active = page.url.pathname === item.href}
					<a
						href={item.href}
						class="flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors
							{active
								? 'bg-primary/10 text-primary font-medium'
								: 'text-foreground hover:bg-muted'}"
					>
						<item.icon class="size-4 shrink-0" />
						{item.label}
					</a>
				{/each}
			</nav>
		</aside>

		<!-- Right content -->
		<div class="min-w-0 flex-1 pl-8">
			{@render children()}
		</div>
	</div>
</div>
