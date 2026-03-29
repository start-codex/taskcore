<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import type { Issue, Status } from '$lib/api';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import LayoutIcon from '@lucide/svelte/icons/layout';
	import { statuses as statusesApi } from '$lib/api';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			noIssues:       m.board_no_issues(),
			noStatuses:     m.board_no_statuses(),
			noStatusesDesc: m.board_no_statuses_desc(),
			createStatus:   m.status_create(),
			name:           m.status_name(),
			category:       m.status_category(),
			creating:       m.status_creating(),
			cancel:         m.workspace_cancel(),
			catTodo:        m.status_category_todo(),
			catInProgress:  m.status_category_in_progress(),
			catDone:        m.status_category_done()
		};
	});

	const priorityColors: Record<string, string> = {
		urgent: 'bg-red-100 text-red-700',
		high: 'bg-orange-100 text-orange-700',
		medium: 'bg-yellow-100 text-yellow-700',
		low: 'bg-blue-100 text-blue-700'
	};

	// local reactive list — syncs from data on navigation, allows optimistic adds
	let localStatuses = $state<Status[]>([]);
	$effect(() => { localStatuses = [...data.statuses]; });

	const sortedStatuses = $derived([...localStatuses].sort((a, b) => a.position - b.position));

	function issuesForStatus(statusId: string): Issue[] {
		return data.issues.filter((i) => i.status_id === statusId);
	}

	// --- Create status sheet ---
	let sheetOpen    = $state(false);
	let statusName   = $state('');
	let category     = $state<'todo' | 'doing' | 'done'>('todo');
	let saving       = $state(false);
	let error        = $state('');

	function resetForm() { statusName = ''; category = 'todo'; error = ''; saving = false; }

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const created = await statusesApi.create(data.board.project_id, {
				name: statusName.trim(),
				category
			});
			localStatuses = [...localStatuses, created];
			sheetOpen = false;
			resetForm();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create status';
		} finally {
			saving = false;
		}
	}
</script>

{#if sortedStatuses.length === 0}
	<Empty.Root class="border border-dashed">
		<Empty.Header>
			<Empty.Media variant="icon"><LayoutIcon /></Empty.Media>
			<Empty.Title>{t.noStatuses}</Empty.Title>
			<Empty.Description>{t.noStatusesDesc}</Empty.Description>
		</Empty.Header>
		<Empty.Content>
			<Button onclick={() => { sheetOpen = true; }}>{t.createStatus}</Button>
		</Empty.Content>
	</Empty.Root>
{:else}
	<div class="flex h-full min-h-0 gap-4 overflow-x-auto pb-4">
		{#each sortedStatuses as status}
			{@const columnIssues = issuesForStatus(status.id)}
			<div class="flex w-72 shrink-0 flex-col gap-3">
				<div class="flex items-center gap-2 px-1">
					<span class="text-sm font-medium">{status.name}</span>
					<span class="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
						{columnIssues.length}
					</span>
				</div>
				<div class="flex flex-1 flex-col gap-2 overflow-y-auto rounded-lg bg-muted/40 p-2">
					{#if columnIssues.length === 0}
						<div class="flex items-center justify-center py-8 text-xs text-muted-foreground">
							{t.noIssues}
						</div>
					{:else}
						{#each columnIssues as issue}
							<div class="flex flex-col gap-2 rounded-md border bg-background p-3 shadow-xs">
								<div class="flex items-start justify-between gap-2">
									<span class="text-sm leading-snug">{issue.title}</span>
									{#if issue.priority && issue.priority !== 'none'}
										<span
											class="shrink-0 rounded px-1.5 py-0.5 text-xs font-medium {priorityColors[issue.priority] ?? 'bg-muted text-muted-foreground'}"
										>
											{issue.priority}
										</span>
									{/if}
								</div>
								<span class="text-xs text-muted-foreground">#{issue.number}</span>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		{/each}
	</div>
{/if}

<Sheet.Root bind:open={sheetOpen} onOpenChange={(open) => { if (!open) resetForm(); }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-96">
			<Sheet.Header>
				<Sheet.Title>{t.createStatus}</Sheet.Title>
			</Sheet.Header>
			<form onsubmit={handleCreate} class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1.5">
					<label for="status-name" class="text-sm font-medium">{t.name}</label>
					<Input id="status-name" placeholder="To Do" bind:value={statusName} required />
				</div>
				<div class="flex flex-col gap-1.5">
					<label for="status-cat" class="text-sm font-medium">{t.category}</label>
					<select
						id="status-cat"
						bind:value={category}
						class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
					>
						<option value="todo">{t.catTodo}</option>
						<option value="doing">{t.catInProgress}</option>
						<option value="done">{t.catDone}</option>
					</select>
				</div>
				{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
				<div class="flex justify-end gap-2 pt-2">
					<Sheet.Close><Button variant="outline" type="button">{t.cancel}</Button></Sheet.Close>
					<Button type="submit" disabled={saving || !statusName.trim()}>
						{saving ? t.creating : t.createStatus}
					</Button>
				</div>
			</form>
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
