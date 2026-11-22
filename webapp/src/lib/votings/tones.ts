import type { Tone } from '#lib/types.ts';

export function voteTone(vote: string): Tone {
	switch (vote) {
		case 'Afirmativo':
			return 'positive';
		case 'En Contra':
			return 'negative';
		case 'Abstención':
			return 'warning';
		default:
			return 'neutral';
	}
}

export function resultTone(result: string): Tone {
	switch (result) {
		case 'Aprobado':
			return 'positive';
		case 'Rechazado':
			return 'negative';
		default:
			return 'neutral';
	}
}
