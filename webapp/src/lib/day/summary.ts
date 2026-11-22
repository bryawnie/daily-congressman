import type { Day, DaySummary } from '#lib/types.ts';
import { fullName } from '#lib/format/person.ts';

export function summarizeDay(day: Day): DaySummary {
	return {
		date: day.date,
		fullName: fullName(day.congressman),
		chamber: day.congressman.chamber
	};
}
