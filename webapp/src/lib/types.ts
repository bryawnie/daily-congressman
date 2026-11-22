import type { Snippet } from 'svelte';

export type Chamber = 'deputies' | 'senate';
export type Gender = 'MALE' | 'FEMALE' | 'OTHER';
export type ConstituencyKind = 'district' | 'region' | '';
export type PendingSection = 'attendance' | 'votings' | 'activity' | 'expenses';
export type StatsGroup = 'all' | 'district' | 'region';

export interface Beacon {
	chainId: string;
	pulseId: string;
	randOut: string;
}

export interface Constituency {
	kind: ConstituencyKind;
	name: string;
	detail: string;
}

export interface Period {
	from: number;
	to: number;
}

export interface Profile {
	id: number;
	chamber: Chamber;
	externalId: number;
	name: string;
	firstSurname: string;
	secondSurname: string;
	picture: string;
	profileUrl: string;
	gender: Gender;
	birthdate: string;
	profession: string;
	party: string;
	partyAlias: string;
	twitterUsername: string;
	instagramUsername: string;
	constituency: Constituency;
	periods: Period[];
}

export interface Attendance {
	attended: number;
	justifiedAbsent: number;
	unjustifiedAbsent: number;
	total: number;
}

export interface Activity {
	all: number;
	byStatus: Record<string, number>;
}

export interface Voting {
	votingId: number;
	votingDate: string;
	bulletinNumber: string;
	documentTitle: string;
	articleText: string;
	voted: string;
	totalApproved: number;
	totalRejected: number;
	totalAbstention: number;
	quorum: string;
	result: string;
}

export interface Stats {
	_type: StatsGroup;
	count: number;
	ranking?: number;
	avg: number | null;
	min: number | null;
	max: number | null;
}

export interface Comparison {
	global: Stats | null;
	local: Stats | null;
}

export interface ExpenseCategory {
	subtype: string;
	amount: number;
	stats: Comparison;
}

export interface ExpenseDetail {
	type: string;
	amount: number | null;
	supportStaff?: number;
	stats: Comparison;
	expenses?: ExpenseCategory[];
}

export interface MonthExpenses {
	year: number;
	month: string;
	total: number;
	detail: ExpenseDetail[];
}

export interface Day {
	date: string;
	beacon: Beacon;
	congressman: Profile;
	attendance: Attendance | null;
	activity: Activity | null;
	votings: Voting[];
	expenses: MonthExpenses[];
	pending: PendingSection[];
}

export interface DaySummary {
	date: string;
	fullName: string;
	chamber: Chamber;
}

export type Tone = 'positive' | 'warning' | 'negative' | 'neutral' | 'accent';

export interface ToneSegment {
	label: string;
	value: number;
	tone: Tone;
}

export interface BarItem {
	label: string;
	value: number;
}

export interface PeerStats {
	count: number;
	ranking?: number;
	avg: number;
	min: number;
	max: number;
}

export interface LabeledPeers {
	label: string;
	peers: PeerStats;
}

export type FetchFunction = typeof fetch;

export interface DayPageProps {
	day: Day;
	daySummaries: DaySummary[];
}

export interface DayPickerProps {
	days: DaySummary[];
	currentDate: string;
}

export interface CategoryBreakdownProps {
	categories: ExpenseCategory[];
	chamber: Chamber;
}

export interface CategoryRowProps {
	category: ExpenseCategory;
	largestAmount: number;
	chamber: Chamber;
}

export interface ComparisonListProps {
	amount: number;
	comparison: Comparison;
	chamber: Chamber;
}

export interface ComparisonRangeProps {
	label: string;
	amount: number;
	peers: PeerStats;
}

export interface ExpenseExplorerProps {
	months: MonthExpenses[];
	chamber: Chamber;
	source?: SourceLink | null;
}

export interface ExpenseTypePanelProps {
	detail: ExpenseDetail;
	chamber: Chamber;
}

export interface MonthSummaryProps {
	month: MonthExpenses;
}

export interface MonthBreakdownProps {
	month: MonthExpenses;
	chamber: Chamber;
}

export interface MonthTotalsChartProps {
	months: MonthExpenses[];
	selectedPosition: number;
	onSelect: (position: number) => void;
}

export interface SourceLink {
	label: string;
	href: string;
}

export interface SectionCardProps {
	title: string;
	source?: SourceLink | null;
	description?: string;
	pending?: boolean;
	children: Snippet;
}
