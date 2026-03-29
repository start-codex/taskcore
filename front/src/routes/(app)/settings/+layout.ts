// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import type { LayoutLoad } from './$types';

export const load: LayoutLoad = () => {
	return {
		breadcrumb: [{ labelKey: 'settings_nav' }]
	};
};
