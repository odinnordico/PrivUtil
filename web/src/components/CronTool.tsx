import { useState, useEffect } from 'react';
import { client } from '../lib/client';
import { Clock } from 'lucide-react';

export function CronTool() {
  const [expr, setExpr] = useState('*/5 * * * *');
  const [res, setRes] = useState<{desc: string, next: string[], error?: string} | null>(null);

  useEffect(() => {
    const timer = setTimeout(async () => {
      if (!expr) return;
      try {
        const resp = await client.cronExplain({ expression: expr } as Parameters<typeof client.cronExplain>[0]);
        setRes({ 
          desc: resp.description, 
          next: resp.nextRuns.split('\n').filter(Boolean), 
          error: resp.error 
        });
      } catch (e) { console.error(e); }
    }, 500);
    return () => clearTimeout(timer);
  }, [expr]);

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <Clock className="w-6 h-6 text-kawa-500" /> 
        Cron Expression Tester
      </h2>

      <div className="space-y-2">
        <label className="text-slate-600 dark:text-gray-400 text-sm font-bold">Cron Expression</label>
        <input 
          type="text" 
          value={expr}
          onChange={e => setExpr(e.target.value)}
          placeholder="* * * * *"
          className="w-full bg-white dark:bg-gray-800 text-slate-900 dark:text-white px-4 py-3 rounded-lg border border-slate-300 dark:border-gray-700 focus:ring-2 focus:ring-kawa-500 focus:outline-none font-mono text-lg shadow-inner"
        />
        <div className="grid grid-cols-5 gap-2 text-xs text-gray-500 font-mono mt-1 text-center">
          <div className="bg-slate-50 dark:bg-gray-900/50 p-1 rounded border border-slate-300 dark:border-gray-700">Minute<br/>(0-59)</div>
          <div className="bg-slate-50 dark:bg-gray-900/50 p-1 rounded border border-slate-300 dark:border-gray-700">Hour<br/>(0-23)</div>
          <div className="bg-slate-50 dark:bg-gray-900/50 p-1 rounded border border-slate-300 dark:border-gray-700">Day<br/>(1-31)</div>
          <div className="bg-slate-50 dark:bg-gray-900/50 p-1 rounded border border-slate-300 dark:border-gray-700">Month<br/>(1-12)</div>
          <div className="bg-slate-50 dark:bg-gray-900/50 p-1 rounded border border-slate-300 dark:border-gray-700">WeekDay<br/>(0=Sun)</div>
        </div>
      </div>

      {res && (
        <div className={`p-6 rounded-lg border shadow-sm ${res.error ? 'bg-red-50 dark:bg-red-900/20 border-red-300 dark:border-red-800' : 'bg-white dark:bg-gray-800 border-slate-300 dark:border-gray-700'}`}>
          {res.error ? (
            <div className="text-red-400">{res.error}</div>
          ) : (
            <div className="space-y-4">
              <div>
                <div className="text-sm text-slate-500 dark:text-gray-400 uppercase font-bold mb-1">Human Description</div>
                <div className="text-kawa-700 dark:text-kawa-300 font-bold text-lg">{res.desc}</div>
              </div>
              
              <div>
                <div className="text-sm text-slate-500 dark:text-gray-400 uppercase font-bold mb-2">Next 5 Executions</div>
                <div className="space-y-2 font-mono text-sm text-slate-700 dark:text-gray-300">
                  {res.next.map((t, i) => (
                    <div key={i} className="flex items-center gap-4 bg-slate-50 dark:bg-black/30 px-3 py-2 rounded border border-slate-100 dark:border-transparent">
                      <span className="text-gray-500">Run #{i+1}</span>
                      <span>{t}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
