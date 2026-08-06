const STORAGE_KEY = 'mustang-webui:theme';

function loadInitial(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === 'dark';
  } catch {
    return false;
  }
}

let dark = $state(loadInitial());

/**
 * Dark mode is an explicit, in-app opt-in — it never follows
 * `prefers-color-scheme`. The choice persists across launches via
 * localStorage and is applied as a `data-theme` attribute scoped to the
 * app shell (see App.svelte), not to <html>.
 */
export const theme = {
  get dark() {
    return dark;
  },
  toggle() {
    dark = !dark;
    try {
      localStorage.setItem(STORAGE_KEY, dark ? 'dark' : 'light');
    } catch {
      // Storage unavailable (private mode, etc.) — the toggle still works
      // for the session, it just won't persist.
    }
  },
};
