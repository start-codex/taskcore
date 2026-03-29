// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { projects, boards } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const [project, rawBoards] = await Promise.all([
		projects.get(params.id),
		boards.list(params.id)
	]);
	const boardList = rawBoards ?? [];

	if (boardList.length > 0) {
		redirect(302, `/boards/${boardList[0].id}`);
	}

	return {
		project,
		breadcrumb: [{ label: project.name }]
	};
};
