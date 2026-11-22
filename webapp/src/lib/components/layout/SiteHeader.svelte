<script lang="ts">
  import "@fontsource/kaushan-script/latin-400.css";
  import Logo from "./Logo.svelte";
</script>

<header>
  <div class="content">
    <a class="brand" href="/">
      <Logo />
      <span>Parlamentario del día</span>
    </a>
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
    align-items: center;
    gap: var(--space-2) var(--space-5);
    max-inline-size: var(--page-width);
    margin-inline: auto;
    padding: var(--space-4) var(--page-gutter);
  }

  .brand {
    display: inline-flex;
    align-items: center;
    gap: var(--space-3);
    color: var(--color-text);
    font-family: "Kaushan Script", "Segoe Script", "Bradley Hand", cursive;
    font-size: 1.875rem;
    font-weight: 400;
    letter-spacing: 0.01em;
    text-decoration: none;
  }

  .tagline {
    color: var(--color-text-muted);
    font-size: 0.9375rem;
  }
</style>
