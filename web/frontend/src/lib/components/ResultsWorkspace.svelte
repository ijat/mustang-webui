<script lang="ts">
  import ValidationRail from './ValidationRail.svelte';
  import HumanReadablePanel from './HumanReadablePanel.svelte';
  import RawXmlPanel from './RawXmlPanel.svelte';
  import PdfPreviewPanel from './PdfPreviewPanel.svelte';
  import PdfMetadataPanel from './PdfMetadataPanel.svelte';
  import { fadeIn } from '../motion';
  import type { InspectResponse } from '../types';

  let { result, file }: { result: InspectResponse; file: File } = $props();

  type TabId = 'human' | 'xml' | 'preview' | 'metadata';
  const tabs: { id: TabId; label: string }[] = [
    { id: 'human', label: 'Human-readable' },
    { id: 'xml', label: 'Raw XML' },
    { id: 'preview', label: 'PDF preview' },
    { id: 'metadata', label: 'PDF metadata' },
  ];

  let active = $state<TabId>('human');
  let tabEls: Record<string, HTMLButtonElement> = {};

  function onTabKeydown(event: KeyboardEvent, index: number) {
    if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
    event.preventDefault();
    const next = (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length;
    active = tabs[next].id;
    tabEls[tabs[next].id]?.focus();
  }
</script>

<div class="grid min-h-[560px] flex-1 grid-cols-1 sm:grid-cols-[280px_1fr]">
  <ValidationRail findings={result.findings} valid={result.valid} />

  <main
    class="relative z-10 min-w-0 px-6.5 py-[22px] pb-[30px]"
    style="background: color-mix(in srgb, var(--surface) 70%, transparent); backdrop-filter: blur(14px) saturate(140%); -webkit-backdrop-filter: blur(14px) saturate(140%);"
  >
    <div class="mb-[22px] flex gap-[22px] border-b border-hairline" role="tablist" aria-label="Result views">
      {#each tabs as tab, i (tab.id)}
        <button
          bind:this={tabEls[tab.id]}
          role="tab"
          id={`tab-${tab.id}`}
          aria-selected={active === tab.id}
          aria-controls={`panel-${tab.id}`}
          tabindex={active === tab.id ? 0 : -1}
          class="border-b-2 pb-3 text-base font-medium transition-colors"
          class:border-accent={active === tab.id}
          class:border-transparent={active !== tab.id}
          class:text-ink={active === tab.id}
          class:text-muted={active !== tab.id}
          onclick={() => (active = tab.id)}
          onkeydown={(e) => onTabKeydown(e, i)}
        >
          {tab.label}
        </button>
      {/each}
    </div>

    {#key active}
      <div
        id={`panel-${active}`}
        role="tabpanel"
        aria-labelledby={`tab-${active}`}
        tabindex="0"
        use:fadeIn
      >
        {#if active === 'human'}
          <HumanReadablePanel invoice={result.invoice} />
        {:else if active === 'xml'}
          <RawXmlPanel rawXml={result.rawXml} />
        {:else if active === 'preview'}
          <PdfPreviewPanel {file} />
        {:else}
          <PdfMetadataPanel metadata={result.metadata} />
        {/if}
      </div>
    {/key}
  </main>
</div>
