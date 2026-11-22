<script lang="ts">
	import { wipe } from '#lib/motion/actions.ts';

	let { ratio }: { ratio: number } = $props();
</script>

<span class="track" aria-hidden="true">
	<span class="fill" style:inline-size="{ratio * 100}%" use:wipe></span>
</span>

<style>
	.track {
		display: block;
		block-size: 0.5rem;
		border-radius: 999px;
		background: var(--color-surface-sunken);
		overflow: hidden;
	}

	.fill {
		display: block;
		block-size: 100%;
		border-radius: 999px;
		background: var(--color-accent);
	}
</style>
