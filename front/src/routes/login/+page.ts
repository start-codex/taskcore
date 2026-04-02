// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance, auth, ApiError } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const { initialized } = await instance.status();
	if (!initialized) redirect(302, '/setup');

	let providers: Awaited<ReturnType<typeof auth.oidcProviders>> = [];
	let oidcLoadFailed = false;
	try {
		providers = await auth.oidcProviders();
	} catch (err) {
		// 404 means endpoint not available (e.g., table not migrated yet) — safe to ignore
		if (!(err instanceof ApiError && err.status === 404)) {
			oidcLoadFailed = true;
		}
	}

	return { providers, oidcLoadFailed };
};
