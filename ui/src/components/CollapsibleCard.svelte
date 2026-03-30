<script lang="ts">
  import type { Snippet } from 'svelte';

  let { title, open = $bindable(false), onopen, children }: {
    title: string;
    open?: boolean;
    onopen?: () => void;
    children: Snippet;
  } = $props();

  $effect(() => {
    if (open && onopen) {
      onopen();
    }
  });
</script>

<details bind:open>
  <summary>{title}</summary>
  <div class="card-body">
    {@render children()}
  </div>
</details>

<style>
  details {
    border: 1px solid var(--pico-muted-border-color);
    border-radius: var(--pico-border-radius);
    padding: 0;
    margin-bottom: 1.5rem;
  }

  summary {
    cursor: pointer;
    font-weight: 600;
    font-size: 1.1rem;
    padding: 0.75rem 1rem;
    color: var(--pico-color);
  }

  .card-body {
    padding: 0 1rem 1rem;
  }
</style>
