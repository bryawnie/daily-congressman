import type { Profile, SourceLink } from '#lib/types.ts';

const SENATE_LABEL = 'Ver en sitio oficial del Senado';
const SENATE_EXPENSES_URL = 'https://www.senado.cl/transparencia/gastos-operacionales-senadores';
const SENATE_VOTINGS_URL = 'https://www.senado.cl/actividad-legislativa/sala/votaciones';

export interface SourceLinks {
	attendance: SourceLink | null;
	activity: SourceLink | null;
	votings: SourceLink | null;
	expenses: SourceLink | null;
}

// Where each section can be checked on the chamber's own website. Only the
// Senate is linked for now; a missing address means no link rather than a wrong one.
export function sourceLinksOf(profile: Profile): SourceLinks {
	if (profile.chamber !== 'senate') {
		return { attendance: null, activity: null, votings: null, expenses: null };
	}
	const profileLink = profile.profileUrl ? { label: SENATE_LABEL, href: profile.profileUrl } : null;
	return {
		attendance: profileLink,
		activity: profileLink,
		votings: { label: SENATE_LABEL, href: SENATE_VOTINGS_URL },
		expenses: { label: SENATE_LABEL, href: SENATE_EXPENSES_URL }
	};
}
