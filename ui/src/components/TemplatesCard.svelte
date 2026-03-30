<script lang="ts">
  import { api } from '../lib/api';
  import CollapsibleCard from './CollapsibleCard.svelte';

  let resume = $state('');
  let cover = $state('');
  let loaded = $state(false);
  let loading = $state(false);
  let resumeSaved = $state(false);
  let coverSaved = $state(false);

  async function load() {
    if (loaded) return;
    loading = true;
    try {
      const tmpl = await api.getTemplates();
      resume = tmpl.resume;
      cover = tmpl.cover;
      loaded = true;
    } finally {
      loading = false;
    }
  }

  async function saveResume() {
    await api.saveTemplates({ resume });
    resumeSaved = true;
    setTimeout(() => resumeSaved = false, 3000);
  }

  async function saveCover() {
    await api.saveTemplates({ cover });
    coverSaved = true;
    setTimeout(() => coverSaved = false, 3000);
  }

</script>

<CollapsibleCard title="Templates" onopen={load}>
    {#if loading}
      <p aria-busy="true">Loading templates...</p>
    {:else if loaded}
      <label>
        Resume Template
        <textarea rows={8} bind:value={resume}></textarea>
      </label>
      <button onclick={saveResume}>Save Resume</button>
      {#if resumeSaved}<small class="success">Saved!</small>{/if}

      <label>
        Cover Letter Template
        <textarea rows={8} bind:value={cover}></textarea>
      </label>
      <button onclick={saveCover}>Save Cover</button>
      {#if coverSaved}<small class="success">Saved!</small>{/if}
    {/if}
  </CollapsibleCard>

<style>
  .success {
    color: var(--pico-ins-color);
    margin-left: 0.5rem;
  }
</style>
