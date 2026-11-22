<ul aria-label="Cómo leer los gráficos de comparación">
	<li><span class="dot" aria-hidden="true"></span>Gasto de este parlamentario</li>
	<li><span class="line" aria-hidden="true"></span>Promedio del grupo</li>
</ul>

<style>
	ul {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1) var(--space-5);
		color: var(--color-text-muted);
		font-size: 0.8125rem;
	}

	li {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.dot {
		inline-size: 0.75rem;
		block-size: 0.75rem;
		border-radius: 50%;
		background: var(--color-accent);
	}

	.line {
		inline-size: 0.1875rem;
		block-size: 0.875rem;
		border-radius: 999px;
		background: var(--color-text-muted);
	}
</style>
