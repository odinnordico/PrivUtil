import { useState } from 'react';
import { ConvertRequest, DataFormat } from '../proto/proto/privutil';
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
      const request = ConvertRequest.create({ 
        data, 
        sourceFormat, 
        targetFormat 
      });
      const response = await client.convert(request as any);
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
      <h2 className="text-2xl font-bold text-white">Universal Converter</h2>

      <div className="flex gap-4 items-center bg-gray-800/50 p-4 rounded-lg">
        <div className="flex items-center gap-2">
          <span className="text-gray-400">From</span>
          <select 
            value={sourceFormat}
            onChange={(e) => setSourceFormat(parseInt(e.target.value))}
            className="bg-gray-700 text-white rounded px-3 py-2 border-none focus:ring-2 focus:ring-blue-500"
          >
            <option value={DataFormat.JSON}>JSON</option>
            <option value={DataFormat.YAML}>YAML</option>
            <option value={DataFormat.XML}>XML</option>
          </select>
        </div>

        <ArrowLeftRight className="text-gray-500 w-5 h-5" />

        <div className="flex items-center gap-2">
          <span className="text-gray-400">To</span>
          <select 
             value={targetFormat}
             onChange={(e) => setTargetFormat(parseInt(e.target.value))}
             className="bg-gray-700 text-white rounded px-3 py-2 border-none focus:ring-2 focus:ring-blue-500"
          >
            <option value={DataFormat.JSON}>JSON</option>
            <option value={DataFormat.YAML}>YAML</option>
            <option value={DataFormat.XML}>XML</option>
          </select>
        </div>

        <button
          onClick={handleConvert}
          className="ml-auto flex items-center gap-2 px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white font-medium transition-colors"
        >
          Convert
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-gray-400">{getFormatName(sourceFormat)} Input</label>
          <textarea
            className="w-full h-[500px] bg-gray-800 p-4 rounded-lg border border-gray-700 text-gray-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50"
            value={data}
            onChange={(e) => setData(e.target.value)}
            placeholder={`Paste ${getFormatName(sourceFormat)} here...`}
          />
        </div>
        
        <div className="space-y-2">
           <label className="text-sm font-medium text-gray-400">{getFormatName(targetFormat)} Output</label>
           {error ? (
              <div className="w-full h-[500px] bg-red-900/20 p-4 rounded-lg border border-red-500/50 text-red-400 font-mono text-sm">
                {error}
              </div>
           ) : (
             <textarea
               readOnly
               className="w-full h-[500px] bg-black/30 p-4 rounded-lg border border-gray-800 text-gray-100 font-mono text-sm focus:outline-none"
               value={output}
             />
           )}
        </div>
      </div>
    </div>
  );
}
