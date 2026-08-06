<script lang="ts">
  import BackgroundBlobs from './lib/components/BackgroundBlobs.svelte';
  import TitleBar from './lib/components/TitleBar.svelte';
  import Dropzone from './lib/components/Dropzone.svelte';
  import LoadingState from './lib/components/LoadingState.svelte';
  import ErrorState from './lib/components/ErrorState.svelte';
  import ResultsWorkspace from './lib/components/ResultsWorkspace.svelte';
  import { theme } from './lib/theme.svelte';
  import { inspectPdf, ApiRequestError } from './lib/api';
  import { formatBytes } from './lib/format';
  import type { InspectResponse } from './lib/types';

  type Status = 'empty' | 'loading' | 'results' | 'error';

  let status = $state<Status>('empty');
  let file = $state<File | null>(null);
  let result = $state<InspectResponse | null>(null);
  let errorMessage = $state('');

  async function handleFile(picked: File) {
    file = picked;
    status = 'loading';
    try {
      result = await inspectPdf(picked);
      status = 'results';
    } catch (err) {
      errorMessage = err instanceof ApiRequestError ? err.message : 'Something went wrong reading this file.';
      status = 'error';
    }
  }

  function reset() {
    status = 'empty';
    file = null;
    result = null;
    errorMessage = '';
  }

  const fileMeta = $derived(file ? formatBytes(file.size) : null);
</script>

<main
  class="relative flex min-h-screen flex-col overflow-hidden bg-paper text-ink transition-colors duration-300"
  data-theme={theme.dark ? 'dark' : undefined}
>
  <BackgroundBlobs />
  <TitleBar filename={file?.name ?? null} fileMeta={fileMeta} onNewFile={reset} />

  <div class="flex flex-1 flex-col">
    {#if status === 'empty'}
      <Dropzone onFile={handleFile} />
    {:else if status === 'loading'}
      <LoadingState />
    {:else if status === 'error'}
      <ErrorState message={errorMessage} onRetry={reset} />
    {:else if status === 'results' && result && file}
      <ResultsWorkspace {result} {file} />
    {/if}
  </div>
</main>
