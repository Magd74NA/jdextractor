<script lang="ts">
  import { link, push } from 'svelte-spa-router';
  import active from 'svelte-spa-router/active';
  import { api } from '../lib/api';
  import type { SearchResult } from '../lib/types';

  let query = $state('');
  let results = $state<SearchResult | null>(null);
  let open = $state(false);
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  let containerEl: HTMLDivElement;

  function onInput() {
    if (debounceTimer) clearTimeout(debounceTimer);
    if (!query.trim()) {
      results = null;
      open = false;
      return;
    }
    debounceTimer = setTimeout(async () => {
      results = await api.searchAll(query);
      open = true;
    }, 300);
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      open = false;
      query = '';
    }
  }

  function navigate(path: string) {
    open = false;
    query = '';
    push(path);
  }

  function handleClickOutside(e: MouseEvent) {
    if (containerEl && !containerEl.contains(e.target as Node)) {
      open = false;
    }
  }

  $effect(() => {
    window.addEventListener('click', handleClickOutside);
    return () => window.removeEventListener('click', handleClickOutside);
  });

  const jobCount = $derived(results?.jobs?.length ?? 0);
  const contactCount = $derived(results?.contacts?.length ?? 0);
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
    <div class="search-container" bind:this={containerEl}>
      <input
        class="search-input"
        type="search"
        placeholder="Search…"
        bind:value={query}
        oninput={onInput}
        onkeydown={onKeydown}
      />
      {#if open && results}
        <div class="search-dropdown">
          {#if jobCount === 0 && contactCount === 0}
            <p class="search-empty">No results</p>
          {/if}
          {#if results.jobs && results.jobs.length > 0}
            <p class="search-group-label">Jobs ({jobCount})</p>
            {#each results.jobs.slice(0, 5) as job}
              <button class="search-result" onclick={() => navigate(`/jobs?q=${encodeURIComponent(query)}`)}>
                <span class="result-primary">{job.company}</span>
                <span class="result-secondary">{job.role} · score {job.score} · {job.status}</span>
              </button>
            {/each}
          {/if}
          {#if results.contacts && results.contacts.length > 0}
            <p class="search-group-label">Contacts ({contactCount})</p>
            {#each results.contacts.slice(0, 5) as contact}
              <button class="search-result" onclick={() => navigate(`/contacts?q=${encodeURIComponent(query)}`)}>
                <span class="result-primary">{contact.name}</span>
                <span class="result-secondary">{[contact.company, contact.role].filter(Boolean).join(' · ')} · {contact.status}</span>
              </button>
            {/each}
          {/if}
          {#if jobCount > 0 || contactCount > 0}
            <div class="search-footer">
              <button class="search-see-all" onclick={() => navigate(`/jobs?q=${encodeURIComponent(query)}`)}>
                See all results →
              </button>
            </div>
          {/if}
        </div>
      {/if}
    </div>
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

  .search-container {
    position: relative;
    margin-left: 1.5rem;
  }

  .search-input {
    width: 14rem;
    padding: 0.25rem 0.6rem;
    font-size: 0.85rem;
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
    background: var(--pico-background-color);
    color: var(--pico-color);
    margin-bottom: 0;
  }

  .search-input:focus {
    outline: none;
    border-color: var(--pico-primary);
  }

  .search-dropdown {
    position: absolute;
    top: calc(100% + 0.4rem);
    right: 0;
    width: 22rem;
    background: var(--pico-card-background-color, var(--pico-background-color));
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 6px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    z-index: 100;
    padding: 0.5rem 0;
  }

  .search-group-label {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--pico-muted-color);
    padding: 0.4rem 0.75rem 0.2rem;
    margin: 0;
  }

  .search-result {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    padding: 0.35rem 0.75rem;
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
    border-radius: 0;
    color: var(--pico-color);
  }

  .search-result:hover {
    background: var(--pico-muted-border-color);
  }

  .result-primary {
    font-size: 0.875rem;
    font-weight: 600;
  }

  .result-secondary {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
  }

  .search-empty {
    font-size: 0.85rem;
    color: var(--pico-muted-color);
    padding: 0.5rem 0.75rem;
    margin: 0;
  }

  .search-footer {
    border-top: 1px solid var(--pico-muted-border-color);
    margin-top: 0.25rem;
    padding-top: 0.25rem;
  }

  .search-see-all {
    width: 100%;
    padding: 0.35rem 0.75rem;
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
    font-size: 0.82rem;
    color: var(--pico-primary);
    font-weight: 600;
    border-radius: 0;
  }

  .search-see-all:hover {
    background: var(--pico-muted-border-color);
  }
</style>
