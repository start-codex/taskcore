// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';

export function load() {
	if (browser && !localStorage.getItem('user')) {
		redirect(302, '/login');
	}
}
