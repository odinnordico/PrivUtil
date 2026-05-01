import { useState, useCallback, useRef, useEffect } from 'react';
import { DataFormat } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { ArrowLeftRight, Copy, Check } from 'lucide-react';
import { cn } from '../lib/utils';

const FORMATS = [
  { value: DataFormat.JSON,  label: 'JSON' },
  { value: DataFormat.YAML,  label: 'YAML' },
  { value: DataFormat.XML,   label: 'XML' },
  { value: DataFormat.TOML,  label: 'TOML' },
  { value: DataFormat.CSV,   label: 'CSV' },
] as const;

const DELIMITERS = [
  { value: ',',  label: 'Comma (,)' },
  { value: ';',  label: 'Semicolon (;)' },
  { value: '\\t', label: 'Tab (TSV)' },
  { value: '|',  label: 'Pipe (|)' },
] as const;

const PLACEHOLDERS: Record<number, string> = {
  [DataFormat.JSON]: '{\n  "name": "Alice",\n  "age": 30\n}',
  [DataFormat.YAML]: 'name: Alice\nage: 30',
  [DataFormat.XML]:  '<person>\n  <name>Alice</name>\n  <age>30</age>\n</person>',
  [DataFormat.TOML]: 'name = "Alice"\nage = 30',
  [DataFormat.CSV]:  'name,age\nAlice,30\nBob,25',
};

function formatLabel(fmt: DataFormat): string {
  return FORMATS.find(f => f.value === fmt)?.label ?? 'Unknown';
}

