import { writable } from 'svelte/store';
import type { Workspace, Project } from '$lib/api';

const _workspace = writable<Workspace | null>(null);
const _projects  = writable<Project[]>([]);

export const selectedWorkspace = { subscribe: _workspace.subscribe };
export const workspaceProjects = { subscribe: _projects.subscribe };

export function selectWorkspace(workspace: Workspace, projects: Project[]): void {
	try { localStorage.setItem('workspace_id', workspace.id); } catch {}
	_workspace.set(workspace);
	_projects.set(projects);
}

export function getStoredWorkspaceId(): string | null {
	try { return localStorage.getItem('workspace_id'); } catch { return null; }
}
