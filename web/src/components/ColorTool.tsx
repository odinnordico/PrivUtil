import { useState } from 'react';
import { ColorRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Palette, Copy } from 'lucide-react';

export function ColorTool() {
  const [input, setInput] = useState('#3b82f6');
  const [res, setRes] = useState<any>(null);

  const convert = async (val: string) => {
    setInput(val);
    // Simple debounce or just call on change if light enough? Let's verify input length for hex
    if (val.startsWith('#') && (val.length === 4 || val.length === 7)) {
       doConvert(val);
    } else if (val.startsWith('rgb')) {
       doConvert(val);
    }
  };

  const doConvert = async (val: string) => {
    try {
      const resp = await client.colorConvert(ColorRequest.create({ input: val }) as any);
      if (!resp.error) setRes(resp);
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6 max-w-xl mx-auto">
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <Palette className="w-6 h-6 text-purple-400" /> 
        Color Converter
      </h2>
      
      <p className="text-gray-400">
        Convert colors between common formats. Enter a HEX code (e.g. #3b82f6) or RGB value (e.g. rgb(59, 130, 246)) to see conversions and previews.
      </p>

      <div className="flex gap-4 items-center">
        <input 
          type="color" 
          value={res?.hex || '#000000'}
          onChange={e => convert(e.target.value)}
          className="w-16 h-16 rounded cursor-pointer bg-transparent border-0 p-0"
        />
        <input 
          type="text" 
          value={input} 
          onChange={e => convert(e.target.value)}
          placeholder="#RRGGBB or rgb(r,g,b)"
          className="flex-1 bg-gray-800 text-white px-4 py-3 rounded-lg border border-gray-700 focus:ring-2 focus:ring-purple-500 focus:outline-none font-mono text-lg uppercase"
        />
      </div>

      {res && (
        <div className="grid gap-3">
          <ColorRow label="HEX" value={res.hex} color={res.hex} />
          <ColorRow label="RGB" value={res.rgb} color={res.hex} />
          <ColorRow label="HSL" value={res.hsl} color={res.hex} />
        </div>
      )}
    </div>
  );
}

function ColorRow({ label, value, color }: { label: string, value: string, color: string }) {
  return (
    <div className="flex items-center justify-between p-4 bg-gray-800 rounded-lg border border-gray-700 hover:border-gray-600 transition-colors group">
      <div className="flex items-center gap-4">
        <div className="w-2 h-8 rounded bg-gray-600" style={{ backgroundColor: color }} />
        <span className="text-gray-400 font-bold w-12">{label}</span>
      </div>
      <div className="flex items-center gap-4">
        <code className="font-mono text-gray-200">{value}</code>
        <button 
          onClick={() => navigator.clipboard.writeText(value)}
          className="opacity-0 group-hover:opacity-100 p-2 text-gray-400 hover:text-white transition-opacity"
        >
          <Copy className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
