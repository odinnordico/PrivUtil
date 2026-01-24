import { useState } from 'react';
import { IpRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';
import { Network } from 'lucide-react';

export function IpTool() {
  const [cidr, setCidr] = useState('');
  const [res, setRes] = useState<any>(null);
  const [error, setError] = useState('');

  const calc = async () => {
    try {
      const resp = await client.ipCalc(IpRequest.create({ cidr }) as any);
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
      <h2 className="text-2xl font-bold text-white flex items-center gap-2">
        <Network className="w-6 h-6 text-cyan-400" />
        IP Subnet Calculator
      </h2>

      <div className="flex gap-4">
        <input 
          type="text" 
          value={cidr} 
          onChange={e => setCidr(e.target.value)}
          placeholder="192.168.1.0/24" 
          className="flex-1 bg-gray-800 text-white px-4 py-3 rounded-lg border border-gray-700 focus:ring-2 focus:ring-cyan-500 focus:outline-none font-mono"
        />
        <button onClick={calc} className="bg-cyan-600 hover:bg-cyan-700 text-white px-6 py-3 rounded-lg font-medium">Calculate</button>
      </div>

      {error && <div className="text-red-400 bg-red-900/20 p-4 rounded-lg">{error}</div>}

      {res && (
        <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-hidden">
          <div className="grid grid-cols-2 divide-x divide-gray-700 border-b border-gray-700 bg-gray-700/30">
            <div className="p-4 text-center">
              <div className="text-xs text-gray-400 uppercase font-bold mb-1">Network</div>
              <div className="text-lg font-mono text-cyan-300">{res.network}</div>
            </div>
            <div className="p-4 text-center">
              <div className="text-xs text-gray-400 uppercase font-bold mb-1">Broadcast</div>
              <div className="text-lg font-mono text-purple-300">{res.broadcast}</div>
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
    <div className="flex justify-between items-center border-b border-gray-700/50 pb-2 last:border-0 last:pb-0">
      <span className="text-gray-400">{label}</span>
      <span className="font-mono text-gray-200">{value}</span>
    </div>
  );
}
