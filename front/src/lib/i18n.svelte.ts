// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { browser } from '$app/environment';
import { overwriteGetLocale, locales, baseLocale } from '$lib/paraglide/runtime';
import type { Locale } from '$lib/paraglide/runtime';

const STORAGE_KEY = 'PARAGLIDE_LOCALE';

// Wrap in an object so Svelte allows exporting — mutate the property, never reassign
export const i18n = $state({ locale: baseLocale as Locale });

// Override Paraglide's getLocale to read from the reactive object
overwriteGetLocale(() => i18n.locale);

export function initLocale(): void {
	if (!browser) return;

	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored && (locales as readonly string[]).includes(stored)) {
		i18n.locale = stored as Locale;
		return;
	}

	const browserLang = navigator.language.split('-')[0];
	if ((locales as readonly string[]).includes(browserLang)) {
		i18n.locale = browserLang as Locale;
	}
}

export function switchLocale(lang: Locale): void {
	i18n.locale = lang;
	if (browser) localStorage.setItem(STORAGE_KEY, lang);
}
