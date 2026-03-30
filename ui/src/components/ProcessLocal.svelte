<script lang="ts">
  import { api } from '../lib/api';
  import { refreshJobs } from '../lib/stores.svelte';
  import CollapsibleCard from './CollapsibleCard.svelte';

  let content = $state('');
  let loading = $state(false);
  let result = $state('');
  let error = $state('');

  async function submit() {
    if (!content) return;
    loading = true;
    result = '';
    error = '';
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

<CollapsibleCard title="Process Local Text">
  <label>
    Paste job description
    <textarea rows={6} bind:value={content} placeholder="Paste the full job description here..."></textarea>
  </label>
  <button onclick={submit} disabled={loading || !content.trim()}>
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
