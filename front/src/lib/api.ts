const BASE = '/api';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method,
		headers: body ? { 'Content-Type': 'application/json' } : {},
		body: body ? JSON.stringify(body) : undefined
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error ?? res.statusText);
	}

	if (res.status === 204) return undefined as T;
	return res.json();
}

const get = <T>(path: string) => request<T>('GET', path);
const post = <T>(path: string, body: unknown) => request<T>('POST', path, body);
const put = <T>(path: string, body: unknown) => request<T>('PUT', path, body);
const del = (path: string) => request<void>('DELETE', path);

// --- Users ---
export const users = {
	create: (body: { email: string; name: string }) => post<User>('/users', body),
	get: (userID: string) => get<User>(`/users/${userID}`)
};

// --- Workspaces ---
export const workspaces = {
	create: (body: { name: string; slug: string }) => post<Workspace>('/workspaces', body),
	get: (workspaceID: string) => get<Workspace>(`/workspaces/${workspaceID}`),
	archive: (workspaceID: string) => del(`/workspaces/${workspaceID}`),
	members: {
		list: (workspaceID: string) => get<WorkspaceMember[]>(`/workspaces/${workspaceID}/members`),
		add: (workspaceID: string, body: { user_id: string; role: string }) =>
			post<WorkspaceMember>(`/workspaces/${workspaceID}/members`, body),
		updateRole: (workspaceID: string, userID: string, body: { role: string }) =>
			put<WorkspaceMember>(`/workspaces/${workspaceID}/members/${userID}`, body),
		remove: (workspaceID: string, userID: string) =>
			del(`/workspaces/${workspaceID}/members/${userID}`)
	}
};

// --- Projects ---
export const projects = {
	create: (workspaceID: string, body: { name: string; key: string; description?: string }) =>
		post<Project>(`/workspaces/${workspaceID}/projects`, body),
	list: (workspaceID: string) => get<Project[]>(`/workspaces/${workspaceID}/projects`),
	get: (projectID: string) => get<Project>(`/projects/${projectID}`),
	archive: (projectID: string) => del(`/projects/${projectID}`),
	members: {
		list: (projectID: string) => get<ProjectMember[]>(`/projects/${projectID}/members`),
		add: (projectID: string, body: { user_id: string; role: string }) =>
			post<ProjectMember>(`/projects/${projectID}/members`, body),
		updateRole: (projectID: string, userID: string, body: { role: string }) =>
			put<ProjectMember>(`/projects/${projectID}/members/${userID}`, body),
		remove: (projectID: string, userID: string) =>
			del(`/projects/${projectID}/members/${userID}`)
	}
};

// --- Statuses ---
export const statuses = {
	create: (projectID: string, body: { name: string; category: string }) =>
		post<Status>(`/projects/${projectID}/statuses`, body),
	list: (projectID: string) => get<Status[]>(`/projects/${projectID}/statuses`),
	update: (projectID: string, statusID: string, body: { name: string; category: string }) =>
		put<Status>(`/projects/${projectID}/statuses/${statusID}`, body),
	archive: (projectID: string, statusID: string) =>
		del(`/projects/${projectID}/statuses/${statusID}`)
};

// --- Issue Types ---
export const issueTypes = {
	create: (projectID: string, body: { name: string; icon?: string; level: number }) =>
		post<IssueType>(`/projects/${projectID}/issue-types`, body),
	list: (projectID: string) => get<IssueType[]>(`/projects/${projectID}/issue-types`),
	archive: (projectID: string, issueTypeID: string) =>
		del(`/projects/${projectID}/issue-types/${issueTypeID}`)
};

// --- Boards ---
export const boards = {
	create: (projectID: string, body: { name: string; type: string; filter_query?: string }) =>
		post<Board>(`/projects/${projectID}/boards`, body),
	list: (projectID: string) => get<Board[]>(`/projects/${projectID}/boards`),
	get: (boardID: string) => get<Board>(`/boards/${boardID}`),
	archive: (boardID: string) => del(`/boards/${boardID}`),
	columns: {
		list: (boardID: string) => get<BoardColumn[]>(`/boards/${boardID}/columns`),
		add: (boardID: string, body: { name: string }) =>
			post<BoardColumn>(`/boards/${boardID}/columns`, body),
		archive: (columnID: string) => del(`/columns/${columnID}`),
		assignStatus: (columnID: string, body: { status_id: string }) =>
			post<void>(`/columns/${columnID}/statuses`, body),
		unassignStatus: (columnID: string, statusID: string) =>
			del(`/columns/${columnID}/statuses/${statusID}`)
	}
};

// --- Issues ---
export const issues = {
	create: (projectID: string, body: CreateIssueBody) =>
		post<Issue>(`/projects/${projectID}/issues`, body),
	list: (projectID: string, params?: { status_id?: string; assignee_id?: string }) => {
		const qs = new URLSearchParams(
			Object.entries(params ?? {}).filter(([, v]) => v) as [string, string][]
		).toString();
		return get<Issue[]>(`/projects/${projectID}/issues${qs ? `?${qs}` : ''}`);
	},
	get: (projectID: string, issueID: string) =>
		get<Issue>(`/projects/${projectID}/issues/${issueID}`),
	update: (projectID: string, issueID: string, body: UpdateIssueBody) =>
		put<Issue>(`/projects/${projectID}/issues/${issueID}`, body),
	archive: (projectID: string, issueID: string) =>
		del(`/projects/${projectID}/issues/${issueID}`),
	move: (projectID: string, issueID: string, body: { target_status_id: string; target_position: number }) =>
		post<void>(`/projects/${projectID}/issues/${issueID}/move`, body)
};

// --- Types ---
export interface User {
	id: string; email: string; name: string;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface Workspace {
	id: string; name: string; slug: string;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface WorkspaceMember {
	workspace_id: string; user_id: string; role: string;
	created_at: string; updated_at: string;
}
export interface Project {
	id: string; workspace_id: string; name: string; key: string; description: string;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface ProjectMember {
	project_id: string; user_id: string; role: string;
	created_at: string; updated_at: string;
}
export interface Status {
	id: string; project_id: string; name: string; category: string; position: number;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface IssueType {
	id: string; project_id: string; name: string; icon: string; level: number;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface Board {
	id: string; project_id: string; name: string; type: string; filter_query: string;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface BoardColumn {
	id: string; board_id: string; name: string; position: number;
	created_at: string; updated_at: string; archived_at?: string;
}
export interface Issue {
	id: string; project_id: string; number: number;
	issue_type_id: string; status_id: string; parent_issue_id?: string;
	title: string; description: string; priority: string;
	assignee_id?: string; reporter_id: string; due_date?: string;
	status_position: number; created_at: string; updated_at: string; archived_at?: string;
}
export interface CreateIssueBody {
	issue_type_id: string; status_id: string; title: string;
	reporter_id: string; description?: string; priority?: string;
	assignee_id?: string; parent_issue_id?: string; due_date?: string;
}
export interface UpdateIssueBody {
	title: string; description?: string; priority: string; assignee_id?: string;
}
