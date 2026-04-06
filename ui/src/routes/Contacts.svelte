<script lang="ts">
  import ContactsTable from "../components/ContactsTable.svelte";
  import { api } from "../lib/api";
  import { getContacts, refreshContacts } from "../lib/stores.svelte";
  import { CONTACT_STATUSES } from "../lib/types";
  import { push } from "svelte-spa-router";

  let showForm = $state(false);
  let saving = $state(false);
  let formError = $state("");

  let name = $state("");
  let company = $state("");
  let role = $state("");
  let email = $state("");
  let phone = $state("");
  let linkedin = $state("");
  let source = $state("");
  let status = $state("new");
  let followUpDate = $state("");
  let notes = $state("");

  function resetForm() {
    name = "";
    company = "";
    role = "";
    email = "";
    phone = "";
    linkedin = "";
    source = "";
    status = "new";
    followUpDate = "";
    notes = "";
    formError = "";
  }

  async function createContact() {
    if (!name.trim()) {
      formError = "Name is required";
      return;
    }
    saving = true;
    formError = "";
    try {
      const data: Record<string, string> = { name: name.trim(), status };
      if (company) data.company = company;
      if (role) data.role = role;
      if (email) data.email = email;
      if (phone) data.phone = phone;
      if (linkedin) data.linkedin = linkedin;
      if (source) data.source = source;
      if (followUpDate) data.follow_up_date = followUpDate;
      if (notes) data.notes = notes;
      await api.createContact(data);
      resetForm();
      showForm = false;
      await refreshContacts();
    } catch (e) {
      formError = e instanceof Error ? e.message : "Failed to create contact";
    } finally {
      saving = false;
    }
  }

  // --- Filter state ---

  let fq = $state("");
  let fStatus = $state("");
  let fTag = $state("");
  let fFollowup = $state<"" | "overdue" | "this-week" | "has-date">("");

  function readFiltersFromURL() {
    const qs = window.location.hash.split("?")[1] ?? "";
    const search = new URLSearchParams(qs);
    fq = search.get("q") ?? "";
    fStatus = search.get("status") ?? "";
    fTag = search.get("tag") ?? "";
    fFollowup = (search.get("followup") ?? "") as typeof fFollowup;
  }

  // Read on mount + re-read whenever the hash changes (e.g. nav search push)
  $effect(() => {
    readFiltersFromURL();
    window.addEventListener("hashchange", readFiltersFromURL);
    return () => window.removeEventListener("hashchange", readFiltersFromURL);
  });

  function syncURL() {
    const p = new URLSearchParams();
    if (fq) p.set("q", fq);
    if (fStatus) p.set("status", fStatus);
    if (fTag) p.set("tag", fTag);
    if (fFollowup) p.set("followup", fFollowup);
    const qs = p.toString();
    push(qs ? `/contacts?${qs}` : "/contacts");
  }

  function clearFilters() {
    push("/contacts");
  }

  const hasFilters = $derived(fq !== "" || fStatus !== "" || fTag !== "" || fFollowup !== "");

  const allTags = $derived(
    [...new Set(getContacts().flatMap((c) => c.tags ?? []))].sort()
  );

  const filteredContacts = $derived.by(() => {
    const today = new Date().toISOString().slice(0, 10);
    const weekAhead = new Date(Date.now() + 7 * 86400000).toISOString().slice(0, 10);
    const qLow = fq.trim().toLowerCase();
    return getContacts().filter((c) => {
      if (qLow) {
        const msgText = c.conversations
          .flatMap((cv) => cv.messages)
          .map((m) => m.content)
          .join(" ");
        const haystack = `${c.name} ${c.company ?? ""} ${c.role ?? ""} ${c.email ?? ""} ${c.notes ?? ""} ${msgText}`.toLowerCase();
        if (!haystack.includes(qLow)) return false;
      }
      if (fStatus && c.status !== fStatus) return false;
      if (fTag && !(c.tags ?? []).some((t) => t.toLowerCase() === fTag.toLowerCase())) return false;
      if (fFollowup === "has-date" && !c.follow_up_date) return false;
      if (fFollowup === "overdue" && (!c.follow_up_date || c.follow_up_date >= today)) return false;
      if (fFollowup === "this-week" && (!c.follow_up_date || c.follow_up_date < today || c.follow_up_date > weekAhead)) return false;
      return true;
    });
  });
