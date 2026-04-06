<script lang="ts">
  import { api } from "../lib/api";
  import type { Contact, FollowupResult } from "../lib/types";
  import { CHANNELS } from "../lib/types";

  let overdue = $state<Contact[]>([]);
  let upcoming = $state<Contact[]>([]);
  let loading = $state(true);
  let error = $state("");

  let generating = $state<Record<string, boolean>>({});
  let streamingText = $state<Record<string, string>>({});
  let generatedResults = $state<Record<string, FollowupResult>>({});
  let generateErrors = $state<Record<string, string>>({});

  // Mark Sent modal state
  let markSentDir = $state<string | null>(null);
  let markSentContact = $state<Contact | null>(null);
  let markSentContent = $state("");
  let markSentChannel = $state("");
  let markSentNextDate = $state("");
  let markSentSaving = $state(false);
  let markSentError = $state("");

  function openMarkSent(contact: Contact) {
    markSentDir = contact.dir;
    markSentContact = contact;
    markSentContent = "";
    markSentChannel = contact.conversations.at(-1)?.channel ?? "";
    markSentNextDate = "";
    markSentError = "";
  }

  function closeMarkSent() {
    markSentDir = null;
    markSentContact = null;
  }

  async function submitMarkSent() {
    if (!markSentDir || !markSentContent.trim() || !markSentChannel) return;
    markSentSaving = true;
    markSentError = "";
    try {
      const body: { content: string; channel: string; next_followup_date?: string } = {
        content: markSentContent,
        channel: markSentChannel,
      };
      if (markSentNextDate) body.next_followup_date = markSentNextDate;
      await api.sendFollowup(markSentDir, body);
      closeMarkSent();
      await loadQueue();
    } catch (e) {
      markSentError = e instanceof Error ? e.message : "Failed to log";
    } finally {
      markSentSaving = false;
    }
  }

  async function loadQueue() {
    try {
      const [o, u] = await Promise.all([
        api.getOverdueFollowups(),
        api.getUpcomingFollowups(7),
      ]);
      overdue = o.sort((a, b) =>
        (a.follow_up_date ?? "").localeCompare(b.follow_up_date ?? ""),
      );
      upcoming = u.sort((a, b) =>
        (a.follow_up_date ?? "").localeCompare(b.follow_up_date ?? ""),
      );
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load follow-ups";
    } finally {
      loading = false;
    }
  }

  function daysFromToday(dateStr: string): number {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const target = new Date(dateStr + "T00:00:00");
    return Math.round((target.getTime() - today.getTime()) / 86400000);
  }

  function lastConvSnippet(contact: Contact): string {
    if (!contact.conversations || contact.conversations.length === 0) return "";
    const conv = contact.conversations.at(-1);
    if (!conv) return "";
    return conv.summary.length > 80 ?
        conv.summary.slice(0, 77) + "..."
      : conv.summary;
  }

  function lastChannel(contact: Contact): string {
    if (!contact.conversations || contact.conversations.length === 0) return "";
    return contact.conversations.at(-1)?.channel ?? "";
  }

  async function handleGenerate(dir: string) {
    generating = { ...generating, [dir]: true };
    streamingText = { ...streamingText, [dir]: "" };
    generateErrors = { ...generateErrors, [dir]: "" };
    try {
      const result = await api.generateFollowupStream(dir, (event) => {
        if (event.stage === "content" && event.delta) {
          streamingText = {
            ...streamingText,
            [dir]: (streamingText[dir] ?? "") + event.delta,
          };
        }
      });
      generatedResults = { ...generatedResults, [dir]: result };
    } catch (e) {
      generateErrors = {
        ...generateErrors,
        [dir]: e instanceof Error ? e.message : "Generation failed",
      };
    } finally {
      generating = { ...generating, [dir]: false };
    }
  }

  function copyToClipboard(dir: string) {
    const result = generatedResults[dir];
    if (result) navigator.clipboard.writeText(result.message);
  }

  loadQueue();
</script>