export function ConverterTool() {
  const [input, setInput]           = useState('');
  const [output, setOutput]         = useState('');
  const [source, setSource]         = useState<DataFormat>(DataFormat.JSON);
  const [target, setTarget]         = useState<DataFormat>(DataFormat.YAML);
  const [delimiter, setDelimiter]   = useState(',');
  const [noHeader, setNoHeader]     = useState(false);
  const [error, setError]           = useState<string | null>(null);
  const [converting, setConverting] = useState(false);
  const [copied, setCopied]         = useState(false);
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);

  const csvActive = source === DataFormat.CSV || target === DataFormat.CSV;

  const convert = useCallback(async (
    data: string,
    src: DataFormat,
    tgt: DataFormat,
    delim: string,
    noHdr: boolean,
  ) => {
    if (!data.trim()) { setOutput(''); setError(null); return; }
    setConverting(true);
    try {
      const resp = await client.convert({
        data,
        sourceFormat: src,
        targetFormat: tgt,
        csvDelimiter: delim,
        csvNoHeader: noHdr,
      } as Parameters<typeof client.convert>[0]);
      if (resp.error) {
        setError(resp.error);
        setOutput('');
      } else {
        setError(null);
        setOutput(resp.data);
      }
    } catch {
      setError('Conversion failed — is the server running?');
    } finally {
      setConverting(false);
    }
  }, []);

  const scheduleConvert = useCallback((
    data: string,
    src: DataFormat,
    tgt: DataFormat,
    delim: string,
    noHdr: boolean,
  ) => {
    if (debounce.current) clearTimeout(debounce.current);
    debounce.current = setTimeout(() => convert(data, src, tgt, delim, noHdr), 300);
  }, [convert]);

  useEffect(() => () => { if (debounce.current) clearTimeout(debounce.current); }, []);

  const handleInput = (v: string) => {
    setInput(v);
    scheduleConvert(v, source, target, delimiter, noHeader);
  };

  const handleSource = (v: DataFormat) => {
    setSource(v);
    scheduleConvert(input, v, target, delimiter, noHeader);
  };

  const handleTarget = (v: DataFormat) => {
    setTarget(v);
    scheduleConvert(input, source, v, delimiter, noHeader);
  };

  const handleDelimiter = (v: string) => {
    setDelimiter(v);
    scheduleConvert(input, source, target, v, noHeader);
  };

  const handleNoHeader = (v: boolean) => {
    setNoHeader(v);
    scheduleConvert(input, source, target, delimiter, v);
  };

  const swap = () => {
    setSource(target);
    setTarget(source);
    setInput(output);
    setOutput('');
    setError(null);
    scheduleConvert(output, target, source, delimiter, noHeader);
  };

  const copy = async () => {
    if (!output) return;
    await navigator.clipboard.writeText(output);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const selectClass = "bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 text-sm";
  const textareaClass = "w-full h-[480px] bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-none shadow-sm";

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Universal Converter</h2>

      {/* Controls row */}
      <div className="flex flex-wrap gap-3 items-center bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex items-center gap-2">
          <span className="text-sm font-bold text-slate-600 dark:text-slate-400">From</span>
          <select value={source} onChange={e => handleSource(parseInt(e.target.value) as DataFormat)} className={selectClass}>
            {FORMATS.map(f => <option key={f.value} value={f.value}>{f.label}</option>)}
          </select>
        </div>

        <button
          onClick={swap}
          title="Swap formats"
          className="p-2 rounded-lg bg-slate-100 dark:bg-slate-700 hover:bg-kawa-500/10 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
        >
          <ArrowLeftRight className="w-4 h-4" />
        </button>

        <div className="flex items-center gap-2">
          <span className="text-sm font-bold text-slate-600 dark:text-slate-400">To</span>
          <select value={target} onChange={e => handleTarget(parseInt(e.target.value) as DataFormat)} className={selectClass}>
            {FORMATS.map(f => <option key={f.value} value={f.value}>{f.label}</option>)}
          </select>
        </div>

        {converting && (
          <span className="text-xs text-slate-400 dark:text-slate-500 ml-2 animate-pulse">Converting…</span>
        )}

        {/* CSV options */}
        {csvActive && (
          <div className="flex items-center gap-3 ml-auto flex-wrap">
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold text-slate-500 dark:text-slate-400">Delimiter</span>
              <select
                value={delimiter}
                onChange={e => handleDelimiter(e.target.value)}
                className={cn(selectClass, 'py-1 text-xs')}
              >
                {DELIMITERS.map(d => <option key={d.value} value={d.value}>{d.label}</option>)}
              </select>
            </div>
            {source === DataFormat.CSV && (
              <label className="flex items-center gap-1.5 text-xs text-slate-600 dark:text-slate-400 cursor-pointer select-none">
                <input
                  type="checkbox"
                  checked={noHeader}
                  onChange={e => handleNoHeader(e.target.checked)}
                  className="accent-kawa-500"
                />
                No header row
              </label>
            )}
          </div>
        )}
      </div>

      {/* Panels */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
            {formatLabel(source)} — Input
          </label>
          <textarea
            className={textareaClass}
            value={input}
            onChange={e => handleInput(e.target.value)}
            placeholder={PLACEHOLDERS[source]}
            spellCheck={false}
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
              {formatLabel(target)} — Output
            </label>
            <button
              onClick={copy}
              disabled={!output}
              className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 disabled:opacity-30 transition-colors"
            >
              {copied
                ? <><Check className="w-3.5 h-3.5" /> Copied</>
                : <><Copy className="w-3.5 h-3.5" /> Copy</>
              }
            </button>
          </div>
          {error ? (
            <div className="w-full h-[480px] bg-red-50 dark:bg-red-900/20 p-4 rounded-lg border border-red-300 dark:border-red-500/40 text-red-600 dark:text-red-400 font-mono text-sm overflow-auto">
              {error}
            </div>
          ) : (
            <textarea
              readOnly
              className={cn(textareaClass, 'bg-slate-50 dark:bg-black/30 focus:ring-0 cursor-default')}
              value={output}
              placeholder="Output appears here automatically…"
              spellCheck={false}
            />
          )}
        </div>
      </div>
    </div>
  );
}
