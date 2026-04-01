// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
	const { user } = await parent();
	if (!user?.is_instance_admin) redirect(302, '/');

	let smtpConfig = { host: '', port: 587, from: '', username: '', password: '' };
	try {
		const config = await instance.smtp.get();
		if (config && config.host) smtpConfig = { ...smtpConfig, ...config };
	} catch {
		// SMTP not configured yet — use defaults
	}

	return { smtpConfig };
};
