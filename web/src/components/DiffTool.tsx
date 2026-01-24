import { useState } from 'react';
import { DiffRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { cn } from '../lib/utils';
import { ArrowRightLeft } from 'lucide-react';

export function DiffTool() {
  const [text1, setText1] = useState('');
  const [text2, setText2] = useState('');
  const [diffHtml, setDiffHtml] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleDiff = async () => {
    setLoading(true);
    try {
      const request = DiffRequest.create({ text1, text2 });
      const response = await client.diff(request as any);
      setDiffHtml(response.diffHtml);
    } catch (error) {
      console.error('Error fetching diff:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-white">Diff Viewer</h2>
        <button
          onClick={handleDiff}
          disabled={loading}
          className={cn(
            "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
            "bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white"
          )}
        >
          <ArrowRightLeft className="w-4 h-4" />
          {loading ? 'Comparing...' : 'Compare'}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-400">Original Text</label>
          <textarea
            className="w-full h-64 bg-gray-800 p-4 rounded-lg border border-gray-700 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
            value={text1}
            onChange={(e) => setText1(e.target.value)}
            placeholder="Paste original text here..."
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-400">Modified Text</label>
          <textarea
             className="w-full h-64 bg-gray-800 p-4 rounded-lg border border-gray-700 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
            value={text2}
            onChange={(e) => setText2(e.target.value)}
            placeholder="Paste modified text here..."
          />
        </div>
      </div>

      {diffHtml && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-400">Result</label>
          <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
            <div 
              data-testid="diff-output"
              className="p-6 font-mono text-sm overflow-auto"
              dangerouslySetInnerHTML={{ __html: diffHtml }} 
            />
          </div>
        </div>
      )}
    </div>
  );
}
