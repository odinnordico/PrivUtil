import { useState } from 'react';
import { SqlRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Database } from 'lucide-react';

export function SqlTool() {
  const [query, setQuery] = useState('');
  const [formatted, setFormatted] = useState('');

  const format = async () => {
    try {
      const resp = await client.sqlFormat(SqlRequest.create({ query }) as any);
      setFormatted(resp.formatted);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <Database className="w-6 h-6 text-blue-300" />
        SQL Formatter
      </h2>

      <div className="grid lg:grid-cols-2 gap-4">
        <div className="flex flex-col gap-2">
          <label className="text-gray-400 text-sm">Input Query</label>
          <textarea 
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="SELECT * FROM table"
            className="flex-1 min-h-[400px] bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-blue-500"
          />
        </div>
        <div className="flex flex-col gap-2">
          <label className="text-gray-400 text-sm">Formatted Output</label>
          <textarea 
            readOnly
            value={formatted}
            placeholder="Formatted SQL..."
            className="flex-1 min-h-[400px] bg-gray-900 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-gray-700 text-blue-300"
          />
        </div>
      </div>

      <button onClick={format} className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 rounded-lg font-medium">
        Format SQL
      </button>
    </div>
  );
}
