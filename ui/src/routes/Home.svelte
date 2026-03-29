<script lang="ts">
  import { push } from "svelte-spa-router";
</script>

<div class="home">
  <h1>jdextractor</h1>
  <p>Job description extractor and resume tailor.</p>
  <button onclick={() => push("/hello")}>Hello World</button>
</div>

<style>
  .home {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    text-align: center;
    gap: 0.5rem;
  }
</style>
