<script lang="ts">
  import type { Snippet } from 'svelte';

  let { title, open = $bindable(false), onopen, children }: {
    title: string;
    open?: boolean;
    onopen?: () => void;
    children: Snippet;
  } = $props();

  function handleToggle(e: Event) {
    if ((e.target as HTMLDetailsElement).open && onopen) {
      onopen();
    }
  }
</script>

<details bind:open ontoggle={handleToggle}>
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
    list-style: none;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  summary::after {
    content: '›';
    font-size: 1.3rem;
    transition: transform 0.2s ease;
  }

  details[open] > summary::after {
    transform: rotate(90deg);
  }

  summary::-webkit-details-marker {
    display: none;
  }

  .card-body {
    padding: 0 1rem 1rem;
  }
</style>
