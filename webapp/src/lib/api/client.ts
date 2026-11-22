import type { Day, FetchFunction } from '#lib/types.ts';

async function getJson<Body>(fetchFunction: FetchFunction, url: string): Promise<Body> {
	const response = await fetchFunction(url);
	if (!response.ok) {
		throw new Error(`The API answered ${response.status} for ${url}`);
	}
	return response.json();
}

export function createApiClient(fetchFunction: FetchFunction, baseUrl: string) {
	return {
		async recentDays(limit: number): Promise<Day[]> {
			const body = await getJson<{ days: Day[] }>(fetchFunction, `${baseUrl}/days?limit=${limit}`);
			return body.days;
		},
		day(date: string): Promise<Day> {
			return getJson<Day>(fetchFunction, `${baseUrl}/days/${date}`);
		}
	};
}
