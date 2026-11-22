import { animate, inView } from 'motion';

function prefersReducedMotion() {
	return matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function isBelowTheFold(node: HTMLElement) {
	return node.getBoundingClientRect().top > innerHeight * 0.9;
}

export function reveal(node: HTMLElement, options: { delay?: number } = {}) {
	if (prefersReducedMotion() || !isBelowTheFold(node)) return;

	node.style.opacity = '0';
	const stop = inView(
		node,
		() => {
			stop();
			animate(node, { opacity: [0, 1], y: [28, 0] }, { duration: 0.7, delay: options.delay ?? 0, ease: [0.22, 1, 0.36, 1] });
		},
		{ amount: 0.15 }
	);
	return { destroy: stop };
}

export function wipe(node: HTMLElement, options: { delay?: number } = {}) {
	if (prefersReducedMotion()) return;

	const hidden = 'inset(0 100% 0 0 round 999px)';
	const shown = 'inset(0 0% 0 0 round 999px)';
	node.style.clipPath = hidden;
	const stop = inView(
		node.parentElement ?? node,
		() => {
			stop();
			animate(node, { clipPath: [hidden, shown] }, { duration: 0.9, delay: options.delay ?? 0.1, ease: [0.22, 1, 0.36, 1] }).finished.then(() => {
				node.style.clipPath = '';
			});
		},
		{ amount: 0.3 }
	);
	return { destroy: stop };
}
