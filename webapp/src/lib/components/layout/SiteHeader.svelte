<header>
  <div class="content">
    <a class="brand" href="/">Congresista del día</a>
    <p class="tagline">
      Sorteamos de forma verificable a un parlamentario cada día, exponiendo
      información pública disponible en la página oficial de la cámara.
    </p>
  </div>
</header>

<style>
  header {
    border-block-end: 1px solid var(--color-border);
    background: var(--color-surface);
  }

  .content {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: var(--space-1) var(--space-4);
    max-inline-size: var(--page-width);
    margin-inline: auto;
    padding: var(--space-4) var(--page-gutter);
  }

  .brand {
    color: var(--color-text);
    font-size: 1.25rem;
    font-weight: 700;
    text-decoration: none;
  }

  .tagline {
    color: var(--color-text-muted);
    font-size: 0.9375rem;
  }
</style>
