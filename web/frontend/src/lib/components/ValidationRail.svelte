<script lang="ts">
  import Icon from './Icon.svelte';
  import CheckAccordionItem from './CheckAccordionItem.svelte';
  import { severityTier } from '../findings';
  import type { Finding } from '../types';

  let { findings, valid }: { findings: Finding[]; valid: boolean } = $props();

  const notices = $derived(findings.filter((f) => severityTier(f.severity) === 'notice'));
  const warnings = $derived(findings.filter((f) => severityTier(f.severity) === 'warning'));
  const criticals = $derived(findings.filter((f) => severityTier(f.severity) === 'critical'));
</script>

<aside
  class="relative z-10 border-r border-hairline px-5 py-[22px]"
  style="background: color-mix(in srgb, var(--surface-2) 58%, transparent); backdrop-filter: blur(16px) saturate(160%); -webkit-backdrop-filter: blur(16px) saturate(160%);"
>
  <div
    class="mb-[22px] flex items-center gap-2 border px-3 py-2.5 text-base font-medium"
    class:border-success-soft={valid}
    class:bg-success-soft={valid}
    class:text-success={valid}
    class:border-critical-soft={!valid}
    class:bg-critical-soft={!valid}
    class:text-critical={!valid}
  >
    <Icon name={valid ? 'check' : 'critical'} class="h-4 w-4 flex-none" />
    {valid ? 'Document is valid' : 'Document has failures'}
  </div>

  <h3 class="m-0 mb-3.5 text-xs font-medium uppercase tracking-[0.08em] text-muted">Validation summary</h3>
  <ul class="mb-[22px] flex list-none gap-2.5 p-0">
    <li class="flex-1 border border-hairline bg-surface py-2.5 text-center text-xl font-medium tabular-nums text-accent">
      {notices.length}
      <span class="mt-0.5 block text-2xs font-medium uppercase tracking-[0.04em]">Notices</span>
    </li>
    <li class="flex-1 border border-hairline bg-surface py-2.5 text-center text-xl font-medium tabular-nums text-warning">
      {warnings.length}
      <span class="mt-0.5 block text-2xs font-medium uppercase tracking-[0.04em]">Warnings</span>
    </li>
    <li class="flex-1 border border-hairline bg-surface py-2.5 text-center text-xl font-medium tabular-nums text-critical">
      {criticals.length}
      <span class="mt-0.5 block text-2xs font-medium uppercase tracking-[0.04em]">Errors</span>
    </li>
  </ul>

  <h3 class="m-0 mb-3.5 text-xs font-medium uppercase tracking-[0.08em] text-muted">Checks</h3>
  {#if findings.length === 0}
    <p class="m-0 text-base text-muted">No issues found — every check mustangproject ran came back clean.</p>
  {:else}
    <ul class="flex list-none flex-col gap-2 p-0">
      {#each findings as finding (finding.id + finding.section + finding.message)}
        <CheckAccordionItem {finding} />
      {/each}
    </ul>
  {/if}
</aside>
