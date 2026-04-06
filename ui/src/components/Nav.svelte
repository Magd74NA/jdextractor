<script lang="ts">
  import { link } from 'svelte-spa-router';
  import active from 'svelte-spa-router/active';
</script>

<header>
  <nav>
    <ul>
      <li>
        <a href="/" use:link class="brand">
          <strong>JD</strong><span class="brand-dim">extractor</span>
        </a>
      </li>
    </ul>
    <ul class="nav-links">
      <li><a href="/" use:link use:active={{path: '/', className: 'active'}}>Dashboard</a></li>
      <li><a href="/jobs" use:link use:active={'/jobs'}>Jobs</a></li>
      <li><a href="/process" use:link use:active={'/process'}>Process</a></li>
      <li><a href="/contacts" use:link use:active={'/contacts'}>Contacts</a></li>
      <li><a href="/settings" use:link use:active={'/settings'}>Settings</a></li>
    </ul>
  </nav>
</header>

<style>
  header {
    border-bottom: 2px solid var(--pico-muted-border-color);
    margin-bottom: 2rem;
    padding-bottom: 0.25rem;
  }

  .brand {
    text-decoration: none;
    font-size: 1.25rem;
    letter-spacing: -0.02em;
  }

  .brand strong {
    color: var(--pico-primary);
  }

  .brand-dim {
    color: var(--pico-muted-color);
  }

  .brand:hover .brand-dim {
    color: var(--pico-color);
  }

  .nav-links li {
    margin-left: 0.75rem;
  }

  .nav-links a {
    font-weight: 500;
    font-size: 0.95rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 0.25rem 0;
    color: var(--pico-muted-color);
  }

  .nav-links a:hover {
    color: var(--pico-color);
  }

  .nav-links :global(a.active) {
    color: var(--pico-primary);
    font-weight: 700;
    border-bottom: 2px solid var(--pico-primary);
  }
</style>
