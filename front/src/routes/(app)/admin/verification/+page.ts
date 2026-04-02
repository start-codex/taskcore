// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
	const { user } = await parent();
	if (!user?.is_instance_admin) redirect(302, '/');

	let required = false;
	try {
		const config = await instance.verification.get();
		required = config.required;
	} catch {
		// default false
	}

	return { verificationRequired: required };
};
