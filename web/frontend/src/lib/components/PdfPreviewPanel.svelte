<script lang="ts">
  let { file }: { file: File } = $props();

  let url = $state<string>('');

  $effect(() => {
    const objectUrl = URL.createObjectURL(file);
    url = objectUrl;
    return () => URL.revokeObjectURL(objectUrl);
  });
</script>

<div class="flex justify-center pb-1.5 pt-5">
  {#if url}
    <iframe
      src={url}
      title="PDF preview of {file.name}"
      class="h-[70vh] w-full max-w-[820px] border border-hairline bg-surface"
      style={`box-shadow: 0 16px 32px -22px var(--shadow);`}
    ></iframe>
  {/if}
</div>
