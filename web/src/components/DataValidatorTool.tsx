import { useState, useCallback, useRef, useEffect } from 'react';
import { DataFormat } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { CheckCircle2, XCircle, AlertCircle } from 'lucide-react';
import { cn } from '../lib/utils';

const FORMATS = [
  { value: DataFormat.JSON, label: 'JSON' },
  { value: DataFormat.YAML, label: 'YAML' },
  { value: DataFormat.XML,  label: 'XML' },
  { value: DataFormat.TOML, label: 'TOML' },
] as const;

const PLACEHOLDERS: Record<number, string> = {
  [DataFormat.JSON]: '{\n  "name": "Alice",\n  "age": 30\n}',
  [DataFormat.YAML]: 'name: Alice\nage: 30',
  [DataFormat.XML]:  '<person>\n  <name>Alice</name>\n  <age>30</age>\n</person>',
  [DataFormat.TOML]: 'name = "Alice"\nage = 30',
};

type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

export function DataValidatorTool() {
  const [input, setInput]         = useState('');
  const [format, setFormat]       = useState<DataFormat>(DataFormat.JSON);
  const [state, setState]         = useState<ValidationState>('idle');
  const [errorMsg, setErrorMsg]   = useState('');
  const [line, setLine]           = useState<number | null>(null);
  const [column, setColumn]       = useState<number | null>(null);
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);

  const validate = useCallback(async (data: string, fmt: DataFormat) => {
    if (!data.trim()) { setState('idle'); setErrorMsg(''); setLine(null); setColumn(null); return; }
    setState('validating');
    try {
      const resp = await client.validateData({ data, format: fmt } as Parameters<typeof client.validateData>[0]);
      if (resp.valid) {
        setState('valid');
        setErrorMsg('');
        setLine(null);
        setColumn(null);
      } else {
        setState('invalid');
        setErrorMsg(resp.error ?? 'Invalid input');
        setLine(resp.line > 0 ? resp.line : null);
        setColumn(resp.column > 0 ? resp.column : null);
      }
    } catch {
      setState('invalid');
      setErrorMsg('Validation failed — is the server running?');
      setLine(null);
      setColumn(null);
    }
  }, []);

  const schedule = useCallback((data: string, fmt: DataFormat) => {
    if (debounce.current) clearTimeout(debounce.current);
    debounce.current = setTimeout(() => validate(data, fmt), 300);
  }, [validate]);

  useEffect(() => () => { if (debounce.current) clearTimeout(debounce.current); }, []);

  const handleInput = (v: string) => {
    setInput(v);
    schedule(v, format);
  };

  const handleFormat = (v: DataFormat) => {
    setFormat(v);
    schedule(input, v);
  };

  const borderClass = {
    idle:       'border-slate-300 dark:border-neutral-700',
    validating: 'border-slate-300 dark:border-neutral-700',
    valid:      'border-emerald-400 dark:border-emerald-500',
    invalid:    'border-red-400 dark:border-red-500',
  }[state];

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Data Validator</h2>

      {/* Controls */}
      <div className="flex flex-wrap gap-3 items-center bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <span className="text-sm font-bold text-slate-600 dark:text-slate-400">Format</span>
        <div className="flex gap-2">
          {FORMATS.map(f => (
            <button
              key={f.value}
              onClick={() => handleFormat(f.value)}
              className={cn(
                'px-3 py-1.5 rounded text-sm font-medium transition-colors',
                format === f.value
                  ? 'bg-kawa-500 text-slate-900'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-kawa-500/20'
              )}
            >
              {f.label}
            </button>
          ))}
        </div>

        <div className="ml-auto flex items-center gap-2">
          {state === 'validating' && (
            <span className="text-xs text-slate-400 animate-pulse">Validating…</span>
          )}
          {state === 'valid' && (
            <span className="flex items-center gap-1.5 text-sm font-semibold text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="w-4 h-4" /> Valid
            </span>
          )}
          {state === 'invalid' && (
            <span className="flex items-center gap-1.5 text-sm font-semibold text-red-600 dark:text-red-400">
              <XCircle className="w-4 h-4" /> Invalid
            </span>
          )}
          {state === 'idle' && (
            <span className="flex items-center gap-1.5 text-sm text-slate-400 dark:text-slate-500">
              <AlertCircle className="w-4 h-4" /> Enter input to validate
            </span>
          )}
        </div>
      </div>

      {/* Editor */}
      <div className="space-y-2">
        <textarea
          className={cn(
            'w-full h-[480px] bg-white dark:bg-neutral-800 p-4 rounded-lg border-2 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y shadow-sm transition-colors duration-200',
            borderClass
          )}
          value={input}
          onChange={e => handleInput(e.target.value)}
          placeholder={PLACEHOLDERS[format]}
          spellCheck={false}
        />

        {/* Error detail */}
        {state === 'invalid' && errorMsg && (
          <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 text-red-700 dark:text-red-300 text-sm font-mono">
            <XCircle className="w-4 h-4 mt-0.5 shrink-0" />
            <div className="space-y-0.5">
              <div>{errorMsg}</div>
              {(line != null || column != null) && (
                <div className="text-xs text-red-500 dark:text-red-400">
                  {line != null && `Line ${line}`}{line != null && column != null && ', '}{column != null && `Column ${column}`}
                </div>
              )}
            </div>
          </div>
        )}

        {state === 'valid' && (
          <div className="flex items-center gap-2 p-3 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-500/30 text-emerald-700 dark:text-emerald-300 text-sm">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            Input is valid {FORMATS.find(f => f.value === format)?.label}
          </div>
        )}
      </div>
    </div>
  );
}
