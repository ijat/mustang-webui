<script lang="ts">
  import Icon from './Icon.svelte';
  import { fadeIn } from '../motion';

  let { onFile }: { onFile: (file: File) => void } = $props();

  let dragOver = $state(false);
  let inputEl: HTMLInputElement;

  const specs = ['PDF/A-3', 'EN 16931', 'ZUGFeRD 2.x', 'Factur-X', 'XRechnung 3.0'];

  function isPdf(file: File): boolean {
    return file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf');
  }

  function handleFiles(files: FileList | null) {
    const file = files?.[0];
    if (file && isPdf(file)) {
      onFile(file);
    }
  }

  function onDrop(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
    handleFiles(event.dataTransfer?.files ?? null);
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      inputEl.click();
    }
  }
</script>

<div class="relative z-10 flex justify-center px-8 py-16 sm:py-[74px]" use:fadeIn>
  <div
    role="button"
    tabindex="0"
    aria-label="Drop a PDF invoice here, or activate to choose one"
    class="w-full max-w-[460px] border-[1.5px] border-dashed border-hairline px-8 py-13 text-center transition-colors"
    class:border-accent-border={dragOver}
    style={`background: color-mix(in srgb, var(--surface) ${dragOver ? '20%' : '55%'}, ${dragOver ? 'var(--accent-soft) 80%,' : ''} transparent); backdrop-filter: blur(14px) saturate(140%); -webkit-backdrop-filter: blur(14px) saturate(140%);`}
    ondragover={(e) => {
      e.preventDefault();
      dragOver = true;
    }}
    ondragleave={() => (dragOver = false)}
    ondrop={onDrop}
    onclick={() => inputEl.click()}
    onkeydown={onKeydown}
  >
    <Icon name="upload" class="mx-auto mb-[18px] h-[34px] w-[34px] text-muted" />
    <p class="mb-2 text-lg font-medium text-ink">Drop a PDF invoice here, or click to choose one</p>
    <p class="m-0 text-base text-muted">Processed entirely on this device. Nothing is uploaded anywhere.</p>

    <ul class="my-5 flex flex-wrap justify-center gap-1.5 p-0">
      {#each specs as spec (spec)}
        <li
          class="list-none border border-hairline px-2.5 py-1 text-xs font-medium tracking-[0.02em] text-muted"
          style="background: color-mix(in srgb, var(--surface) 45%, transparent);"
        >
          {spec}
        </li>
      {/each}
    </ul>

    <p class="m-0 text-sm leading-[1.5] text-faint">
      Checked against the real specs with <strong class="font-semibold text-muted">mustangproject</strong>.
    </p>

    <input
      bind:this={inputEl}
      type="file"
      accept=".pdf,application/pdf"
      class="hidden"
      onchange={(e) => handleFiles((e.currentTarget as HTMLInputElement).files)}
    />
  </div>
</div>
