<script lang="ts">
  import { api } from "../lib/api";
  import { refreshContacts, getJobs } from "../lib/stores.svelte";
  import type {
    Contact,
    ContactStatus,
    Conversation,
    Message,
  } from "../lib/types";
  import { CONTACT_STATUSES, CHANNELS } from "../lib/types";

  let allJobs = $derived(getJobs());
  let showJobPicker = $state(false);
  let jobSearchQuery = $state("");
  let availableJobs = $derived(
    allJobs.filter(
      (j) =>
        !(contact.linked_jobs ?? []).includes(j.dir) &&
        (jobSearchQuery === "" ||
          j.company.toLowerCase().includes(jobSearchQuery.toLowerCase()) ||
          j.role.toLowerCase().includes(jobSearchQuery.toLowerCase())),
    ),
  );

  let { contact }: { contact: Contact } = $props();

  let expanded = $state(false);
  let generating = $state(false);
  let generatedMessage = $state("");
  let generatedSubject = $state("");
  let generatedChannel = $state("");
  let generatedTiming = $state("");
  let generateError = $state("");

  let editing = $state(false);
  let editName = $state("");
  let editCompany = $state("");
  let editRole = $state("");
  let editFollowUp = $state("");
  let editNotes = $state("");

  // New conversation form
  let showNewConvForm = $state(false);
  let newConvSummary = $state("");
  let newConvChannel = $state("");
  let newConvMsgSender = $state("");
  let newConvMsgContent = $state("");
  let newConvSaving = $state(false);

  // Expanded conversation threads
  let expandedConvs = $state<Set<number>>(new Set());

  // Add message form state (per conversation index)
  let addMsgConvIdx = $state<number | null>(null);
  let addMsgSender = $state("");
  let addMsgContent = $state("");
  let addMsgSaving = $state(false);

  // Inline summary editing
  let editSummaryIdx = $state<number | null>(null);
  let editSummaryText = $state("");
  let summarizing = $state<number | null>(null);

  function startEdit() {
    editName = contact.name;
    editCompany = contact.company ?? "";
    editRole = contact.role ?? "";
    editFollowUp = contact.follow_up_date ?? "";
    editNotes = contact.notes ?? "";
    editing = true;
  }

  function cancelEdit() {
    editing = false;
  }

  async function saveMeta() {
    const patch: Partial<Contact> = {
      name: editName,
      company: editCompany,
      role: editRole,
      notes: editNotes,
    };
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

  function toggleConv(idx: number) {
    const next = new Set(expandedConvs);
    if (next.has(idx)) next.delete(idx);
    else next.add(idx);
    expandedConvs = next;
  }

  async function createConversation() {
    if (!newConvSummary.trim()) return;
    newConvSaving = true;
    try {
      const conv: Conversation = {
        summary: newConvSummary.trim(),
        messages: [],
        created: new Date().toISOString().slice(0, 10),
      };
      if (newConvChannel) conv.channel = newConvChannel;
      if (newConvMsgSender.trim() && newConvMsgContent.trim()) {
        conv.messages.push({
          sender: newConvMsgSender.trim(),
          content: newConvMsgContent.trim(),
          date: new Date().toISOString().slice(0, 10),
        });
      }
      await api.addConversation(contact.dir, conv);
      newConvSummary = "";
      newConvChannel = "";
      newConvMsgSender = "";
      newConvMsgContent = "";
      showNewConvForm = false;
      await refreshContacts();
    } finally {
      newConvSaving = false;
    }
  }

  async function deleteConversation(index: number) {
    await api.deleteConversation(contact.dir, index);
    await refreshContacts();
  }

  async function addMessage(convIdx: number) {
    if (!addMsgSender.trim() || !addMsgContent.trim()) return;
    addMsgSaving = true;
    try {
      const msg: Message = {
        sender: addMsgSender.trim(),
        content: addMsgContent.trim(),
        date: new Date().toISOString().slice(0, 10),
      };
      await api.addMessage(contact.dir, convIdx, msg);
      addMsgSender = "";
      addMsgContent = "";
      addMsgConvIdx = null;
      await refreshContacts();
    } finally {
      addMsgSaving = false;
    }
  }

  async function deleteMessage(convIdx: number, msgIdx: number) {
    await api.deleteMessage(contact.dir, convIdx, msgIdx);
    await refreshContacts();
  }

  function startEditSummary(idx: number, current: string) {
    editSummaryIdx = idx;
    editSummaryText = current;
  }

  async function saveSummary(idx: number) {
    await api.updateConversationSummary(contact.dir, idx, editSummaryText);
    editSummaryIdx = null;
    await refreshContacts();
  }

  async function summarizeConv(idx: number) {
    summarizing = idx;
    try {
      await api.summarizeConversation(contact.dir, idx);
      await refreshContacts();
    } finally {
      summarizing = null;
    }
  }

  async function generateFollowup() {
    generating = true;
    generatedMessage = "";
    generatedSubject = "";
    generatedChannel = "";
    generatedTiming = "";
    generateError = "";
    try {
      const result = await api.generateFollowupStream(contact.dir, (event) => {
        if (event.stage === "content" && event.delta) {
          generatedMessage += event.delta;
        }
      });
      generatedMessage = result.message;
      generatedSubject = result.subject ?? "";
      generatedChannel = result.channel ?? "";
      generatedTiming = result.timing ?? "";
    } catch (e) {
      generateError = e instanceof Error ? e.message : "Generation failed";
    } finally {
      generating = false;
    }
  }

  async function addLinkedJob(jobDir: string) {
    const current = contact.linked_jobs ?? [];
    await api.updateContact(contact.dir, { linked_jobs: [...current, jobDir] });
    jobSearchQuery = "";
    showJobPicker = false;
    await refreshContacts();
  }

  async function removeLinkedJob(jobDir: string) {
    const current = contact.linked_jobs ?? [];
    await api.updateContact(contact.dir, {
      linked_jobs: current.filter((j) => j !== jobDir),
    });
    await refreshContacts();
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
  <td class="truncate">
    {#if editing}
      <input class="edit-input" bind:value={editName} />
    {:else}
      {contact.name}
    {/if}
  </td>
  <td class="truncate">
    {#if editing}
      <input
        class="edit-input"
        bind:value={editCompany}
        placeholder="Company"
      />
    {:else}
      {contact.company ?? ""}
    {/if}
  </td>
  <td>
    <select value={contact.status} onchange={updateStatus}>
      {#each CONTACT_STATUSES as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
  </td>
  <td class:overdue={isOverdue(contact.follow_up_date)} class="truncate">
    {#if editing}
      <input class="edit-input" type="date" bind:value={editFollowUp} />
    {:else}
      {contact.follow_up_date ?? "-"}
    {/if}
  </td>
  <td class="center">{contact.conversations.length}</td>
  <td class="actions-cell">
    <button class="outline btn-sm" onclick={() => (expanded = !expanded)}
      >{expanded ? "▲" : "▼"}</button
    >
    {#if editing}
      <button class="outline btn-sm save-btn" onclick={saveMeta} title="Save"
        >💾</button
      >
      <button class="outline btn-sm" onclick={cancelEdit} title="Cancel"
        >✕</button
      >
    {:else}
      <button class="outline btn-sm" onclick={startEdit} title="Edit">✏</button
      >
      <button
        class="outline btn-sm danger-btn"
        onclick={deleteContact}
        title="Delete">✕</button
      >
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
            <label
              >Role <input
                class="edit-input"
                bind:value={editRole}
                placeholder="Role/Title"
              /></label
            >
            <label
              >Follow-up <input
                class="edit-input"
                type="date"
                bind:value={editFollowUp}
              /></label
            >
            <label class="full-width"
              >Notes <textarea class="mono" rows={2} bind:value={editNotes}
              ></textarea></label
            >
          </div>
        {:else}
          <div class="detail-row">
            {#if contact.role}<span><strong>Role:</strong> {contact.role}</span
              >{/if}
            {#if contact.email}<span
                ><strong>Email:</strong> {contact.email}</span
              >{/if}
            {#if contact.linkedin}<span
                ><strong>LinkedIn:</strong>
                <a href={contact.linkedin} target="_blank">{contact.linkedin}</a
                ></span
              >{/if}
            {#if contact.source}<span
                ><strong>Met:</strong> {contact.source}</span
              >{/if}
          </div>
          {#if contact.notes}<p class="notes">{contact.notes}</p>{/if}
          {#if contact.tags && contact.tags.length > 0}
            <div class="tags">
              {#each contact.tags as tag}
                <span class="tag">{tag}</span>
              {/each}
            </div>
          {/if}
          <div class="linked-jobs">
            <strong>Linked jobs:</strong>
            {#each contact.linked_jobs ?? [] as job}
              <span class="tag job-tag">
                {job}
                <button
                  class="tag-remove"
                  onclick={() => removeLinkedJob(job)}
                  title="Unlink">x</button
                >
              </span>
            {/each}
            <button
              class="outline btn-sm"
              onclick={() => {
                showJobPicker = !showJobPicker;
                jobSearchQuery = "";
              }}
            >
              {showJobPicker ? "Cancel" : "+ Link Job"}
            </button>
          </div>
          {#if showJobPicker}
            <div class="job-picker">
              <input
                class="edit-input"
                bind:value={jobSearchQuery}
                placeholder="Search by company or role..."
              />
              {#if availableJobs.length > 0}
                <div class="job-picker-list">
                  {#each availableJobs.slice(0, 8) as job}
                    <button
                      class="job-picker-item"
                      onclick={() => addLinkedJob(job.dir)}
                    >
                      <span class="job-picker-company">{job.company}</span>
                      <span class="job-picker-role">{job.role}</span>
                      <span class="job-picker-date">{job.date}</span>
                    </button>
                  {/each}
                </div>
              {:else}
                <p class="muted">No matching jobs found.</p>
              {/if}
            </div>
          {/if}
        {/if}

        <!-- Conversations -->
        <div class="section">
          <div class="section-header">
            <h4>Conversations</h4>
            <button
              class="outline btn-sm"
              onclick={() => {
                showNewConvForm = !showNewConvForm;
              }}
            >
              {showNewConvForm ? "Cancel" : "+ New Thread"}
            </button>
          </div>

          {#if showNewConvForm}
            <div class="log-form">
              <div class="form-row">
                <select bind:value={newConvChannel}>
                  <option value="">Channel (optional)</option>
                  {#each CHANNELS as ch}
                    <option value={ch}>{ch}</option>
                  {/each}
                </select>
              </div>
              <input
                bind:value={newConvSummary}
                placeholder="Summary (required)"
              />
              <div class="form-sub-header">First message (optional)</div>
              <input bind:value={newConvMsgSender} placeholder="Sender name" />
              <textarea
                rows={2}
                bind:value={newConvMsgContent}
                placeholder="Message content"
              ></textarea>
              <button
                onclick={createConversation}
                aria-busy={newConvSaving}
                disabled={!newConvSummary.trim() || newConvSaving}
              >
                Create Thread
              </button>
            </div>
          {/if}

          {#if contact.conversations.length === 0}
            <p class="muted">No conversations logged yet.</p>
          {:else}
            <div class="conv-list">
              {#each contact.conversations as conv, i}
                <div class="conv-card">
                  <!-- Conversation header (always visible, clickable) -->
                  <div
                    class="conv-header"
                    role="button"
                    tabindex="0"
                    onclick={() => toggleConv(i)}
                    onkeydown={(e) => {
                      if (e.key === "Enter" || e.key === " ") toggleConv(i);
                    }}
                  >
                    <div class="conv-header-left">
                      <span class="conv-toggle"
                        >{expandedConvs.has(i) ? "▾" : "▸"}</span
                      >
                      {#if editSummaryIdx === i}
                        <!-- svelte-ignore a11y_autofocus -->
                        <input
                          class="edit-input summary-edit"
                          bind:value={editSummaryText}
                          onclick={(e) => e.stopPropagation()}
                          onkeydown={(e) => {
                            if (e.key === "Enter") saveSummary(i);
                            if (e.key === "Escape") editSummaryIdx = null;
                          }}
                          autofocus
                        />
                      {:else}
                        <span class="conv-summary">{conv.summary}</span>
                      {/if}
                    </div>
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <div
                      class="conv-header-right"
                      onclick={(e) => e.stopPropagation()}
                    >
                      {#if conv.channel}<span class="channel-badge"
                          >{conv.channel}</span
                        >{/if}
                      <span class="conv-meta">{conv.messages.length} msg</span>
                      <span class="conv-date">{conv.created}</span>
                      {#if editSummaryIdx === i}
                        <button
                          class="outline btn-sm save-btn"
                          onclick={() => saveSummary(i)}
                          title="Save">💾</button
                        >
                        <button
                          class="outline btn-sm"
                          onclick={() => (editSummaryIdx = null)}
                          title="Cancel">✕</button
                        >
                      {:else}
                        <button
                          class="outline btn-sm"
                          onclick={() => startEditSummary(i, conv.summary)}
                          title="Edit summary">✏</button
                        >
                        {#if conv.messages.length > 0}
                          <button
                            class="outline btn-sm"
                            onclick={() => summarizeConv(i)}
                            aria-busy={summarizing === i}
                            disabled={summarizing === i}
                            title="AI Summarize">✦</button
                          >
                        {/if}
                        <button
                          class="outline btn-sm danger-btn"
                          onclick={() => deleteConversation(i)}
                          title="Delete thread">✕</button
                        >
                      {/if}
                    </div>
                  </div>

                  <!-- Expanded: message thread -->
                  {#if expandedConvs.has(i)}
                    <div class="msg-thread">
                      {#if conv.messages.length === 0}
                        <p class="muted">No messages yet.</p>
                      {:else}
                        {#each conv.messages as msg, mi}
                          <div class="msg-entry">
                            <div class="msg-header">
                              <span class="msg-sender">{msg.sender}</span>
                              <span class="msg-date">{msg.date}</span>
                              <button
                                class="outline btn-xs danger-btn"
                                onclick={() => deleteMessage(i, mi)}
                                title="Delete">✕</button
                              >
                            </div>
                            <p class="msg-content">{msg.content}</p>
                          </div>
                        {/each}
                      {/if}

                      <!-- Add message form -->
                      {#if addMsgConvIdx === i}
                        <div class="add-msg-form">
                          <input
                            bind:value={addMsgSender}
                            placeholder="Sender name (e.g. me, Jane)"
                          />
                          <textarea
                            rows={2}
                            bind:value={addMsgContent}
                            placeholder="Message content"
                          ></textarea>
                          <div class="form-actions">
                            <button
                              onclick={() => addMessage(i)}
                              aria-busy={addMsgSaving}
                              disabled={!addMsgSender.trim() ||
                                !addMsgContent.trim() ||
                                addMsgSaving}>Add</button
                            >
                            <button
                              class="outline"
                              onclick={() => (addMsgConvIdx = null)}
                              >Cancel</button
                            >
                          </div>
                        </div>
                      {:else}
                        <button
                          class="outline btn-sm"
                          onclick={() => {
                            addMsgConvIdx = i;
                            addMsgSender = "";
                            addMsgContent = "";
                          }}
                        >
                          + Add Message
                        </button>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Follow-up generator -->
        <div class="section">
          <div class="section-header">
            <h4>AI Follow-up</h4>
            <button
              class="outline btn-sm"
              onclick={generateFollowup}
              aria-busy={generating}
              disabled={generating}
            >
              Generate
            </button>
          </div>

          {#if generateError}
            <p class="error">{generateError}</p>
          {/if}

          {#if generatedMessage}
            <div class="generated">
              {#if generatedSubject}<p>
                  <strong>Subject:</strong>
                  {generatedSubject}
                </p>{/if}
              <pre class="message-box">{generatedMessage}</pre>
              {#if generatedChannel || generatedTiming}
                <p class="muted">
                  Channel: {generatedChannel} · Timing: {generatedTiming}
                </p>
              {/if}
              <button class="outline btn-sm" onclick={copyToClipboard}
                >Copy</button
              >
            </div>
          {/if}
        </div>
      </div>
    </td>
  </tr>
{/if}

<style>
  /* Component-specific styles only - shared styles moved to app.css */
  .overdue {
    color: var(--pico-del-color);
    font-weight: 600;
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

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
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

  .log-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
    padding: 0.75rem;
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
  }

  .log-form input,
  .log-form textarea {
    font-size: 0.85rem;
    margin-bottom: 0;
  }

  .log-form button {
    align-self: flex-start;
    margin-bottom: 0;
  }

  .form-row {
    display: flex;
    gap: 0.5rem;
  }

  .form-row select {
    flex: 1;
    font-size: 0.85rem;
  }

  .form-sub-header {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
    font-weight: 600;
    margin-top: 0.25rem;
  }

  /* Conversation cards */
  .conv-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .conv-card {
    border: 1px solid var(--pico-muted-border-color);
    border-radius: 4px;
  }

  .conv-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    gap: 0.5rem;
  }

  .conv-header:hover {
    background: color-mix(
      in srgb,
      var(--pico-muted-border-color) 20%,
      transparent
    );
  }

  .conv-header-left {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    flex: 1;
  }

  .conv-toggle {
    font-size: 0.78rem;
    color: var(--pico-muted-color);
    flex-shrink: 0;
  }

  .conv-summary {
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .summary-edit {
    flex: 1;
    font-size: 0.85rem;
  }

  .conv-header-right {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
  }

  .conv-meta {
    font-size: 0.72rem;
    color: var(--pico-muted-color);
  }

  .conv-date {
    font-size: 0.72rem;
    color: var(--pico-muted-color);
    font-family: monospace;
  }

  /* Message thread */
  .msg-thread {
    padding: 0.5rem 0.75rem 0.75rem;
    border-top: 1px solid var(--pico-muted-border-color);
    max-height: 400px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .msg-entry {
    border-left: 2px solid var(--pico-muted-border-color);
    padding-left: 0.6rem;
  }

  .msg-header {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-bottom: 0.15rem;
  }

  .msg-sender {
    font-size: 0.82rem;
    font-weight: 600;
  }

  .msg-date {
    font-size: 0.72rem;
    color: var(--pico-muted-color);
    font-family: monospace;
  }

  .msg-content {
    font-size: 0.82rem;
    margin: 0;
    white-space: pre-wrap;
  }

  .add-msg-form {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0.5rem;
    border: 1px dashed var(--pico-muted-border-color);
    border-radius: 4px;
  }

  .add-msg-form input,
  .add-msg-form textarea {
    font-size: 0.82rem;
    margin-bottom: 0;
  }

  .form-actions {
    display: flex;
    gap: 0.5rem;
  }

  .form-actions button {
    margin-bottom: 0;
    font-size: 0.82rem;
    padding: 0.25em 0.6em;
  }

  /* Follow-up generator */
  .generated {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* Job linking */
  .job-picker {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    margin-top: 0.5rem;
    padding: 0.5rem;
    border: 1px dashed var(--pico-muted-border-color);
    border-radius: 4px;
  }

  .job-picker-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    max-height: 200px;
    overflow-y: auto;
  }

  .job-picker-item {
    all: unset;
    cursor: pointer;
    display: flex;
    gap: 0.75rem;
    padding: 0.3rem 0.5rem;
    border-radius: 3px;
    font-size: 0.8rem;
  }

  .job-picker-item:hover {
    background: color-mix(in srgb, var(--pico-primary) 10%, transparent);
  }

  .job-picker-company {
    font-weight: 600;
  }

  .job-picker-role {
    color: var(--pico-muted-color);
    flex: 1;
  }

  .job-picker-date {
    font-family: monospace;
    font-size: 0.72rem;
    color: var(--pico-muted-color);
  }
</style>
