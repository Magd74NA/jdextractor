<script lang="ts">
  import { api } from '../lib/api';
  import { getNetworkingPromptConfig, loadNetworkingPromptConfig } from '../lib/stores.svelte';

  let saving = $state(false);
  let saved = $state(false);
  let error = $state('');

  let promptConfig = $derived(getNetworkingPromptConfig());

  let systemPrompt = $state('');
  let taskList = $state('');

  let loaded = $state(false);

  async function init() {
    await loadNetworkingPromptConfig();
    loaded = true;
  }

  $effect(() => {
    if (promptConfig) {
      systemPrompt = promptConfig.system_prompt;
      taskList = promptConfig.task_list;
    }
  });

  async function save() {
    saving = true;
    error = '';
    try {
      await api.saveNetworkingPromptConfig({
        system_prompt: systemPrompt,
        task_list: taskList,
      });
      await loadNetworkingPromptConfig();
      saved = true;
      setTimeout(() => saved = false, 3000);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Save failed';
    } finally {
      saving = false;
    }
  }

  init();
</script>

<section>
  <h3>Networking Prompts</h3>
  <p class="description">Controls how AI follow-up messages and conversation summaries are generated.</p>

  {#if !loaded}
    <p aria-busy="true">Loading...</p>
  {:else}
    <label>
      <h4>System Prompt</h4>
      <textarea class="mono" rows={4} bind:value={systemPrompt}></textarea>
    </label>

    <label>
      <h4>Task List</h4>
      <textarea class="mono" rows={4} bind:value={taskList}></textarea>
    </label>

    <button onclick={save} disabled={saving}>
      {saving ? 'Saving...' : 'Save Networking Prompts'}
    </button>
    {#if saved}<small class="success">Saved!</small>{/if}
    {#if error}<small class="error">{error}</small>{/if}
  {/if}
</section>

<style>
  .description {
    color: var(--pico-muted-color);
    font-size: 0.85rem;
    margin-bottom: 1rem;
  }

  label {
    margin-bottom: 0.75rem;
    display: block;
  }

  small {
    margin-left: 0.5rem;
  }
</style>
