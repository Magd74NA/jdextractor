<script lang="ts">
  import { api } from "../lib/api";
  import {
    getConfig,
    getPromptConfig,
    loadConfig,
    loadPromptConfig,
  } from "../lib/stores.svelte";

  let saving = $state(false);
  let saved = $state(false);
  let error = $state("");
  let showApiKey = $state(false);

  let config = $derived(getConfig());
  let promptConfig = $derived(getPromptConfig());

  let backend = $state("deepseek");
  let deepseekModel = $state("deepseek-chat");
  let deepseekApiKey = $state("");
  let kimiApiKey = $state("");
  let kimiModel = $state("moonshotai/Kimi-K2.5");
  let port = $state(8080);
  let systemPrompt = $state("");
  let taskList = $state("");

  $effect(() => {
    if (config) {
      backend = config.backend;
      deepseekModel = config.deepseek_model;
      deepseekApiKey = config.deepseek_api_key;
      kimiApiKey = config.kimi_api_key;
      kimiModel = config.kimi_model;
      port = config.port || 8080;
    }
  });

  $effect(() => {
    if (promptConfig) {
      systemPrompt = promptConfig.system_prompt;
      taskList = promptConfig.task_list;
    }
  });

  async function save() {
    saving = true;
    error = "";
    try {
      const configUpdate: Record<string, any> = {
        backend,
        port,
      };
      if (backend === "deepseek") {
        configUpdate.deepseek_model = deepseekModel;
        configUpdate.deepseek_api_key = deepseekApiKey;
      } else {
        configUpdate.kimi_api_key = kimiApiKey;
        configUpdate.kimi_model = kimiModel;
      }
      await Promise.all([
        api.saveConfig(configUpdate),
        api.savePromptConfig({
          system_prompt: systemPrompt,
          task_list: taskList,
        }),
      ]);
      await Promise.all([loadConfig(), loadPromptConfig()]);
      saved = true;
      setTimeout(() => (saved = false), 3000);
    } catch (e) {
      error = e instanceof Error ? e.message : "Save failed";
    } finally {
      saving = false;
    }
  }
</script>

<section>
  <h3>Configuration</h3>
  <label>
    <h4>Backend</h4>
    <select bind:value={backend}>
      <option value="deepseek">DeepSeek</option>
      <option value="kimi">Kimi K2.5 (experimental)</option>
    </select>
  </label>

  {#if backend === "deepseek"}
    <label>
      <h4>Model</h4>
      <select bind:value={deepseekModel}>
        <option value="deepseek-chat">deepseek-chat</option>
        <option value="deepseek-reasoner">deepseek-reasoner</option>
      </select>
    </label>

    <label>
      <h4>API Key</h4>
      <div class="key-row">
        {#if showApiKey}
          <input type="text" bind:value={deepseekApiKey} placeholder="sk-..." />
        {:else}
          <input
            type="password"
            bind:value={deepseekApiKey}
            placeholder="sk-..."
          />
        {/if}
        <button class="outline" onclick={() => (showApiKey = !showApiKey)}>
          {showApiKey ? "Hide" : "Show"}
        </button>
      </div>
    </label>
  {:else}
    <label>
      <h4>Model</h4>
      <select bind:value={kimiModel}>
        <option value="moonshotai/Kimi-K2.5">Kimi K2.5</option>
      </select>
    </label>

    <label>
      <h4>API Key</h4>
      <div class="key-row">
        {#if showApiKey}
          <input type="text" bind:value={kimiApiKey} placeholder="Co1I3p..." />
        {:else}
          <input
            type="password"
            bind:value={kimiApiKey}
            placeholder="Co1I3p..."
          />
        {/if}
        <button class="outline" onclick={() => (showApiKey = !showApiKey)}>
          {showApiKey ? "Hide" : "Show"}
        </button>
      </div>
    </label>
  {/if}

  <label>
    <h4>Port</h4>
    <input type="number" bind:value={port} />
    <small>Changes require server restart.</small>
  </label>

  <label>
    <h4>System Prompt</h4>
    <textarea class="mono" rows={4} bind:value={systemPrompt}></textarea>
  </label>

  <label>
    <h4>Task List</h4>
    <textarea rows={3} bind:value={taskList}></textarea>
  </label>

  <button onclick={save} disabled={saving}>
    {saving ? "Saving..." : "Save Configuration"}
  </button>
  {#if saved}<small class="success">Saved!</small>{/if}
  {#if error}<small class="error">{error}</small>{/if}
</section>

<style>
  .key-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .key-row input {
    flex: 1;
    margin-bottom: 0;
  }

  .key-row button {
    width: auto;
    margin-bottom: 0;
    padding: 0.4rem 0.75rem;
  }

  label {
    margin-bottom: 0.75rem;
    display: block;
  }

  textarea {
    font-family: monospace;
    font-size: 0.85rem;
  }

  small {
    margin-left: 0.5rem;
  }
</style>
