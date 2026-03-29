<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { PageData } from './$types';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import LayoutIcon from '@lucide/svelte/icons/layout';
	import { boards as boardsApi } from '$lib/api';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title:    m.project_no_boards_title(),
			desc:     m.project_no_boards_desc({ name: data.project.name }),
			create:   m.board_create(),
			name:     m.board_name(),
			creating: m.board_creating(),
			cancel:   m.workspace_cancel()
		};
	});

	let sheetOpen  = $state(false);
	let boardName  = $state('');
	let saving     = $state(false);
	let error      = $state('');

	function resetForm() { boardName = ''; error = ''; saving = false; }

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const board = await boardsApi.create(data.project.id, { name: boardName.trim(), type: 'kanban' });
			sheetOpen = false;
			resetForm();
			goto(`/${page.params.workspace}/boards/${board.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create board';
		} finally {
			saving = false;
		}
	}
</script>

<Empty.Root class="border border-dashed">
	<Empty.Header>
		<Empty.Media variant="icon"><LayoutIcon /></Empty.Media>
		<Empty.Title>{t.title}</Empty.Title>
		<Empty.Description>{t.desc}</Empty.Description>
	</Empty.Header>
	<Empty.Content>
		<Button onclick={() => { sheetOpen = true; }}>{t.create}</Button>
	</Empty.Content>
</Empty.Root>

<Sheet.Root bind:open={sheetOpen} onOpenChange={(open) => { if (!open) resetForm(); }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-96">
			<Sheet.Header>
				<Sheet.Title>{t.create}</Sheet.Title>
			</Sheet.Header>
			<form onsubmit={handleCreate} class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1.5">
					<label for="board-name" class="text-sm font-medium">{t.name}</label>
					<Input id="board-name" placeholder="My Board" bind:value={boardName} required />
				</div>
				{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
				<div class="flex justify-end gap-2 pt-2">
					<Sheet.Close><Button variant="outline" type="button">{t.cancel}</Button></Sheet.Close>
					<Button type="submit" disabled={saving || !boardName.trim()}>
						{saving ? t.creating : t.create}
					</Button>
				</div>
			</form>
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
