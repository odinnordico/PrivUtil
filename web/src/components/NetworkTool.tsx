import { useState, useCallback, useRef, useEffect } from 'react';
import { client } from '../lib/client';
import { Copy, Check, RefreshCw, Shield, Network, Globe, Hash } from 'lucide-react';
import { cn } from '../lib/utils';

// ─── shared helpers ────────────────────────────────────────────────────────────

function useCopy() {
  const [copied, setCopied] = useState<string | null>(null);
  const copy = useCallback(async (text: string, key: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 1500);
  }, []);
  return { copied, copy };
}

function CopyBtn({ text, id, copied, copy, className }: {
  text: string; id: string; copied: string | null;
  copy: (t: string, k: string) => void; className?: string;
}) {
  return (
    <button
      onClick={() => copy(text, id)}
      disabled={!text}
      className={cn(
        'flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 disabled:opacity-30 transition-colors',
        className
      )}
    >
      {copied === id ? <><Check className="w-3.5 h-3.5" /> Copied</> : <><Copy className="w-3.5 h-3.5" /> Copy</>}
    </button>
  );
}

function OutputCard({ label, value, id, copied, copy }: {
  label: string; value: string; id: string; copied: string | null;
  copy: (t: string, k: string) => void;
}) {
  return (
    <div className="bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-3">
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{label}</span>
        <CopyBtn text={value} id={id} copied={copied} copy={copy} />
      </div>
      <code className="text-sm font-mono text-slate-800 dark:text-slate-200 break-all">{value || '—'}</code>
    </div>
  );
}

