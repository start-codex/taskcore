// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import { instance, auth } from '$lib/api';
import { login } from '$lib/stores/auth';

export const ssr = false;

export async function load() {
	if (!browser) return { user: null, emailVerificationRequired: false };

	const { initialized } = await instance.status();
	if (!initialized) redirect(302, '/setup');

	const me = await auth.me();
	if (!me.authenticated || !me.user) redirect(302, '/login');

	login(me.user!);

	return {
		user: me.user!,
		emailVerificationRequired: me.email_verification_required ?? false
	};
}
