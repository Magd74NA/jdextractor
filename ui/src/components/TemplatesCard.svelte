<script lang="ts">
  import { api } from '../lib/api';

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

  load();
</script>

<section>
  <h3>Templates</h3>

  {#if loading}
    <p aria-busy="true">Loading templates...</p>
  {:else if loaded}
    <div class="file-section">
      <div class="file-header">
        <h4>Resume Template</h4>
        <div class="file-actions">
          {#if resumeSaved}<small class="success">Saved!</small>{/if}
          <button class="outline" onclick={saveResume}>Save</button>
        </div>
      </div>
      <textarea rows={8} bind:value={resume}></textarea>
    </div>

    <div class="file-section">
      <div class="file-header">
        <h4>Cover Letter Template</h4>
        <div class="file-actions">
          {#if coverSaved}<small class="success">Saved!</small>{/if}
          <button class="outline" onclick={saveCover}>Save</button>
        </div>
      </div>
      <textarea rows={8} bind:value={cover}></textarea>
    </div>
  {/if}
</section>

<style>
  .file-section {
    margin-bottom: 1.5rem;
  }

  .file-section:last-child {
    margin-bottom: 0;
  }

  .file-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .file-header h4 {
    margin: 0;
    font-size: 1rem;
  }

  .file-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .file-actions button {
    padding: 0.25rem 0.5rem;
    margin-bottom: 0;
  }

  textarea {
    font-family: monospace;
    font-size: 0.85rem;
    margin-bottom: 0;
  }

  .success {
    color: var(--pico-ins-color);
  }
</style>
