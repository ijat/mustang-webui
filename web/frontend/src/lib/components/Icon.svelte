<script lang="ts">
  type IconName =
    | 'check'
    | 'warning'
    | 'critical'
    | 'notice'
    | 'chevron'
    | 'upload'
    | 'sun'
    | 'moon'
    | 'code'
    | 'file-x';

  let { name, class: className = '' }: { name: IconName; class?: string } = $props();
</script>

{#if name === 'check'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <path d="M4 10.5l3.5 3.5L16 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
{:else if name === 'warning'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <path d="M10 3.5l7.5 13h-15l7.5-13z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
    <path d="M10 8.25v3.25" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
    <circle cx="10" cy="14" r="0.9" fill="currentColor" />
  </svg>
{:else if name === 'critical'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <circle cx="10" cy="10" r="7.25" stroke="currentColor" stroke-width="1.6" />
    <path d="M7.5 7.5l5 5m0-5l-5 5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
  </svg>
{:else if name === 'notice'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <circle cx="10" cy="10" r="7.25" stroke="currentColor" stroke-width="1.6" />
    <path d="M10 9v4.25" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
    <circle cx="10" cy="6.5" r="0.9" fill="currentColor" />
  </svg>
{:else if name === 'chevron'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <path d="M5 7.5l5 5 5-5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
{:else if name === 'upload'}
  <svg class={className} width="34" height="34" viewBox="0 0 24 24" fill="none" aria-hidden="true">
    <path d="M12 3v13m0-13l-4.5 4.5M12 3l4.5 4.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
    <path d="M4 16.5V19a2 2 0 002 2h12a2 2 0 002-2v-2.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
  </svg>
{:else if name === 'sun'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <circle cx="10" cy="10" r="3.5" stroke="currentColor" stroke-width="1.5" />
    <path d="M10 2v2M10 16v2M3.5 10h-2M18.5 10h-2M5.05 5.05L3.6 3.6M16.4 16.4l-1.45-1.45M5.05 14.95L3.6 16.4M16.4 3.6l-1.45 1.45" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
  </svg>
{:else if name === 'moon'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <path d="M17 11.5A7 7 0 018.5 3 7.5 7.5 0 1017 11.5z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
  </svg>
{:else if name === 'code'}
  <svg class={className} viewBox="0 0 20 20" fill="none" aria-hidden="true">
    <path d="M7 5.5L2.5 10 7 14.5M13 5.5L17.5 10 13 14.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
{:else if name === 'file-x'}
  <svg class={className} width="34" height="34" viewBox="0 0 24 24" fill="none" aria-hidden="true">
    <path d="M6 2.5h8l4 4V20a1.5 1.5 0 01-1.5 1.5h-11A1.5 1.5 0 014 20V4A1.5 1.5 0 016 2.5z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
    <path d="M14 2.5V7h4" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
    <path d="M9 12l6 6m0-6l-6 6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
  </svg>
{/if}
