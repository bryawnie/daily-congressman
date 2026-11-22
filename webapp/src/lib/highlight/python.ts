export type PythonTokenKind = 'comment' | 'string' | 'keyword' | 'builtin' | 'number' | 'plain';

export interface PythonToken {
	kind: PythonTokenKind;
	text: string;
}

const KEYWORDS = new Set([
	'import', 'from', 'as', 'def', 'return', 'if', 'else', 'elif', 'for', 'in', 'while', 'with',
	'try', 'except', 'class', 'and', 'or', 'not', 'is', 'None', 'True', 'False', 'lambda', 'pass'
]);
const BUILTINS = new Set(['print', 'sorted', 'len', 'int', 'range', 'open', 'str', 'list', 'dict']);

const TOKEN =
	/(?<comment>#[^\n]*)|(?<string>[frbFRB]{0,2}(?:"""[\s\S]*?"""|'''[\s\S]*?'''|"(?:\\.|[^"\\\n])*"|'(?:\\.|[^'\\\n])*'))|(?<word>[A-Za-z_][A-Za-z0-9_]*)|(?<number>\d+(?:\.\d+)?)/gy;

export function tokenizePython(source: string): PythonToken[] {
	const tokens: PythonToken[] = [];
	let plain = '';
	const flush = () => {
		if (plain) tokens.push({ kind: 'plain', text: plain });
		plain = '';
	};

	let position = 0;
	while (position < source.length) {
		TOKEN.lastIndex = position;
		const match = TOKEN.exec(source);
		if (!match) {
			plain += source[position++];
			continue;
		}
		const { comment, string, word, number } = match.groups ?? {};
		let kind: PythonTokenKind = 'plain';
		if (comment) kind = 'comment';
		else if (string) kind = 'string';
		else if (number) kind = 'number';
		else if (word) kind = KEYWORDS.has(word) ? 'keyword' : BUILTINS.has(word) ? 'builtin' : 'plain';

		if (kind === 'plain') {
			plain += match[0];
		} else {
			flush();
			tokens.push({ kind, text: match[0] });
		}
		position += match[0].length;
	}
	flush();
	return tokens;
}
