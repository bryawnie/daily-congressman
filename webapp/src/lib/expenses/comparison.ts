import type { Chamber, Comparison, LabeledPeers, PeerStats, Stats, StatsGroup } from '#lib/types.ts';
import { formatPercent } from '#lib/format/number.ts';

export function peerStatsOf(stats: Stats | null): PeerStats | null {
	if (!stats || stats.count < 2 || stats.avg === null || stats.min === null || stats.max === null) {
		return null;
	}
	return { count: stats.count, ranking: stats.ranking, avg: stats.avg, min: stats.min, max: stats.max };
}

export function peerComparisonsOf(comparison: Comparison, chamber: Chamber): LabeledPeers[] {
	return [comparison.global, comparison.local].flatMap((stats) => {
		const peers = peerStatsOf(stats);
		return stats && peers ? [{ label: groupLabel(stats._type, chamber), peers }] : [];
	});
}

export function groupLabel(group: StatsGroup, chamber: Chamber): string {
	switch (group) {
		case 'district':
			return 'Su distrito';
		case 'region':
			return 'Su región';
		default:
			return chamber === 'senate' ? 'Todo el Senado' : 'Toda la Cámara';
	}
}

export function positionOnScale(value: number, min: number, max: number): number {
	if (max === min) {
		return 50;
	}
	const ratio = (value - min) / (max - min);
	return Math.min(100, Math.max(0, ratio * 100));
}

export function differenceFromAverage(amount: number, average: number): string {
	if (average === 0) {
		return amount === 0 ? 'igual al promedio' : 'sobre un promedio de cero';
	}
	const ratio = (amount - average) / average;
	if (Math.abs(ratio) < 0.005) {
		return 'igual al promedio';
	}
	const magnitude = formatPercent(Math.abs(ratio));
	return ratio > 0 ? `${magnitude} sobre el promedio` : `${magnitude} bajo el promedio`;
}
