import { useState } from 'react';
import { CertRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { FileBadge, ShieldCheck, Calendar, Globe } from 'lucide-react';

export function CertTool() {
  const [pem, setPem] = useState('');
  const [cert, setCert] = useState<any>(null);
  const [error, setError] = useState('');

  const parse = async (data: string) => {
    setPem(data);
    if (!data.trim()) {
      setCert(null);
      return;
    }
    try {
      const resp = await client.certParse(CertRequest.create({ data }) as any);
      if (resp.error) {
        setError(resp.error);
        setCert(null);
      } else {
        setError('');
        setCert(resp);
      }
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <FileBadge className="w-6 h-6 text-yellow-400" /> 
        Certificate Inspector
      </h2>

      <div className="grid lg:grid-cols-2 gap-6">
        <div className="space-y-2">
          <label className="text-gray-400 text-sm">PEM Certificate</label>
          <textarea 
            value={pem} 
            onChange={e => parse(e.target.value)} 
            placeholder="-----BEGIN CERTIFICATE-----..." 
            className="w-full h-[500px] bg-black/30 p-4 rounded-lg font-mono text-xs resize-none focus:outline-none border border-gray-700 focus:border-yellow-500 text-gray-300"
          />
        </div>

        <div className="space-y-4">
          {error && (
            <div className="bg-red-900/20 border border-red-800 text-red-400 p-4 rounded-lg">
              {error}
            </div>
          )}
          
          {cert && (
            <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
              <div className="p-4 bg-gray-700/50 border-b border-gray-700 font-medium text-white flex items-center gap-2">
                <ShieldCheck className="w-4 h-4 text-green-400"/> Certificate Details
              </div>
              
              <div className="p-4 space-y-6">
                <DetailGroup label="Subject" value={cert.subject} />
                <DetailGroup label="Issuer" value={cert.issuer} />
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                      <Calendar className="w-3 h-3"/> Not Before
                    </div>
                    <div className="text-sm font-mono text-gray-200">{cert.notBefore}</div>
                  </div>
                  <div className="space-y-1">
                    <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                      <Calendar className="w-3 h-3"/> Not After
                    </div>
                    <div className="text-sm font-mono text-gray-200">{cert.notAfter}</div>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                    <Globe className="w-3 h-3"/> SANs (DNS Names)
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {cert.sans && cert.sans.length > 0 ? (
                      cert.sans.map((s: string, i: number) => (
                        <span key={i} className="px-2 py-1 bg-gray-700 rounded text-xs text-gray-300 font-mono">
                          {s}
                        </span>
                      ))
                    ) : (
                      <span className="text-gray-600 text-sm italic">None</span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}
          
          {!cert && !error && (
            <div className="h-full flex items-center justify-center text-gray-600 border-2 border-dashed border-gray-800 rounded-lg">
              Paste a certificate to inspect
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function DetailGroup({ label, value }: { label: string, value: string }) {
  // Simple parser to make Subject/Issuer string more readable? 
  // Normally comes as "CN=foo,O=bar". Let's just display as is for now or maybe split by comma.
  return (
    <div className="space-y-1">
      <div className="text-xs text-gray-500 uppercase font-bold">{label}</div>
      <div className="text-sm text-gray-200 font-mono bg-black/20 p-2 rounded break-all border border-gray-700/50">
        {value}
      </div>
    </div>
  );
}
