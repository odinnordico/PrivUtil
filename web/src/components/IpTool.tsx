import { useState } from 'react';
import { IpResponse } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Network } from 'lucide-react';

export function IpTool() {
  const [cidr, setCidr] = useState('');
  const [res, setRes] = useState<IpResponse | null>(null);
  const [error, setError] = useState('');

  const calc = async () => {
    try {
      const resp = await client.ipCalc({ cidr } as Parameters<typeof client.ipCalc>[0]);
      if (resp.error) {
        setError(resp.error);
        setRes(null);
      } else {
        setError('');
        setRes(resp);
      }
    } catch (e) { console.error(e); }
  };

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <Network className="w-6 h-6 text-cyan-400" />
        IP Subnet Calculator
      </h2>

      <div className="flex gap-4">
        <input 
          type="text" 
          value={cidr} 
          onChange={e => setCidr(e.target.value)}
          placeholder="192.168.1.0/24" 
          className="flex-1 bg-white dark:bg-gray-800 text-slate-900 dark:text-white px-4 py-3 rounded-lg border border-slate-300 dark:border-gray-700 focus:ring-2 focus:ring-kawa-500 focus:outline-none font-mono shadow-inner"
        />
        <button onClick={calc} className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-6 py-3 rounded-lg font-bold transition-all shadow-md active:scale-95">Calculate</button>
      </div>

      {error && <div className="text-red-400 bg-red-900/20 p-4 rounded-lg">{error}</div>}

      {res && (
        <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 overflow-hidden shadow-sm">
          <div className="grid grid-cols-2 divide-x divide-slate-200 dark:divide-gray-700 border-b border-slate-200 dark:border-gray-700 bg-slate-50 dark:bg-gray-700/30">
            <div className="p-4 text-center">
              <div className="text-xs text-slate-500 dark:text-gray-400 uppercase font-bold mb-1">Network</div>
              <div className="text-lg font-mono text-cyan-600 dark:text-cyan-300">{res.network}</div>
            </div>
            <div className="p-4 text-center">
              <div className="text-xs text-slate-500 dark:text-gray-400 uppercase font-bold mb-1">Broadcast</div>
              <div className="text-lg font-mono text-purple-600 dark:text-purple-300">{res.broadcast}</div>
            </div>
          </div>
          
          <div className="p-6 space-y-4">
            <IpRow label="Netmask" value={res.netmask} />
            <IpRow label="First IP" value={res.firstIp} />
            <IpRow label="Last IP" value={res.lastIp} />
            <IpRow label="Total Hosts" value={res.numHosts.toString()} />
          </div>
        </div>
      )}
    </div>
  );
}

function IpRow({ label, value }: { label: string, value: string }) {
  return (
    <div className="flex justify-between items-center border-b border-slate-100 dark:border-gray-700/50 pb-2 last:border-0 last:pb-0">
      <span className="text-slate-500 dark:text-gray-400 font-bold">{label}</span>
      <span className="font-mono text-slate-700 dark:text-gray-200">{value}</span>
    </div>
  );
}
