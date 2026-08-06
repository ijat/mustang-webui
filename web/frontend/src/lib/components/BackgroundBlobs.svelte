<!--
  Slow ambient drift behind the glass chrome — visible proof the blur
  panels above it are real backdrop-filter glass, not opaque decoration.
  Pure CSS keyframes (transform/opacity only) so it costs nothing on
  the main thread; `prefers-reduced-motion` freezes it via app.css's
  global animation-duration override.
-->
<div class="pointer-events-none absolute inset-0 z-0 overflow-hidden" aria-hidden="true">
  <div
    class="absolute -left-[70px] -top-[90px] h-[320px] w-[320px] rounded-full bg-accent opacity-[0.16] blur-[64px]"
    style="animation: drift 32s ease-in-out infinite alternate;"
  ></div>
  <div
    class="absolute -bottom-[110px] right-[12%] h-[280px] w-[280px] rounded-full bg-warning opacity-[0.12] blur-[64px]"
    style="animation: drift 26s ease-in-out infinite alternate; animation-delay: -8s;"
  ></div>
  <div
    class="absolute right-[-70px] top-[38%] h-[240px] w-[240px] rounded-full bg-success opacity-10 blur-[64px]"
    style="animation: drift 36s ease-in-out infinite alternate; animation-delay: -14s;"
  ></div>
</div>

<style>
  @keyframes drift {
    0% {
      transform: translate(0, 0) scale(1);
    }
    50% {
      transform: translate(22px, -16px) scale(1.08);
    }
    100% {
      transform: translate(-18px, 18px) scale(0.95);
    }
  }
</style>
