import { useState } from 'react';
import { UuidRequest, LoremRequest, HashRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { RefreshCw, Copy, Hash, FileText } from 'lucide-react';

export function GeneratorTool() {
  const [activeTab, setActiveTab] = useState<'uuid' | 'lorem' | 'hash'>('uuid');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white">Generators</h2>

      <div className="flex gap-4 border-b border-gray-700">
        <button
          onClick={() => setActiveTab('uuid')}
          className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'uuid' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}
        >
          UUIDs
        </button>
        <button
          onClick={() => setActiveTab('lorem')}
          className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'lorem' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}
        >
          Lorem Ipsum
        </button>
        <button
          onClick={() => setActiveTab('hash')}
          className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'hash' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}
        >
          Hash Calculator
        </button>
      </div>

      <div className="p-6 bg-gray-800 rounded-lg border border-gray-700">
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
      const resp = await client.generateUuid(UuidRequest.create({ count, hyphen, uppercase, version }) as any);
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
          <label className="block text-xs text-gray-400 mb-1">Version</label>
          <select value={version} onChange={e => setVersion(e.target.value)} className="bg-gray-700 text-white rounded px-2 py-1">
            <option value="v4">v4 (Random)</option>
            <option value="v1">v1 (Time)</option>
          </select>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={hyphen} onChange={e => setHyphen(e.target.checked)} />
          <span className="text-gray-300 text-sm">Hyphens</span>
        </label>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={uppercase} onChange={e => setUppercase(e.target.checked)} />
          <span className="text-gray-300 text-sm">Uppercase</span>
        </label>
        <button onClick={generate} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded flex items-center gap-2">
          <RefreshCw className="w-4 h-4" /> Generate
        </button>
      </div>
      
      {uuids.length > 0 && (
        <div className="bg-black/30 p-4 rounded font-mono text-sm space-y-1">
          {uuids.map((u, i) => (
            <div key={i} className="flex justify-between hover:bg-gray-700/50 p-1 rounded group">
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
      const resp = await client.generateLorem(LoremRequest.create({ type, count }) as any);
      setText(resp.text);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-4 items-end">
        <div>
          <label className="block text-xs text-gray-400 mb-1">Type</label>
          <select value={type} onChange={e => setType(e.target.value)} className="bg-gray-700 text-white rounded px-2 py-1">
            <option value="paragraph">Paragraphs</option>
            <option value="sentence">Sentences</option>
            <option value="word">Words</option>
          </select>
        </div>
        <div>
          <label className="block text-xs text-gray-400 mb-1">Count</label>
          <input type="number" value={count} onChange={e => setCount(parseInt(e.target.value))} className="bg-gray-700 text-white rounded px-2 py-1 w-20" min="1" max="100"/>
        </div>
        <button onClick={generate} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded flex items-center gap-2">
          <FileText className="w-4 h-4" /> Generate
        </button>
      </div>
      <textarea readOnly value={text} className="w-full h-64 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none" />
    </div>
  );
}

function HashGenerator() {
  const [input, setInput] = useState('');
  const [hash, setHash] = useState('');
  const [algo, setAlgo] = useState('sha256');

  const calculate = async () => {
    try {
      const resp = await client.calculateHash(HashRequest.create({ text: input, algo }) as any);
      setHash(resp.hash);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Algorithm</label>
          <select value={algo} onChange={e => setAlgo(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2 w-full">
            <option value="md5">MD5</option>
            <option value="sha1">SHA1</option>
            <option value="sha256">SHA256</option>
            <option value="sha512">SHA512</option>
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Input Text</label>
          <textarea 
            value={input} 
            onChange={e => setInput(e.target.value)} 
            className="w-full h-24 bg-gray-700 p-2 rounded text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Text to hash..."
          />
        </div>
        <button onClick={calculate} className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded flex items-center gap-2 w-fit">
          <Hash className="w-4 h-4" /> Calculate Hash
        </button>
        {hash && (
          <div>
            <label className="block text-sm text-gray-400 mb-1">Result</label>
            <div className="bg-black/30 p-3 rounded font-mono text-break-all text-blue-300 relative group">
              {hash}
              <button 
                onClick={() => navigator.clipboard.writeText(hash)}
                className="absolute right-2 top-2 opacity-0 group-hover:opacity-100 text-gray-400 hover:text-white bg-gray-800 p-1 rounded"
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
