import type { Chamber, ConstituencyKind, Profile } from '#lib/types.ts';

export function fullName(profile: Profile): string {
	return [profile.name, profile.firstSurname, profile.secondSurname].filter(Boolean).join(' ');
}

export function initialsOf(profile: Profile): string {
	return [profile.name, profile.firstSurname]
		.filter(Boolean)
		.map((word) => word.charAt(0).toUpperCase())
		.join('');
}

export function roleTitle(profile: Profile): string {
	const feminine = profile.gender === 'FEMALE';
	if (profile.chamber === 'senate') {
		return feminine ? 'Senadora' : 'Senador';
	}
	return feminine ? 'Diputada' : 'Diputado';
}

export function chamberName(chamber: Chamber): string {
	return chamber === 'senate' ? 'Senado' : 'Cámara de Diputadas y Diputados';
}

export function constituencyTitle(kind: ConstituencyKind): string {
	return kind === 'district' ? 'Distrito' : 'Región';
}
