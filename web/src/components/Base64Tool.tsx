import { useState } from 'react';
import { Base64Request } from '../proto/proto/privutil';
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
      const request = Base64Request.create({ text: input });
      const response = action === 'encode' 
        ? await client.base64Encode(request as any)
        : await client.base64Decode(request as any);
      
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
      <h2 className="text-2xl font-bold text-white">Base64 Encoder/Decoder</h2>

      <div className="space-y-2">
        <label className="text-sm font-medium text-gray-400">Input</label>
        <textarea
          className="w-full h-40 bg-gray-800 p-4 rounded-lg border border-gray-700 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
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
            "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
            "bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white"
          )}
        >
          <ArrowDownUp className="w-4 h-4" />
          Encode
        </button>
        <button
          onClick={() => handleAction('decode')}
          disabled={loading || !input}
          className={cn(
            "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
            "bg-gray-700 hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed text-white"
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
          <label className="text-sm font-medium text-gray-400">Output</label>
          <textarea
            readOnly
            className="w-full h-40 bg-black/30 p-4 rounded-lg border border-gray-800 text-gray-100 font-mono text-sm focus:outline-none"
            value={output}
          />
        </div>
      )}
    </div>
  );
}
