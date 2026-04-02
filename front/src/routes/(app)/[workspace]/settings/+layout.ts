// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { workspaces, ApiError } from '$lib/api';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ parent }) => {
	const { workspace, user } = await parent();

	let canManageInvitations = false;
	try {
		const members = await workspaces.members.list(workspace.id);
		const me = members.find((m) => m.user_id === user?.id);
		canManageInvitations = me?.role === 'admin' || me?.role === 'owner';
	} catch (err) {
		if (err instanceof ApiError && (err.status === 403 || err.status === 401)) {
			canManageInvitations = false;
		} else {
			throw err; // Propagate unexpected errors
		}
	}

	return {
		breadcrumb: [{ labelKey: 'settings_nav' }],
		canManageInvitations
	};
};
