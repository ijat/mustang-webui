<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import gsap from 'gsap';
  import { fadeIn } from '../motion';

  const steps = ['Checking PDF/A structure', 'Parsing embedded XML', 'Running Schematron rules'];

  let activeIndex = $state(0);
  let stepEls: HTMLLIElement[] = [];
  let tl: gsap.core.Timeline | undefined;
  let reduced = $state(false);

  onMount(() => {
    reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (reduced) return;

    // A fixed, precisely-sequenced choreography (not user-interruptible),
    // so GSAP rather than Motion — it just loops for as long as the
    // sidecar takes, since the backend gives no real per-step progress.
    tl = gsap.timeline({ repeat: -1 });
    steps.forEach((_, i) => {
      tl!.call(() => (activeIndex = i)).to({}, { duration: 0.9 });
    });
  });

  onDestroy(() => {
    tl?.kill();
  });
</script>

<div class="relative z-10 flex justify-center px-8 py-20" use:fadeIn>
  <div class="w-full max-w-[340px]">
    <ul class="flex flex-col gap-3 p-0">
      {#each steps as step, i (step)}
        <li
          bind:this={stepEls[i]}
          class="flex items-center gap-3 border border-hairline px-4 py-3 text-base transition-colors duration-300"
          class:border-accent-border={activeIndex === i && !reduced}
          class:text-ink={activeIndex >= i || reduced}
          class:text-faint={activeIndex < i && !reduced}
        >
          <span
            class="h-2 w-2 flex-none rounded-full transition-colors duration-300"
            style={`background: ${activeIndex >= i || reduced ? 'var(--accent)' : 'var(--hairline)'};`}
          ></span>
          {step}
        </li>
      {/each}
    </ul>
  </div>
</div>
