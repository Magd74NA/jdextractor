<script lang="ts">
  import { api } from '../lib/api';
  import { refreshContacts } from '../lib/stores.svelte';
  import type { Contact, ContactStatus, ConversationEntry } from '../lib/types';
  import { CONTACT_STATUSES, CHANNELS } from '../lib/types';

  let { contact }: { contact: Contact } = $props();

  let expanded = $state(false);
  let generating = $state(false);
  let generatedMessage = $state('');
  let generatedSubject = $state('');
  let generatedChannel = $state('');
  let generatedTiming = $state('');
  let generateError = $state('');

  let editing = $state(false);
  let editName = $state('');
  let editCompany = $state('');
  let editRole = $state('');
  let editFollowUp = $state('');
  let editNotes = $state('');

  // New conversation form
  let showLogForm = $state(false);
  let newSummary = $state('');
  let newChannel = $state('');
  let newNotes = $state('');
  let logSaving = $state(false);

  function startEdit() {
    editName = contact.name;
    editCompany = contact.company ?? '';
    editRole = contact.role ?? '';
    editFollowUp = contact.follow_up_date ?? '';
    editNotes = contact.notes ?? '';
    editing = true;
  }

  function cancelEdit() {
    editing = false;
  }

  async function saveMeta() {
    const patch: Partial<Contact> = { name: editName, company: editCompany, role: editRole, notes: editNotes };
    if (editFollowUp) patch.follow_up_date = editFollowUp;
    await api.updateContact(contact.dir, patch);
    editing = false;
    await refreshContacts();
  }

  async function updateStatus(e: Event) {
    const status = (e.target as HTMLSelectElement).value as ContactStatus;
    await api.updateContact(contact.dir, { status });
    await refreshContacts();
  }

  async function deleteContact() {
    if (!confirm(`Delete contact "${contact.name}"?`)) return;
    await api.deleteContact(contact.dir);
    await refreshContacts();
  }

  async function logConversation() {
    if (!newSummary.trim()) return;
    logSaving = true;
    try {
      const entry: ConversationEntry = { date: new Date().toISOString().slice(0, 10), summary: newSummary.trim() };
      if (newChannel) entry.channel = newChannel;
      if (newNotes.trim()) entry.notes = newNotes.trim();
      await api.addConversation(contact.dir, entry);
      newSummary = '';
      newChannel = '';
      newNotes = '';
      showLogForm = false;
      await refreshContacts();
    } finally {
      logSaving = false;
    }
  }

  async function deleteConversation(index: number) {
    await api.deleteConversation(contact.dir, index);
    await refreshContacts();
  }

  async function generateFollowup() {
    generating = true;
    generatedMessage = '';
    generatedSubject = '';
    generatedChannel = '';
    generatedTiming = '';
    generateError = '';
    try {
      const result = await api.generateFollowupStream(contact.dir, (event) => {
        if (event.stage === 'content' && event.delta) {
          generatedMessage += event.delta;
        }
      });
      generatedMessage = result.message;
      generatedSubject = result.subject ?? '';
      generatedChannel = result.channel ?? '';
      generatedTiming = result.timing ?? '';
    } catch (e) {
      generateError = e instanceof Error ? e.message : 'Generation failed';
    } finally {
      generating = false;
    }
  }

  function isOverdue(dateStr?: string): boolean {
    if (!dateStr) return false;
    return dateStr <= new Date().toISOString().slice(0, 10);
  }

  function copyToClipboard() {
    navigator.clipboard.writeText(generatedMessage);
  }
</script>

