import { useState, useEffect } from 'react';
import { TimeResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Clock } from 'lucide-react';

export function TimeTool() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState<TimeResponse | null>(null);

  const convert = async (val: string) => {
    try {
      const resp = await client.timeConvert({ input: val } as Parameters<typeof client.timeConvert>[0]);
      setResult(resp);
    } catch (e) { console.error(e); }
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      void convert('now');
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <Clock className="w-6 h-6 text-blue-400" /> 
        Time Converter
      </h2>

      <div className="flex gap-2">
        <input 
          type="text" 
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Unix timestamp or ISO date string (leave empty for Now)..."
          className="flex-1 bg-white dark:bg-gray-800 text-slate-900 dark:text-white px-4 py-3 rounded-lg border border-slate-300 dark:border-gray-700 focus:ring-2 focus:ring-kawa-500 focus:outline-none font-mono shadow-inner"
        />
        <button 
          onClick={() => convert(input)}
          className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-6 py-2 rounded-lg font-bold transition-all shadow-md active:scale-95"
        >
          Convert
        </button>
        <button 
          onClick={() => { setInput('now'); void convert('now'); }}
          className="bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 text-slate-700 dark:text-white px-4 py-2 rounded-lg font-bold transition-colors border border-slate-300 dark:border-transparent"
        >
          Now
        </button>
      </div>

      {result && (
        <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden border border-slate-300 dark:border-gray-700 shadow-sm">
          <ResultRow label="Unix Timestamp" value={result.unix.toString()} />
          <ResultRow label="ISO 8601" value={result.iso} copy />
          <ResultRow label="UTC" value={result.utc} copy />
          <ResultRow label="Local" value={result.local} copy />
        </div>
      )}
    </div>
  );
}

function ResultRow({ label, value, copy }: { label: string, value: string, copy?: boolean }) {
  return (
    <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-700 last:border-0 hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors">
      <span className="text-slate-600 dark:text-gray-400 font-bold w-32">{label}</span>
      <code className="flex-1 font-mono text-kawa-700 dark:text-kawa-300 truncate px-2">{value}</code>
      {copy && (
        <button 
          onClick={() => navigator.clipboard.writeText(value)}
          className="text-xs bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 text-slate-600 dark:text-white px-2 py-1 rounded ml-2 border border-slate-300 dark:border-transparent font-bold transition-colors"
        >
          Copy
        </button>
      )}
    </div>
  );
}
