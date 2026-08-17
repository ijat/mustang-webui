<script lang="ts">
  import Icon from './Icon.svelte';
  import { theme } from '../theme.svelte';

  let {
    filename = null,
    fileMeta = null,
    formatLabel = null,
    onNewFile,
  }: {
    filename?: string | null;
    fileMeta?: string | null;
    formatLabel?: string | null;
    onNewFile: () => void;
  } = $props();
</script>

<div
  class="relative z-10 flex items-center gap-4 border-b border-hairline px-5 py-3.5 backdrop-blur-2xl backdrop-saturate-150"
  style="background: color-mix(in srgb, var(--surface-2) 62%, transparent);"
>
  <span class="inline-flex items-baseline gap-[3px]">
    <span class="text-3xl font-bold tracking-[-0.015em] text-ink">mustang</span>
    <span class="text-2xs font-medium uppercase tracking-[0.03em] text-faint">webui</span>
  </span>

  {#if filename}
    <span class="flex-1 truncate text-md font-medium text-ink">
      {filename}
      {#if fileMeta}
        <span class="text-sm font-normal text-faint"> · {fileMeta}</span>
      {/if}
    </span>
  {:else}
    <span class="flex-1"></span>
  {/if}

  {#if formatLabel}
    <span
      class="hidden flex-none border border-accent-border bg-accent-soft px-2.5 py-1 text-xs font-medium tracking-[0.02em] text-accent sm:inline-block"
    >
      {formatLabel}
    </span>
  {/if}

  {#if filename}
    <button
      type="button"
      class="flex items-center gap-1.5 border border-hairline px-2.5 py-1.5 text-xs font-medium uppercase tracking-[0.04em] text-ink transition-colors hover:border-accent-border hover:bg-accent-soft"
      onclick={onNewFile}
    >
      <Icon name="upload" class="h-3.5 w-3.5" />
      New file
    </button>
  {/if}

  <button
    type="button"
    class="flex items-center gap-1.5 border border-hairline px-2.5 py-1.5 text-xs font-medium uppercase tracking-[0.04em] text-ink transition-colors hover:border-accent-border hover:bg-accent-soft"
    aria-pressed={theme.dark}
    onclick={() => theme.toggle()}
  >
    <Icon name={theme.dark ? 'moon' : 'sun'} class="h-3.5 w-3.5" />
    {theme.dark ? 'Dark' : 'Light'}
  </button>
</div>
