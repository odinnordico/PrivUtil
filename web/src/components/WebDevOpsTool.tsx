import { useState, useCallback, useEffect, useRef } from 'react';
import { client } from '../lib/client';
import { Copy, Check, Globe, Cpu, Hash, File, Container, GitBranch, Search, ChevronDown, ChevronUp } from 'lucide-react';
import { cn } from '../lib/utils';

// ─── Shared helpers ────────────────────────────────────────────────────────────

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
      {copied === id ? <Check size={13} className="text-emerald-500" /> : <Copy size={13} />}
    </button>
  );
}

function useDebounce<T>(value: T, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);
  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(handler);
  }, [value, delay]);
  return debouncedValue;
}

// ─── Tabs ──────────────────────────────────────────────────────────────────────

const tabs = [
  { id: 'url',     label: 'URL Parser',       icon: Globe      },
  { id: 'ua',      label: 'User-Agent',        icon: Cpu        },
  { id: 'status',  label: 'HTTP Status',       icon: Hash       },
  { id: 'mime',    label: 'MIME Types',        icon: File       },
  { id: 'docker',  label: 'Docker → Compose',  icon: Container  },
  { id: 'git',     label: 'Git Cheat Sheet',   icon: GitBranch  },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const labelClass = "block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1";

// ─── URL Parser ────────────────────────────────────────────────────────────────

type UrlResult = {
  scheme: string; username: string; password: string;
  host: string; hostname: string; port: string;
  path: string; query: string; fragment: string; normalized: string;
  isValid: boolean; error: string;
  queryParams: Array<{ key: string; value: string }>;
};

const SAMPLE_URLS = [
  'https://user:pass@api.example.com:8443/v1/search?q=hello+world&lang=en&page=2#results',
  'ftp://files.example.com/pub/downloads/archive.tar.gz',
  'postgres://admin:secret@db.internal:5432/myapp?sslmode=require&connect_timeout=10',
  'mailto:user@example.com?subject=Hello&body=World',
];

function UrlParserTab() {
  const [url, setUrl] = useState('https://user:pass@api.example.com:8443/v1/search?q=hello+world&lang=en&page=2#results');
  const [result, setResult] = useState<UrlResult | null>(null);
  const debouncedUrl = useDebounce(url, 300);
  const { copied, copy } = useCopy();

  useEffect(() => {
    if (!debouncedUrl.trim()) { setResult(null); return; }
    client.urlParse({ url: debouncedUrl } as Parameters<typeof client.urlParse>[0]).then(res => {
      setResult({
        scheme: res.scheme, username: res.username, password: res.password,
        host: res.host, hostname: res.hostname, port: res.port,
        path: res.path, query: res.query, fragment: res.fragment,
        normalized: res.normalized, isValid: res.isValid, error: res.error,
        queryParams: res.queryParams.map(p => ({ key: p.key, value: p.value })),
      });
    }).catch(() => {});
  }, [debouncedUrl]);

  const rows = result ? [
    { label: 'Scheme',    value: result.scheme,    id: 'url-scheme' },
    { label: 'Username',  value: result.username,  id: 'url-user'   },
    { label: 'Password',  value: result.password,  id: 'url-pass'   },
    { label: 'Hostname',  value: result.hostname,  id: 'url-host'   },
    { label: 'Port',      value: result.port,      id: 'url-port'   },
    { label: 'Path',      value: result.path,      id: 'url-path'   },
    { label: 'Fragment',  value: result.fragment,  id: 'url-frag'   },
    { label: 'Normalized',value: result.normalized,id: 'url-norm'   },
  ].filter(r => r.value) : [];

  return (
    <div className="space-y-4">
      <div>
        <label className={labelClass}>URL to parse</label>
        <textarea
          value={url}
          onChange={e => setUrl(e.target.value)}
          rows={2}
          placeholder="https://example.com/path?key=value#fragment"
          className={cn(inputClass, 'font-mono resize-none')}
        />
        <div className="flex flex-wrap gap-1 mt-1.5">
          {SAMPLE_URLS.map((s, i) => (
            <button key={i} onClick={() => setUrl(s)}
              className="text-xs px-2 py-0.5 rounded bg-slate-100 dark:bg-neutral-700 text-slate-600 dark:text-slate-300 hover:bg-kawa-100 dark:hover:bg-kawa-900 transition-colors font-mono truncate max-w-[200px]">
              {s.substring(0, 40)}…
            </button>
          ))}
        </div>
      </div>

      {result?.error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{result.error}</div>
      )}

      {result?.isValid && (
        <div className="space-y-3">
          <div className="overflow-hidden rounded-xl border border-slate-200 dark:border-neutral-700">
            <table className="w-full text-sm">
              <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                {rows.map(r => (
                  <tr key={r.id} className="hover:bg-slate-50 dark:hover:bg-neutral-800/50">
                    <td className="px-4 py-2 text-xs font-bold text-slate-400 uppercase tracking-wide w-28 whitespace-nowrap">{r.label}</td>
                    <td className="px-4 py-2 font-mono text-slate-800 dark:text-slate-200 break-all">{r.value}</td>
                    <td className="px-3 py-2 w-10">
                      <CopyBtn text={r.value} id={r.id} copied={copied} copy={copy} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {result.queryParams.length > 0 && (
            <div>
              <p className={labelClass}>Query Parameters ({result.queryParams.length})</p>
              <div className="overflow-hidden rounded-xl border border-slate-200 dark:border-neutral-700">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 dark:bg-neutral-800">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase w-1/3">Key</th>
                      <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase">Value</th>
                      <th className="w-10" />
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                    {result.queryParams.map((p, i) => (
                      <tr key={i} className="hover:bg-slate-50 dark:hover:bg-neutral-800/50">
                        <td className="px-4 py-2 font-mono text-kawa-600 dark:text-kawa-400 break-all">{p.key}</td>
                        <td className="px-4 py-2 font-mono text-slate-700 dark:text-slate-300 break-all">{p.value}</td>
                        <td className="px-3 py-2">
                          <CopyBtn text={p.value} id={`qp-${i}`} copied={copied} copy={copy} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ─── User-Agent Parser ────────────────────────────────────────────────────────

const SAMPLE_UAS = [
  { label: 'Chrome / Windows', value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36' },
  { label: 'Firefox / Linux', value: 'Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0' },
  { label: 'Safari / macOS', value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15' },
  { label: 'Edge / Windows', value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0' },
  { label: 'Chrome / Android', value: 'Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36' },
  { label: 'Googlebot', value: 'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)' },
  { label: 'curl', value: 'curl/8.4.0' },
];

type UAResult = {
  browserName: string; browserVersion: string;
  osName: string; osVersion: string;
  deviceType: string; engine: string; engineVersion: string;
  isBot: boolean; isMobile: boolean;
  fields: Array<{ label: string; value: string }>;
};

function deviceIcon(type: string) {
  switch (type) {
    case 'mobile': return '📱';
    case 'tablet': return '📟';
    case 'bot':    return '🤖';
    default:       return '🖥️';
  }
}

function UserAgentTab() {
  const [ua, setUa] = useState(SAMPLE_UAS[0].value);
  const [result, setResult] = useState<UAResult | null>(null);
  const [error, setError] = useState('');
  const debouncedUa = useDebounce(ua, 300);

  useEffect(() => {
    if (!debouncedUa.trim()) { setResult(null); setError(''); return; }
    client.userAgentParse({ userAgent: debouncedUa } as Parameters<typeof client.userAgentParse>[0])
      .then(res => {
        if (res.error) { setError(res.error); setResult(null); return; }
        setError('');
        setResult({
          browserName: res.browserName, browserVersion: res.browserVersion,
          osName: res.osName, osVersion: res.osVersion,
          deviceType: res.deviceType, engine: res.engine, engineVersion: res.engineVersion,
          isBot: res.isBot, isMobile: res.isMobile,
          fields: res.fields.map(f => ({ label: f.label, value: f.value })),
        });
      }).catch(() => {});
  }, [debouncedUa]);

  return (
    <div className="space-y-4">
      <div>
        <label className={labelClass}>User-Agent string</label>
        <textarea
          value={ua}
          onChange={e => setUa(e.target.value)}
          rows={3}
          placeholder="Mozilla/5.0 ..."
          className={cn(inputClass, 'font-mono resize-none text-xs')}
        />
        <div className="flex flex-wrap gap-1 mt-1.5">
          {SAMPLE_UAS.map((s, i) => (
            <button key={i} onClick={() => setUa(s.value)}
              className="text-xs px-2 py-0.5 rounded bg-slate-100 dark:bg-neutral-700 text-slate-600 dark:text-slate-300 hover:bg-kawa-100 dark:hover:bg-kawa-900 transition-colors">
              {s.label}
            </button>
          ))}
        </div>
      </div>

      {error && <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{error}</div>}

      {result && (
        <div className="space-y-3">
          {/* Summary cards */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {[
              { label: 'Browser', value: `${result.browserName} ${result.browserVersion}`.trim(), sub: result.engine ? `${result.engine} ${result.engineVersion}`.trim() : undefined },
              { label: 'OS', value: `${result.osName} ${result.osVersion}`.trim() || '—' },
              { label: 'Device', value: `${deviceIcon(result.deviceType)} ${result.deviceType}` },
              { label: 'Type', value: result.isBot ? '🤖 Bot' : result.isMobile ? '📱 Mobile' : '🖥️ Desktop' },
            ].map((card, i) => (
              <div key={i} className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3 text-center">
                <div className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-1">{card.label}</div>
                <div className="font-semibold text-slate-800 dark:text-slate-200 text-sm">{card.value || '—'}</div>
                {card.sub && <div className="text-xs text-slate-400 mt-0.5">{card.sub}</div>}
              </div>
            ))}
          </div>

          {/* Raw fields */}
          <div className="overflow-hidden rounded-xl border border-slate-200 dark:border-neutral-700">
            <table className="w-full text-sm">
              <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                {result.fields.map((f, i) => (
                  <tr key={i} className="hover:bg-slate-50 dark:hover:bg-neutral-800/50">
                    <td className="px-4 py-2 text-xs font-bold text-slate-400 uppercase tracking-wide w-28">{f.label}</td>
                    <td className="px-4 py-2 text-slate-700 dark:text-slate-300 font-mono text-xs">{f.value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

// ─── HTTP Status Codes ────────────────────────────────────────────────────────

const STATUS_CATEGORIES = ['All', '1xx', '2xx', '3xx', '4xx', '5xx'];
const STATUS_COLORS: Record<string, string> = {
  '1xx': 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300',
  '2xx': 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300',
  '3xx': 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800 text-amber-700 dark:text-amber-300',
  '4xx': 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300',
  '5xx': 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-300',
};
const CODE_COLORS: Record<string, string> = {
  '1xx': 'text-blue-600 dark:text-blue-400',
  '2xx': 'text-emerald-600 dark:text-emerald-400',
  '3xx': 'text-amber-600 dark:text-amber-400',
  '4xx': 'text-orange-600 dark:text-orange-400',
  '5xx': 'text-red-600 dark:text-red-400',
};

type StatusEntry = { code: number; name: string; description: string; category: string };

function HttpStatusTab() {
  const [search, setSearch] = useState('');
  const [cat, setCat] = useState('All');
  const [entries, setEntries] = useState<StatusEntry[]>([]);
  const [expanded, setExpanded] = useState<number | null>(null);
  const debouncedSearch = useDebounce(search, 250);

  useEffect(() => {
    client.httpStatusSearch({
      query: debouncedSearch,
      category: cat === 'All' ? '' : cat,
    } as Parameters<typeof client.httpStatusSearch>[0]).then(res => {
      setEntries(res.entries.map(e => ({
        code: e.code, name: e.name, description: e.description, category: e.category,
      })));
    }).catch(() => {});
  }, [debouncedSearch, cat]);

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-2">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Search code, name, or description…"
            className={cn(inputClass, 'pl-8')}
          />
        </div>
        <div className="flex gap-1 flex-wrap">
          {STATUS_CATEGORIES.map(c => (
            <button key={c} onClick={() => setCat(c)}
              className={cn(
                'px-3 py-2 rounded-lg text-sm font-medium transition-colors border',
                cat === c
                  ? 'bg-kawa-600 text-white border-kawa-600'
                  : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700 hover:border-kawa-400'
              )}>
              {c}
            </button>
          ))}
        </div>
      </div>

      <div className="text-xs text-slate-400">{entries.length} result{entries.length !== 1 ? 's' : ''}</div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        {entries.map(e => (
          <button
            key={e.code}
            onClick={() => setExpanded(expanded === e.code ? null : e.code)}
            className={cn(
              'text-left rounded-xl border p-3 transition-all',
              STATUS_COLORS[e.category] ?? 'bg-slate-50 dark:bg-neutral-800 border-slate-200 dark:border-neutral-700'
            )}>
            <div className="flex items-center justify-between">
              <span className={cn('text-2xl font-bold font-mono', CODE_COLORS[e.category])}>{e.code}</span>
              {expanded === e.code ? <ChevronUp size={14} className="opacity-50" /> : <ChevronDown size={14} className="opacity-50" />}
            </div>
            <div className="font-semibold text-sm mt-0.5">{e.name}</div>
            {expanded === e.code && (
              <p className="text-xs mt-2 leading-relaxed opacity-80">{e.description}</p>
            )}
          </button>
        ))}
      </div>
    </div>
  );
}

// ─── MIME Types ───────────────────────────────────────────────────────────────

const MIME_CAT_COLORS: Record<string, string> = {
  Text:        'bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-300',
  Image:       'bg-violet-100 dark:bg-violet-900/30 text-violet-700 dark:text-violet-300',
  Audio:       'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300',
  Video:       'bg-rose-100 dark:bg-rose-900/30 text-rose-700 dark:text-rose-300',
  Application: 'bg-slate-100 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300',
  Font:        'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300',
  Archive:     'bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300',
  Feed:        'bg-teal-100 dark:bg-teal-900/30 text-teal-700 dark:text-teal-300',
  Web:         'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300',
};

type MimeEntry = { mimeType: string; extensions: string; category: string; description: string };

function MimeTab() {
  const [query, setQuery] = useState('');
  const [entries, setEntries] = useState<MimeEntry[]>([]);
  const { copied, copy } = useCopy();
  const debouncedQ = useDebounce(query, 250);

  useEffect(() => {
    client.mimeLookup({ query: debouncedQ } as Parameters<typeof client.mimeLookup>[0]).then(res => {
      setEntries(res.entries.map(e => ({
        mimeType: e.mimeType, extensions: e.extensions,
        category: e.category, description: e.description,
      })));
    }).catch(() => {});
  }, [debouncedQ]);

  return (
    <div className="space-y-4">
      <div>
        <label className={labelClass}>Search by extension, MIME type, or category</label>
        <div className="relative">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="pdf, image/png, audio, .js, font…"
            className={cn(inputClass, 'pl-8')}
          />
        </div>
      </div>

      <div className="text-xs text-slate-400">{entries.length} type{entries.length !== 1 ? 's' : ''} found</div>

      <div className="overflow-hidden rounded-xl border border-slate-200 dark:border-neutral-700">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 dark:bg-neutral-800">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase">MIME Type</th>
              <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase w-28">Extensions</th>
              <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase w-24 hidden sm:table-cell">Category</th>
              <th className="px-4 py-2 text-left text-xs font-bold text-slate-500 uppercase hidden md:table-cell">Description</th>
              <th className="w-10" />
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
            {entries.map((e, i) => (
              <tr key={i} className="hover:bg-slate-50 dark:hover:bg-neutral-800/50">
                <td className="px-4 py-2 font-mono text-xs text-slate-800 dark:text-slate-200 break-all">{e.mimeType}</td>
                <td className="px-4 py-2 text-xs text-slate-500 dark:text-slate-400 font-mono">{e.extensions || '—'}</td>
                <td className="px-4 py-2 hidden sm:table-cell">
                  <span className={cn('text-xs px-2 py-0.5 rounded-full font-medium', MIME_CAT_COLORS[e.category] ?? 'bg-slate-100 text-slate-600')}>
                    {e.category}
                  </span>
                </td>
                <td className="px-4 py-2 text-xs text-slate-500 dark:text-slate-400 hidden md:table-cell">{e.description}</td>
                <td className="px-3 py-2">
                  <CopyBtn text={e.mimeType} id={`mime-${i}`} copied={copied} copy={copy} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── Docker run → Compose ─────────────────────────────────────────────────────

const DOCKER_EXAMPLES = [
  {
    label: 'nginx with ports',
    cmd: 'docker run -d --name nginx-web --restart always -p 80:80 -p 443:443 -v /var/www:/usr/share/nginx/html nginx:1.25',
  },
  {
    label: 'postgres with env',
    cmd: 'docker run -d --name postgres-db --restart unless-stopped -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=myapp -v pgdata:/var/lib/postgresql/data -p 5432:5432 postgres:16',
  },
  {
    label: 'redis with memory',
    cmd: 'docker run -d --name redis-cache --restart always --memory 256m -p 6379:6379 redis:7-alpine',
  },
  {
    label: 'full options',
    cmd: 'docker run -d --name myapp --restart unless-stopped -p 8080:8080 -v /app/data:/data -e DATABASE_URL=postgres://localhost/db -e SECRET_KEY=mysecret --network backend --hostname app-server --memory 512m --cpus 1.5 --log-driver json-file --log-opt max-size=10m --log-opt max-file=3 myimage:latest',
  },
];

function DockerTab() {
  const [cmd, setCmd] = useState(DOCKER_EXAMPLES[0].cmd);
  const [result, setResult] = useState<{ yaml: string; warnings: string[]; error: string } | null>(null);
  const { copied, copy } = useCopy();

  const convert = useCallback(() => {
    if (!cmd.trim()) return;
    client.dockerRunToCompose({ command: cmd } as Parameters<typeof client.dockerRunToCompose>[0]).then(res => {
      setResult({ yaml: res.composeYaml, warnings: res.warnings, error: res.error });
    }).catch(() => {});
  }, [cmd]);

  useEffect(() => { convert(); }, [convert]);

  return (
    <div className="space-y-4">
      <div>
        <label className={labelClass}>docker run command</label>
        <textarea
          value={cmd}
          onChange={e => setCmd(e.target.value)}
          rows={4}
          placeholder="docker run -d --name myapp -p 8080:80 nginx:latest"
          className={cn(inputClass, 'font-mono resize-none text-xs')}
        />
        <div className="flex flex-wrap gap-1 mt-1.5">
          {DOCKER_EXAMPLES.map((ex, i) => (
            <button key={i} onClick={() => setCmd(ex.cmd)}
              className="text-xs px-2 py-0.5 rounded bg-slate-100 dark:bg-neutral-700 text-slate-600 dark:text-slate-300 hover:bg-kawa-100 dark:hover:bg-kawa-900 transition-colors">
              {ex.label}
            </button>
          ))}
        </div>
      </div>

      <button onClick={convert}
        className="px-4 py-2 bg-kawa-600 hover:bg-kawa-700 text-white rounded-lg text-sm font-semibold transition-colors">
        Convert to Compose
      </button>

      {result?.error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{result.error}</div>
      )}

      {result?.warnings && result.warnings.length > 0 && (
        <div className="rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 p-3 space-y-1">
          <p className="text-xs font-bold text-amber-600 dark:text-amber-400 uppercase tracking-wide">Warnings</p>
          {result.warnings.map((w, i) => (
            <p key={i} className="text-xs text-amber-700 dark:text-amber-300">⚠ {w}</p>
          ))}
        </div>
      )}

      {result?.yaml && !result.error && (
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <p className={labelClass}>docker-compose.yml</p>
            <CopyBtn text={result.yaml} id="compose-yaml" copied={copied} copy={copy} className="text-sm" />
          </div>
          <pre className="bg-slate-900 text-slate-100 rounded-xl p-4 text-xs font-mono overflow-x-auto whitespace-pre leading-relaxed">
            {result.yaml}
          </pre>
        </div>
      )}
    </div>
  );
}

// ─── Git Cheat Sheet ──────────────────────────────────────────────────────────

const GIT_CATEGORIES = [
  'All', 'Setup & Config', 'Creating & Cloning', 'Staging & Committing',
  'Branching', 'Merging & Rebasing', 'Remote Repositories', 'Inspection & Diff',
  'Stashing', 'Tags', 'Undoing & Resetting', 'Advanced',
];

type GitCmdItem = { command: string; description: string; examples: string[] };
type GitCatItem = { name: string; commands: GitCmdItem[] };

function GitTab() {
  const [search, setSearch] = useState('');
  const [activeCat, setActiveCat] = useState('All');
  const [categories, setCategories] = useState<GitCatItem[]>([]);
  const [expanded, setExpanded] = useState<string | null>(null);
  const { copied, copy } = useCopy();
  const debouncedSearch = useDebounce(search, 250);

  useEffect(() => {
    client.gitCheatSheet({
      query: debouncedSearch,
      category: activeCat === 'All' ? '' : activeCat,
    } as Parameters<typeof client.gitCheatSheet>[0]).then(res => {
      setCategories(res.categories.map(c => ({
        name: c.name,
        commands: c.commands.map(cmd => ({
          command: cmd.command, description: cmd.description, examples: cmd.examples,
        })),
      })));
      if (debouncedSearch) setExpanded('all');
    }).catch(() => {});
  }, [debouncedSearch, activeCat]);

  const totalCmds = categories.reduce((n, c) => n + c.commands.length, 0);

  // Compact category pill row
  const catScrollRef = useRef<HTMLDivElement>(null);

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-2">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Search commands or descriptions…"
            className={cn(inputClass, 'pl-8')}
          />
        </div>
      </div>

      {/* Category pills — scrollable */}
      <div ref={catScrollRef} className="flex gap-1.5 overflow-x-auto pb-1 scrollbar-none">
        {GIT_CATEGORIES.map(c => (
          <button key={c} onClick={() => setActiveCat(c)}
            className={cn(
              'px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors border',
              activeCat === c
                ? 'bg-kawa-600 text-white border-kawa-600'
                : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700 hover:border-kawa-400'
            )}>
            {c}
          </button>
        ))}
      </div>

      <div className="text-xs text-slate-400">{totalCmds} command{totalCmds !== 1 ? 's' : ''}</div>

      <div className="space-y-3">
        {categories.map(cat => {
          const isOpen = expanded === 'all' || expanded === cat.name || (activeCat !== 'All' && categories.length === 1);
          return (
            <div key={cat.name} className="rounded-xl border border-slate-200 dark:border-neutral-700 overflow-hidden">
              <button
                onClick={() => setExpanded(isOpen && expanded !== 'all' ? null : cat.name)}
                className="w-full flex items-center justify-between px-4 py-3 bg-slate-50 dark:bg-neutral-800 hover:bg-slate-100 dark:hover:bg-neutral-700 transition-colors">
                <span className="font-semibold text-sm text-slate-800 dark:text-slate-200">{cat.name}</span>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-400">{cat.commands.length} commands</span>
                  {isOpen ? <ChevronUp size={14} className="text-slate-400" /> : <ChevronDown size={14} className="text-slate-400" />}
                </div>
              </button>
              {isOpen && (
                <div className="divide-y divide-slate-100 dark:divide-neutral-700">
                  {cat.commands.map((cmd, i) => (
                    <div key={i} className="px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-neutral-800/50 group">
                      <div className="flex items-start justify-between gap-2">
                        <code className="font-mono text-xs text-kawa-700 dark:text-kawa-400 break-all leading-relaxed">{cmd.command}</code>
                        <CopyBtn text={cmd.command} id={`git-${cat.name}-${i}`} copied={copied} copy={copy} className="shrink-0 mt-0.5 opacity-0 group-hover:opacity-100" />
                      </div>
                      <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 leading-relaxed">{cmd.description}</p>
                      {cmd.examples.length > 0 && (
                        <div className="mt-1.5 space-y-0.5">
                          {cmd.examples.map((ex, j) => (
                            <div key={j} className="flex items-center gap-2">
                              <code className="text-xs font-mono text-slate-500 dark:text-slate-500 italic">{ex}</code>
                              <CopyBtn text={ex} id={`git-ex-${cat.name}-${i}-${j}`} copied={copied} copy={copy} className="opacity-0 group-hover:opacity-100" />
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ─── Root component ────────────────────────────────────────────────────────────

export function WebDevOpsTool() {
  const [activeTab, setActiveTab] = useState<TabId>('url');

  return (
    <div className="max-w-5xl mx-auto p-4 space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Web & DevOps</h1>
        <p className="text-slate-500 dark:text-slate-400 text-sm">URL parser, User-Agent detector, HTTP status codes, MIME types, Docker → Compose, Git cheat sheet</p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 overflow-x-auto pb-1 scrollbar-none border-b border-slate-200 dark:border-neutral-700">
        {tabs.map(tab => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-2 text-sm font-medium whitespace-nowrap transition-colors border-b-2 -mb-px',
                activeTab === tab.id
                  ? 'border-kawa-600 text-kawa-700 dark:text-kawa-400'
                  : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'
              )}>
              <Icon size={15} />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <div>
        {activeTab === 'url'    && <UrlParserTab />}
        {activeTab === 'ua'     && <UserAgentTab />}
        {activeTab === 'status' && <HttpStatusTab />}
        {activeTab === 'mime'   && <MimeTab />}
        {activeTab === 'docker' && <DockerTab />}
        {activeTab === 'git'    && <GitTab />}
      </div>
    </div>
  );
}
