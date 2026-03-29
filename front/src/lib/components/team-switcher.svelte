<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import * as Sheet from "$lib/components/ui/sheet/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { useSidebar } from "$lib/components/ui/sidebar/index.js";
	import ChevronsUpDownIcon from "@lucide/svelte/icons/chevrons-up-down";
	import PlusIcon from "@lucide/svelte/icons/plus";
	import type { Workspace } from '$lib/api';
	import { workspaces as workspacesApi } from '$lib/api';
	import { currentUser } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';
	import { workspaceSheetOpen } from '$lib/stores/workspace';

	$effect(() => { if ($workspaceSheetOpen) sheetOpen = true; });

	let {
		workspaces,
		selected,
		onSelect,
		onCreate
	}: {
		workspaces: Workspace[];
		selected: Workspace | null;
		onSelect: (w: Workspace) => void;
		onCreate?: (w: Workspace) => void;
	} = $props();

	const sidebar = useSidebar();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			workspaces: m.workspace_workspaces(),
			noWorkspace: m.workspace_no_workspace(),
			create: m.workspace_create(),
			createTitle: m.workspace_create_title(),
			createDescription: m.workspace_create_description(),
			name: m.workspace_name(),
			slug: m.workspace_slug(),
			slugHint: m.workspace_slug_hint(),
			cancel: m.workspace_cancel(),
			creating: m.workspace_creating()
		};
	});

	let sheetOpen = $state(false);
	let name = $state('');
	let slug = $state('');
	let error = $state('');
	let saving = $state(false);

	function toSlug(value: string): string {
		return value.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
	}

	function onNameInput(e: Event) {
		name = (e.target as HTMLInputElement).value;
		slug = toSlug(name);
	}

	function resetForm() {
		name = '';
		slug = '';
		error = '';
		saving = false;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const workspace = await workspacesApi.create({ name: name.trim(), slug: slug.trim(), owner_id: $currentUser!.id });
			sheetOpen = false;
			resetForm();
			onCreate?.(workspace);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create workspace';
			saving = false;
		}
	}
</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						{...props}
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
					>
						<div
							class="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg font-semibold"
						>
							{selected ? selected.name[0].toUpperCase() : '?'}
						</div>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-medium">
								{selected ? selected.name : t.noWorkspace}
							</span>
							<span class="truncate text-xs">{selected ? selected.slug : ''}</span>
						</div>
						<ChevronsUpDownIcon class="ms-auto" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				align="start"
				side={sidebar.isMobile ? "bottom" : "right"}
				sideOffset={4}
			>
				<DropdownMenu.Label class="text-muted-foreground text-xs">{t.workspaces}</DropdownMenu.Label>
				{#each workspaces as workspace, index (workspace.id)}
					<DropdownMenu.Item onSelect={() => onSelect(workspace)} class="gap-2 p-2">
						<div class="flex size-6 items-center justify-center rounded-md border font-semibold text-xs">
							{workspace.name[0].toUpperCase()}
						</div>
						{workspace.name}
						<DropdownMenu.Shortcut>⌘{index + 1}</DropdownMenu.Shortcut>
					</DropdownMenu.Item>
				{/each}
				<DropdownMenu.Separator />
				<DropdownMenu.Item onSelect={() => { sheetOpen = true; }} class="gap-2 p-2">
					<div class="flex size-6 items-center justify-center rounded-md border">
						<PlusIcon class="size-4" />
					</div>
					{t.create}
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>

<Sheet.Root bind:open={sheetOpen} onOpenChange={(open) => { if (!open) { resetForm(); workspaceSheetOpen.set(false); } }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-96">
			<Sheet.Header>
				<Sheet.Title>{t.createTitle}</Sheet.Title>
				<Sheet.Description>{t.createDescription}</Sheet.Description>
			</Sheet.Header>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1.5">
					<label for="ws-name" class="text-sm font-medium">{t.name}</label>
					<Input
						id="ws-name"
						placeholder="Acme Corp"
						value={name}
						oninput={onNameInput}
						required
					/>
				</div>
				<div class="flex flex-col gap-1.5">
					<label for="ws-slug" class="text-sm font-medium">{t.slug}</label>
					<Input
						id="ws-slug"
						placeholder="acme-corp"
						bind:value={slug}
						required
					/>
					<p class="text-xs text-muted-foreground">{t.slugHint}</p>
				</div>
				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
				<div class="flex justify-end gap-2 pt-2">
					<Sheet.Close>
						<Button variant="outline" type="button">{t.cancel}</Button>
					</Sheet.Close>
					<Button type="submit" disabled={saving || !name.trim() || !slug.trim()}>
						{saving ? t.creating : t.create}
					</Button>
				</div>
			</form>
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
