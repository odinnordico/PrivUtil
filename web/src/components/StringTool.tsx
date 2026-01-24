import { useState } from 'react';
import { CaseResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { CaseSensitive, Code2, Copy } from 'lucide-react';

export function StringTool() {
  const [activeTab, setActiveTab] = useState<'case' | 'escape'>('case');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <CaseSensitive className="w-6 h-6 text-pink-400" />
        String Utilities
      </h2>

      <div className="flex gap-4 border-b border-gray-700">
        <button onClick={() => setActiveTab('case')} className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'case' ? 'text-pink-400 border-b-2 border-pink-400' : 'text-gray-400 hover:text-gray-200'}`}>
          Case Converter
        </button>
        <button onClick={() => setActiveTab('escape')} className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'escape' ? 'text-pink-400 border-b-2 border-pink-400' : 'text-gray-400 hover:text-gray-200'}`}>
          String Escaper
        </button>
      </div>

      <div className="p-6 bg-gray-800 rounded-lg border border-gray-700">
        {activeTab === 'case' && <CaseConverter />}
        {activeTab === 'escape' && <StringEscaper />}
      </div>
    </div>
  );
}

function CaseConverter() {
  const [text, setText] = useState('');
  const [res, setRes] = useState<CaseResponse | null>(null);

  const convert = async (val: string) => {
    setText(val);
    try {
      const resp = await client.caseConvert({ text: val } as Parameters<typeof client.caseConvert>[0]);
      setRes(resp);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6">
      <textarea 
        value={text}
        onChange={e => convert(e.target.value)}
        placeholder="Type variable name to convert..."
        className="w-full h-32 bg-black/30 p-4 rounded-lg font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-pink-500"
      />
      
      {res && (
        <div className="grid md:grid-cols-2 gap-4">
          <CopyInput label="camelCase" value={res.camel} />
          <CopyInput label="PascalCase" value={res.pascal} />
          <CopyInput label="snake_case" value={res.snake} />
          <CopyInput label="kebab-case" value={res.kebab} />
          <CopyInput label="CONSTANT_CASE" value={res.constant} />
          <CopyInput label="Title Case" value={res.title} />
        </div>
      )}
    </div>
  );
}

function CopyInput({ label, value }: { label: string, value: string }) {
  return (
    <div className="space-y-1 group">
      <label className="text-xs text-gray-500 uppercase font-bold">{label}</label>
      <div className="flex gap-2">
        <input 
          readOnly 
          value={value} 
          className="flex-1 bg-black/20 text-gray-200 px-3 py-2 rounded font-mono text-sm border border-gray-700/50"
        />
        <button 
          onClick={() => navigator.clipboard.writeText(value)}
          className="bg-gray-700 hover:bg-gray-600 text-white p-2 rounded opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <Copy className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

function StringEscaper() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState('');
  const [mode, setMode] = useState('json');

  const process = async (action: 'escape' | 'unescape') => {
    try {
      const resp = await client.stringEscape({ text: input, mode, action } as Parameters<typeof client.stringEscape>[0]);
      if (resp.error) setResult(`Error: ${resp.error}`);
      else setResult(resp.result);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <Code2 className="w-5 h-5 text-gray-400" />
          <select 
            value={mode} 
            onChange={e => setMode(e.target.value)}
            className="bg-gray-900 text-white px-3 py-1.5 rounded border border-gray-700 text-sm"
          >
            <option value="json">JSON String</option>
            <option value="java">Java/C/Go String</option>
            <option value="sql">SQL String</option>
            <option value="html_entity">HTML Entities</option>
            <option value="url">URL Encoded</option>
          </select>
        </div>
        <div className="flex gap-2">
          <button onClick={() => process('escape')} className="bg-pink-600 hover:bg-pink-700 text-white px-4 py-1.5 rounded text-sm font-medium">Escape</button>
          <button onClick={() => process('unescape')} className="bg-gray-700 hover:bg-gray-600 text-white px-4 py-1.5 rounded text-sm font-medium">Unescape</button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <textarea 
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Input text..."
          className="w-full h-64 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-pink-500"
        />
        <textarea 
          readOnly
          value={result}
          placeholder="Result..."
          className="w-full h-64 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700"
        />
      </div>
    </div>
  );
}
