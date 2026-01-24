import { useState } from 'react';
import { JwtRequest, RegexRequest, JsonToGoRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Terminal, Braces, Search } from 'lucide-react';

export function DevTools() {
  const [activeTab, setActiveTab] = useState<'jwt' | 'regex' | 'json2go'>('jwt');

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white">Developer Utilities</h2>

      <div className="flex gap-4 border-b border-gray-700">
        <button onClick={() => setActiveTab('jwt')} className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'jwt' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}>
          JWT Debugger
        </button>
        <button onClick={() => setActiveTab('regex')} className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'regex' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}>
          Regex Tester
        </button>
        <button onClick={() => setActiveTab('json2go')} className={`pb-2 px-4 font-medium transition-colors ${activeTab === 'json2go' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-200'}`}>
          JSON to Go
        </button>
      </div>

      <div className="p-6 bg-gray-800 rounded-lg border border-gray-700">
        {activeTab === 'jwt' && <JwtDebugger />}
        {activeTab === 'regex' && <RegexTester />}
        {activeTab === 'json2go' && <JsonToGo />}
      </div>
    </div>
  );
}

function JwtDebugger() {
  const [token, setToken] = useState('');
  const [res, setRes] = useState<{header: string, payload: string, error?: string} | null>(null);

  const decode = async (t: string) => {
    setToken(t);
    try {
      const resp = await client.jwtDecode(JwtRequest.create({ token: t }) as any);
      setRes({ header: resp.header, payload: resp.payload, error: resp.error });
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-gray-400 mb-2">
        <Terminal className="w-5 h-5"/> Encode JWT
      </div>
      <textarea 
        value={token} 
        onChange={e => decode(e.target.value)} 
        placeholder="Paste JWT here..." 
        className="w-full h-24 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700 focus:border-blue-500"
      />
      {res?.error ? (
        <div className="text-red-400 bg-red-900/20 p-3 rounded">{res.error}</div>
      ) : (
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-xs text-gray-500 uppercase font-bold">Header</label>
            <pre className="mt-1 bg-black/30 p-4 rounded text-green-400 font-mono text-sm overflow-auto h-64">{res?.header}</pre>
          </div>
          <div>
            <label className="text-xs text-gray-500 uppercase font-bold">Payload</label>
            <pre className="mt-1 bg-black/30 p-4 rounded text-blue-400 font-mono text-sm overflow-auto h-64">{res?.payload}</pre>
          </div>
        </div>
      )}
    </div>
  );
}

function RegexTester() {
  const [pattern, setPattern] = useState('');
  const [text, setText] = useState('');
  const [matches, setMatches] = useState<string[]>([]);
  const [error, setError] = useState('');

  const test = async () => {
    try {
      const resp = await client.regexTest(RegexRequest.create({ pattern, text }) as any);
      if (resp.error) setError(resp.error);
      else {
        setError('');
        setMatches(resp.matches);
      }
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-gray-400 mb-2">
        <Search className="w-5 h-5"/> Regex Pattern
      </div>
      <input 
        type="text" 
        value={pattern}
        onChange={e => setPattern(e.target.value)}
        placeholder="e.g. ^[a-z]+$"
        className="w-full bg-gray-900 text-white px-4 py-2 rounded border border-gray-700 font-mono"
      />
      <div className="flex gap-4">
        <textarea 
          value={text} 
          onChange={e => setText(e.target.value)} 
          placeholder="Test string..." 
          className="flex-1 h-48 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700"
        />
        <div className="flex-1 h-48 bg-black/30 p-4 rounded font-mono text-sm overflow-auto border border-gray-700">
          {error ? (
            <span className="text-red-400">{error}</span>
          ) : (
            <div className="space-y-1">
              <div className="text-gray-500 mb-2">{matches.length} matches found</div>
              {matches.map((m, i) => (
                <div key={i} className="bg-blue-500/20 text-blue-300 px-2 py-1 rounded inline-block mr-2 mb-1">{m}</div>
              ))}
            </div>
          )}
        </div>
      </div>
      <button onClick={test} className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded">Test Regex</button>
    </div>
  );
}

function JsonToGo() {
  const [json, setJson] = useState('');
  const [goCode, setGoCode] = useState('');
  const [error, setError] = useState('');

  const convert = async () => {
    try {
      const resp = await client.jsonToGo(JsonToGoRequest.create({ json, structName: 'AutoGenerated' }) as any);
      if (resp.error) setError(resp.error);
      else {
        setError('');
        setGoCode(resp.goCode);
      }
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-gray-400 mb-2">
        <Braces className="w-5 h-5"/> JSON to Go Struct
      </div>
      <div className="grid grid-cols-2 gap-4">
        <textarea 
          value={json} 
          onChange={e => setJson(e.target.value)} 
          placeholder='{"key": "value"}' 
          className="w-full h-96 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700"
        />
        <div className="relative">
          <textarea 
            readOnly 
            value={error || goCode} 
            className={`w-full h-96 bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-gray-700 ${error ? 'text-red-400' : 'text-green-400'}`}
          />
        </div>
      </div>
      <button onClick={convert} className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded">Convert</button>
    </div>
  );
}
