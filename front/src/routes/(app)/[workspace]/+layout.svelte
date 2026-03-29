<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';
	import {
		selectWorkspace,
		setWorkspaceList,
		markWorkspaceReady
	} from '$lib/stores/workspace';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

	$effect(() => {
		setWorkspaceList(data.workspaceList);
		selectWorkspace(data.workspace, data.workspaceProjects);
		markWorkspaceReady();
	});

	type RawCrumb = { label?: string; labelKey?: keyof typeof m; href?: string };
	type Crumb = { label: string; href?: string };

	function resolveCrumbs(raw: RawCrumb[]): Crumb[] {
		return raw.map((c) => ({
			label: c.labelKey ? String((m[c.labelKey] as () => string)()) : (c.label ?? ''),
			href: c.href
		}));
	}

	const defaultCrumb = $derived.by<Crumb[]>(() => {
		i18n.locale;
		return [{ label: String(m.nav_dashboard()) }];
	});

	const crumbs = $derived.by<Crumb[]>(() => {
		i18n.locale;
		const raw = page.data.breadcrumb as RawCrumb[] | undefined;
		return raw ? resolveCrumbs(raw) : defaultCrumb;
	});
</script>

<Sidebar.Provider>
	<AppSidebar />
	<Sidebar.Inset>
		<header
			class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
		>
			<div class="flex items-center gap-2 px-4">
				<Sidebar.Trigger class="-ms-1" />
				<Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
				<Breadcrumb.Root>
					<Breadcrumb.List>
						<Breadcrumb.Item class="hidden md:block">
							<Breadcrumb.Link href="/{data.workspace.slug}">Traza Work</Breadcrumb.Link>
						</Breadcrumb.Item>
						<Breadcrumb.Separator class="hidden md:block" />
						{#each crumbs as crumb, i}
							{#if i > 0}
								<Breadcrumb.Separator />
							{/if}
							<Breadcrumb.Item>
								{#if crumb.href && i < crumbs.length - 1}
									<Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
								{:else}
									<Breadcrumb.Page>{crumb.label}</Breadcrumb.Page>
								{/if}
							</Breadcrumb.Item>
						{/each}
					</Breadcrumb.List>
				</Breadcrumb.Root>
			</div>
		</header>
		<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
			{@render children()}
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
