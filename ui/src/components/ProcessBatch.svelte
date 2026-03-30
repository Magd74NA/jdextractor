<script lang="ts">
  import { api } from '../lib/api';
  import { refreshJobs } from '../lib/stores.svelte';
  import type { BatchResult } from '../lib/types';
  import CollapsibleCard from './CollapsibleCard.svelte';

  let urlsText = $state('');
  let loading = $state(false);
  let results = $state<BatchResult[]>([]);
  let error = $state('');

  async function submit() {
    const urls = urlsText.split('\n').map(u => u.trim()).filter(Boolean);
    if (urls.length === 0) return;
    loading = true;
    results = [];
    error = '';
    try {
      results = await api.processBatch(urls);
      await refreshJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Batch processing failed';
    } finally {
      loading = false;
    }
  }
</script>

<CollapsibleCard title="Process Batch (URLs)">
  <label>
    One URL per line
    <textarea rows={4} bind:value={urlsText} placeholder="https://...&#10;https://..."></textarea>
  </label>
  <button onclick={submit} disabled={loading || !urlsText.trim()}>
    {loading ? 'Processing...' : 'Generate All'}
  </button>

  {#if error}<small class="error">{error}</small>{/if}

  {#if results.length > 0}
    <ul class="results">
      {#each results as r}
        <li>
          {#if r.error}
            <span class="fail">✗</span> {r.url} — {r.error}
          {:else}
            <span class="ok">✓</span> {r.url}
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</CollapsibleCard>

<style>
  .results {
    list-style: none;
    padding: 0;
    margin-top: 1rem;
  }

  .ok {
    color: var(--pico-ins-color);
  }

  .fail, .error {
    color: var(--pico-del-color);
  }
</style>
