import { RECENT_DAYS, apiFor } from '#lib/server/api.ts';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = async () => {
	const days = await apiFor(fetch).recentDays(RECENT_DAYS);
	return days.map((day) => ({ date: day.date }));
};

export const load: PageServerLoad = async ({ fetch, params }) => {
	return { day: await apiFor(fetch).day(params.date) };
};
