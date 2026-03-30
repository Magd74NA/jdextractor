<script lang="ts">
  import { link } from 'svelte-spa-router';
  import { api } from '../lib/api';
  import { refreshJobs } from '../lib/stores.svelte';
  import type { BatchResult } from '../lib/types';

  let mode = $state<'url' | 'batch' | 'local'>('url');

  let url = $state('');
  let urlsText = $state('');
  let content = $state('');
  let loading = $state(false);
  let result = $state('');
  let batchResults = $state<BatchResult[]>([]);
  let error = $state('');

  function reset() {
    result = '';
    batchResults = [];
    error = '';
  }

  async function submitUrl() {
    if (!url) return;
    loading = true;
    reset();
    try {
      const res = await api.process(url);
      result = res.dir;
      await refreshJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Process failed';
    } finally {
      loading = false;
    }
  }

  async function submitBatch() {
    const urls = urlsText.split('\n').map(u => u.trim()).filter(Boolean);
    if (urls.length === 0) return;
    loading = true;
    reset();
    try {
      batchResults = await api.processBatch(urls);
      await refreshJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Batch processing failed';
    } finally {
      loading = false;
    }
  }

  async function submitLocal() {
    if (!content) return;
    loading = true;
    reset();
    try {
      const res = await api.processLocal(content);
      result = res.dir;
      await refreshJobs();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Process failed';
    } finally {
      loading = false;
    }
  }
</script>

<h2>Process Job Description</h2>

<div role="group">
  <button class:outline={mode !== 'url'} onclick={() => { mode = 'url'; reset(); }}>URL</button>
  <button class:outline={mode !== 'batch'} onclick={() => { mode = 'batch'; reset(); }}>Batch</button>
  <button class:outline={mode !== 'local'} onclick={() => { mode = 'local'; reset(); }}>Paste Text</button>
</div>

{#if mode === 'url'}
  <label>
    Job Posting URL
    <input type="url" bind:value={url} placeholder="https://..." />
  </label>
  <button onclick={submitUrl} disabled={loading || !url}>
    {loading ? 'Generating...' : 'Generate'}
  </button>

{:else if mode === 'batch'}
  <label>
    One URL per line
    <textarea rows={6} bind:value={urlsText} placeholder={"https://...\nhttps://..."}></textarea>
  </label>
  <button onclick={submitBatch} disabled={loading || !urlsText.trim()}>
    {loading ? 'Processing...' : 'Generate All'}
  </button>

{:else}
  <label>
    Paste job description
    <textarea rows={8} bind:value={content} placeholder="Paste the full job description here..."></textarea>
  </label>
  <button onclick={submitLocal} disabled={loading || !content.trim()}>
    {loading ? 'Generating...' : 'Generate'}
  </button>
{/if}

{#if error}<p class="error">{error}</p>{/if}

{#if result}
  <p class="success">Created: {result} — <a href="/" use:link>View Applications</a></p>
{/if}

{#if batchResults.length > 0}
  <ul class="results">
    {#each batchResults as r}
      <li>
        {#if r.error}
          <span class="fail">✗</span> {r.url} — {r.error}
        {:else}
          <span class="ok">✓</span> {r.url}
        {/if}
      </li>
    {/each}
  </ul>
  <p><a href="/" use:link>View Applications</a></p>
{/if}

<style>
  .success {
    color: var(--pico-ins-color);
  }

  .error {
    color: var(--pico-del-color);
  }

  .results {
    list-style: none;
    padding: 0;
    margin-top: 1rem;
  }

  .ok {
    color: var(--pico-ins-color);
  }

  .fail {
    color: var(--pico-del-color);
  }
</style>
