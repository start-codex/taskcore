<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import LayoutIcon from "@lucide/svelte/icons/layout";
	import type { Project } from '$lib/api';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { projects }: { projects: Project[] } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			projects: m.nav_projects(),
			noProjects: m.nav_no_projects()
		};
	});
</script>

<Sidebar.Group class="group-data-[collapsible=icon]:hidden">
	<Sidebar.GroupLabel>{t.projects}</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each projects as project (project.id)}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton>
					{#snippet child({ props })}
						<a href={`/${page.params.workspace}/projects/${project.id}`} {...props}>
							<LayoutIcon />
							<span>{project.name}</span>
							<span class="ml-auto text-xs text-muted-foreground">{project.key}</span>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		{/each}
		{#if projects.length === 0}
			<Sidebar.MenuItem>
				<span class="px-2 py-1 text-xs text-muted-foreground">{t.noProjects}</span>
			</Sidebar.MenuItem>
		{/if}
	</Sidebar.Menu>
</Sidebar.Group>
