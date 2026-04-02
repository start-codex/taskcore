// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance, auth, invitations, ApiError } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const { initialized } = await instance.status();
	if (!initialized) redirect(302, '/setup');

	const token = url.searchParams.get('token') ?? '';
	if (!token) return { token: '', invitation: null, authedUser: null, serverError: false };

	let invitation = null;
	let serverError = false;
	try {
		invitation = await invitations.getAccept(token);
	} catch (err) {
		if (err instanceof ApiError && (err.status === 400 || err.status === 404)) {
			// Invalid/expired token — page will show invalid state
		} else {
			serverError = true;
		}
	}

	let authedUser = null;
	try {
		const me = await auth.me();
		if (me.authenticated && me.user) authedUser = me.user;
	} catch {
		// Not authenticated
	}

	return { token, invitation, authedUser, serverError };
};
