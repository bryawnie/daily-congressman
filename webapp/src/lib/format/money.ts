const MILLION = 1_000_000;
const THOUSAND = 1_000;

const pesosFormatter = new Intl.NumberFormat('es-CL', {
	style: 'currency',
	currency: 'CLP',
	maximumFractionDigits: 0
});

const oneDecimalFormatter = new Intl.NumberFormat('es-CL', { maximumFractionDigits: 1 });

export function formatPesos(amount: number): string {
	return pesosFormatter.format(amount);
}

export function formatCompactPesos(amount: number): string {
	if (amount >= MILLION) {
		return `$${oneDecimalFormatter.format(amount / MILLION)} M`;
	}
	if (amount >= THOUSAND) {
		return `$${oneDecimalFormatter.format(amount / THOUSAND)} mil`;
	}
	return pesosFormatter.format(amount);
}
