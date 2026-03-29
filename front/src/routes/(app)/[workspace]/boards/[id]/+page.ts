// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { boards, statuses, issues } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const board = await boards.get(params.id);

	const [statusList, issueList] = await Promise.all([
		statuses.list(board.project_id).then((r) => r ?? []),
		issues.list(board.project_id).then((r) => r ?? [])
	]);

	return {
		board,
		statuses: statusList,
		issues: issueList,
		breadcrumb: [{ label: board.name }]
	};
};
