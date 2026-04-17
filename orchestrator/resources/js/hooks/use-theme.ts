import { useCallback, useEffect, useState } from 'react';

export type Theme = 'light' | 'dark' | 'system';

function applyTheme(theme: Theme): void {
  const resolved =
    theme === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      : theme;

  document.documentElement.classList.toggle('dark', resolved === 'dark');
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem('theme') as Theme) ?? 'system'
  );

  const setTheme = useCallback((next: Theme) => {
    localStorage.setItem('theme', next);
    setThemeState(next);
    applyTheme(next);
  }, []);

  useEffect(() => {
    if (theme !== 'system') return;

    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => applyTheme('system');
    media.addEventListener('change', handler);
    return () => media.removeEventListener('change', handler);
  }, [theme]);

  return { theme, setTheme };
}
