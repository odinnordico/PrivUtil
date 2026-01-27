import { useState } from 'react';
import { DataFormat } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { ArrowLeftRight } from 'lucide-react';

export function ConverterTool() {
  const [data, setData] = useState('');
  const [output, setOutput] = useState('');
  const [sourceFormat, setSourceFormat] = useState<DataFormat>(DataFormat.JSON);
  const [targetFormat, setTargetFormat] = useState<DataFormat>(DataFormat.YAML);
  const [error, setError] = useState<string | null>(null);

  const handleConvert = async () => {
    setError(null);
    try {
      const response = await client.convert({ 
        data, 
        sourceFormat, 
        targetFormat 
      } as Parameters<typeof client.convert>[0]);
      if (response.error) {
        setError(response.error);
      } else {
        setOutput(response.data);
      }
    } catch (err) {
      console.error(err);
      setError('Conversion failed');
    }
  };

  const getFormatName = (fmt: DataFormat) => {
    switch(fmt) {
      case DataFormat.JSON: return 'JSON';
      case DataFormat.YAML: return 'YAML';
      case DataFormat.XML: return 'XML';
      default: return 'Unknown';
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Universal Converter</h2>

      <div className="flex gap-4 items-center bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex items-center gap-2">
          <span className="text-slate-600 dark:text-gray-400 font-bold">From</span>
          <select 
            value={sourceFormat}
            onChange={(e) => setSourceFormat(parseInt(e.target.value))}
            className="bg-gray-100 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-gray-200 dark:border-transparent focus:ring-2 focus:ring-kawa-500"
          >
            <option value={DataFormat.JSON}>JSON</option>
            <option value={DataFormat.YAML}>YAML</option>
            <option value={DataFormat.XML}>XML</option>
          </select>
        </div>

        <ArrowLeftRight className="text-gray-500 w-5 h-5" />

        <div className="flex items-center gap-2">
          <span className="text-slate-600 dark:text-gray-400 font-bold">To</span>
          <select 
             value={targetFormat}
             onChange={(e) => setTargetFormat(parseInt(e.target.value))}
             className="bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500"
          >
            <option value={DataFormat.JSON}>JSON</option>
            <option value={DataFormat.YAML}>YAML</option>
            <option value={DataFormat.XML}>XML</option>
          </select>
        </div>

        <button
          onClick={handleConvert}
          className="ml-auto flex items-center gap-2 px-6 py-2 bg-kawa-500 hover:bg-kawa-600 rounded-lg text-slate-900 font-medium transition-colors"
        >
          Convert
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">{getFormatName(sourceFormat)} Input</label>
          <textarea
            className="w-full h-[500px] bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
            value={data}
            onChange={(e) => setData(e.target.value)}
            placeholder={`Paste ${getFormatName(sourceFormat)} here...`}
          />
        </div>
        
        <div className="space-y-2">
           <label className="text-sm font-bold text-slate-600 dark:text-slate-400">{getFormatName(targetFormat)} Output</label>
           {error ? (
              <div className="w-full h-[500px] bg-red-900/20 p-4 rounded-lg border border-red-500/50 text-red-400 font-mono text-sm">
                {error}
              </div>
           ) : (
             <textarea
               readOnly
                className="w-full h-[500px] bg-slate-50 dark:bg-black/30 p-4 rounded-lg border border-slate-300 dark:border-neutral-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none shadow-inner"
               value={output}
             />
           )}
        </div>
      </div>
    </div>
  );
}
