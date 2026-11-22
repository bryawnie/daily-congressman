<script lang="ts">
	import type { IconProps } from '#lib/types.ts';

	let { icon }: IconProps = $props();

	const [width, height, , , path] = $derived(icon.icon);
</script>

<svg
	class="icon"
	viewBox="0 0 {width} {height}"
	aria-hidden="true"
	focusable="false"
	fill="currentColor"
>
	<path d={Array.isArray(path) ? path.join('') : path} />
</svg>

<style>
	.icon {
		inline-size: 1em;
		block-size: 1em;
		flex: none;
		vertical-align: -0.125em;
	}
</style>
