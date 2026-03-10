import { useState, useCallback, useRef, useEffect } from 'react';
import { Hash } from 'lucide-react';
import { client } from '../lib/client';

export function BaseTool() {
  const [valDecimal, setValDecimal] = useState('');
  const [valHex, setValHex] = useState('');
  const [valBinary, setValBinary] = useState('');
  const [valOctal, setValOctal] = useState('');
  const [valBase64, setValBase64] = useState('');

  const debounceTimer = useRef<NodeJS.Timeout | null>(null);

  // Only call the API when explicitly typed by the user, not dynamically
  const performConversion = useCallback(async (input: string, sourceBase: number) => {
    input = input.trim();
    if (!input) {
      setValDecimal('');
      setValHex('');
      setValBinary('');
      setValOctal('');
      setValBase64('');
      return;
    }

    try {
      const resp = await client.baseConvert({ input, sourceBase } as Parameters<typeof client.baseConvert>[0]);
      if (resp.error) {
         // Optionally show error, but invalid typing just ignores update
         return;
      }
      setValDecimal(resp.decimal);
      setValHex(resp.hex);
      setValBinary(resp.binary);
      setValOctal(resp.octal);
      setValBase64(resp.base64);
    } catch(err) {
      console.error(err);
    }
  }, []);

  const handleUpdate = (input: string, sourceBase: number, setter: (val: string) => void) => {
    // Optimistic UI update for the typed box
    setter(input);

    if (debounceTimer.current) {
       clearTimeout(debounceTimer.current);
    }

    // Since RPCs are heavier, limit the dispatch to a debounce 
    debounceTimer.current = setTimeout(() => {
       performConversion(input, sourceBase);
    }, 200);
  };
  
  useEffect(() => {
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, []);

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <Hash className="w-6 h-6 text-kawa-500" /> Number Base Converter (Backend)
      </h2>

      <div className="grid md:grid-cols-2 gap-4">
        {/* Decimal */}
        <div className="bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Decimal (Base 10)</label>
          <textarea
            className="w-full min-h-[100px] bg-slate-50 dark:bg-neutral-900 p-3 rounded border border-slate-200 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y"
            value={valDecimal}
            onChange={(e) => handleUpdate(e.target.value, 10, setValDecimal)}
            placeholder="e.g. 255"
          />
        </div>

        {/* Hexadecimal */}
        <div className="bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Hexadecimal (Base 16)</label>
          <textarea
            className="w-full min-h-[100px] bg-slate-50 dark:bg-neutral-900 p-3 rounded border border-slate-200 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y"
            value={valHex}
            onChange={(e) => handleUpdate(e.target.value, 16, setValHex)}
            placeholder="e.g. FF"
          />
        </div>

        {/* Binary */}
        <div className="bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Binary (Base 2)</label>
          <textarea
            className="w-full min-h-[100px] bg-slate-50 dark:bg-neutral-900 p-3 rounded border border-slate-200 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y"
            value={valBinary}
            onChange={(e) => handleUpdate(e.target.value, 2, setValBinary)}
            placeholder="e.g. 11111111"
          />
        </div>

        {/* Octal */}
        <div className="bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Octal (Base 8)</label>
          <textarea
            className="w-full min-h-[100px] bg-slate-50 dark:bg-neutral-900 p-3 rounded border border-slate-200 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y"
            value={valOctal}
            onChange={(e) => handleUpdate(e.target.value, 8, setValOctal)}
            placeholder="e.g. 377"
          />
        </div>

        {/* Base64 */}
        <div className="md:col-span-2 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Base64 (Base 64)</label>
          <textarea
            className="w-full min-h-[100px] bg-slate-50 dark:bg-neutral-900 p-3 rounded border border-slate-200 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 resize-y"
            value={valBase64}
            onChange={(e) => handleUpdate(e.target.value, 64, setValBase64)}
            placeholder="e.g. /w"
          />
        </div>
      </div>
    </div>
  );
}
