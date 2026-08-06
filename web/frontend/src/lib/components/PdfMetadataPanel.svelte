<script lang="ts">
  import Icon from './Icon.svelte';
  import { formatDate } from '../format';
  import type { PdfMetadata } from '../types';

  let { metadata }: { metadata: PdfMetadata | null } = $props();
</script>

{#if !metadata}
  <div class="flex flex-col items-center px-8 py-16 text-center">
    <Icon name="file-x" class="mb-4 h-[34px] w-[34px] text-faint" />
    <p class="m-0 mb-2 text-lg font-medium text-ink">No PDF metadata available</p>
    <p class="m-0 max-w-[46ch] text-base text-muted">
      The PDF structure couldn't be inspected for this file.
    </p>
  </div>
{:else}
  <div
    class="mb-[22px] flex items-center gap-2 border px-3 py-2.5 text-base font-medium"
    class:border-success-soft={metadata.pdfaCompliant}
    class:bg-success-soft={metadata.pdfaCompliant}
    class:text-success={metadata.pdfaCompliant}
    class:border-critical-soft={!metadata.pdfaCompliant}
    class:bg-critical-soft={!metadata.pdfaCompliant}
    class:text-critical={!metadata.pdfaCompliant}
  >
    <Icon name={metadata.pdfaCompliant ? 'check' : 'critical'} class="h-4 w-4 flex-none" />
    {#if metadata.pdfaFlavour}
      {metadata.pdfaFlavour} {metadata.pdfaCompliant ? 'compliant' : '— not fully compliant'}
    {:else}
      Not a PDF/A document
    {/if}
  </div>

  <dl class="mb-[22px] grid grid-cols-2 gap-4 border border-hairline bg-surface-2 px-[18px] py-4 sm:grid-cols-3">
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">PDF version</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{metadata.pdfVersion}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Pages</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{metadata.pageCount}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Encrypted</dt>
      <dd class="m-0 text-md font-medium text-ink">{metadata.encrypted ? 'Yes' : 'No'}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">XMP metadata</dt>
      <dd class="m-0 text-md font-medium text-ink">{metadata.hasXmpMetadata ? 'Present' : 'Not present'}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Producer</dt>
      <dd class="m-0 text-md font-medium text-ink">{metadata.producer ?? '—'}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Creator</dt>
      <dd class="m-0 text-md font-medium text-ink">{metadata.creator ?? '—'}</dd>
    </div>
    <div>
      <dt class="m-0 mb-1 text-xs uppercase tracking-[0.06em] text-muted">Created</dt>
      <dd class="m-0 text-md font-medium tabular-nums text-ink">{formatDate(metadata.creationDate)}</dd>
    </div>
  </dl>

  <div>
    <h4 class="m-0 mb-2 text-xs font-medium uppercase tracking-[0.08em] text-muted">Embedded files</h4>
    {#if metadata.embeddedFiles.length === 0}
      <p class="m-0 text-base text-muted">No files embedded in this PDF.</p>
    {:else}
      <ul class="flex flex-wrap gap-1.5 p-0">
        {#each metadata.embeddedFiles as file (file)}
          <li class="list-none border border-hairline px-2.5 py-1 text-xs font-medium tracking-[0.01em] text-muted">

            {file}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}
