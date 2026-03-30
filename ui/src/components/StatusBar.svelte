<script lang="ts">
  import type { StatusCounts } from '../lib/dashboard';

  let { counts }: { counts: StatusCounts } = $props();

  let total = $derived(
    counts.draft + counts.applied + counts.interviewing + counts.offer + counts.rejected
  );

  function pct(n: number): string {
    return total > 0 ? ((n / total) * 100).toFixed(1) + '%' : '0%';
  }

  const segments: { key: keyof StatusCounts; label: string; cssClass: string }[] = [
    { key: 'draft',        label: 'Draft',        cssClass: 'seg-draft' },
    { key: 'applied',      label: 'Applied',      cssClass: 'seg-applied' },
    { key: 'interviewing', label: 'Interviewing', cssClass: 'seg-interviewing' },
    { key: 'offer',        label: 'Offer',        cssClass: 'seg-offer' },
    { key: 'rejected',     label: 'Rejected',     cssClass: 'seg-rejected' },
  ];
</script>

<div class="status-section">
  <h4>Status Breakdown</h4>
  {#if total > 0}
    <div class="bar">
      {#each segments as seg}
        {#if counts[seg.key] > 0}
          <div
            class="segment {seg.cssClass}"
            style="width: {pct(counts[seg.key])}"
            title="{seg.label}: {counts[seg.key]}"
          ></div>
        {/if}
      {/each}
    </div>
    <div class="legend">
      {#each segments as seg}
        {#if counts[seg.key] > 0}
          <span class="legend-item">
            <span class="dot {seg.cssClass}"></span>
            {seg.label} ({counts[seg.key]})
          </span>
        {/if}
      {/each}
    </div>
  {:else}
    <p class="empty">No applications yet.</p>
  {/if}
</div>

<style>
  h4 {
    margin-bottom: 0.75rem;
  }

  .bar {
    display: flex;
    height: 1.5rem;
    border-radius: var(--pico-border-radius);
    overflow: hidden;
    margin-bottom: 0.75rem;
  }

  .segment {
    transition: width 0.3s ease;
  }

  .seg-draft        { background: var(--pico-muted-color); }
  .seg-applied      { background: var(--pico-primary); }
  .seg-interviewing { background: var(--pico-primary-focus); }
  .seg-offer        { background: var(--pico-ins-color); }
  .seg-rejected     { background: var(--pico-del-color); }

  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    font-size: 0.8rem;
    color: var(--pico-muted-color);
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  .dot {
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    display: inline-block;
  }

  .empty {
    color: var(--pico-muted-color);
  }
</style>