</script>

<div class="page-header">
  <h2>Contacts</h2>
  <button
    class="outline"
    onclick={() => {
      showForm = !showForm;
      if (!showForm) resetForm();
    }}
  >
    {showForm ? "Cancel" : "+ New Contact"}
  </button>
</div>

{#if showForm}
  <article class="contact-form">
    <h3>New Contact</h3>
    {#if formError}<p class="error">{formError}</p>{/if}
    <div class="form-grid">
      <label>Name * <input bind:value={name} placeholder="Jane Doe" /></label>
      <label
        >Company <input bind:value={company} placeholder="Acme Corp" /></label
      >
      <label
        >Role <input
          bind:value={role}
          placeholder="Engineering Manager"
        /></label
      >
      <label
        >Email <input
          type="email"
          bind:value={email}
          placeholder="jane@acme.com"
        /></label
      >
      <label>Phone <input bind:value={phone} placeholder="+1 555 0100" /></label
      >
      <label
        >LinkedIn <input
          bind:value={linkedin}
          placeholder="https://linkedin.com/in/..."
        /></label
      >
      <label
        >How we met
        <input
          bind:value={source}
          placeholder="conference, referral, cold outreach…"
        />
      </label>
      <label
        >Status
        <select bind:value={status}>
          {#each CONTACT_STATUSES as s}
            <option value={s}>{s}</option>
          {/each}
        </select>
      </label>
      <label
        >Follow-up date <input type="date" bind:value={followUpDate} /></label
      >
      <label class="full-width"
        >Notes <textarea rows={2} bind:value={notes}></textarea></label
      >
    </div>
    <button onclick={createContact} aria-busy={saving} disabled={saving}
      >Create Contact</button
    >
  </article>
{/if}

<div class="filter-bar">
  <input
    class="filter-search"
    type="text"
    placeholder="Search name / company / notes…"
    bind:value={fq}
    oninput={syncURL}
  />
  <select bind:value={fStatus} onchange={syncURL}>
    <option value="">All statuses</option>
    {#each CONTACT_STATUSES as s}
      <option value={s}>{s}</option>
    {/each}
  </select>
  <select bind:value={fTag} onchange={syncURL}>
    <option value="">All tags</option>
    {#each allTags as t}
      <option value={t}>{t}</option>
    {/each}
  </select>
  <select bind:value={fFollowup} onchange={syncURL}>
    <option value="">Any follow-up</option>
    <option value="has-date">Has date</option>
    <option value="overdue">Overdue</option>
    <option value="this-week">This week</option>
  </select>
  {#if hasFilters}
    <button class="outline secondary" onclick={clearFilters}>Clear</button>
  {/if}
</div>

<ContactsTable contacts={filteredContacts} />

<style>
  .contact-form {
    margin-bottom: 2rem;
    padding: 1.5rem;
  }

  .contact-form h3 {
    margin-top: 0;
    margin-bottom: 1rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .full-width {
    grid-column: span 2;
  }

  .form-grid label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.85rem;
    font-weight: 600;
    margin-bottom: 0;
  }

  .form-grid input,
  .form-grid select,
  .form-grid textarea {
    margin-bottom: 0;
    font-size: 0.85rem;
    font-weight: 400;
  }

  .filter-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .filter-bar select,
  .filter-bar input {
    width: auto;
    margin-bottom: 0;
    font-size: 0.85rem;
    padding: 0.3rem 0.5rem;
    height: auto;
  }

  .filter-search {
    flex: 1;
    min-width: 14rem;
  }

  .filter-bar button {
    padding: 0.3rem 0.75rem;
    font-size: 0.85rem;
  }
</style>
