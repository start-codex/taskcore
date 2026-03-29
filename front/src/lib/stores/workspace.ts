// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { writable } from 'svelte/store';
import type { Workspace, Project } from '$lib/api';

const _workspace    = writable<Workspace | null>(null);
const _projects     = writable<Project[]>([]);
const _initialized  = writable(false);
const _workspaces   = writable<Workspace[]>([]);

export const selectedWorkspace    = { subscribe: _workspace.subscribe };
export const workspaceProjects    = { subscribe: _projects.subscribe };
export const workspaceReady       = { subscribe: _initialized.subscribe };
export const workspaceList        = { subscribe: _workspaces.subscribe };
export const workspaceSheetOpen   = writable(false);

export function selectWorkspace(workspace: Workspace, projects: Project[]): void {
	try { localStorage.setItem('workspace_id', workspace.id); } catch {}
	_workspace.set(workspace);
	_projects.set(projects);
	_initialized.set(true);
}

export function markWorkspaceReady(): void {
	_initialized.set(true);
}

export function setWorkspaceList(list: Workspace[]): void {
	_workspaces.set(list);
}

export function addToWorkspaceList(workspace: Workspace): void {
	_workspaces.update(ws => [...ws, workspace]);
}

export function addWorkspaceProject(project: Project): void {
	_projects.update(p => [...p, project]);
}

export function getStoredWorkspaceId(): string | null {
	try { return localStorage.getItem('workspace_id'); } catch { return null; }
}
