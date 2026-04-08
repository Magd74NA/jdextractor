<script lang="ts">
  import { api } from "../lib/api";
  import type { Contact } from "../lib/types";

  let { contact, onSent }: { contact: Contact; onSent?: () => void } = $props();

  let generating = $state(false);
  let generatedMessage = $state("");
  let generatedSubject = $state("");
  let generatedChannel = $state("");
  let generatedTiming = $state("");
  let generatedNextDate = $state("");
  let generateError = $state("");
  let sending = $state(false);
  let sendError = $state("");
  let sendSuccess = $state(false);
  let lastSentNextDate = $state("");
  let guidance = $state("");

  async function generate() {
    generating = true;
    generatedMessage = "";
    generatedSubject = "";
    generatedChannel = "";
    generatedTiming = "";
    generatedNextDate = "";
    generateError = "";
    sendError = "";
    sendSuccess = false;
    try {
      const trimmedGuidance = guidance.trim();
      const result = await api.generateFollowupStream(
        contact.dir,
        (event) => {
          if (event.stage === "content" && event.delta) {
            generatedMessage += event.delta;
          }
        },
        trimmedGuidance ? { guidance: trimmedGuidance } : {},
      );
      generatedMessage = result.message;
      generatedSubject = result.subject ?? "";
      generatedChannel = result.channel ?? "";
      generatedTiming = result.timing ?? "";
      generatedNextDate = result.suggested_next_date ?? "";
    } catch (e) {
      generateError = e instanceof Error ? e.message : "Generation failed";
    } finally {
      generating = false;
    }
  }

  async function markAsSent() {
    if (!generatedMessage.trim() || !generatedChannel) return;
    sending = true;
    sendError = "";
    try {
      const body: { content: string; channel: string; next_followup_date?: string } = {
        content: generatedMessage,
        channel: generatedChannel,
      };
      if (generatedNextDate) body.next_followup_date = generatedNextDate;
      await api.sendFollowup(contact.dir, body);
      lastSentNextDate = generatedNextDate;
      sendSuccess = true;
      generatedMessage = "";
      generatedSubject = "";
      generatedChannel = "";
      generatedTiming = "";
      generatedNextDate = "";
      guidance = "";
      onSent?.();
    } catch (e) {
      sendError = e instanceof Error ? e.message : "Failed to mark as sent";
    } finally {
      sending = false;
    }
  }

  function copy() {
    navigator.clipboard.writeText(generatedMessage);
  }
</script>

<div class="followup-panel">
  <textarea
    class="guidance-input"
    rows={2}
    bind:value={guidance}
    placeholder="Optional: extra guidance for the LLM (e.g. 'be more casual', 'mention the React opening')"
    disabled={generating}
  ></textarea>
  <div class="panel-header">
    <button
      class="outline btn-sm"
      onclick={generate}
      aria-busy={generating}
      disabled={generating}
    >
      Generate
    </button>
  </div>

  {#if generateError}
    <p class="error">{generateError}</p>
  {/if}

  {#if sendSuccess}
    <p class="send-success">
      Logged. {lastSentNextDate ? `Next follow-up: ${lastSentNextDate}` : "Follow-up date cleared."}
    </p>
  {/if}

  {#if generatedMessage || generating}
    <div class="generated">
      {#if generatedSubject}
        <p class="gen-meta"><strong>Subject:</strong> {generatedSubject}</p>
      {/if}
      {#if generatedChannel || generatedTiming}
        <p class="muted">Channel: {generatedChannel} · Timing: {generatedTiming}</p>
      {/if}
      <!-- svelte-ignore a11y_autofocus -->
      <textarea
        class="mono"
        rows={6}
        bind:value={generatedMessage}
        disabled={generating}
        autofocus={!generating}
      ></textarea>
      <div class="send-row">
        <label class="send-date-label">
          Next follow-up
          <input type="date" bind:value={generatedNextDate} class="edit-input" />
        </label>
        <div class="send-actions">
          <button class="outline btn-sm" onclick={copy} disabled={generating}>Copy</button>
          <button
            class="btn-sm"
            onclick={markAsSent}
            aria-busy={sending}
            disabled={generating || sending || !generatedMessage.trim()}
          >Mark as Sent</button>
        </div>
      </div>
      {#if sendError}
        <p class="error">{sendError}</p>
      {/if}
    </div>
  {/if}
</div>

<style>
  .followup-panel {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .panel-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .guidance-input {
    font-size: 0.82rem;
    margin: 0;
    resize: vertical;
  }

  .generated {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .gen-meta {
    font-size: 0.82rem;
    margin: 0;
  }

  .send-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .send-date-label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.82rem;
    flex: 1;
  }

  .send-date-label input {
    margin: 0;
    font-size: 0.82rem;
  }

  .send-actions {
    display: flex;
    gap: 0.4rem;
    flex-shrink: 0;
  }

  .send-success {
    font-size: 0.82rem;
    color: var(--pico-ins-color);
    margin: 0;
  }
</style>
