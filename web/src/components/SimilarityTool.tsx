import { useState } from 'react';
import { SimilarityResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { GitCompareArrows } from 'lucide-react';

export function SimilarityTool() {
  const [text1, setText1] = useState('');
  const [text2, setText2] = useState('');
  const [res, setRes] = useState<SimilarityResponse | null>(null);

  const compare = async () => {
    try {
      const resp = await client.textSimilarity({ text1, text2 } as Parameters<typeof client.textSimilarity>[0]);
      setRes(resp);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <GitCompareArrows className="w-6 h-6 text-orange-400" />
        Similarity Counter
      </h2>

      <div className="grid gap-4">
        <textarea 
          value={text1}
          onChange={e => setText1(e.target.value)}
          placeholder="First text..."
          className="w-full h-32 bg-white dark:bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
        />
        <textarea 
          value={text2}
          onChange={e => setText2(e.target.value)}
          placeholder="Second text..."
          className="w-full h-32 bg-white dark:bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
        />
      </div>

      <button onClick={compare} className="w-full bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold py-3 rounded-lg transition-colors shadow-lg shadow-kawa-500/20 active:scale-[0.98]">
        Calculate Similarity
      </button>

      {res && !res.error && (
        <div className="grid grid-cols-2 gap-4 bg-white dark:bg-neutral-800 p-6 rounded-lg border border-slate-300 dark:border-neutral-700 text-center shadow-sm">
          <div>
            <div className="text-slate-500 dark:text-gray-400 text-sm uppercase font-bold mb-1">Levenshtein Distance</div>
            <div className="text-3xl font-mono text-slate-900 dark:text-white">{res.distance}</div>
          </div>
          <div>
            <div className="text-slate-600 dark:text-gray-400 text-sm uppercase font-bold mb-1">Similarity Score</div>
            <div className="text-3xl font-mono text-kawa-600 dark:text-kawa-400">{(res.similarity * 100).toFixed(1)}%</div>
          </div>
        </div>
      )}
    </div>
  );
}
