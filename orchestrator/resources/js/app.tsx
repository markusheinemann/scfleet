import { TooltipProvider } from '@/components/ui/tooltip';
import { createInertiaApp } from '@inertiajs/react';
import { createRoot } from 'react-dom/client';

createInertiaApp({
  strictMode: true,
  resolve: name => {
    const pages = import.meta.glob('./pages/**/*.tsx', { eager: true });
    return pages[`./pages/${name}.tsx`] as never;
  },
  setup({ el, App, props }) {
    createRoot(el).render(
      <TooltipProvider>
        <App {...props} />
      </TooltipProvider>
    );
  },
});
