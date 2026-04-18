import Editor, { useMonaco } from '@monaco-editor/react';
import { useEffect, useRef, useState } from 'react';
import templateSchema from '@/schemas/template.v1.json';

const SCHEMA_URI = 'https://github.com/markusheinemann/scfleet/api/schemas/template.v1.json';
const MODEL_PATH = 'inmemory://scfleet/extraction-schema.json';

function isDark() {
  return typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
}

function useMonacoTheme() {
  const [theme, setTheme] = useState<'vs' | 'vs-dark'>('vs');

  useEffect(() => {
    setTheme(isDark() ? 'vs-dark' : 'vs');
    const observer = new MutationObserver(() => {
      setTheme(isDark() ? 'vs-dark' : 'vs');
    });
    observer.observe(document.documentElement, { attributeFilter: ['class'] });
    return () => observer.disconnect();
  }, []);

  return theme;
}

type Props = {
  value: string;
  onChange: (value: string) => void;
  invalid?: boolean;
};

export default function SchemaEditor({ value, onChange, invalid }: Props) {
  const monaco = useMonaco();
  const hiddenRef = useRef<HTMLInputElement>(null);
  const editorTheme = useMonacoTheme();

  useEffect(() => {
    if (!monaco) return;
    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
      validate: true,
      schemaValidation: 'error',
      schemas: [{ uri: SCHEMA_URI, fileMatch: [MODEL_PATH], schema: templateSchema }],
    });
  }, [monaco]);

  function handleChange(v: string | undefined) {
    const next = v ?? '';
    onChange(next);
    if (hiddenRef.current) hiddenRef.current.value = next;
  }

  return (
    <>
      <input type="hidden" name="schema" ref={hiddenRef} defaultValue={value} />
      <div
        className={`overflow-hidden rounded-md border ${invalid ? 'border-destructive' : 'border-input'}`}
      >
        <Editor
          height="320px"
          defaultLanguage="json"
          path={MODEL_PATH}
          theme={editorTheme}
          value={value}
          onChange={handleChange}
          options={{
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            fontSize: 13,
            tabSize: 2,
            formatOnPaste: true,
            formatOnType: true,
            lineNumbers: 'off',
            folding: true,
            wordWrap: 'on',
            quickSuggestions: true,
            suggestOnTriggerCharacters: true,
          }}
        />
      </div>
    </>
  );
}