const tabs = [
  { id: 'chmod',   label: 'chmod',        icon: Shield },
  { id: 'ipv4',    label: 'IPv4 Convert', icon: Globe },
  { id: 'range',   label: 'IPv4 Range',   icon: Network },
  { id: 'port',    label: 'Port Gen',     icon: Hash },
  { id: 'mac',     label: 'MAC Gen',      icon: Network },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const selectClass = "bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 text-sm";

// ─── chmod ────────────────────────────────────────────────────────────────────

const PERM_ROWS = [
  { label: 'Owner', prefix: 'owner' as const },
  { label: 'Group', prefix: 'group' as const },
  { label: 'Others', prefix: 'other' as const },
];

type ChmodBits = {
  ownerRead: boolean; ownerWrite: boolean; ownerExecute: boolean;
  groupRead: boolean; groupWrite: boolean; groupExecute: boolean;
  otherRead: boolean; otherWrite: boolean; otherExecute: boolean;
  setuid: boolean; setgid: boolean; sticky: boolean;
};

function bitsToOctal(b: ChmodBits): string {
  let special = 0;
  if (b.setuid) special += 4;
  if (b.setgid) special += 2;
  if (b.sticky) special += 1;
  const owner = (b.ownerRead ? 4 : 0) + (b.ownerWrite ? 2 : 0) + (b.ownerExecute ? 1 : 0);
  const group = (b.groupRead ? 4 : 0) + (b.groupWrite ? 2 : 0) + (b.groupExecute ? 1 : 0);
  const other = (b.otherRead ? 4 : 0) + (b.otherWrite ? 2 : 0) + (b.otherExecute ? 1 : 0);
  const val = special > 0
    ? `${special}${owner}${group}${other}`
    : `${owner}${group}${other}`;
  return val;
}

function ChmodTab() {
  const [bits, setBits] = useState<ChmodBits>({
    ownerRead: true, ownerWrite: true, ownerExecute: true,
    groupRead: true, groupWrite: false, groupExecute: true,
    otherRead: true, otherWrite: false, otherExecute: true,
    setuid: false, setgid: false, sticky: false,
  });
  const [textInput, setTextInput] = useState('');
  const [result, setResult] = useState<{
    octal: string; symbolic: string; value: number; description: string; error?: string;
  } | null>(null);
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { copied, copy } = useCopy();

  const callServer = useCallback((input: string) => {
    if (debounce.current) clearTimeout(debounce.current);
    debounce.current = setTimeout(async () => {
      try {
        const resp = await client.chmodCalc({ input } as Parameters<typeof client.chmodCalc>[0]);
        if (resp.error) {
          setResult({ octal: '', symbolic: '', value: 0, description: '', error: resp.error });
        } else {
          setResult({
            octal: resp.octal,
            symbolic: resp.symbolic,
            value: resp.value,
            description: resp.description,
          });
          // Sync checkboxes from server response
          setBits({
            ownerRead: resp.ownerRead, ownerWrite: resp.ownerWrite, ownerExecute: resp.ownerExecute,
            groupRead: resp.groupRead, groupWrite: resp.groupWrite, groupExecute: resp.groupExecute,
            otherRead: resp.otherRead, otherWrite: resp.otherWrite, otherExecute: resp.otherExecute,
            setuid: resp.setuid, setgid: resp.setgid, sticky: resp.sticky,
          });
        }
      } catch {
        setResult({ octal: '', symbolic: '', value: 0, description: '', error: 'Server error' });
      }
    }, 150);
  }, []);

  useEffect(() => () => { if (debounce.current) clearTimeout(debounce.current); }, []);

  const handleCheckbox = (key: keyof ChmodBits) => {
    const newBits = { ...bits, [key]: !bits[key] };
    setBits(newBits);
    setTextInput('');
    callServer(bitsToOctal(newBits));
  };

  const handleTextInput = (v: string) => {
    setTextInput(v);
    if (v.trim()) callServer(v.trim());
  };

  // Run once on mount
  useEffect(() => { callServer(bitsToOctal(bits)); }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const permKeys: Record<string, [keyof ChmodBits, keyof ChmodBits, keyof ChmodBits]> = {
    owner: ['ownerRead', 'ownerWrite', 'ownerExecute'],
    group: ['groupRead', 'groupWrite', 'groupExecute'],
    other: ['otherRead', 'otherWrite', 'otherExecute'],
  };

  return (
    <div className="space-y-5">
      {/* Grid */}
      <div className="bg-white dark:bg-neutral-800/50 rounded-lg border border-slate-300 dark:border-neutral-700 overflow-hidden shadow-sm">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200 dark:border-neutral-700">
              <th className="text-left px-4 py-2.5 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase w-20">Who</th>
              {['Read', 'Write', 'Execute'].map(h => (
                <th key={h} className="text-center px-4 py-2.5 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">{h}</th>
              ))}
              <th className="text-center px-4 py-2.5 text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">Value</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-neutral-700/50">
            {PERM_ROWS.map(({ label, prefix }) => {
              const [rk, wk, xk] = permKeys[prefix];
              const val = (bits[rk] ? 4 : 0) + (bits[wk] ? 2 : 0) + (bits[xk] ? 1 : 0);
              return (
                <tr key={prefix} className="hover:bg-slate-50 dark:hover:bg-neutral-800 transition-colors">
                  <td className="px-4 py-3 font-medium text-slate-700 dark:text-slate-300">{label}</td>
                  {[rk, wk, xk].map(k => (
                    <td key={k} className="text-center px-4 py-3">
                      <input
                        type="checkbox"
                        checked={bits[k]}
                        onChange={() => handleCheckbox(k)}
                        className="w-4 h-4 accent-kawa-500 cursor-pointer"
                      />
                    </td>
                  ))}
                  <td className="text-center px-4 py-3 font-mono font-bold text-slate-700 dark:text-slate-200">{val}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {/* Special bits */}
        <div className="flex gap-6 px-4 py-3 border-t border-slate-200 dark:border-neutral-700 bg-slate-50 dark:bg-neutral-800">
          {[
            { key: 'setuid' as keyof ChmodBits, label: 'setuid (4000)' },
            { key: 'setgid' as keyof ChmodBits, label: 'setgid (2000)' },
            { key: 'sticky' as keyof ChmodBits, label: 'sticky (1000)' },
          ].map(({ key, label }) => (
            <label key={key} className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={bits[key]}
                onChange={() => handleCheckbox(key)}
                className="w-4 h-4 accent-kawa-500"
              />
              {label}
            </label>
          ))}
        </div>
      </div>

      {/* Text input */}
      <div className="flex items-center gap-3">
        <span className="text-sm text-slate-500 dark:text-slate-400 whitespace-nowrap">Or enter directly</span>
        <input
          type="text"
          value={textInput}
          onChange={e => handleTextInput(e.target.value)}
          placeholder='octal "755" or symbolic "rwxr-xr-x"'
          className={inputClass}
        />
      </div>

      {/* Output */}
      {result && !result.error && (
        <div className="grid grid-cols-2 gap-3">
          <div className="bg-kawa-500/10 dark:bg-kawa-500/20 rounded-xl border border-kawa-400/30 p-4 text-center">
            <div className="text-4xl font-mono font-black text-kawa-600 dark:text-kawa-400 mb-1">{result.octal}</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Octal</div>
            <CopyBtn text={result.octal} id="octal" copied={copied} copy={copy} className="justify-center mt-2" />
          </div>
          <div className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-4 text-center">
            <div className="text-2xl font-mono font-bold text-slate-700 dark:text-slate-200 mb-1">{result.symbolic}</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Symbolic</div>
            <CopyBtn text={result.symbolic} id="sym" copied={copied} copy={copy} className="justify-center mt-2" />
          </div>
          <div className="col-span-2 bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 px-4 py-3 text-sm text-slate-600 dark:text-slate-400">
            {result.description || 'No permissions'}
          </div>
        </div>
      )}
      {result?.error && (
        <div className="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">
          {result.error}
        </div>
      )}
    </div>
  );
}

// ─── IPv4 converter ───────────────────────────────────────────────────────────

function Ipv4ConvertTab() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState<{ dotted: string; decimal: string; hex: string; binary: string } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);

  const call = useCallback((v: string) => {
    if (debounce.current) clearTimeout(debounce.current);
    if (!v.trim()) { setResult(null); setError(''); return; }
    debounce.current = setTimeout(async () => {
      try {
        const resp = await client.ipv4Convert({ input: v.trim() } as Parameters<typeof client.ipv4Convert>[0]);
        if (resp.error) { setError(resp.error); setResult(null); }
        else { setError(''); setResult({ dotted: resp.dotted, decimal: resp.decimal, hex: resp.hex, binary: resp.binary }); }
      } catch { setError('Server error'); }
    }, 300);
  }, []);

  useEffect(() => () => { if (debounce.current) clearTimeout(debounce.current); }, []);

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">
          Input — any format
        </label>
        <input
          type="text"
          value={input}
          onChange={e => { setInput(e.target.value); call(e.target.value); }}
          placeholder='192.168.1.1 · 3232235777 · 0xC0A80101 · 11000000.10101000.00000001.00000001'
          className={inputClass}
          spellCheck={false}
        />
        <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
          Accepts dotted decimal, integer, hex (0x…), binary (dotted or plain 32-bit)
        </p>
      </div>

      {error && (
        <div className="text-sm text-red-500 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">{error}</div>
      )}

      {result && (
        <div className="grid grid-cols-2 gap-3">
          <OutputCard label="Dotted Decimal" value={result.dotted} id="d" copied={copied} copy={copy} />
          <OutputCard label="Decimal Integer" value={result.decimal} id="dec" copied={copied} copy={copy} />
          <OutputCard label="Hexadecimal" value={result.hex} id="hex" copied={copied} copy={copy} />
          <div className="col-span-2">
            <OutputCard label="Binary" value={result.binary} id="bin" copied={copied} copy={copy} />
          </div>
        </div>
      )}
    </div>
  );
}

// ─── IPv4 range expander ──────────────────────────────────────────────────────

function Ipv4RangeTab() {
  const [start, setStart] = useState('');
  const [end, setEnd] = useState('');
  const [result, setResult] = useState<{ addresses: string[]; cidrs: string[]; total: number } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (s: string, e: string) => {
    if (!s.trim() || !e.trim()) { setResult(null); setError(''); return; }
    try {
      const resp = await client.ipv4RangeExpand({ start: s.trim(), end: e.trim() } as Parameters<typeof client.ipv4RangeExpand>[0]);
      if (resp.error) { setError(resp.error); setResult(null); }
      else {
        setError('');
        setResult({
          addresses: resp.addresses as string[],
          cidrs: resp.cidrs as string[],
          total: Number(resp.total),
        });
      }
    } catch { setError('Server error'); }
  }, []);

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Start IP</label>
          <input type="text" value={start} onChange={e => { setStart(e.target.value); call(e.target.value, end); }}
            placeholder="10.0.0.0" className={inputClass} spellCheck={false} />
        </div>
        <div>
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">End IP</label>
          <input type="text" value={end} onChange={e => { setEnd(e.target.value); call(start, e.target.value); }}
            placeholder="10.0.0.255" className={inputClass} spellCheck={false} />
        </div>
      </div>

      {error && (
        <div className="text-sm text-red-500 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">{error}</div>
      )}

      {result && (
        <div className="space-y-3">
          {/* Summary */}
          <div className="flex items-center gap-4 bg-kawa-500/10 dark:bg-kawa-500/20 rounded-lg border border-kawa-400/30 px-4 py-3">
            <div className="text-center">
              <div className="text-2xl font-mono font-black text-kawa-600 dark:text-kawa-400">{result.total.toLocaleString()}</div>
              <div className="text-xs text-slate-500 dark:text-slate-400">Total IPs</div>
            </div>
            <div className="h-10 w-px bg-kawa-400/30" />
            <div className="flex-1">
              <div className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">CIDR Summary</div>
              <div className="flex flex-wrap gap-1.5">
                {result.cidrs.map(cidr => (
                  <button
                    key={cidr}
                    onClick={() => copy(cidr, `cidr-${cidr}`)}
                    className="font-mono text-xs bg-white dark:bg-neutral-800 border border-slate-200 dark:border-neutral-600 rounded px-2 py-0.5 text-slate-700 dark:text-slate-300 hover:border-kawa-400 transition-colors"
                  >
                    {copied === `cidr-${cidr}` ? '✓' : cidr}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* IP list */}
          {result.addresses.length > 0 ? (
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                  Individual IPs ({result.addresses.length})
                </span>
                <CopyBtn text={result.addresses.join('\n')} id="allips" copied={copied} copy={copy} />
              </div>
              <div className="h-48 overflow-y-auto bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-3 font-mono text-sm text-slate-700 dark:text-slate-300 grid grid-cols-4 gap-1 content-start">
                {result.addresses.map(ip => (
                  <button
                    key={ip}
                    onClick={() => copy(ip, `ip-${ip}`)}
                    className="text-left px-1 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
                  >
                    {copied === `ip-${ip}` ? '✓' : ip}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-500 dark:text-slate-400 italic">
              Range contains {result.total.toLocaleString()} IPs — individual listing capped at 256. Use the CIDR summary above.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

// ─── port generator ───────────────────────────────────────────────────────────

function PortGenTab() {
  const [count, setCount] = useState(10);
  const [min, setMin] = useState(1024);
  const [max, setMax] = useState(65535);
  const [excludeSys, setExcludeSys] = useState(true);
  const [ports, setPorts] = useState<number[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { copied, copy } = useCopy();

  const generate = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const resp = await client.generatePort({
        count,
        min,
        max,
        excludeSystem: excludeSys,
      } as Parameters<typeof client.generatePort>[0]);
      if (resp.error) { setError(resp.error); setPorts([]); }
      else { setPorts(resp.ports as number[]); }
    } catch { setError('Server error'); }
    finally { setLoading(false); }
  }, [count, min, max, excludeSys]);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-end bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="w-24">
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Count</label>
          <input type="number" min={1} max={100} value={count} onChange={e => setCount(Math.min(100, Math.max(1, +e.target.value)))} className={inputClass} />
        </div>
        <div className="w-28">
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Min port</label>
          <input type="number" min={0} max={65535} value={min} onChange={e => setMin(+e.target.value)} className={inputClass} />
        </div>
        <div className="w-28">
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Max port</label>
          <input type="number" min={0} max={65535} value={max} onChange={e => setMax(+e.target.value)} className={inputClass} />
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none pb-2">
          <input type="checkbox" checked={excludeSys} onChange={e => setExcludeSys(e.target.checked)} className="w-4 h-4 accent-kawa-500" />
          Exclude system ports (0–1023)
        </label>
        <button
          onClick={generate}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold rounded-lg shadow-md transition-all active:scale-95 disabled:opacity-50 ml-auto"
        >
          <RefreshCw className={cn("w-4 h-4", loading && "animate-spin")} />
          Generate
        </button>
      </div>

      {error && (
        <div className="text-sm text-red-500 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">{error}</div>
      )}

      {ports.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{ports.length} ports</span>
            <CopyBtn text={ports.join('\n')} id="allports" copied={copied} copy={copy} />
          </div>
          <div className="flex flex-wrap gap-2">
            {ports.map(p => (
              <button
                key={p}
                onClick={() => copy(String(p), `p-${p}`)}
                className="font-mono text-sm bg-white dark:bg-neutral-800 border border-slate-200 dark:border-neutral-600 rounded-lg px-3 py-1.5 text-slate-700 dark:text-slate-200 hover:border-kawa-400 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
              >
                {copied === `p-${p}` ? '✓' : p}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── MAC address generator ────────────────────────────────────────────────────

function MacGenTab() {
  const [count, setCount] = useState(5);
  const [sep, setSep] = useState(':');
  const [upper, setUpper] = useState(false);
  const [oui, setOui] = useState('');
  const [unicast, setUnicast] = useState(true);
  const [local, setLocal] = useState(false);
  const [macs, setMacs] = useState<string[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { copied, copy } = useCopy();

  const generate = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const resp = await client.generateMac({
        count,
        separator: sep,
        uppercase: upper,
        oui,
        unicast,
        local,
      } as Parameters<typeof client.generateMac>[0]);
      if (resp.error) { setError(resp.error); setMacs([]); }
      else { setMacs(resp.addresses as string[]); }
    } catch { setError('Server error'); }
    finally { setLoading(false); }
  }, [count, sep, upper, oui, unicast, local]);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-end bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="w-24">
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Count</label>
          <input type="number" min={1} max={100} value={count} onChange={e => setCount(Math.min(100, Math.max(1, +e.target.value)))} className={inputClass} />
        </div>
        <div>
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Separator</label>
          <select value={sep} onChange={e => setSep(e.target.value)} className={selectClass}>
            <option value=":">Colon  AA:BB:CC</option>
            <option value="-">Dash   AA-BB-CC</option>
            <option value=".">Dot (Cisco)</option>
            <option value="">None   AABBCC</option>
          </select>
        </div>
        <div className="w-36">
          <label className="block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">OUI Prefix</label>
          <input type="text" value={oui} onChange={e => setOui(e.target.value)} placeholder="00:1A:2B" className={inputClass} spellCheck={false} />
        </div>
        <div className="flex flex-col gap-2 pb-1">
          <label className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none">
            <input type="checkbox" checked={upper} onChange={e => setUpper(e.target.checked)} className="w-4 h-4 accent-kawa-500" />
            Uppercase
          </label>
          <label className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none">
            <input type="checkbox" checked={unicast} onChange={e => setUnicast(e.target.checked)} className="w-4 h-4 accent-kawa-500" />
            Force unicast
          </label>
          <label className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none">
            <input type="checkbox" checked={local} onChange={e => setLocal(e.target.checked)} className="w-4 h-4 accent-kawa-500" />
            Locally administered
          </label>
        </div>
        <button
          onClick={generate}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold rounded-lg shadow-md transition-all active:scale-95 disabled:opacity-50 ml-auto self-end"
        >
          <RefreshCw className={cn("w-4 h-4", loading && "animate-spin")} />
          Generate
        </button>
      </div>

      {error && (
        <div className="text-sm text-red-500 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">{error}</div>
      )}

      {macs.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{macs.length} addresses</span>
            <CopyBtn text={macs.join('\n')} id="allmacs" copied={copied} copy={copy} />
          </div>
          <div className="space-y-1.5">
            {macs.map((mac, i) => (
              <div key={i} className="flex items-center justify-between bg-white dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 px-4 py-2">
                <code className="font-mono text-sm text-slate-800 dark:text-slate-200">{mac}</code>
                <CopyBtn text={mac} id={`mac-${i}`} copied={copied} copy={copy} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── main component ───────────────────────────────────────────────────────────

export function NetworkTool() {
  const [activeTab, setActiveTab] = useState<TabId>('chmod');

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Network & System</h2>

      {/* Tab bar */}
      <div className="flex flex-wrap gap-1 bg-white dark:bg-neutral-800/50 p-1 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        {tabs.map(tab => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 rounded text-sm font-medium transition-colors',
                activeTab === tab.id
                  ? 'bg-kawa-500 text-slate-900'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-700'
              )}
            >
              <Icon className="w-3.5 h-3.5" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <div className="bg-white dark:bg-neutral-900/30 rounded-lg border border-slate-200 dark:border-neutral-700 p-5 shadow-sm">
        {activeTab === 'chmod'  && <ChmodTab />}
        {activeTab === 'ipv4'   && <Ipv4ConvertTab />}
        {activeTab === 'range'  && <Ipv4RangeTab />}
        {activeTab === 'port'   && <PortGenTab />}
        {activeTab === 'mac'    && <MacGenTab />}
      </div>
    </div>
  );
}
