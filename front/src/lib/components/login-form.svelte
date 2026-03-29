<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
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
	import { auth } from '$lib/api';
	import { login } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { class: className, ...restProps }: HTMLAttributes<HTMLDivElement> = $props();

	const id = $props.id();

	let email = $state('');
	let password = $state('');
	let errorMessage = $state('');
	let loading = $state(false);

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
			terms: m.login_terms()
		};
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		errorMessage = '';
		loading = true;
		try {
			const user = await auth.login({ email, password });
			login(user);
			goto('/');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : t.submit;
		} finally {
			loading = false;
		}
	}
</script>

<div class={cn('flex flex-col gap-6', className)} {...restProps}>
	<Card.Root>
		<Card.Header class="text-center">
			<Card.Title class="text-xl">{t.welcomeBack}</Card.Title>
			<Card.Description>{t.signIn}</Card.Description>
		</Card.Header>
		<Card.Content>
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
						<div class="flex items-center">
							<FieldLabel for="password-{id}">{t.password}</FieldLabel>
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
