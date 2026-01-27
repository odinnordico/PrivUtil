import { useState } from 'react';
import { client } from '../lib/client';
import { Link, Code } from 'lucide-react';

export function EncoderTool() {
  const [activeTab, setActiveTab] = useState<'url' | 'html'>('url');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Encoders / Decoders</h2>

      <div className="flex gap-4 border-b border-slate-300 dark:border-gray-700">
        <button
          onClick={() => setActiveTab('url')}
          className={`pb-2 px-4 font-bold transition-colors ${activeTab === 'url' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          URL
        </button>
        <button
          onClick={() => setActiveTab('html')}
          className={`pb-2 px-4 font-bold transition-colors ${activeTab === 'html' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          HTML
        </button>
      </div>

      <div className="p-6 bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <EncoderProcessor type={activeTab} />
      </div>
    </div>
  );
}

function EncoderProcessor({ type }: { type: 'url' | 'html' }) {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');

  const handleAction = async (action: 'encode' | 'decode') => {
    try {
      const req = { text: input } as Parameters<typeof client.urlEncode>[0];
      let resp;
      if (type === 'url') {
        resp = action === 'encode' 
          ? await client.urlEncode(req) 
          : await client.urlDecode(req);
      } else {
        resp = action === 'encode'
          ? await client.htmlEncode(req)
          : await client.htmlDecode(req);
      }
      setOutput(resp.text);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-bold text-slate-800 dark:text-gray-200 flex items-center gap-2 capitalize">
          {type === 'url' ? <Link className="w-5 h-5"/> : <Code className="w-5 h-5"/>}
          {type} Encoder
        </h3>
        <div className="flex gap-2">
          <button onClick={() => handleAction('encode')} className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-4 py-1.5 rounded text-sm font-medium transition-colors">Encode</button>
          <button onClick={() => handleAction('decode')} className="bg-slate-100 dark:bg-gray-700 hover:bg-slate-200 dark:hover:bg-gray-600 text-slate-700 dark:text-white px-4 py-1.5 rounded text-sm font-bold transition-colors border border-slate-300 dark:border-transparent">Decode</button>
        </div>
      </div>
      
      <div className="grid grid-cols-2 gap-4">
        <textarea 
          value={input} 
          onChange={e => setInput(e.target.value)} 
          placeholder="Input..." 
          className="w-full h-64 bg-white dark:bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
        />
        <textarea 
          readOnly 
          value={output} 
          placeholder="Result..." 
          className="w-full h-64 bg-slate-50 dark:bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100"
        />
      </div>
    </div>
  );
}
