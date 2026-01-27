import { useState } from 'react';
import { client } from '../lib/client';
import { AlignLeft, Minimize2, Check } from 'lucide-react';

export function JsonTool() {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [indent, setIndent] = useState('2');
  const [error, setError] = useState<string | null>(null);

  const handleFormat = async () => {
    setError(null);
    try {
      const response = await client.jsonFormat({ 
        text: input, 
        indent: indent, 
        sortKeys: true 
      } as Parameters<typeof client.jsonFormat>[0]);
      if (response.error) {
        setError(response.error);
      } else {
        setOutput(response.text);
      }
    } catch (err) {
      console.error(err);
      setError('Format failed');
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">JSON Formatter</h2>

      <div className="flex gap-4 items-center bg-white dark:bg-neutral-800/50 p-2 rounded-lg w-fit border border-slate-300 dark:border-neutral-700 shadow-sm">
        <select 
          value={indent}
          onChange={(e) => setIndent(e.target.value)}
          className="bg-slate-50 dark:bg-neutral-700 text-slate-900 dark:text-white rounded px-3 py-1 text-sm border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500"
        >
          <option value="2">2 Spaces</option>
          <option value="4">4 Spaces</option>
          <option value="tab">Tab</option>
          <option value="min">Minify</option>
        </select>

        <button
          onClick={handleFormat}
          className="flex items-center gap-2 px-4 py-1.5 bg-kawa-500 hover:bg-kawa-600 rounded text-sm font-medium text-slate-900 transition-colors"
        >
          {indent === 'min' ? <Minimize2 className="w-4 h-4"/> : <AlignLeft className="w-4 h-4"/>}
          {indent === 'min' ? 'Minify' : 'Format'}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Input JSON</label>
          <textarea
            className="w-full h-[500px] bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder='{"key": "value"}'
          />
        </div>
        
        <div className="space-y-2">
           <label className="text-sm font-bold text-slate-600 dark:text-slate-400">
             Output 
             {output && !error && <span className="ml-2 text-green-400 text-xs flex items-center inline-flex gap-1"><Check className="w-3 h-3"/> Valid</span>}
           </label>
           {error ? (
              <div className="w-full h-[500px] bg-red-900/20 p-4 rounded-lg border border-red-500/50 text-red-400 font-mono text-sm">
                {error}
              </div>
           ) : (
             <textarea
               readOnly
               className="w-full h-[500px] bg-slate-50 dark:bg-black/30 p-4 rounded-lg border border-slate-300 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none shadow-inner"
               value={output}
             />
           )}
        </div>
      </div>
    </div>
  );
}
