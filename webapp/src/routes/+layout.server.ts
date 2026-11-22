import { summarizeDay } from '#lib/day/summary.ts';
import { RECENT_DAYS, apiFor } from '#lib/server/api.ts';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ fetch }) => {
	const days = await apiFor(fetch).recentDays(RECENT_DAYS);
	return { daySummaries: days.map(summarizeDay) };
};
