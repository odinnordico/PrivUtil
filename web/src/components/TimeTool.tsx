import { useState, useEffect } from 'react';
import { TimeRequest, TimeResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Clock } from 'lucide-react';

export function TimeTool() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState<TimeResponse | null>(null);

  const convert = async (val: string) => {
    try {
      const resp = await client.timeConvert(TimeRequest.create({ input: val }) as any);
      setResult(resp);
    } catch (e) { console.error(e); }
  };

  useEffect(() => {
    convert('now');
  }, []);

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <Clock className="w-6 h-6 text-blue-400" /> 
        Time Converter
      </h2>

      <div className="flex gap-2">
        <input 
          type="text" 
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Unix timestamp or ISO date string (leave empty for Now)..."
          className="flex-1 bg-gray-800 text-white px-4 py-3 rounded-lg border border-gray-700 focus:ring-2 focus:ring-blue-500 focus:outline-none"
        />
        <button 
          onClick={() => convert(input)}
          className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium"
        >
          Convert
        </button>
        <button 
          onClick={() => { setInput('now'); convert('now'); }}
          className="bg-gray-700 hover:bg-gray-600 text-white px-4 py-2 rounded-lg font-medium"
        >
          Now
        </button>
      </div>

      {result && (
        <div className="bg-gray-800 rounded-lg overflow-hidden border border-gray-700">
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
    <div className="flex items-center justify-between p-4 border-b border-gray-700 last:border-0 hover:bg-gray-700/30 transition-colors">
      <span className="text-gray-400 font-medium w-32">{label}</span>
      <code className="flex-1 font-mono text-blue-300 truncate px-2">{value}</code>
      {copy && (
        <button 
          onClick={() => navigator.clipboard.writeText(value)}
          className="text-xs bg-gray-700 hover:bg-gray-600 text-white px-2 py-1 rounded ml-2"
        >
          Copy
        </button>
      )}
    </div>
  );
}
