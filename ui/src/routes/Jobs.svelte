<script lang="ts">
  import JobsTable from "../components/JobsTable.svelte";
  import { getJobs } from "../lib/stores.svelte";
  import { JOB_STATUSES } from "../lib/types";
  import { push } from "svelte-spa-router";

  let q = $state("");
  let status = $state("");
  let scoreMin = $state("");
  let scoreMax = $state("");
  let dateFrom = $state("");
  let dateTo = $state("");

  function readFiltersFromURL() {
    const qs = window.location.hash.split("?")[1] ?? "";
    const search = new URLSearchParams(qs);
    q = search.get("q") ?? "";
    status = search.get("status") ?? "";
    scoreMin = search.get("score_min") ?? "";
    scoreMax = search.get("score_max") ?? "";
    dateFrom = search.get("date_from") ?? "";
    dateTo = search.get("date_to") ?? "";
  }

  // Read on mount + re-read whenever the hash changes (e.g. nav search push)
  $effect(() => {
    readFiltersFromURL();
    window.addEventListener("hashchange", readFiltersFromURL);
    return () => window.removeEventListener("hashchange", readFiltersFromURL);
  });

  function syncURL() {
    const p = new URLSearchParams();
    if (q) p.set("q", q);
    if (status) p.set("status", status);
    if (scoreMin) p.set("score_min", scoreMin);
    if (scoreMax) p.set("score_max", scoreMax);
    if (dateFrom) p.set("date_from", dateFrom);
    if (dateTo) p.set("date_to", dateTo);
    const qs = p.toString();
    push(qs ? `/jobs?${qs}` : "/jobs");
  }

  function clearFilters() {
    push("/jobs");
  }

  const hasFilters = $derived(
    q !== "" || status !== "" || scoreMin !== "" || scoreMax !== "" || dateFrom !== "" || dateTo !== ""
  );

  const filteredJobs = $derived.by(() => {
    const all = getJobs();
    const qLow = q.trim().toLowerCase();
    const sMin = scoreMin !== "" ? parseInt(scoreMin, 10) : null;
    const sMax = scoreMax !== "" ? parseInt(scoreMax, 10) : null;
    return all.filter((j) => {
      if (qLow && !(`${j.company} ${j.role}`).toLowerCase().includes(qLow)) return false;
      if (status && j.status !== status) return false;
      if (sMin !== null && !isNaN(sMin) && j.score < sMin) return false;
      if (sMax !== null && !isNaN(sMax) && j.score > sMax) return false;
      if (dateFrom && j.date < dateFrom) return false;
      if (dateTo && j.date > dateTo) return false;
      return true;
    });
  });
</script>

<div class="page-header">
  <h2>Applications</h2>
</div>

<div class="filter-bar">
  <input
    class="filter-search"
    type="text"
    placeholder="Search company / role…"
    bind:value={q}
    oninput={syncURL}
  />
  <select bind:value={status} onchange={syncURL}>
    <option value="">All statuses</option>
    {#each JOB_STATUSES as s}
      <option value={s}>{s}</option>
    {/each}
  </select>
  <input
    class="filter-score"
    type="number"
    min="0"
    max="10"
    placeholder="Score ≥"
    bind:value={scoreMin}
    oninput={syncURL}
  />
  <input
    class="filter-score"
    type="number"
    min="0"
    max="10"
    placeholder="Score ≤"
    bind:value={scoreMax}
    oninput={syncURL}
  />
  <input type="date" bind:value={dateFrom} onchange={syncURL} title="From date" />
  <input type="date" bind:value={dateTo} onchange={syncURL} title="To date" />
  {#if hasFilters}
    <button class="outline secondary" onclick={clearFilters}>Clear</button>
  {/if}
</div>

<JobsTable jobs={filteredJobs} />

<style>
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

  .filter-score {
    width: 6rem;
  }

  .filter-bar button {
    padding: 0.3rem 0.75rem;
    font-size: 0.85rem;
  }
</style>
