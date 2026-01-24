import { useState } from 'react';
import { TextRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Link, Code } from 'lucide-react';

export function EncoderTool() {
  const [activeTab, setActiveTab] = useState<'url' | 'html'>('url');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white">Encoders / Decoders</h2>

      <div className="flex gap-4 border-b border-gray-700">
        <button
          onClick={() => setActiveTab('url')}
          className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'url' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}
        >
          URL
        </button>
        <button
          onClick={() => setActiveTab('html')}
          className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'html' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}
        >
          HTML
        </button>
      </div>

      <div className="p-6 bg-gray-800 rounded-lg border border-gray-700">
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
      const req = TextRequest.create({ text: input });
      let resp;
      if (type === 'url') {
        resp = action === 'encode' 
          ? await client.urlEncode(req as any) 
          : await client.urlDecode(req as any);
      } else {
        resp = action === 'encode'
          ? await client.htmlEncode(req as any)
          : await client.htmlDecode(req as any);
      }
      setOutput(resp.text);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-medium text-gray-200 flex items-center gap-2 capitalize">
          {type === 'url' ? <Link className="w-5 h-5"/> : <Code className="w-5 h-5"/>}
          {type} Encoder
        </h3>
        <div className="flex gap-2">
          <button onClick={() => handleAction('encode')} className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded text-sm">Encode</button>
          <button onClick={() => handleAction('decode')} className="bg-gray-700 hover:bg-gray-600 text-white px-4 py-1.5 rounded text-sm">Decode</button>
        </div>
      </div>
      
      <div className="grid grid-cols-2 gap-4">
        <textarea 
          value={input} 
          onChange={e => setInput(e.target.value)} 
          placeholder="Input..." 
          className="w-full h-64 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-blue-500"
        />
        <textarea 
          readOnly 
          value={output} 
          placeholder="Result..." 
          className="w-full h-64 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700"
        />
      </div>
    </div>
  );
}
