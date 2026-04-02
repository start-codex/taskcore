<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import { instance } from '$lib/api';
	import type { OIDCProvider } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.admin_oidc_title(),
			add: m.admin_oidc_add(),
			name: m.admin_oidc_name(),
			slug: m.admin_oidc_slug(),
			slugHint: m.admin_oidc_slug_hint(),
			issuer: m.admin_oidc_issuer(),
			issuerHint: m.admin_oidc_issuer_hint(),
			clientId: m.admin_oidc_client_id(),
			clientSecret: m.admin_oidc_client_secret(),
			redirectUri: m.admin_oidc_redirect_uri(),
			redirectUriHint: m.admin_oidc_redirect_uri_hint(),
			scopes: m.admin_oidc_scopes(),
			autoRegister: m.admin_oidc_auto_register(),
			enabled: m.admin_oidc_enabled(),
			save: m.admin_oidc_save(),
			saving: m.admin_oidc_saving(),
			saved: m.admin_oidc_saved(),
			deleteLabel: m.admin_oidc_delete(),
			deleteConfirm: m.admin_oidc_delete_confirm(),
			deleted: m.admin_oidc_deleted(),
			noProviders: m.admin_oidc_no_providers(),
			noProvidersDesc: m.admin_oidc_no_providers_desc(),
			edit: m.admin_oidc_edit(),
			cancel: m.admin_oidc_cancel(),
			disabledLabel: m.admin_oidc_disabled(),
			saveError: m.admin_oidc_save_error(),
			deleteError: m.admin_oidc_delete_error()
		};
	});

	let providers = $state<OIDCProvider[]>([]);
	let showForm = $state(false);
	let editingId = $state<string | null>(null);
	let saving = $state(false);
	let error = $state('');
	let successMsg = $state('');

	// Form fields
	let fname = $state('');
	let fslug = $state('');
	let fissuer = $state('');
	let fclientId = $state('');
	let fclientSecret = $state('');
	let fredirectUri = $state('');
	let fscopes = $state('openid email profile');
	let fautoRegister = $state(false);
	let fenabled = $state(true);

	$effect(() => {
		providers = data.providers;
	});

	function computeRedirectUri(slug: string): string {
		return `${window.location.origin}/api/auth/oidc/${slug}/callback`;
	}

	function resetForm() {
		fname = ''; fslug = ''; fissuer = ''; fclientId = ''; fclientSecret = '';
		fredirectUri = ''; fscopes = 'openid email profile';
		fautoRegister = false; fenabled = true;
		editingId = null; showForm = false;
	}

	function startCreate() {
		resetForm();
		showForm = true;
	}

	function startEdit(p: OIDCProvider) {
		fname = p.name; fslug = p.slug; fissuer = p.issuer_url; fclientId = p.client_id;
		fclientSecret = '********'; fredirectUri = p.redirect_uri; fscopes = p.scopes;
		fautoRegister = p.auto_register; fenabled = p.enabled;
		editingId = p.id; showForm = true;
	}

	async function handleSave() {
		error = ''; successMsg = ''; saving = true;
		try {
			if (editingId) {
				const updated = await instance.oidc.update(editingId, {
					name: fname, issuer_url: fissuer, client_id: fclientId,
					client_secret: fclientSecret, redirect_uri: fredirectUri,
					scopes: fscopes, auto_register: fautoRegister, enabled: fenabled
				});
				providers = providers.map(p => p.id === editingId ? updated : p);
			} else {
				const uri = fredirectUri || computeRedirectUri(fslug);
				const created = await instance.oidc.create({
					name: fname, slug: fslug, issuer_url: fissuer, client_id: fclientId,
					client_secret: fclientSecret, redirect_uri: uri,
					scopes: fscopes, auto_register: fautoRegister, enabled: fenabled
				});
				providers = [...providers, created];
			}
			successMsg = t.saved;
			setTimeout(() => { successMsg = ''; }, 3000);
			resetForm();
		} catch (err) {
			error = err instanceof Error ? err.message : t.saveError;
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm(t.deleteConfirm)) return;
		try {
			await instance.oidc.delete(id);
			providers = providers.filter(p => p.id !== id);
			successMsg = t.deleted;
			setTimeout(() => { successMsg = ''; }, 3000);
			if (editingId === id) resetForm();
		} catch (err) {
			error = err instanceof Error ? err.message : t.deleteError;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h2 class="text-lg font-semibold">{t.title}</h2>
		{#if !showForm}
			<Button onclick={startCreate}>{t.add}</Button>
		{/if}
	</div>

	{#if successMsg}
		<div class="rounded-md bg-green-50 p-3 text-sm text-green-700 border border-green-200">
			{successMsg}
		</div>
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Content class="space-y-4 pt-6">
				<div class="space-y-1.5">
					<label for="oidc-name" class="text-sm font-medium">{t.name}</label>
					<Input id="oidc-name" bind:value={fname} placeholder="Google" />
				</div>

				{#if !editingId}
					<div class="space-y-1.5">
						<label for="oidc-slug" class="text-sm font-medium">{t.slug}</label>
						<Input id="oidc-slug" bind:value={fslug} placeholder="google" />
						<p class="text-xs text-muted-foreground">{t.slugHint}</p>
					</div>
				{/if}

				<div class="space-y-1.5">
					<label for="oidc-issuer" class="text-sm font-medium">{t.issuer}</label>
					<Input id="oidc-issuer" bind:value={fissuer} placeholder="https://accounts.google.com" />
					<p class="text-xs text-muted-foreground">{t.issuerHint}</p>
				</div>

				<Separator />

				<div class="grid grid-cols-2 gap-4">
					<div class="space-y-1.5">
						<label for="oidc-client-id" class="text-sm font-medium">{t.clientId}</label>
						<Input id="oidc-client-id" bind:value={fclientId} />
					</div>
					<div class="space-y-1.5">
						<label for="oidc-client-secret" class="text-sm font-medium">{t.clientSecret}</label>
						<Input id="oidc-client-secret" type="password" bind:value={fclientSecret} />
					</div>
				</div>

				<div class="space-y-1.5">
					<label for="oidc-redirect" class="text-sm font-medium">{t.redirectUri}</label>
					<Input id="oidc-redirect" bind:value={fredirectUri} placeholder={fslug ? computeRedirectUri(fslug) : ''} />
					<p class="text-xs text-muted-foreground">{t.redirectUriHint}</p>
				</div>

				<div class="space-y-1.5">
					<label for="oidc-scopes" class="text-sm font-medium">{t.scopes}</label>
					<Input id="oidc-scopes" bind:value={fscopes} placeholder="openid email profile" />
				</div>

				<Separator />

				<div class="flex items-center gap-6">
					<label class="flex items-center gap-2 text-sm">
						<input type="checkbox" bind:checked={fautoRegister} class="rounded" />
						{t.autoRegister}
					</label>
					<label class="flex items-center gap-2 text-sm">
						<input type="checkbox" bind:checked={fenabled} class="rounded" />
						{t.enabled}
					</label>
				</div>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<div class="flex items-center gap-3 pt-2">
					<Button onclick={handleSave} disabled={saving || !fname || (!editingId && !fslug) || !fissuer || !fclientId}>
						{saving ? t.saving : t.save}
					</Button>
					<Button variant="outline" onclick={resetForm}>{t.cancel}</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if providers.length === 0 && !showForm}
		<Card.Root>
			<Card.Content class="py-12 text-center">
				<p class="text-sm font-medium text-muted-foreground">{t.noProviders}</p>
				<p class="mt-1 text-xs text-muted-foreground">{t.noProvidersDesc}</p>
			</Card.Content>
		</Card.Root>
	{:else if providers.length > 0}
		<div class="space-y-2">
			{#each providers as provider}
				<Card.Root>
					<Card.Content class="flex items-center justify-between py-4">
						<div>
							<p class="text-sm font-medium">{provider.name}</p>
							<p class="text-xs text-muted-foreground">{provider.slug} &middot; {provider.issuer_url}</p>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs {provider.enabled ? 'text-green-600' : 'text-muted-foreground'}">
								{provider.enabled ? t.enabled : t.disabledLabel}
							</span>
							<Button variant="outline" size="sm" onclick={() => startEdit(provider)}>{t.edit}</Button>
							<Button variant="destructive" size="sm" onclick={() => handleDelete(provider.id)}>
								{t.deleteLabel}
							</Button>
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
