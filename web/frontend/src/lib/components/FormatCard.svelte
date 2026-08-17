<script lang="ts">
  import Icon from './Icon.svelte';
  import type { En16931Status, FormatDescription } from '../eInvoiceFormat';

  let { format }: { format: FormatDescription | null } = $props();

  type Compliance = { label: string; note: string | null; icon: 'check' | 'warning'; tone: string };

  const COMPLIANCE: Record<En16931Status, Compliance | null> = {
    compliant: { label: 'EN 16931 compliant', note: null, icon: 'check', tone: 'text-success' },
    extension: { label: 'EN 16931 conformant extension', note: null, icon: 'check', tone: 'text-success' },
    no: {
      label: 'Not EN 16931 compliant',
      note: 'This profile makes no EN 16931 claim.',
      icon: 'warning',
      tone: 'text-warning',
    },
    'not-applicable': null,
    unknown: null,
  };

  const compliance = $derived(format ? COMPLIANCE[format.en16931] : null);
</script>

<div class="mb-[22px] border border-hairline bg-surface px-3 py-2.5">
  <h3 class="m-0 mb-1.5 text-xs font-medium uppercase tracking-[0.08em] text-muted">Document format</h3>

  {#if !format}
    <p class="m-0 flex items-start gap-1.5 text-base text-muted">
      <Icon name="file-x" class="mt-0.5 h-3.5 w-3.5 flex-none text-faint" />
      No e-invoice XML detected — plain PDF.
    </p>
  {:else}
    <p class="m-0 mb-2 text-md font-semibold leading-snug text-ink">{format.standard}</p>

    <ul class="m-0 mb-2 flex list-none flex-wrap gap-1.5 p-0">
      {#if format.profileLevel}
        <li class="border border-accent-border bg-accent-soft px-2 py-0.5 text-2xs font-medium uppercase tracking-[0.05em] text-accent">
          {format.profileLevel}
        </li>
      {/if}
      {#if format.syntax}
        <li class="border border-hairline px-2 py-0.5 text-2xs font-medium uppercase tracking-[0.05em] text-muted">
          {format.syntax}
        </li>
      {/if}
    </ul>

    {#if compliance}
      <p class={`m-0 mb-2 flex items-start gap-1.5 text-base font-medium leading-snug ${compliance.tone}`}>
        <Icon name={compliance.icon} class="mt-px h-3.5 w-3.5 flex-none" />
        <span>
          {compliance.label}
          {#if compliance.note}
            <span class="block font-normal text-muted">{compliance.note}</span>
          {/if}
        </span>
      </p>
    {/if}

    {#if format.jurisdiction}
      <p class="m-0 mb-2 text-base leading-snug text-muted">{format.jurisdiction}</p>
    {/if}

    {#if format.attachment}
      <p class="m-0 mb-2 text-base leading-snug text-muted">
        <span class="text-ink">{format.attachment.filename}</span> · {format.attachment.naming}
      </p>
    {/if}

    {#if format.specificationId}
      <dl class="m-0 border-t border-hairline pt-2">
        <dt class="m-0 mb-0.5 text-2xs font-medium uppercase tracking-[0.06em] text-faint">Specification identifier</dt>
        <dd class="m-0 break-all text-xs leading-snug text-muted">{format.specificationId}</dd>
      </dl>
    {/if}
  {/if}
</div>
