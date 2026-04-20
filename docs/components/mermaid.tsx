'use client';

import { useEffect, useRef } from 'react';

export function Mermaid({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;

    const render = async () => {
      const node = ref.current;
      if (!node) return;

      const isDark = document.documentElement.classList.contains('dark');
      const { default: mermaid } = await import('mermaid');

      if (cancelled || !ref.current) return;

      mermaid.initialize({ startOnLoad: false, theme: isDark ? 'dark' : 'neutral' });

      node.removeAttribute('data-processed');
      node.innerHTML = chart;

      try {
        await mermaid.run({ nodes: [node] });
      } catch (err) {
        console.error('Mermaid render error:', err);
      }
    };

    render();

    const observer = new MutationObserver(() => render());
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    });

    return () => {
      cancelled = true;
      observer.disconnect();
    };
  }, [chart]);

  return <div className="mermaid" ref={ref} />;
}