{#if loading}
  <p aria-busy="true">Loading follow-ups...</p>
{:else if error}
  <p class="error">{error}</p>
{:else if overdue.length === 0 && upcoming.length === 0}
  <p class="muted">No follow-ups scheduled.</p>
{:else}
  {#if overdue.length > 0}
    <div class="queue-group">
      <h4 class="queue-title overdue-title">Overdue ({overdue.length})</h4>
      {#each overdue as contact (contact.dir)}
        {@const days = daysFromToday(contact.follow_up_date ?? "")}
        <div class="queue-item overdue-item">
          <div class="queue-row">
            <div class="queue-left">
              <a href="#/contacts" class="queue-name">{contact.name}</a>
              {#if contact.company}<span class="queue-company"
                  >{contact.company}</span
                >{/if}
              {#if lastChannel(contact)}<span class="channel-badge"
                  >{lastChannel(contact)}</span
                >{/if}
            </div>
            <div class="queue-center">
              <span class="queue-snippet">{lastConvSnippet(contact)}</span>
            </div>
            <div class="queue-right">
              <span class="queue-date">{contact.follow_up_date}</span>
              <span class="days-badge overdue-badge"
                >{Math.abs(days)}d overdue</span
              >
              <button
                class="outline btn-sm"
                onclick={() => handleGenerate(contact.dir)}
                aria-busy={generating[contact.dir]}
                disabled={generating[contact.dir]}
              >
                Generate
              </button>
              <button
                class="outline btn-sm"
                onclick={() => openMarkSent(contact)}
              >
                Mark Sent
              </button>
            </div>
          </div>

          {#if streamingText[contact.dir] && !generatedResults[contact.dir]}
            <pre class="message-box">{streamingText[contact.dir]}</pre>
          {/if}

          {#if generateErrors[contact.dir]}
            <p class="error">{generateErrors[contact.dir]}</p>
          {/if}

          {#if generatedResults[contact.dir]}
            {@const result = generatedResults[contact.dir]!}
            <div class="generated">
              {#if result.subject}
                <p class="gen-meta">
                  <strong>Subject:</strong>
                  {result.subject}
                </p>
              {/if}
              <pre class="message-box">{result.message}</pre>
              <div class="gen-footer">
                <span class="muted"
                  >Channel: {result.channel} · Timing: {result.timing}</span
                >
                <button
                  class="outline btn-sm"
                  onclick={() => copyToClipboard(contact.dir)}>Copy</button
                >
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  {#if upcoming.length > 0}
    <div class="queue-group">
      <h4 class="queue-title">Upcoming ({upcoming.length})</h4>
      {#each upcoming as contact (contact.dir)}
        {@const days = daysFromToday(contact.follow_up_date ?? "")}
        <div class="queue-item">
          <div class="queue-row">
            <div class="queue-left">
              <a href="#/contacts" class="queue-name">{contact.name}</a>
              {#if contact.company}<span class="queue-company"
                  >{contact.company}</span
                >{/if}
              {#if lastChannel(contact)}<span class="channel-badge"
                  >{lastChannel(contact)}</span
                >{/if}
            </div>
            <div class="queue-center">
              <span class="queue-snippet">{lastConvSnippet(contact)}</span>
            </div>
            <div class="queue-right">
              <span class="queue-date">{contact.follow_up_date}</span>
              <span class="days-badge">{days}d</span>
              <button
                class="outline btn-sm"
                onclick={() => handleGenerate(contact.dir)}
                aria-busy={generating[contact.dir]}
                disabled={generating[contact.dir]}
              >
                Generate
              </button>
              <button
                class="outline btn-sm"
                onclick={() => openMarkSent(contact)}
              >
                Mark Sent
              </button>
            </div>
          </div>

          {#if streamingText[contact.dir] && !generatedResults[contact.dir]}
            <pre class="message-box">{streamingText[contact.dir]}</pre>
          {/if}

          {#if generateErrors[contact.dir]}
            <p class="error">{generateErrors[contact.dir]}</p>
          {/if}

          {#if generatedResults[contact.dir]}
            {@const upResult = generatedResults[contact.dir]!}
            <div class="generated">
              {#if upResult.subject}
                <p class="gen-meta">
                  <strong>Subject:</strong>
                  {upResult.subject}
                </p>
              {/if}
              <pre class="message-box">{upResult.message}</pre>
              <div class="gen-footer">
                <span class="muted"
                  >Channel: {upResult.channel} · Timing: {upResult.timing}</span
                >
                <button
                  class="outline btn-sm"
                  onclick={() => copyToClipboard(contact.dir)}>Copy</button
                >
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
{/if}

<!-- Mark Sent modal -->
{#if markSentDir}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-backdrop" onclick={closeMarkSent}>
    <div class="modal" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
      <h4 class="modal-title">Log Sent Message — {markSentContact?.name}</h4>
      <div class="modal-body">
        <label>
          Channel
          <select bind:value={markSentChannel}>
            <option value="">Select channel</option>
            {#each CHANNELS as ch}
              <option value={ch}>{ch}</option>
            {/each}
          </select>
        </label>
        <label>
          Message
          <textarea rows={5} bind:value={markSentContent} placeholder="What did you send?"></textarea>
        </label>
        <label>
          Next follow-up date (optional)
          <input type="date" bind:value={markSentNextDate} />
        </label>
        {#if markSentError}
          <p class="error">{markSentError}</p>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="outline btn-sm" onclick={closeMarkSent}>Cancel</button>
        <button
          class="btn-sm"
          onclick={submitMarkSent}
          aria-busy={markSentSaving}
          disabled={markSentSaving || !markSentContent.trim() || !markSentChannel}
        >Log</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Component-specific styles only - shared styles moved to app.css */
  .queue-group {
    margin-bottom: 1.5rem;
  }

  .queue-title {
    font-size: 0.82rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--pico-muted-color);
    margin-bottom: 0.5rem;
  }

  .overdue-title {
    color: var(--pico-del-color);
  }

  .queue-item {
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
    padding: 0.6rem 0.75rem;
    margin-bottom: 0.4rem;
  }

  .overdue-item {
    border-left: 3px solid var(--pico-del-color);
  }

  .queue-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .queue-left {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
    min-width: 180px;
  }

  .queue-name {
    font-weight: 600;
    font-size: 0.85rem;
    text-decoration: none;
  }

  .queue-company {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
  }

  .queue-center {
    flex: 1;
    min-width: 0;
  }

  .queue-snippet {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }

  .queue-right {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
  }

  .queue-date {
    font-size: 0.72rem;
    font-family: monospace;
    color: var(--pico-muted-color);
  }

  .days-badge {
    font-size: 0.68rem;
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
    background: var(--pico-secondary-background);
    color: var(--pico-muted-color);
    white-space: nowrap;
  }

  .overdue-badge {
    background: var(--badge-low-bg);
    color: var(--badge-low-fg);
  }

  .generated {
    margin-top: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .gen-meta {
    font-size: 0.82rem;
    margin: 0;
  }

  .gen-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  /* Mark Sent modal */
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--pico-background-color);
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 6px;
    padding: 1.25rem;
    width: min(480px, 90vw);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .modal-title {
    margin: 0;
    font-size: 0.95rem;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .modal-body label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.82rem;
  }

  .modal-body select,
  .modal-body textarea,
  .modal-body input {
    font-size: 0.82rem;
    margin: 0;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
</style>
