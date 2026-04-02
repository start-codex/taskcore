<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { invitations as invApi, auth as authApi, ApiError } from '$lib/api';
	import { signIn } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Field, FieldGroup, FieldLabel } from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			invalidToken: m.accept_invalid_token(),
			invitedBy: m.accept_invited_by,
			acceptBtn: m.accept_button(),
			accepting: m.accept_accepting(),
			registerTitle: m.accept_register_title(),
			name: m.accept_name(),
			email: m.accept_email(),
			password: m.accept_password(),
			confirmPw: m.accept_confirm_password(),
			register: m.accept_register(),
			registering: m.accept_registering(),
			loginLink: m.accept_login_link(),
			emailMismatch: m.accept_email_mismatch,
			error: m.accept_error(),
			pwMismatch: m.accept_passwords_mismatch(),
			pwShort: m.accept_password_too_short()
		};
	});

	let name = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');

	const hasToken = $derived(!!data.token);
	const hasInvitation = $derived(!!data.invitation);
	const isAuthenticated = $derived(!!data.authedUser);
	const emailMatches = $derived(
		isAuthenticated && data.authedUser?.email === data.invitation?.email
	);

	const title = $derived(
		hasInvitation ? m.accept_title({ workspace: data.invitation!.workspace_name }) : ''
	);

	async function handleAcceptAuth() {
		loading = true;
		error = '';
		try {
			const result = await invApi.accept({ token: data.token });
			goto('/' + result.workspace_slug);
		} catch (err) {
			error = err instanceof Error ? err.message : t.error;
		} finally {
			loading = false;
		}
	}

	async function handleRegisterAndAccept(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) { error = t.pwMismatch; return; }
		if (password.length < 8) { error = t.pwShort; return; }

		loading = true;
		try {
			const result = await invApi.accept({
				token: data.token,
				email: data.invitation!.email,
				name,
				password
			});
			// Auto-login after registration
			try {
				await signIn(data.invitation!.email, password);
				goto('/' + result.workspace_slug);
			} catch {
				// Login failed — redirect to login with workspace
				goto('/login?next=/' + result.workspace_slug);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : t.error;
		} finally {
			loading = false;
		}
	}

	async function handleSignOutAndRedirect() {
		try { await authApi.logout(); } catch {}
		window.location.href = loginWithNext;
	}

	const loginWithNext = $derived(
		`/login?next=/invitations/accept?token=${encodeURIComponent(data.token)}`
	);
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

		{#if data.serverError}
			<!-- Server error — not an invalid token -->
			<Card.Root>
				<Card.Content class="py-8 text-center">
					<p class="text-sm text-destructive">Something went wrong. Please try again later.</p>
				</Card.Content>
			</Card.Root>

		{:else if !hasToken || !hasInvitation}
			<!-- Invalid or missing token -->
			<Card.Root>
				<Card.Content class="py-8 text-center">
					<p class="text-sm text-destructive">{t.invalidToken}</p>
					<a href="/login" class="mt-4 inline-block text-sm text-primary underline-offset-4 hover:underline">
						{t.loginLink}
					</a>
				</Card.Content>
			</Card.Root>

		{:else if isAuthenticated && emailMatches}
			<!-- Authenticated user, email matches — show accept button -->
			<Card.Root>
				<Card.Header class="text-center">
					<Card.Title class="text-xl">{title}</Card.Title>
					<Card.Description>
						{t.invitedBy({ name: data.invitation!.inviter_name, role: data.invitation!.role })}
					</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-4">
					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}
					<Button class="w-full" onclick={handleAcceptAuth} disabled={loading}>
						{loading ? t.accepting : t.acceptBtn}
					</Button>
				</Card.Content>
			</Card.Root>

		{:else if isAuthenticated && !emailMatches}
			<!-- Authenticated but wrong email — sign out preserves token via next param -->
			<Card.Root>
				<Card.Content class="py-8 space-y-4 text-center">
					<p class="text-sm text-muted-foreground">
						{t.emailMismatch({ email: data.invitation!.email })}
					</p>
					<Button variant="outline" onclick={handleSignOutAndRedirect}>Sign out and switch account</Button>
				</Card.Content>
			</Card.Root>

		{:else}
			<!-- Not authenticated — registration form -->
			<Card.Root>
				<Card.Header class="text-center">
					<Card.Title class="text-xl">{title}</Card.Title>
					<Card.Description>
						{t.invitedBy({ name: data.invitation!.inviter_name, role: data.invitation!.role })}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<p class="mb-4 text-sm font-medium">{t.registerTitle}</p>
					<form onsubmit={handleRegisterAndAccept}>
						<FieldGroup>
							<Field>
								<FieldLabel for="accept-name">{t.name}</FieldLabel>
								<Input id="accept-name" type="text" required bind:value={name} />
							</Field>
							<Field>
								<FieldLabel for="accept-email">{t.email}</FieldLabel>
								<Input id="accept-email" type="email" value={data.invitation!.email} disabled />
							</Field>
							<Field>
								<FieldLabel for="accept-pw">{t.password}</FieldLabel>
								<Input id="accept-pw" type="password" required bind:value={password} />
							</Field>
							<Field>
								<FieldLabel for="accept-cpw">{t.confirmPw}</FieldLabel>
								<Input id="accept-cpw" type="password" required bind:value={confirmPassword} />
							</Field>
							{#if error}
								<p class="text-sm text-destructive">{error}</p>
							{/if}
							<Field>
								<Button type="submit" class="w-full" disabled={loading || !name || !password}>
									{loading ? t.registering : t.register}
								</Button>
							</Field>
						</FieldGroup>
					</form>
					<div class="mt-4 text-center">
						<a href={loginWithNext} class="text-sm text-muted-foreground underline-offset-4 hover:underline">
							{t.loginLink}
						</a>
					</div>
				</Card.Content>
			</Card.Root>
		{/if}
	</div>
</div>
