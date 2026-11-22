<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Tone } from '#lib/types.ts';

	let { tone, children }: { tone: Tone; children: Snippet } = $props();
</script>

<span class="badge" data-tone={tone}>{@render children()}</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		padding: 0.125rem 0.625rem;
		border: 1px solid color-mix(in oklch, var(--tone) 50%, transparent);
		border-radius: 999px;
		background: color-mix(in oklch, var(--tone) 12%, transparent);
		font-size: 0.8125rem;
		font-weight: 600;
		white-space: nowrap;
	}

	.badge::before {
		content: '';
		inline-size: 0.5rem;
		block-size: 0.5rem;
		border-radius: 50%;
		background: var(--tone);
	}
</style>
