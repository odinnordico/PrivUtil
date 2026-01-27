import { useState } from 'react';
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
      const response = await client.diff({ text1, text2 } as Parameters<typeof client.diff>[0]);
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
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Diff Viewer</h2>
        <button
          onClick={handleDiff}
          disabled={loading}
          className={cn(
            "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
            "bg-kawa-500 hover:bg-kawa-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-900"
          )}
        >
          <ArrowRightLeft className="w-4 h-4" />
          {loading ? 'Comparing...' : 'Compare'}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Original Text</label>
          <textarea
            className="w-full h-64 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-gray-200 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50"
            value={text1}
            onChange={(e) => setText1(e.target.value)}
            placeholder="Paste original text here..."
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Modified Text</label>
          <textarea
             className="w-full h-64 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
            value={text2}
            onChange={(e) => setText2(e.target.value)}
            placeholder="Paste modified text here..."
          />
        </div>
      </div>

      {diffHtml && (
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Result</label>
          <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 overflow-hidden text-slate-900 dark:text-neutral-100 shadow-sm">
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
