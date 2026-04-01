<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import { instance } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.admin_smtp_title(),
			host: m.admin_smtp_host(),
			port: m.admin_smtp_port(),
			from: m.admin_smtp_from(),
			username: m.admin_smtp_username(),
			password: m.admin_smtp_password(),
			save: m.admin_smtp_save(),
			saving: m.admin_smtp_saving(),
			saved: m.admin_smtp_saved(),
			test: m.admin_smtp_test(),
			testSent: m.admin_smtp_test_sent,
			testError: m.admin_smtp_test_error()
		};
	});

	let host = $state('');
	let port = $state(587);
	let from = $state('');
	let username = $state('');
	let password = $state('');

	let saving = $state(false);
	let saved = $state(false);
	let testing = $state(false);
	let error = $state('');
	let testResult = $state('');

	$effect(() => {
		host = data.smtpConfig.host;
		port = data.smtpConfig.port;
		from = data.smtpConfig.from;
		username = data.smtpConfig.username ?? '';
		password = data.smtpConfig.password ?? '';
	});

	async function handleSave() {
		error = '';
		saved = false;
		saving = true;
		try {
			await instance.smtp.save({ host, port, from, username, password });
			saved = true;
			setTimeout(() => { saved = false; }, 3000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save';
		} finally {
			saving = false;
		}
	}

	async function handleTest() {
		testResult = '';
		error = '';
		testing = true;
		try {
			const result = await instance.smtp.test();
			testResult = t.testSent({ email: result.to });
			setTimeout(() => { testResult = ''; }, 5000);
		} catch (err) {
			error = err instanceof Error ? err.message : t.testError;
		} finally {
			testing = false;
		}
	}
</script>

<div class="space-y-6">
	<h2 class="text-lg font-semibold">{t.title}</h2>

	<Card.Root>
		<Card.Content class="space-y-4 pt-6">
			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-1.5">
					<label for="smtp-host" class="text-sm font-medium">{t.host}</label>
					<Input id="smtp-host" bind:value={host} placeholder="smtp.example.com" />
				</div>
				<div class="space-y-1.5">
					<label for="smtp-port" class="text-sm font-medium">{t.port}</label>
					<Input id="smtp-port" type="number" bind:value={port} />
				</div>
			</div>

			<div class="space-y-1.5">
				<label for="smtp-from" class="text-sm font-medium">{t.from}</label>
				<Input id="smtp-from" bind:value={from} placeholder="noreply@example.com" />
			</div>

			<Separator />

			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-1.5">
					<label for="smtp-user" class="text-sm font-medium">{t.username}</label>
					<Input id="smtp-user" bind:value={username} placeholder="(optional)" />
				</div>
				<div class="space-y-1.5">
					<label for="smtp-pass" class="text-sm font-medium">{t.password}</label>
					<Input id="smtp-pass" type="password" bind:value={password} placeholder="(optional)" />
				</div>
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<div class="flex items-center gap-3 pt-2">
				<Button onclick={handleSave} disabled={saving || !host || !from}>
					{saving ? t.saving : t.save}
				</Button>
				<Button variant="outline" onclick={handleTest} disabled={testing || !host}>
					{t.test}
				</Button>
				{#if saved}
					<span class="text-sm text-green-600">{t.saved}</span>
				{/if}
				{#if testResult}
					<span class="text-sm text-green-600">{testResult}</span>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>
</div>
