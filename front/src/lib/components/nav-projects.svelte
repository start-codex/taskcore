<script lang="ts">
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import LayoutIcon from "@lucide/svelte/icons/layout";
	import type { Project } from '$lib/api';

	let { projects }: { projects: Project[] } = $props();
</script>

<Sidebar.Group class="group-data-[collapsible=icon]:hidden">
	<Sidebar.GroupLabel>Projects</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each projects as project (project.id)}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton>
					{#snippet child({ props })}
						<a href={`/projects/${project.id}`} {...props}>
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
				<span class="px-2 py-1 text-xs text-muted-foreground">No projects</span>
			</Sidebar.MenuItem>
		{/if}
	</Sidebar.Menu>
</Sidebar.Group>
