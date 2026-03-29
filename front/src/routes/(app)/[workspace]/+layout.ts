// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import { workspaces as workspacesApi, projects as projectsApi } from '$lib/api';
import type { LayoutLoad } from './$types';

export const ssr = false;

export const load: LayoutLoad = async ({ params }) => {
	const raw = browser ? localStorage.getItem('user') : null;
	if (!raw) redirect(302, '/login');

	const user = JSON.parse(raw);
	const list = await workspacesApi.list().catch(() => []);
	const workspace = list.find((w) => w.slug === params.workspace);

	if (!workspace) redirect(302, '/');

	const projectList = await projectsApi.list(workspace.id).catch(() => []);

	return {
		workspace,
		workspaceProjects: projectList ?? [],
		workspaceList: list
	};
};
