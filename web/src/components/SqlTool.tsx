import { useState } from 'react';
import { client } from '../lib/client';
import { Database } from 'lucide-react';

export function SqlTool() {
  const [query, setQuery] = useState('');
  const [formatted, setFormatted] = useState('');

  const format = async () => {
    try {
      const resp = await client.sqlFormat({ query } as Parameters<typeof client.sqlFormat>[0]);
      setFormatted(resp.formatted);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <Database className="w-6 h-6 text-blue-300" />
        SQL Formatter
      </h2>

      <div className="grid lg:grid-cols-2 gap-4">
        <div className="flex flex-col gap-2">
          <label className="text-slate-600 dark:text-gray-400 text-sm font-bold">Input Query</label>
          <textarea 
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="SELECT * FROM table"
            className="flex-1 min-h-[400px] bg-white dark:bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
          />
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-slate-600 dark:text-gray-400 text-sm font-bold">Formatted Output</label>
          <textarea 
            readOnly
            value={formatted}
            placeholder="Formatted SQL..."
            className="flex-1 min-h-[400px] bg-slate-50 dark:bg-gray-900 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 text-kawa-700 dark:text-kawa-300"
          />
        </div>
      </div>

      <button onClick={format} className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-8 py-3 rounded-lg font-bold transition-all shadow-lg shadow-kawa-500/20 active:scale-95">
        Format SQL
      </button>
    </div>
  );
}
