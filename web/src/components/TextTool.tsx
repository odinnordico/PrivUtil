import { useState, useEffect } from 'react';
import { TextAction } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { type LucideIcon, 
  ArrowDownAZ, 
  ArrowUpAZ, 
  ArrowDownUp, 
  Scissors, 
  Eraser, 
  FileText 
} from 'lucide-react';

export function TextTool() {
  const [input, setInput] = useState('');
  const [stats, setStats] = useState({ chars: 0, words: 0, lines: 0, bytes: 0 });

  useEffect(() => {
    // Debounce stats calculation
    const timer = setTimeout(async () => {
      try {
        const resp = await client.textInspect({ text: input } as Parameters<typeof client.textInspect>[0]);
        setStats({
          chars: resp.charCount,
          words: resp.wordCount,
          lines: resp.lineCount,
          bytes: resp.byteCount
        });
      } catch (e) { console.error(e); }
    }, 300);
    return () => clearTimeout(timer);
  }, [input]);

  const manipulate = async (action: TextAction) => {
    try {
      const resp = await client.textManipulate({ text: input, action } as Parameters<typeof client.textManipulate>[0]);
      setInput(resp.text);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white">Text Tools</h2>

      <div className="flex flex-wrap gap-2 items-center bg-gray-800/50 p-4 rounded-lg">
        <ActionButton onClick={() => manipulate(TextAction.SORT_AZ)} icon={ArrowDownAZ} label="Sort A-Z" />
        <ActionButton onClick={() => manipulate(TextAction.SORT_ZA)} icon={ArrowUpAZ} label="Sort Z-A" />
        <ActionButton onClick={() => manipulate(TextAction.REVERSE)} icon={ArrowDownUp} label="Reverse Lines" />
        <div className="w-px h-6 bg-gray-600 mx-2" />
        <ActionButton onClick={() => manipulate(TextAction.DEDUPE)} icon={Scissors} label="Dedupe" />
        <ActionButton onClick={() => manipulate(TextAction.TRIM)} icon={Eraser} label="Trim Lines" />
        <ActionButton onClick={() => manipulate(TextAction.REMOVE_EMPTY)} icon={FileText} label="Remove Empty" />
      </div>

      <div className="grid gap-2">
        <div className="flex gap-4 text-xs text-gray-400 font-mono">
          <span>Chars: {stats.chars}</span>
          <span>Words: {stats.words}</span>
          <span>Lines: {stats.lines}</span>
          <span>Bytes: {stats.bytes}</span>
        </div>
        <textarea 
          className="w-full h-[600px] bg-gray-800 p-4 rounded-lg border border-gray-700 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Paste text here to inspect or manipulate..."
        />
      </div>
    </div>
  );
}

function ActionButton({ onClick, icon: Icon, label }: { onClick: () => void, icon: LucideIcon, label: string }) {
  return (
    <button 
      onClick={onClick}
      className="flex items-center gap-2 px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded text-sm font-medium text-gray-200 transition-colors"
    >
      <Icon className="w-4 h-4" />
      {label}
    </button>
  );
}
