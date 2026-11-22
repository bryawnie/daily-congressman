const longDateFormatter = new Intl.DateTimeFormat('es-CL', { dateStyle: 'full', timeZone: 'UTC' });
const shortDateFormatter = new Intl.DateTimeFormat('es-CL', {
	day: 'numeric',
	month: 'short',
	timeZone: 'UTC'
});
const mediumDateFormatter = new Intl.DateTimeFormat('es-CL', {
	dateStyle: 'medium',
	timeZone: 'UTC'
});

const MONTH_NAMES = [
	'enero',
	'febrero',
	'marzo',
	'abril',
	'mayo',
	'junio',
	'julio',
	'agosto',
	'septiembre',
	'octubre',
	'noviembre',
	'diciembre'
];

function parseIsoDate(isoDate: string): Date {
	return new Date(`${isoDate}T00:00:00Z`);
}

export function formatLongDate(isoDate: string): string {
	return capitalize(longDateFormatter.format(parseIsoDate(isoDate)));
}

export function formatShortDate(isoDate: string): string {
	return shortDateFormatter.format(parseIsoDate(isoDate));
}

export function formatMediumDate(isoDate: string): string {
	return mediumDateFormatter.format(parseIsoDate(isoDate));
}

export function ageOn(birthdate: string, isoDate: string): number | null {
	if (!birthdate) {
		return null;
	}
	const born = parseIsoDate(birthdate);
	const reference = parseIsoDate(isoDate);
	const hadBirthdayThisYear =
		reference.getUTCMonth() > born.getUTCMonth() ||
		(reference.getUTCMonth() === born.getUTCMonth() &&
			reference.getUTCDate() >= born.getUTCDate());
	return reference.getUTCFullYear() - born.getUTCFullYear() - (hadBirthdayThisYear ? 0 : 1);
}

export function monthNumber(monthName: string): number {
	return MONTH_NAMES.indexOf(monthName) + 1;
}

export function monthSortKey(year: number, monthName: string): number {
	return year * 12 + monthNumber(monthName);
}

function capitalize(text: string): string {
	return text.charAt(0).toUpperCase() + text.slice(1);
}

export function formatMonth(year: number, monthName: string): string {
	return `${capitalize(monthName)} ${year}`;
}

export function formatShortMonth(monthName: string): string {
	return capitalize(monthName.slice(0, 3));
}
