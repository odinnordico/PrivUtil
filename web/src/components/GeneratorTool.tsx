import { useState } from 'react';
import { client } from '../lib/client';
import { RefreshCw, Copy, Hash, FileText } from 'lucide-react';

export function GeneratorTool() {
  const [activeTab, setActiveTab] = useState<'uuid' | 'lorem' | 'hash'>('uuid');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Generators</h2>

      <div className="flex gap-4 border-b border-slate-300 dark:border-gray-700">
        <button
          onClick={() => setActiveTab('uuid')}
          className={`pb-2 px-4 font-bold transition-colors ${activeTab === 'uuid' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          UUIDs
        </button>
        <button
          onClick={() => setActiveTab('lorem')}
          className={`pb-2 px-4 font-bold transition-colors ${activeTab === 'lorem' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          Lorem Ipsum
        </button>
        <button
          onClick={() => setActiveTab('hash')}
          className={`pb-2 px-4 font-bold transition-colors ${activeTab === 'hash' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          Hash Calculator
        </button>
      </div>

      <div className="p-6 bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        {activeTab === 'uuid' && <UuidGenerator />}
        {activeTab === 'lorem' && <LoremGenerator />}
        {activeTab === 'hash' && <HashGenerator />}
      </div>
    </div>
  );
}

function UuidGenerator() {
  const [uuids, setUuids] = useState<string[]>([]);
  const [count, setCount] = useState(5);
  const [hyphen, setHyphen] = useState(true);
  const [uppercase, setUppercase] = useState(false);
  const [version, setVersion] = useState('v4');

  const generate = async () => {
    try {
      const resp = await client.generateUuid({ count, hyphen, uppercase, version } as Parameters<typeof client.generateUuid>[0]);
      setUuids(resp.uuids);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-end">
        <div>
          <label className="block text-xs text-gray-400 mb-1">Count</label>
          <input type="number" value={count} onChange={e => setCount(parseInt(e.target.value))} className="bg-gray-700 text-white rounded px-2 py-1 w-20" min="1" max="100"/>
        </div>
        <div>
          <label className="block text-xs text-slate-500 dark:text-gray-400 mb-1 font-bold italic">Version</label>
          <select value={version} onChange={e => setVersion(e.target.value)} className="bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-2 py-1 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500">
            <option value="v4">v4 (Random)</option>
            <option value="v1">v1 (Time)</option>
          </select>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={hyphen} onChange={e => setHyphen(e.target.checked)} />
          <span className="text-slate-700 dark:text-gray-300 text-sm font-medium">Hyphens</span>
        </label>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={uppercase} onChange={e => setUppercase(e.target.checked)} />
          <span className="text-slate-700 dark:text-gray-300 text-sm font-medium">Uppercase</span>
        </label>
        <button onClick={generate} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded flex items-center gap-2">
          <RefreshCw className="w-4 h-4" /> Generate
        </button>
      </div>
      
      {uuids.length > 0 && (
        <div className="bg-slate-50 dark:bg-black/30 p-4 rounded font-mono text-sm space-y-1 border border-slate-200 dark:border-transparent shadow-inner">
          {uuids.map((u, i) => (
            <div key={i} className="flex justify-between hover:bg-gray-200 dark:hover:bg-gray-700/50 p-1 rounded group text-slate-800 dark:text-neutral-100">
              <span>{u}</span>
              <button 
                onClick={() => navigator.clipboard.writeText(u)}
                className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-white"
                title="Copy"
              >
                <Copy className="w-4 h-4"/>
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function LoremGenerator() {
  const [text, setText] = useState('');
  const [type, setType] = useState('paragraph');
  const [count, setCount] = useState(3);

  const generate = async () => {
    try {
      const resp = await client.generateLorem({ type, count } as Parameters<typeof client.generateLorem>[0]);
      setText(resp.text);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-4 items-end">
        <div>
          <label className="block text-xs text-slate-500 dark:text-gray-400 mb-1 font-bold italic">Type</label>
          <select value={type} onChange={e => setType(e.target.value)} className="bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-2 py-1 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500">
            <option value="paragraph">Paragraphs</option>
            <option value="sentence">Sentences</option>
            <option value="word">Words</option>
          </select>
        </div>
        <div>
          <label className="block text-xs text-gray-400 mb-1">Count</label>
          <input type="number" value={count} onChange={e => setCount(parseInt(e.target.value))} className="bg-gray-700 text-white rounded px-2 py-1 w-20" min="1" max="100"/>
        </div>
        <button onClick={generate} className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-4 py-1.5 rounded flex items-center gap-2 font-bold shadow-md transition-all active:scale-95">
          <FileText className="w-4 h-4" /> Generate
        </button>
      </div>
      <textarea readOnly value={text} className="w-full h-64 bg-slate-50 dark:bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 shadow-inner" />
    </div>
  );
}

function HashGenerator() {
  const [input, setInput] = useState('');
  const [hash, setHash] = useState('');
  const [algo, setAlgo] = useState('sha256');

  const calculate = async () => {
    try {
      const resp = await client.calculateHash({ text: input, algo } as Parameters<typeof client.calculateHash>[0]);
      setHash(resp.hash);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4">
        <div>
          <label className="block text-sm text-slate-600 dark:text-gray-400 mb-1 font-bold">Algorithm</label>
          <select value={algo} onChange={e => setAlgo(e.target.value)} className="bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 w-full border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 shadow-sm">
            <option value="md5">MD5</option>
            <option value="sha1">SHA1</option>
            <option value="sha256">SHA256</option>
            <option value="sha512">SHA512</option>
          </select>
        </div>
        <div>
          <label className="block text-sm text-slate-600 dark:text-gray-400 mb-1 font-bold">Input Text</label>
          <textarea 
            value={input} 
            onChange={e => setInput(e.target.value)} 
            className="w-full h-24 bg-white dark:bg-gray-700 p-2 rounded text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-kawa-500 border border-gray-200 dark:border-transparent shadow-inner"
            placeholder="Text to hash..."
          />
        </div>
        <button onClick={calculate} className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-6 py-2 rounded flex items-center gap-2 w-fit font-bold transition-all shadow-md active:scale-95">
          <Hash className="w-4 h-4" /> Calculate Hash
        </button>
        {hash && (
          <div>
            <label className="block text-sm text-slate-600 dark:text-gray-400 mb-1 font-bold">Result</label>
            <div className="bg-slate-50 dark:bg-black/30 p-3 rounded font-mono break-all text-kawa-700 dark:text-kawa-300 relative group border border-slate-300 dark:border-transparent min-h-[4rem] shadow-inner">
              {hash}
              <button 
                onClick={() => navigator.clipboard.writeText(hash)}
                className="absolute right-2 top-2 opacity-0 group-hover:opacity-100 text-slate-500 dark:text-gray-400 hover:text-slate-900 dark:hover:text-white bg-white dark:bg-gray-800 p-1 rounded border border-gray-200 dark:border-transparent transition-all"
              >
                <Copy className="w-4 h-4"/>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
