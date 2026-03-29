<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import LayoutIcon from '@lucide/svelte/icons/layout';
	import FolderIcon from '@lucide/svelte/icons/folder';
	import { selectedWorkspace, workspaceProjects, addWorkspaceProject } from '$lib/stores/workspace';
	import { projects as projectsApi } from '$lib/api';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const ws = $derived(page.params.workspace);

	const t = $derived.by(() => {
		i18n.locale;
		return {
			noProjectsTitle: m.dashboard_no_projects_title(),
			noProjectsDesc:  m.dashboard_no_projects_desc(),
			createProject:   m.project_create(),
			name:            m.workspace_name(),
			key:             m.project_key(),
			keyHint:         m.project_key_hint(),
			cancel:          m.workspace_cancel(),
			creating:        m.workspace_creating(),
			projects:        m.nav_projects(),
			template:        m.project_template(),
			tmplKanban:      m.project_template_kanban(),
			tmplKanbanDesc:  m.project_template_kanban_desc(),
			tmplScrum:       m.project_template_scrum(),
			tmplScrumDesc:   m.project_template_scrum_desc(),
			next:            m.project_next(),
			back:            m.project_back()
		};
	});

	let projSheetOpen = $state(false);
	let projStep      = $state<'template' | 'details'>('template');
	let projTemplate  = $state<'kanban' | 'scrum'>('kanban');
	let projName      = $state('');
	let projKey       = $state('');
	let projError     = $state('');
	let projSaving    = $state(false);

	function toKey(v: string) { return v.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 6); }
	function onProjNameInput(e: Event) {
		projName = (e.target as HTMLInputElement).value;
		projKey  = toKey(projName);
	}
	function resetProjForm() {
		projStep = 'template'; projTemplate = 'kanban';
		projName = ''; projKey = ''; projError = ''; projSaving = false;
	}

	async function handleCreateProject(e: SubmitEvent) {
		e.preventDefault();
		projError = '';
		projSaving = true;
		try {
			const project = await projectsApi.create($selectedWorkspace!.id, {
				name:     projName.trim(),
				key:      projKey.trim(),
				template: projTemplate,
				locale:   i18n.locale
			});
			addWorkspaceProject(project);
			projSheetOpen = false;
			resetProjForm();
		} catch (err) {
			projError = err instanceof Error ? err.message : 'Failed to create project';
		} finally {
			projSaving = false;
		}
	}

	const projects = $derived($workspaceProjects ?? []);
</script>

{#if projects.length === 0}
	<Empty.Root class="border border-dashed">
		<Empty.Header>
			<Empty.Media variant="icon"><LayoutIcon /></Empty.Media>
			<Empty.Title>{t.noProjectsTitle}</Empty.Title>
			<Empty.Description>{t.noProjectsDesc}</Empty.Description>
		</Empty.Header>
		<Empty.Content>
			<Button onclick={() => { projSheetOpen = true; }}>{t.createProject}</Button>
		</Empty.Content>
	</Empty.Root>
{:else}
	<div class="space-y-4">
		<div class="flex items-center justify-between">
			<h2 class="text-lg font-semibold">{t.projects}</h2>
			<Button variant="outline" size="sm" onclick={() => { projSheetOpen = true; }}>
				{t.createProject}
			</Button>
		</div>
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each projects as project (project.id)}
				<a
					href="/{ws}/projects/{project.id}"
					class="group flex flex-col gap-3 rounded-lg border bg-card p-5 shadow-sm transition-shadow hover:shadow-md"
				>
					<div class="flex items-start justify-between gap-2">
						<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
							<FolderIcon class="h-4 w-4" />
						</div>
						<span class="rounded bg-muted px-2 py-0.5 font-mono text-xs text-muted-foreground">
							{project.key}
						</span>
					</div>
					<div>
						<p class="font-medium leading-tight group-hover:text-primary">{project.name}</p>
						{#if project.description}
							<p class="mt-1 text-sm text-muted-foreground line-clamp-2">{project.description}</p>
						{/if}
					</div>
				</a>
			{/each}
		</div>
	</div>
{/if}

<Sheet.Root bind:open={projSheetOpen} onOpenChange={(open) => { if (!open) resetProjForm(); }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-[440px]">
			<Sheet.Header>
				<Sheet.Title>{t.createProject}</Sheet.Title>
			</Sheet.Header>
			{#if projStep === 'template'}
				<div class="flex flex-col gap-4 p-6">
					<p class="text-sm font-medium">{t.template}</p>
					<div class="flex flex-col gap-3">
						{#each ([['kanban', t.tmplKanban, t.tmplKanbanDesc], ['scrum', t.tmplScrum, t.tmplScrumDesc]] as const) as [value, label, desc]}
							<button
								type="button"
								onclick={() => { projTemplate = value; }}
								class="flex flex-col gap-1 rounded-lg border p-4 text-left transition-colors
								       {projTemplate === value ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'hover:bg-muted/50'}"
							>
								<span class="font-medium">{label}</span>
								<span class="text-xs text-muted-foreground">{desc}</span>
							</button>
						{/each}
					</div>
					<div class="flex justify-end gap-2 pt-2">
						<Sheet.Close><Button variant="outline" type="button">{t.cancel}</Button></Sheet.Close>
						<Button onclick={() => { projStep = 'details'; }}>{t.next}</Button>
					</div>
				</div>
			{:else}
				<form onsubmit={handleCreateProject} class="flex flex-col gap-4 p-6">
					<div class="flex flex-col gap-1.5">
						<label for="proj-name" class="text-sm font-medium">{t.name}</label>
						<Input id="proj-name" placeholder="My Project" value={projName} oninput={onProjNameInput} required />
					</div>
					<div class="flex flex-col gap-1.5">
						<label for="proj-key" class="text-sm font-medium">{t.key}</label>
						<Input id="proj-key" placeholder="PROJ" bind:value={projKey} required />
						<p class="text-xs text-muted-foreground">{t.keyHint}</p>
					</div>
					{#if projError}<p class="text-sm text-destructive">{projError}</p>{/if}
					<div class="flex justify-end gap-2 pt-2">
						<Button variant="outline" type="button" onclick={() => { projStep = 'template'; projError = ''; }}>{t.back}</Button>
						<Button type="submit" disabled={projSaving || !projName.trim() || !projKey.trim()}>
							{projSaving ? t.creating : t.createProject}
						</Button>
					</div>
				</form>
			{/if}
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
