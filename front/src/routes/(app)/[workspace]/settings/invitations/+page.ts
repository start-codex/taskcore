// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { invitations } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent, params }) => {
	const data = await parent();
	if (!data.canManageInvitations) redirect(302, `/${params.workspace}/settings/preferences`);

	const pending = await invitations.listPending(data.workspace.id);
	return { pending: pending ?? [] };
};
