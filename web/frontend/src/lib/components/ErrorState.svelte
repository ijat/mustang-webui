<script lang="ts">
  import Icon from './Icon.svelte';
  import { fadeIn } from '../motion';

  let { message, onRetry }: { message: string; onRetry: () => void } = $props();
</script>

<div class="relative z-10 flex justify-center px-8 py-16 sm:py-[74px]" use:fadeIn>
  <div class="w-full max-w-[460px] border border-hairline bg-surface px-8 py-10 text-center">
    <Icon name="file-x" class="mx-auto mb-4 h-[34px] w-[34px] text-critical" />
    <p class="mb-2 text-lg font-medium text-ink">Couldn't read that file</p>
    <p class="m-0 mb-6 text-base text-muted">{message}</p>
    <button
      type="button"
      class="border border-hairline px-4 py-2 text-sm font-medium uppercase tracking-[0.04em] text-ink transition-colors hover:border-accent-border hover:bg-accent-soft"
      onclick={onRetry}
    >
      Try another file
    </button>
  </div>
</div>
