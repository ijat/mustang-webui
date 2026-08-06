<script lang="ts">
  import Icon from './Icon.svelte';
  import { highlightXml } from '../xmlHighlight';

  let { rawXml }: { rawXml: string | null } = $props();

  const highlighted = $derived(rawXml ? highlightXml(rawXml) : '');
</script>

{#if !rawXml}
  <div class="flex flex-col items-center px-8 py-16 text-center">
    <Icon name="code" class="mb-4 h-[34px] w-[34px] text-faint" />
    <p class="m-0 mb-2 text-lg font-medium text-ink">No embedded XML</p>
    <p class="m-0 max-w-[46ch] text-base text-muted">
      This PDF has no ZUGFeRD/Factur-X XML attachment for mustangproject to extract.
    </p>
  </div>
{:else}
  <div class="overflow-x-auto border border-hairline bg-surface-2">
    <pre class="m-0 min-w-max whitespace-pre px-5 py-[18px] font-mono text-sm leading-[1.7] text-ink xml-code">{@html highlighted}</pre>
  </div>
{/if}

<style>
  :global(.xml-code .xml-tag) {
    color: var(--accent);
  }
  :global(.xml-code .xml-attr) {
    color: var(--warning);
  }
  :global(.xml-code .xml-val) {
    color: var(--success);
  }
  :global(.xml-code .xml-muted) {
    color: var(--faint);
  }
</style>
