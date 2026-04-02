// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance, auth, ApiError } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const { initialized } = await instance.status();
	if (!initialized) redirect(302, '/setup');

	const token = url.searchParams.get('token') ?? '';
	if (!token) return { status: 'invalid' as const };

	try {
		await auth.verifyEmail({ token });
		return { status: 'success' as const };
	} catch (err) {
		if (err instanceof ApiError && err.status === 400) {
			return { status: 'invalid' as const };
		}
		return { status: 'error' as const };
	}
};
