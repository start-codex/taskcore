import { writable } from 'svelte/store';
import type { User } from '$lib/api';

const STORAGE_KEY = 'user';

function loadFromStorage(): User | null {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? (JSON.parse(raw) as User) : null;
	} catch {
		return null;
	}
}

const _store = writable<User | null>(loadFromStorage());

export const currentUser = { subscribe: _store.subscribe };

export function login(user: User): void {
	localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
	_store.set(user);
}

export function logout(): void {
	localStorage.removeItem(STORAGE_KEY);
	_store.set(null);
}
