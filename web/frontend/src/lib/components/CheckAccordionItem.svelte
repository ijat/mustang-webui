<script lang="ts">
  import Icon from './Icon.svelte';
  import { partLabel, severityLabel, severityTier } from '../findings';
  import type { Finding } from '../types';

  let { finding }: { finding: Finding } = $props();

  let open = $state(false);
  const tier = $derived(severityTier(finding.severity));
  const iconName = $derived(tier === 'notice' ? 'notice' : tier === 'warning' ? 'warning' : 'critical');
  const stripeVar = $derived(tier === 'notice' ? 'var(--accent)' : tier === 'warning' ? 'var(--warning)' : 'var(--critical)');
  const iconColorClass = $derived(tier === 'notice' ? 'text-accent' : tier === 'warning' ? 'text-warning' : 'text-critical');
  const detailId = $derived(`finding-detail-${finding.id}-${finding.section}`);
  const label = $derived(partLabel[finding.part]);
</script>

<li class="border border-hairline bg-surface" style={`border-left: 3px solid ${stripeVar};`}>
  <button
    type="button"
    class="flex w-full items-start gap-2.5 px-3 py-2.5 text-left"
    aria-expanded={open}
    aria-controls={detailId}
    onclick={() => (open = !open)}
  >
    <Icon name={iconName} class={`mt-[1px] h-[15px] w-[15px] flex-none ${iconColorClass}`} />
    <span class="min-w-0 flex-1">
      <span class="block text-base font-medium text-ink">{finding.id}</span>
      <span class="mt-[3px] block text-sm text-muted">
        {severityLabel[finding.severity]}{label ? ` · ${label}` : ''}
      </span>
    </span>
    <Icon
      name="chevron"
      class={`h-4 w-4 flex-none text-faint transition-transform duration-[280ms] ease-[cubic-bezier(0.22,0.61,0.36,1)] ${open ? 'rotate-180' : ''}`}
    />
  </button>
  <div
    id={detailId}
    class="grid transition-[grid-template-rows] duration-[320ms] ease-[cubic-bezier(0.22,0.61,0.36,1)]"
    style={`grid-template-rows: ${open ? '1fr' : '0fr'};`}
  >
    <div class="min-h-0 overflow-hidden">
      <p class="m-0 px-3 pb-3.5 pl-[34px] text-base leading-[1.5] text-muted">{finding.message}</p>
    </div>
  </div>
</li>
