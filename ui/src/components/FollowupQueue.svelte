<script lang="ts">
  import { api } from "../lib/api";
  import type { Contact } from "../lib/types";
  import FollowupPanel from "./FollowupPanel.svelte";

  let overdue = $state<Contact[]>([]);
  let upcoming = $state<Contact[]>([]);
  let loading = $state(true);
  let error = $state("");

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

  function lastConvSummary(contact: Contact): string {
    if (!contact.conversations || contact.conversations.length === 0) return "";
    return contact.conversations.at(-1)?.summary ?? "";
  }

  function lastChannel(contact: Contact): string {
    if (!contact.conversations || contact.conversations.length === 0) return "";
    return contact.conversations.at(-1)?.channel ?? "";
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
              {#if contact.company}<span class="queue-company">{contact.company}</span>{/if}
              {#if lastChannel(contact)}<span class="channel-badge">{lastChannel(contact)}</span>{/if}
            </div>
            <div class="queue-center">
              <span class="queue-snippet">{lastConvSummary(contact)}</span>
            </div>
            <div class="queue-right">
              <span class="queue-date">{contact.follow_up_date}</span>
              <span class="days-badge overdue-badge">{Math.abs(days)}d overdue</span>
            </div>
          </div>
          <div class="queue-panel">
            <FollowupPanel contact={contact} onSent={loadQueue} />
          </div>
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
              {#if contact.company}<span class="queue-company">{contact.company}</span>{/if}
              {#if lastChannel(contact)}<span class="channel-badge">{lastChannel(contact)}</span>{/if}
            </div>
            <div class="queue-center">
              <span class="queue-snippet">{lastConvSummary(contact)}</span>
            </div>
            <div class="queue-right">
              <span class="queue-date">{contact.follow_up_date}</span>
              <span class="days-badge">{days}d</span>
            </div>
          </div>
          <div class="queue-panel">
            <FollowupPanel contact={contact} onSent={loadQueue} />
          </div>
        </div>
      {/each}
    </div>
  {/if}
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
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.3;
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

  .queue-panel {
    margin-top: 0.5rem;
  }
</style>
