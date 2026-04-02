<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import type { PageData } from './$types';
	import { invitations as invitationsApi, type Invitation } from '$lib/api';
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
			title: m.invitations_title(),
			invite: m.invitations_invite(),
			email: m.invitations_email(),
			role: m.invitations_role(),
			send: m.invitations_send(),
			sending: m.invitations_sending(),
			sent: m.invitations_sent(),
			noPending: m.invitations_no_pending(),
			resend: m.invitations_resend(),
			revoke: m.invitations_revoke(),
			revoked: m.invitations_revoked(),
			resent: m.invitations_resent(),
			expires: m.invitations_expires()
		};
	});

	let invEmail = $state('');
	let invRole = $state('member');
	let sending = $state(false);
	let error = $state('');
	let success = $state('');
	let pendingList = $state<Invitation[]>([]);

	$effect(() => { pendingList = [...data.pending]; });

	async function refreshList() {
		try {
			const fresh = await invitationsApi.listPending(data.workspace.id);
			pendingList = fresh ?? [];
		} catch { /* keep current */ }
	}

	async function handleInvite(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		success = '';
		sending = true;
		try {
			await invitationsApi.create(data.workspace.id, { email: invEmail, role: invRole });
			success = t.sent;
			invEmail = '';
			invRole = 'member';
			await refreshList();
			setTimeout(() => { success = ''; }, 3000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to send';
		} finally {
			sending = false;
		}
	}

	async function handleResend(id: string) {
		try {
			await invitationsApi.resend(id);
			await refreshList();
			success = t.resent;
			setTimeout(() => { success = ''; }, 3000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to resend';
		}
	}

	async function handleRevoke(id: string) {
		try {
			await invitationsApi.revoke(id);
			pendingList = pendingList.filter((inv) => inv.id !== id);
			success = t.revoked;
			setTimeout(() => { success = ''; }, 3000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to revoke';
		}
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
	}
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-lg font-semibold">{t.title}</h2>
		<hr class="mt-3 border-border" />
	</div>

	<!-- Invite form -->
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-base">{t.invite}</Card.Title>
		</Card.Header>
		<Card.Content>
			<form onsubmit={handleInvite} class="flex items-end gap-3">
				<div class="flex-1 space-y-1.5">
					<label for="inv-email" class="text-sm font-medium">{t.email}</label>
					<Input id="inv-email" type="email" placeholder="user@example.com" required bind:value={invEmail} />
				</div>
				<div class="w-32 space-y-1.5">
					<label for="inv-role" class="text-sm font-medium">{t.role}</label>
					<select
						id="inv-role"
						bind:value={invRole}
						class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
					>
						<option value="member">Member</option>
						<option value="admin">Admin</option>
					</select>
				</div>
				<Button type="submit" disabled={sending || !invEmail}>
					{sending ? t.sending : t.send}
				</Button>
			</form>
			{#if error}
				<p class="mt-2 text-sm text-destructive">{error}</p>
			{/if}
			{#if success}
				<p class="mt-2 text-sm text-green-600">{success}</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Separator />

	<!-- Pending list -->
	{#if pendingList.length === 0}
		<p class="text-sm text-muted-foreground py-4">{t.noPending}</p>
	{:else}
		<div class="space-y-2">
			{#each pendingList as inv (inv.id)}
				<div class="flex items-center justify-between rounded-md border px-4 py-3">
					<div class="space-y-0.5">
						<p class="text-sm font-medium">{inv.email}</p>
						<p class="text-xs text-muted-foreground">
							{inv.role} · {t.expires} {formatDate(inv.expires_at)}
						</p>
					</div>
					<div class="flex gap-2">
						<Button variant="ghost" size="sm" onclick={() => handleResend(inv.id)}>{t.resend}</Button>
						<Button variant="ghost" size="sm" class="text-destructive" onclick={() => handleRevoke(inv.id)}>{t.revoke}</Button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
