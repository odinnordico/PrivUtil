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
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <GitCompareArrows className="w-6 h-6 text-orange-400" />
        Similarity Counter
      </h2>

      <div className="grid gap-4">
        <textarea 
          value={text1}
          onChange={e => setText1(e.target.value)}
          placeholder="First text..."
          className="w-full h-32 bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-orange-500"
        />
        <textarea 
          value={text2}
          onChange={e => setText2(e.target.value)}
          placeholder="Second text..."
          className="w-full h-32 bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-orange-500"
        />
      </div>

      <button onClick={compare} className="w-full bg-orange-600 hover:bg-orange-700 text-white font-bold py-3 rounded-lg transition-colors">
        Calculate Similarity
      </button>

      {res && !res.error && (
        <div className="grid grid-cols-2 gap-4 bg-gray-800 p-6 rounded-lg border border-gray-700 text-center">
          <div>
            <div className="text-gray-400 text-sm uppercase font-bold mb-1">Levenshtein Distance</div>
            <div className="text-3xl font-mono text-white">{res.distance}</div>
          </div>
          <div>
            <div className="text-gray-400 text-sm uppercase font-bold mb-1">Similarity Score</div>
            <div className="text-3xl font-mono text-green-400">{(res.similarity * 100).toFixed(1)}%</div>
          </div>
        </div>
      )}
    </div>
  );
}
