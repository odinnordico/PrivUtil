import { useState } from 'react';
import { client } from '../lib/client';
import { cn } from '../lib/utils';
import { ArrowDownUp } from 'lucide-react';

export function Base64Tool() {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleAction = async (action: 'encode' | 'decode') => {
    setLoading(true);
    setError(null);
    try {
      const response = action === 'encode' 
        ? await client.base64Encode({ text: input } as Parameters<typeof client.base64Encode>[0])
        : await client.base64Decode({ text: input } as Parameters<typeof client.base64Decode>[0]);
      
      if (response.error) {
        setError(response.error);
        setOutput('');
      } else {
        setOutput(response.text);
      }
    } catch (err) {
      console.error(err);
      setError('An unexpected error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Base64 Encoder/Decoder</h2>

      <div className="space-y-2">
        <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Input</label>
        <textarea
          className="w-full h-40 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Enter text to encode or decode..."
        />
      </div>

      <div className="flex gap-4">
        <button
          onClick={() => handleAction('encode')}
          disabled={loading || !input}
          className={cn(
            "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors border border-transparent",
            "bg-kawa-500 hover:bg-kawa-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-900"
          )}
        >
          <ArrowDownUp className="w-4 h-4" />
          Encode
        </button>
        <button
          onClick={() => handleAction('decode')}
          disabled={loading || !input}
          className={cn(
            "flex items-center gap-2 px-6 py-2 rounded-lg font-bold transition-colors border border-slate-300 dark:border-transparent",
            "bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-700 dark:text-white"
          )}
        >
          <ArrowDownUp className="w-4 h-4" />
          Decode
        </button>
      </div>

      {error ? (
        <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500">
          {error}
        </div>
      ) : output && (
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Output</label>
          <textarea
            readOnly
            className="w-full h-40 bg-slate-50 dark:bg-black/30 p-4 rounded-lg border border-slate-300 dark:border-gray-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none shadow-inner"
            value={output}
          />
        </div>
      )}
    </div>
  );
}
