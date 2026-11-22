import { API_URL } from '$app/env/private';
import { createApiClient } from '#lib/api/client.ts';

export const RECENT_DAYS = 8;

export function apiFor(fetchFunction: typeof fetch) {
	return createApiClient(fetchFunction, API_URL);
}
