const percentFormatter = new Intl.NumberFormat('es-CL', {
	style: 'percent',
	maximumFractionDigits: 0
});

export function formatPercent(ratio: number): string {
	return percentFormatter.format(ratio);
}

export function ratioOf(part: number, total: number): number {
	return total === 0 ? 0 : part / total;
}
