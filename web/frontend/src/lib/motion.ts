import { animate } from 'motion';

function prefersReducedMotion(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Svelte action: spring-fades an element in on mount. Used for
 * user-triggered, interruptible transitions (tab switches, state
 * changes) per the animate-transform/opacity-only rule. Reduced-motion
 * users get the end state immediately, no animation.
 */
export function fadeIn(node: HTMLElement, options: { delay?: number } = {}) {
  if (prefersReducedMotion()) {
    node.style.opacity = '1';
    node.style.transform = 'none';
    return {};
  }

  node.style.opacity = '0';
  node.style.transform = 'translateY(6px)';

  const controls = animate(
    node,
    { opacity: [0, 1], transform: ['translateY(6px)', 'translateY(0)'] },
    { type: 'spring', stiffness: 300, damping: 30, delay: options.delay ?? 0 },
  );

  return {
    destroy() {
      controls.stop();
    },
  };
}
