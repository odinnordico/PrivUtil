import { useState, useEffect } from 'react';
import { TimeResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Clock } from 'lucide-react';

export function TimeTool() {
  const [result, setResult] = useState<TimeResponse | null>(null);
  const [error, setError] = useState('');

  const convert = async (val: string) => {
    setError('');
    try {
      const resp = await client.timeConvert({ input: val } as Parameters<typeof client.timeConvert>[0]);
      if (!resp || resp.iso === 'Invalid input format') {
        setError('Invalid input format');
      } else {
        setResult(resp);
      }
    } catch (e) { console.error(e); }
  };

  useEffect(() => {
    void convert('now');
  }, []);

  const rows: { label: string; value: string }[] = result
    ? [
        { label: 'Unix (sec)', value: result.unix.toString() },
        { label: 'Unix (ms)', value: (result.unix * 1000).toString() },
        { label: 'ISO 8601', value: result.iso },
        { label: 'UTC', value: result.utc },
        { label: 'Local', value: result.local },
      ]
    : [];

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <Clock className="w-6 h-6 text-blue-400" />
          Time Converter
        </h2>
        <button
          onClick={() => void convert('now')}
          className="bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 text-slate-700 dark:text-white px-4 py-2 rounded-lg font-bold transition-colors border border-slate-300 dark:border-transparent"
        >
          Now
        </button>
      </div>

      {error && <p className="text-red-500 text-sm font-medium">{error}</p>}

      <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden border border-slate-300 dark:border-gray-700 shadow-sm">
        <p className="px-4 pt-3 pb-1 text-xs text-slate-500 dark:text-gray-400">
          Edit any field and press <kbd className="font-mono bg-slate-100 dark:bg-gray-700 px-1 rounded">Enter</kbd> or click <strong>Use</strong> to convert
        </p>
        {rows.map(({ label, value }) => (
          <EditableRow key={label} label={label} value={value} onConvert={convert} />
        ))}
        {rows.length === 0 && (
          <p className="p-4 text-slate-500 dark:text-gray-400 text-center text-sm">Loading…</p>
        )}
      </div>
    </div>
  );
}

function EditableRow({ label, value, onConvert }: { label: string; value: string; onConvert: (val: string) => void }) {
  const [localVal, setLocalVal] = useState(value);

  useEffect(() => {
    setLocalVal(value);
  }, [value]);

  return (
    <div className="flex items-center p-3 border-b border-gray-100 dark:border-gray-700 last:border-0 hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors gap-3">
      <span className="text-slate-600 dark:text-gray-400 font-bold w-28 shrink-0 text-sm">{label}</span>
      <input
        type="text"
        value={localVal}
        onChange={e => setLocalVal(e.target.value)}
        onKeyDown={e => { if (e.key === 'Enter') onConvert(localVal); }}
        className="flex-1 bg-transparent text-kawa-700 dark:text-kawa-300 font-mono focus:outline-none focus:ring-1 focus:ring-kawa-500 rounded px-1 py-0.5 min-w-0 text-sm"
        aria-label={label}
      />
      <button
        onClick={() => navigator.clipboard.writeText(value)}
        className="text-xs bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 text-slate-600 dark:text-white px-2 py-1 rounded border border-slate-300 dark:border-transparent font-bold transition-colors shrink-0"
      >
        Copy
      </button>
      <button
        onClick={() => onConvert(localVal)}
        className="text-xs bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-2 py-1 rounded font-bold transition-colors shrink-0"
      >
        Use
      </button>
    </div>
  );
}
