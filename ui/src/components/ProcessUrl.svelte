<script lang="ts">
  import { api } from '../lib/api';
  import { refreshJobs } from '../lib/stores.svelte';
  import CollapsibleCard from './CollapsibleCard.svelte';

  let url = $state('');
  let loading = $state(false);
  let result = $state('');
  let error = $state('');

  async function submit() {
    if (!url) return;
    loading = true;
    result = '';
    error = '';
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
</script>

<CollapsibleCard title="Process URL">
  <label>
    Job Posting URL
    <input type="url" bind:value={url} placeholder="https://..." />
  </label>
  <button onclick={submit} disabled={loading || !url}>
    {loading ? 'Generating...' : 'Generate'}
  </button>
  {#if result}<small class="success">Created: {result}</small>{/if}
  {#if error}<small class="error">{error}</small>{/if}
</CollapsibleCard>

<style>
  .success {
    color: var(--pico-ins-color);
  }
  .error {
    color: var(--pico-del-color);
  }
</style>
