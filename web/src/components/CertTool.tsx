import { useState } from 'react';
import { CertResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { FileBadge, ShieldCheck, Calendar, Globe } from 'lucide-react';

export function CertTool() {
  const [pem, setPem] = useState('');
  const [cert, setCert] = useState<CertResponse | null>(null);
  const [error, setError] = useState('');

  const parse = async (data: string) => {
    setPem(data);
    if (!data.trim()) {
      setCert(null);
      return;
    }
    try {
      const resp = await client.certParse({ data } as Parameters<typeof client.certParse>[0]);
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
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <FileBadge className="w-6 h-6 text-yellow-400" /> 
        Certificate Inspector
      </h2>

      <div className="grid lg:grid-cols-2 gap-6">
        <div className="space-y-2">
          <label className="text-slate-600 dark:text-gray-400 text-sm font-bold">PEM Certificate</label>
          <textarea 
            value={pem} 
            onChange={e => parse(e.target.value)} 
            placeholder="-----BEGIN CERTIFICATE-----..." 
            className="w-full h-[500px] bg-white dark:bg-black/30 p-4 rounded-lg font-mono text-xs resize-none focus:outline-none border border-slate-300 dark:border-gray-700 focus:border-kawa-500 text-slate-900 dark:text-gray-300 shadow-inner"
          />
        </div>

        <div className="space-y-4">
          {error && (
            <div className="bg-red-900/20 border border-red-800 text-red-400 p-4 rounded-lg">
              {error}
            </div>
          )}
          
          {cert && (
            <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 overflow-hidden shadow-sm">
              <div className="p-4 bg-slate-50 dark:bg-gray-700/50 border-b border-slate-300 dark:border-gray-700 font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <ShieldCheck className="w-4 h-4 text-kawa-600 dark:text-kawa-400"/> Certificate Details
              </div>
              
              <div className="p-4 space-y-6">
                <DetailGroup label="Subject" value={cert.subject} />
                <DetailGroup label="Issuer" value={cert.issuer} />
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                      <Calendar className="w-3 h-3"/> Not Before
                    </div>
                    <div className="text-sm font-mono text-slate-700 dark:text-gray-200">{cert.notBefore}</div>
                  </div>
                  <div className="space-y-1">
                    <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                      <Calendar className="w-3 h-3"/> Not After
                    </div>
                    <div className="text-sm font-mono text-slate-700 dark:text-gray-200">{cert.notAfter}</div>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="text-xs text-gray-500 uppercase font-bold flex items-center gap-1">
                    <Globe className="w-3 h-3"/> SANs (DNS Names)
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {cert.sans && cert.sans.length > 0 ? (
                      cert.sans.map((s: string, i: number) => (
                        <span key={i} className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs text-slate-700 dark:text-gray-300 font-mono border border-gray-200 dark:border-transparent">
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
            <div className="h-full flex items-center justify-center text-slate-400 dark:text-gray-600 border-2 border-dashed border-slate-300 dark:border-gray-800 rounded-lg min-h-[500px] font-bold">
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
      <div className="text-xs text-slate-500 dark:text-gray-500 uppercase font-bold">{label}</div>
      <div className="text-sm text-slate-700 dark:text-gray-200 font-mono bg-slate-50 dark:bg-black/20 p-2 rounded break-all border border-slate-300 dark:border-gray-700/50 shadow-inner">
        {value}
      </div>
    </div>
  );
}
