<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import BuildingIcon from '@lucide/svelte/icons/building-2';
	import { workspaces as workspacesApi } from '$lib/api';
	import { currentUser } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const t = $derived.by(() => {
		i18n.locale;
		return {
			noWorkspaceTitle: m.dashboard_no_workspace_title(),
			noWorkspaceDesc:  m.dashboard_no_workspace_desc(),
			createWorkspace:  m.workspace_create(),
			createTitle:      m.workspace_create_title(),
			name:             m.workspace_name(),
			slug:             m.workspace_slug(),
			slugHint:         m.workspace_slug_hint(),
			cancel:           m.workspace_cancel(),
			creating:         m.workspace_creating()
		};
	});

	let ready = $state(false);
	let wsSheetOpen = $state(false);
	let wsName = $state('');
	let wsSlug = $state('');
	let wsError = $state('');
	let wsSaving = $state(false);

	onMount(async () => {
		const user = $currentUser;
		if (!user) { goto('/login'); return; }
		try {
			const list = await workspacesApi.list();
			if (list && list.length > 0) {
				goto(`/${list[0].slug}`);
				return;
			}
		} catch {}
		ready = true;
	});

	function toSlug(v: string) { return v.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, ''); }
	function onWsNameInput(e: Event) {
		wsName = (e.target as HTMLInputElement).value;
		wsSlug = toSlug(wsName);
	}
	function resetForm() { wsName = ''; wsSlug = ''; wsError = ''; wsSaving = false; }

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		wsError = '';
		wsSaving = true;
		try {
			const ws = await workspacesApi.create({ name: wsName.trim(), slug: wsSlug.trim() });
			goto(`/${ws.slug}`);
		} catch (err) {
			wsError = err instanceof Error ? err.message : 'Failed to create workspace';
			wsSaving = false;
		}
	}
</script>

{#if ready}
	<div class="flex min-h-screen items-center justify-center">
		<div class="w-full max-w-sm space-y-6 p-8">
			<Empty.Root>
				<Empty.Header>
					<Empty.Media variant="icon"><BuildingIcon /></Empty.Media>
					<Empty.Title>{t.noWorkspaceTitle}</Empty.Title>
					<Empty.Description>{t.noWorkspaceDesc}</Empty.Description>
				</Empty.Header>
				<Empty.Content>
					<Button onclick={() => { wsSheetOpen = true; }}>{t.createWorkspace}</Button>
				</Empty.Content>
			</Empty.Root>
		</div>
	</div>
{/if}

<Sheet.Root bind:open={wsSheetOpen} onOpenChange={(open) => { if (!open) resetForm(); }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-96">
			<Sheet.Header>
				<Sheet.Title>{t.createTitle}</Sheet.Title>
			</Sheet.Header>
			<form onsubmit={handleCreate} class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1.5">
					<label for="ws-name" class="text-sm font-medium">{t.name}</label>
					<Input id="ws-name" placeholder="Acme Corp" value={wsName} oninput={onWsNameInput} required />
				</div>
				<div class="flex flex-col gap-1.5">
					<label for="ws-slug" class="text-sm font-medium">{t.slug}</label>
					<Input id="ws-slug" placeholder="acme-corp" bind:value={wsSlug} required />
					<p class="text-xs text-muted-foreground">{t.slugHint}</p>
				</div>
				{#if wsError}<p class="text-sm text-destructive">{wsError}</p>{/if}
				<div class="flex justify-end gap-2 pt-2">
					<Sheet.Close><Button variant="outline" type="button">{t.cancel}</Button></Sheet.Close>
					<Button type="submit" disabled={wsSaving || !wsName.trim() || !wsSlug.trim()}>
						{wsSaving ? t.creating : t.createWorkspace}
					</Button>
				</div>
			</form>
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
