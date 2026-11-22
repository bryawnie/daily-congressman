import { apiFor } from '#lib/server/api.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [latestDay] = await apiFor(fetch).recentDays(1);
	return { day: latestDay ?? null };
};
