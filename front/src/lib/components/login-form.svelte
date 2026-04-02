<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import {
		FieldGroup,
		Field,
		FieldLabel,
		FieldDescription
	} from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { cn } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';
	import { signIn } from '$lib/stores/auth';
	import type { OIDCPublicProvider } from '$lib/api';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let {
		class: className,
		providers = [],
		...restProps
	}: HTMLAttributes<HTMLDivElement> & { providers?: OIDCPublicProvider[] } = $props();

	const id = $props.id();

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
	let loading = $state(false);

	const next = $derived(page.url.searchParams.get('next') || '/');

	const t = $derived.by(() => {
		i18n.locale;
		return {
			welcomeBack: m.login_welcome_back(),
			signIn: m.login_sign_in(),
			email: m.login_email(),
			password: m.login_password(),
			submit: m.login_submit(),
			signingIn: m.login_signing_in(),
			noAccount: m.login_no_account(),
			terms: m.login_terms(),
			forgotPassword: m.login_forgot_password(),
			signInWith: m.login_sign_in_with(),
			orContinueWith: m.login_or_continue_with()
		};
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		errorMessage = '';
		loading = true;
		try {
			await signIn(email, password);
			goto(next);
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : t.submit;
		} finally {
			loading = false;
		}
	}

	function startOIDC(slug: string) {
		window.location.href = `/api/auth/oidc/${slug}?next=${encodeURIComponent(next)}`;
	}
</script>

<div class={cn('flex flex-col gap-6', className)} {...restProps}>
	<Card.Root>
		<Card.Header class="text-center">
			<Card.Title class="text-xl">{t.welcomeBack}</Card.Title>
			<Card.Description>{t.signIn}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if providers.length > 0}
				<div class="flex flex-col gap-2 mb-4">
					{#each providers as provider}
						<Button variant="outline" class="w-full" onclick={() => startOIDC(provider.slug)}>
							{t.signInWith} {provider.name}
						</Button>
					{/each}
				</div>
				<div class="relative mb-4 text-center text-sm after:absolute after:inset-0 after:top-1/2 after:border-t after:border-border">
					<span class="relative z-10 bg-card px-2 text-muted-foreground">{t.orContinueWith}</span>
				</div>
			{/if}
			<form onsubmit={handleSubmit}>
				<FieldGroup>
					<Field>
						<FieldLabel for="email-{id}">{t.email}</FieldLabel>
						<Input
							id="email-{id}"
							type="email"
							placeholder="m@example.com"
							required
							bind:value={email}
						/>
					</Field>
					<Field>
						<div class="flex items-center justify-between">
							<FieldLabel for="password-{id}">{t.password}</FieldLabel>
							<a href="/forgot-password" class="text-xs underline-offset-4 hover:underline text-muted-foreground">{t.forgotPassword}</a>
						</div>
						<Input id="password-{id}" type="password" required bind:value={password} />
					</Field>
					{#if errorMessage}
						<p class="text-destructive text-sm">{errorMessage}</p>
					{/if}
					<Field>
						<Button type="submit" disabled={loading}>
							{loading ? t.signingIn : t.submit}
						</Button>
						<FieldDescription class="text-center">{t.noAccount}</FieldDescription>
					</Field>
				</FieldGroup>
			</form>
		</Card.Content>
	</Card.Root>
	<FieldDescription class="px-6 text-center">{t.terms}</FieldDescription>
</div>
