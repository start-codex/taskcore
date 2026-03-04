<script lang="ts">
	import { onMount } from 'svelte';
	import NavProjects from "./nav-projects.svelte";
	import NavUser from "./nav-user.svelte";
	import TeamSwitcher from "./team-switcher.svelte";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import type { ComponentProps } from "svelte";
	import { workspaces, projects } from '$lib/api';
	import { currentUser } from '$lib/stores/auth';
	import { selectWorkspace, getStoredWorkspaceId, selectedWorkspace, workspaceProjects } from '$lib/stores/workspace';
	import type { Workspace, Project } from '$lib/api';

	let {
		ref = $bindable(null),
		collapsible = "icon",
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();

	let workspaceList = $state<Workspace[]>([]);
	let activeWorkspace = $state<Workspace | null>(null);
	let projectList = $state<Project[]>([]);

	async function loadProjects(workspace: Workspace): Promise<Project[]> {
		try {
			return await projects.list(workspace.id);
		} catch {
			return [];
		}
	}

	async function handleWorkspaceSelect(workspace: Workspace): Promise<void> {
		const projs = await loadProjects(workspace);
		selectWorkspace(workspace, projs);
		activeWorkspace = workspace;
		projectList = projs;
	}

	onMount(async () => {
		const user = $currentUser;
		if (!user) return;

		try {
			workspaceList = await workspaces.listByUser(user.id);
		} catch {
			workspaceList = [];
		}

		if (workspaceList.length === 0) return;

		const storedId = getStoredWorkspaceId();
		const initial = workspaceList.find(w => w.id === storedId) ?? workspaceList[0];
		await handleWorkspaceSelect(initial);
	});
</script>

<Sidebar.Root {collapsible} {...restProps}>
	<Sidebar.Header>
		<TeamSwitcher
			workspaces={workspaceList}
			selected={activeWorkspace}
			onSelect={handleWorkspaceSelect}
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