<tr>
  <td>
    {#if editing}
      <input class="edit-input" bind:value={editName} />
    {:else}
      {contact.name}
    {/if}
  </td>
  <td>
    {#if editing}
      <input class="edit-input" bind:value={editCompany} placeholder="Company" />
    {:else}
      {contact.company ?? ''}
    {/if}
  </td>
  <td>
    <select value={contact.status} onchange={updateStatus}>
      {#each CONTACT_STATUSES as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
  </td>
  <td class:overdue={isOverdue(contact.follow_up_date)}>
    {#if editing}
      <input class="edit-input" type="date" bind:value={editFollowUp} />
    {:else}
      {contact.follow_up_date ?? '-'}
    {/if}
  </td>
  <td class="center">{contact.conversations.length}</td>
  <td class="actions-cell">
    <button class="outline btn-sm" onclick={() => expanded = !expanded}>{expanded ? '▲' : '▼'}</button>
    {#if editing}
      <button class="outline btn-sm save-btn" onclick={saveMeta} title="Save">💾</button>
      <button class="outline btn-sm" onclick={cancelEdit} title="Cancel">✕</button>
    {:else}
      <button class="outline btn-sm" onclick={startEdit} title="Edit">✏</button>
      <button class="outline btn-sm danger-btn" onclick={deleteContact} title="Delete">✕</button>
    {/if}
  </td>
</tr>

{#if expanded}
  <tr class="expanded-row">
    <td colspan="6">
      <div class="expanded-content">

        <!-- Contact details -->
        {#if editing}
          <div class="detail-grid">
            <label>Role <input class="edit-input" bind:value={editRole} placeholder="Role/Title" /></label>
            <label>Follow-up <input class="edit-input" type="date" bind:value={editFollowUp} /></label>
            <label class="full-width">Notes <textarea rows={2} bind:value={editNotes}></textarea></label>
          </div>
        {:else}
          <div class="detail-row">
            {#if contact.role}<span><strong>Role:</strong> {contact.role}</span>{/if}
            {#if contact.email}<span><strong>Email:</strong> {contact.email}</span>{/if}
            {#if contact.linkedin}<span><strong>LinkedIn:</strong> <a href={contact.linkedin} target="_blank">{contact.linkedin}</a></span>{/if}
            {#if contact.source}<span><strong>Met:</strong> {contact.source}</span>{/if}
          </div>
          {#if contact.notes}<p class="notes">{contact.notes}</p>{/if}
          {#if contact.tags && contact.tags.length > 0}
            <div class="tags">
              {#each contact.tags as tag}
                <span class="tag">{tag}</span>
              {/each}
            </div>
          {/if}
          {#if contact.linked_jobs && contact.linked_jobs.length > 0}
            <div class="linked-jobs">
              <strong>Linked jobs:</strong>
              {#each contact.linked_jobs as job}
                <span class="tag job-tag">{job}</span>
              {/each}
            </div>
          {/if}
        {/if}

        <!-- Conversations -->
        <div class="section">
          <div class="section-header">
            <h4>Conversations</h4>
            <button class="outline btn-sm" onclick={() => showLogForm = !showLogForm}>
              {showLogForm ? 'Cancel' : '+ Log'}
            </button>
          </div>

          {#if showLogForm}
            <div class="log-form">
              <div class="form-row">
                <select bind:value={newChannel}>
                  <option value="">Channel (optional)</option>
                  {#each CHANNELS as ch}
                    <option value={ch}>{ch}</option>
                  {/each}
                </select>
              </div>
              <textarea rows={2} bind:value={newSummary} placeholder="Summary (required)"></textarea>
              <textarea rows={1} bind:value={newNotes} placeholder="Additional notes (optional)"></textarea>
              <button onclick={logConversation} aria-busy={logSaving} disabled={!newSummary.trim() || logSaving}>
                Save
              </button>
            </div>
          {/if}

          {#if contact.conversations.length === 0}
            <p class="muted">No conversations logged yet.</p>
          {:else}
            <div class="timeline">
              {#each contact.conversations as conv, i}
                <div class="timeline-entry">
                  <div class="timeline-meta">
                    <span class="conv-date">{conv.date}</span>
                    {#if conv.channel}<span class="conv-channel">{conv.channel}</span>{/if}
                    <button class="outline btn-sm danger-btn" onclick={() => deleteConversation(i)} title="Delete">✕</button>
                  </div>
                  <p class="conv-summary">{conv.summary}</p>
                  {#if conv.notes}<p class="conv-notes muted">{conv.notes}</p>{/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Follow-up generator -->
        <div class="section">
          <div class="section-header">
            <h4>AI Follow-up</h4>
            <button class="outline btn-sm" onclick={generateFollowup} aria-busy={generating} disabled={generating}>
              Generate
            </button>
          </div>

          {#if generateError}
            <p class="error">{generateError}</p>
          {/if}

          {#if generatedMessage}
            <div class="generated">
              {#if generatedSubject}<p><strong>Subject:</strong> {generatedSubject}</p>{/if}
              <pre class="message-box">{generatedMessage}</pre>
              {#if generatedChannel || generatedTiming}
                <p class="muted">Channel: {generatedChannel} · Timing: {generatedTiming}</p>
              {/if}
              <button class="outline btn-sm" onclick={copyToClipboard}>Copy</button>
            </div>
          {/if}
        </div>

      </div>
    </td>
  </tr>
{/if}

<style>
  .actions-cell {
    white-space: nowrap;
    overflow: visible;
  }

  .btn-sm {
    padding: 0.25em 0.45em;
    margin-bottom: 0;
  }

  .danger-btn {
    color: var(--pico-del-color);
    border-color: var(--pico-del-color);
  }

  .save-btn {
    color: var(--pico-ins-color);
    border-color: var(--pico-ins-color);
  }

  .center {
    text-align: center;
  }

  .overdue {
    color: var(--pico-del-color);
    font-weight: 600;
  }

  td {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  select {
    margin-bottom: 0;
    padding: 0.2em 0.4em;
    width: 100%;
    font-size: inherit;
  }

  .edit-input {
    width: 100%;
    padding: 0.2rem 0.4rem;
    margin-bottom: 0;
    font-size: 0.875rem;
  }

  .expanded-row td {
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--pico-muted-border-color);
  }

  .expanded-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .detail-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    font-size: 0.85rem;
  }

  .detail-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
  }

  .full-width {
    grid-column: span 2;
  }

  .notes {
    font-size: 0.85rem;
    color: var(--pico-muted-color);
    margin: 0;
  }

  .muted {
    color: var(--pico-muted-color);
    font-size: 0.85rem;
    margin: 0;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .tag {
    background: var(--pico-secondary-background);
    color: var(--pico-secondary);
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    font-size: 0.78rem;
  }

  .job-tag {
    font-family: monospace;
    font-size: 0.72rem;
  }

  .linked-jobs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    font-size: 0.85rem;
  }

  .section {
    border-top: 1px solid var(--pico-muted-border-color);
    padding-top: 0.75rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .section-header h4 {
    margin: 0;
    font-size: 0.9rem;
  }

  .log-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
    padding: 0.75rem;
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
  }

  .form-row {
    display: flex;
    gap: 0.5rem;
  }

  .form-row select {
    flex: 1;
    font-size: 0.85rem;
  }

  .log-form textarea {
    font-size: 0.85rem;
    margin-bottom: 0;
  }

  .log-form button {
    align-self: flex-start;
    margin-bottom: 0;
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .timeline-entry {
    border-left: 2px solid var(--pico-muted-border-color);
    padding-left: 0.75rem;
  }

  .timeline-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  }

  .conv-date {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
    font-family: monospace;
  }

  .conv-channel {
    font-size: 0.72rem;
    background: var(--pico-primary-background);
    color: var(--pico-primary);
    padding: 0.1rem 0.35rem;
    border-radius: 3px;
  }

  .conv-summary {
    font-size: 0.85rem;
    margin: 0;
  }

  .conv-notes {
    font-size: 0.78rem;
  }

  .generated {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .message-box {
    white-space: pre-wrap;
    font-family: inherit;
    font-size: 0.85rem;
    background: var(--pico-card-background-color);
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
    padding: 0.75rem;
    margin: 0;
  }

  .error {
    color: var(--pico-del-color);
    font-size: 0.85rem;
  }
</style>
