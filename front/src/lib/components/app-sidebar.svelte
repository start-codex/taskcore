<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import NavProjects from './nav-projects.svelte';
	import NavUser from './nav-user.svelte';
	import TeamSwitcher from './team-switcher.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import type { ComponentProps } from 'svelte';
	import {
		selectedWorkspace,
		workspaceProjects,
		workspaceList,
		addToWorkspaceList
	} from '$lib/stores/workspace';
	import type { Workspace, Project } from '$lib/api';

	let {
		ref = $bindable(null),
		collapsible = 'icon',
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();

	let activeWorkspace = $state<Workspace | null>(null);
	let projectList     = $state<Project[]>([]);

	$effect(() => { activeWorkspace = $selectedWorkspace; });
	$effect(() => { projectList = $workspaceProjects; });

	function handleWorkspaceSelect(workspace: Workspace): void {
		goto(`/${workspace.slug}`);
	}

	function handleWorkspaceCreate(workspace: Workspace): void {
		addToWorkspaceList(workspace);
		goto(`/${workspace.slug}`);
	}
</script>

<Sidebar.Root {collapsible} {...restProps}>
	<Sidebar.Header>
		<TeamSwitcher
			workspaces={$workspaceList}
			selected={activeWorkspace}
			onSelect={handleWorkspaceSelect}
			onCreate={handleWorkspaceCreate}
		/>
	</Sidebar.Header>
	<Sidebar.Content>
		<NavProjects projects={projectList} />
	</Sidebar.Content>
	<Sidebar.Footer>
		<NavUser />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
