import { useState, useCallback, useRef, useEffect, type DragEvent } from 'react';
import { client } from '../lib/client';
import {
  Copy, Check, Upload, Download, FileImage, Code2, Layers,
  RefreshCw, ChevronDown, ChevronUp, MapPin,
} from 'lucide-react';
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
    <button onClick={() => copy(text, id)} disabled={!text}
      className={cn('flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 disabled:opacity-30 transition-colors', className)}>
      {copied === id ? <Check size={13} className="text-emerald-500" /> : <Copy size={13} />}
    </button>
  );
}

function readFileAsText(file: File): Promise<string> {
  return new Promise((res, rej) => {
    const r = new FileReader();
    r.onload = e => res(e.target?.result as string);
    r.onerror = rej;
    r.readAsText(file);
  });
}

function readFileAsBytes(file: File): Promise<Uint8Array> {
  return new Promise((res, rej) => {
    const r = new FileReader();
    r.onload = e => res(new Uint8Array(e.target?.result as ArrayBuffer));
    r.onerror = rej;
    r.readAsArrayBuffer(file);
  });
}

function downloadBytes(bytes: Uint8Array, filename: string, mimeType: string) {
  const blob = new Blob([bytes as Uint8Array<ArrayBuffer>], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function fmtBytes(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(2)} MB`;
}

// ─── Drop Zone ────────────────────────────────────────────────────────────────

function DropZone({ accept, onFile, label, sublabel, file }: {
  accept: string;
  onFile: (file: File) => void;
  label?: string;
  sublabel?: string;
  file?: File | null;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    const f = e.dataTransfer.files[0];
    if (f) onFile(f);
  };

  return (
    <div
      onDragOver={e => { e.preventDefault(); setDragging(true); }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
      onClick={() => inputRef.current?.click()}
      className={cn(
        'border-2 border-dashed rounded-xl p-6 cursor-pointer transition-all text-center',
        dragging
          ? 'border-kawa-500 bg-kawa-50 dark:bg-kawa-900/20'
          : 'border-slate-300 dark:border-neutral-600 hover:border-kawa-400 dark:hover:border-kawa-500 bg-slate-50 dark:bg-neutral-800/50'
      )}>
      <input ref={inputRef} type="file" accept={accept} className="hidden"
        onChange={e => { const f = e.target.files?.[0]; if (f) onFile(f); e.target.value = ''; }} />
      <Upload size={24} className="mx-auto mb-2 text-slate-400" />
      {file ? (
        <div className="space-y-0.5">
          <p className="text-sm font-semibold text-slate-700 dark:text-slate-300">{file.name}</p>
          <p className="text-xs text-slate-400">{fmtBytes(file.size)}</p>
        </div>
      ) : (
        <div className="space-y-0.5">
          <p className="text-sm font-semibold text-slate-600 dark:text-slate-400">{label ?? 'Drop file here or click to browse'}</p>
          {sublabel && <p className="text-xs text-slate-400">{sublabel}</p>}
        </div>
      )}
    </div>
  );
}

// ─── Tabs ──────────────────────────────────────────────────────────────────────

const tabs = [
  { id: 'svg',    label: 'SVG Optimizer',   icon: Code2      },
  { id: 'exif',   label: 'Image Metadata',  icon: FileImage  },
  { id: 'b64',    label: 'Base64 ↔ File',   icon: Layers     },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const labelClass = "block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1";

// ─── SVG Optimizer ────────────────────────────────────────────────────────────

const SAMPLE_SVG = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1//DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg"
     xmlns:xlink="http://www.w3.org/1999/xlink"
     width="100"
     height="100"
     viewBox="0 0 100 100">
  <!-- Generated by Illustrator 2024 -->
  <!-- Copyright (c) Example Corp -->
  <metadata>
    <dc:title>My Icon</dc:title>
    <dc:description>A sample icon for testing</dc:description>
  </metadata>
  <title>My Icon</title>
  <desc>A sample icon for testing purposes</desc>
  <defs>
    <style type="text/css">
      .cls-1 { fill: #333; }
    </style>
  </defs>
  <g id="" style="" class="">
    <g>
    </g>
    <circle cx="50"
            cy="50"
            r="40"
            class="cls-1" />
  </g>
</svg>`;

const SVG_PRESETS = [
  { value: 'safe',       label: 'Safe',       desc: 'Remove comments, declarations, metadata, empty groups' },
  { value: 'aggressive', label: 'Aggressive', desc: 'Safe + remove <title> and <desc>' },
  { value: 'minimal',    label: 'Minimal',    desc: 'Remove comments only' },
  { value: 'custom',     label: 'Custom',     desc: 'Pick individual transforms' },
];

const CUSTOM_OPTS = [
  { key: 'removeComments',    label: 'Remove comments',        field: 'RemoveComments'    },
  { key: 'removeXmlDecl',     label: 'Remove XML declaration', field: 'RemoveXmlDecl'     },
  { key: 'removeDoctype',     label: 'Remove DOCTYPE',         field: 'RemoveDoctype'     },
  { key: 'removeMetadata',    label: 'Remove <metadata>',      field: 'RemoveMetadata'    },
  { key: 'removeTitle',       label: 'Remove <title>',         field: 'RemoveTitle'       },
  { key: 'removeDesc',        label: 'Remove <desc>',          field: 'RemoveDesc'        },
  { key: 'removeEmptyGroups', label: 'Remove empty groups',    field: 'RemoveEmptyGroups' },
  { key: 'collapseWhitespace',label: 'Collapse whitespace',    field: 'CollapseWhitespace'},
  { key: 'removeEmptyAttrs',  label: 'Remove empty attributes',field: 'RemoveEmptyAttrs'  },
] as const;

type CustomOpts = Record<string, boolean>;

function SvgTab() {
  const [svg, setSvg] = useState(SAMPLE_SVG);
  const [preset, setPreset] = useState('safe');
  const [customOpts, setCustomOpts] = useState<CustomOpts>({
    removeComments: true, removeXmlDecl: true, removeDoctype: true,
    removeMetadata: true, removeTitle: false, removeDesc: false,
    removeEmptyGroups: true, collapseWhitespace: true, removeEmptyAttrs: true,
  });
  const [result, setResult] = useState<{
    result: string; originalSize: number; optimizedSize: number;
    savingsPct: number; applied: string[];
  } | null>(null);
  const [error, setError] = useState('');
  const [showPreview, setShowPreview] = useState(false);
  const { copied, copy } = useCopy();

  const optimize = useCallback(() => {
    if (!svg.trim()) { setResult(null); setError(''); return; }
    const req = preset === 'custom'
      ? {
          svg, preset,
          removeComments: customOpts.removeComments, removeXmlDecl: customOpts.removeXmlDecl,
          removeDoctype: customOpts.removeDoctype, removeMetadata: customOpts.removeMetadata,
          removeTitle: customOpts.removeTitle, removeDesc: customOpts.removeDesc,
          removeEmptyGroups: customOpts.removeEmptyGroups, collapseWhitespace: customOpts.collapseWhitespace,
          removeEmptyAttrs: customOpts.removeEmptyAttrs,
        }
      : { svg, preset };
    client.svgOptimize(req as Parameters<typeof client.svgOptimize>[0]).then(res => {
      if (res.error) { setError(res.error); setResult(null); return; }
      setError('');
      setResult({
        result: res.result, originalSize: res.originalSize,
        optimizedSize: res.optimizedSize, savingsPct: res.savingsPct,
        applied: res.applied,
      });
    }).catch(() => {});
  }, [svg, preset, customOpts]);

  // Debounced auto-optimize
  useEffect(() => {
    const t = setTimeout(optimize, 400);
    return () => clearTimeout(t);
  }, [optimize]);

  const handleFileUpload = async (file: File) => {
    const text = await readFileAsText(file);
    setSvg(text);
  };

  const downloadOptimized = () => {
    if (!result?.result) return;
    const blob = new Blob([result.result], { type: 'image/svg+xml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'optimized.svg';
    document.body.appendChild(a); a.click();
    document.body.removeChild(a); URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-4">
      {/* Controls row */}
      <div className="flex flex-wrap gap-3 items-start">
        <div className="flex-1 min-w-[200px]">
          <label className={labelClass}>Preset</label>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-1.5">
            {SVG_PRESETS.map(p => (
              <button key={p.value} onClick={() => setPreset(p.value)}
                title={p.desc}
                className={cn('px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors',
                  preset === p.value
                    ? 'bg-kawa-600 text-white border-kawa-600'
                    : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700 hover:border-kawa-400')}>
                {p.label}
              </button>
            ))}
          </div>
          {preset === 'custom' && (
            <div className="mt-2 grid grid-cols-2 gap-x-4 gap-y-1">
              {CUSTOM_OPTS.map(opt => (
                <label key={opt.key} className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" checked={customOpts[opt.key] ?? false}
                    onChange={e => setCustomOpts(prev => ({ ...prev, [opt.key]: e.target.checked }))}
                    className="accent-kawa-600" />
                  <span className="text-xs text-slate-600 dark:text-slate-400">{opt.label}</span>
                </label>
              ))}
            </div>
          )}
        </div>
        <div>
          <label className={labelClass}>Upload SVG</label>
          <label className="flex items-center gap-2 px-3 py-2 rounded-lg border border-slate-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 text-sm text-slate-600 dark:text-slate-300 cursor-pointer hover:border-kawa-400 transition-colors">
            <Upload size={14} />
            Browse…
            <input type="file" accept=".svg,image/svg+xml" className="hidden"
              onChange={async e => { const f = e.target.files?.[0]; if (f) await handleFileUpload(f); e.target.value = ''; }} />
          </label>
        </div>
      </div>

      {/* Side-by-side editor */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <div>
          <label className={labelClass}>Input SVG</label>
          <textarea
            value={svg} onChange={e => setSvg(e.target.value)}
            rows={14} spellCheck={false}
            className={cn(inputClass, 'font-mono text-xs resize-none')}
            placeholder="Paste SVG markup here…"
          />
          <p className="text-xs text-slate-400 mt-1">{fmtBytes(svg.length)}</p>
        </div>

        <div>
          <div className="flex items-center justify-between mb-1">
            <label className={labelClass}>Optimized output</label>
            <div className="flex gap-2">
              {result && (
                <button onClick={downloadOptimized}
                  className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
                  <Download size={13} /> Save
                </button>
              )}
              {result && <CopyBtn text={result.result} id="svg-out" copied={copied} copy={copy} />}
            </div>
          </div>
          <textarea
            readOnly value={result?.result ?? ''} rows={14} spellCheck={false}
            className={cn(inputClass, 'font-mono text-xs resize-none bg-slate-50 dark:bg-neutral-900')}
            placeholder="Optimized SVG will appear here…"
          />
          {result && <p className="text-xs text-slate-400 mt-1">{fmtBytes(result.optimizedSize)}</p>}
        </div>
      </div>

      {error && <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{error}</div>}

      {/* Stats */}
      {result && (
        <div className="space-y-2">
          <div className="grid grid-cols-3 gap-3">
            {[
              { label: 'Original',  value: fmtBytes(result.originalSize) },
              { label: 'Optimized', value: fmtBytes(result.optimizedSize) },
              { label: 'Savings',   value: `${result.savingsPct.toFixed(1)}%`, highlight: result.savingsPct > 0 },
            ].map((s, i) => (
              <div key={i} className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3 text-center">
                <div className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-1">{s.label}</div>
                <div className={cn('text-lg font-bold font-mono', s.highlight ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-700 dark:text-slate-300')}>{s.value}</div>
              </div>
            ))}
          </div>
          {result.applied.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {result.applied.map((a, i) => (
                <span key={i} className="text-xs px-2 py-0.5 rounded-full bg-kawa-100 dark:bg-kawa-900/40 text-kawa-700 dark:text-kawa-300 font-medium">{a}</span>
              ))}
            </div>
          )}
        </div>
      )}

      {/* SVG Preview */}
      {result?.result && (
        <div>
          <button onClick={() => setShowPreview(!showPreview)}
            className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 transition-colors">
            {showPreview ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            {showPreview ? 'Hide preview' : 'Show SVG preview'}
          </button>
          {showPreview && (
            <div className="mt-2 rounded-xl border border-slate-200 dark:border-neutral-700 p-4 bg-[repeating-conic-gradient(#e2e8f0_0%_25%,transparent_0%_50%)] dark:bg-[repeating-conic-gradient(#404040_0%_25%,transparent_0%_50%)] bg-[length:20px_20px]">
              <div
                className="mx-auto max-w-full max-h-64 flex items-center justify-center overflow-hidden"
                dangerouslySetInnerHTML={{ __html: result.result }}
              />
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ─── Image Metadata / EXIF ────────────────────────────────────────────────────

const GROUP_ORDER = ['Image', 'Camera', 'DateTime', 'Settings', 'GPS', 'Metadata'];
const GROUP_COLORS: Record<string, string> = {
  Image:    'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300',
  Camera:   'bg-violet-100 dark:bg-violet-900/30 text-violet-700 dark:text-violet-300',
  DateTime: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300',
  Settings: 'bg-slate-100 dark:bg-slate-700/60 text-slate-700 dark:text-slate-300',
  GPS:      'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300',
  Metadata: 'bg-rose-100 dark:bg-rose-900/30 text-rose-700 dark:text-rose-300',
};

type ExifResult = {
  format: string; width: number; height: number;
  gpsDecimal: string; gpsDms: string; mapsUrl: string;
  fields: Array<{ label: string; value: string; group: string }>;
  error: string;
};

function ExifTab() {
  const [file, setFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [result, setResult] = useState<ExifResult | null>(null);
  const [loading, setLoading] = useState(false);
  const { copied, copy } = useCopy();

  const handleFile = useCallback(async (f: File) => {
    setFile(f);
    setResult(null);
    // Build preview URL for image types
    if (f.type.startsWith('image/')) {
      const url = URL.createObjectURL(f);
      setPreviewUrl(prev => { if (prev) URL.revokeObjectURL(prev); return url; });
    } else {
      setPreviewUrl(null);
    }
    setLoading(true);
    try {
      const bytes = await readFileAsBytes(f);
      const res = await client.exifRead({ data: bytes, filename: f.name } as Parameters<typeof client.exifRead>[0]);
      setResult({
        format: res.format, width: res.width, height: res.height,
        gpsDecimal: res.gpsDecimal, gpsDms: res.gpsDms, mapsUrl: res.mapsUrl,
        fields: res.fields.map(ff => ({ label: ff.label, value: ff.value, group: ff.group })),
        error: res.error,
      });
    } catch (e: unknown) {
      setResult({ format: '', width: 0, height: 0, gpsDecimal: '', gpsDms: '', mapsUrl: '', fields: [], error: String(e) });
    }
    setLoading(false);
  }, []);

  // Group fields
  const grouped: Record<string, Array<{ label: string; value: string }>> = {};
  if (result) {
    for (const f of result.fields) {
      const g = f.group || 'Other';
      if (!grouped[g]) grouped[g] = [];
      grouped[g].push({ label: f.label, value: f.value });
    }
  }

  const groupOrder = [...GROUP_ORDER, ...Object.keys(grouped).filter(g => !GROUP_ORDER.includes(g))];

  return (
    <div className="space-y-4">
      <DropZone
        accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp"
        onFile={handleFile}
        file={file}
        sublabel="JPEG, PNG, or WebP — up to 10 MB"
      />

      {loading && (
        <div className="flex items-center gap-2 text-slate-500 text-sm">
          <RefreshCw size={14} className="animate-spin" /> Reading metadata…
        </div>
      )}

      {result?.error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{result.error}</div>
      )}

      {result && !result.error && (
        <div className="space-y-4">
          {/* Image preview + summary */}
          <div className="flex flex-col sm:flex-row gap-4">
            {previewUrl && (
              <div className="shrink-0">
                <img src={previewUrl} alt="preview" className="max-h-40 max-w-48 rounded-xl border border-slate-200 dark:border-neutral-700 object-contain bg-slate-100 dark:bg-neutral-800" />
              </div>
            )}
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 flex-1">
              {[
                { label: 'Format', value: result.format.toUpperCase() },
                { label: 'Dimensions', value: result.width && result.height ? `${result.width} × ${result.height} px` : '—' },
                { label: 'File Size', value: file ? fmtBytes(file.size) : '—' },
              ].map((s, i) => (
                <div key={i} className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3">
                  <div className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-1">{s.label}</div>
                  <div className="font-semibold text-sm text-slate-800 dark:text-slate-200">{s.value}</div>
                </div>
              ))}
              {result.gpsDecimal && (
                <div className="bg-emerald-50 dark:bg-emerald-900/20 rounded-xl border border-emerald-200 dark:border-emerald-800 p-3 sm:col-span-2">
                  <div className="text-xs font-bold text-emerald-500 uppercase tracking-wide mb-1 flex items-center gap-1">
                    <MapPin size={11} /> GPS Location
                  </div>
                  <div className="font-mono text-xs text-slate-700 dark:text-slate-300">{result.gpsDms || result.gpsDecimal}</div>
                  {result.mapsUrl && (
                    <a href={result.mapsUrl} target="_blank" rel="noopener noreferrer"
                      className="text-xs text-kawa-600 hover:text-kawa-700 dark:text-kawa-400 hover:underline mt-0.5 block">
                      Open in Google Maps →
                    </a>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Metadata table per group */}
          {groupOrder.filter(g => grouped[g]?.length).map(group => (
            <div key={group}>
              <div className="flex items-center gap-2 mb-1.5">
                <span className={cn('text-xs px-2 py-0.5 rounded-full font-bold uppercase tracking-wide', GROUP_COLORS[group] ?? 'bg-slate-100 text-slate-600')}>{group}</span>
              </div>
              <div className="overflow-hidden rounded-xl border border-slate-200 dark:border-neutral-700">
                <table className="w-full text-sm">
                  <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                    {grouped[group].map((f, i) => (
                      <tr key={i} className="hover:bg-slate-50 dark:hover:bg-neutral-800/50">
                        <td className="px-4 py-2 text-xs font-semibold text-slate-400 w-40 whitespace-nowrap">{f.label}</td>
                        <td className="px-4 py-2 font-mono text-xs text-slate-700 dark:text-slate-300 break-all">{f.value}</td>
                        <td className="px-3 py-2 w-10">
                          <CopyBtn text={f.value} id={`exif-${group}-${i}`} copied={copied} copy={copy} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Base64 ↔ File ────────────────────────────────────────────────────────────

function EncodeSubTab() {
  const [file, setFile] = useState<File | null>(null);
  const [result, setResult] = useState<{ encoded: string; dataUri: string; mimeType: string; size: number } | null>(null);
  const [loading, setLoading] = useState(false);
  const [showUri, setShowUri] = useState(false);
  const { copied, copy } = useCopy();

  const handleFile = useCallback(async (f: File) => {
    setFile(f);
    setLoading(true);
    try {
      const bytes = await readFileAsBytes(f);
      const res = await client.fileToBase64({ data: bytes, filename: f.name } as Parameters<typeof client.fileToBase64>[0]);
      if (res.error) { setResult(null); setLoading(false); return; }
      setResult({ encoded: res.encoded, dataUri: res.dataUri, mimeType: res.mimeType, size: res.size });
    } catch { /* empty */ }
    setLoading(false);
  }, []);

  return (
    <div className="space-y-4">
      <DropZone accept="*/*" onFile={handleFile} file={file} sublabel="Any file type — up to 10 MB" />

      {loading && (
        <div className="flex items-center gap-2 text-slate-500 text-sm">
          <RefreshCw size={14} className="animate-spin" /> Encoding…
        </div>
      )}

      {result && (
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-3">
            {[
              { label: 'MIME Type', value: result.mimeType },
              { label: 'File Size', value: fmtBytes(result.size) },
              { label: 'Base64 Size', value: fmtBytes(result.encoded.length) },
            ].map((s, i) => (
              <div key={i} className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3 text-center">
                <div className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-1">{s.label}</div>
                <div className="font-mono text-sm font-semibold text-slate-700 dark:text-slate-300 truncate">{s.value}</div>
              </div>
            ))}
          </div>

          <div className="flex gap-2 mb-1">
            <button onClick={() => setShowUri(false)}
              className={cn('text-xs px-3 py-1.5 rounded-lg border font-medium transition-colors',
                !showUri ? 'bg-kawa-600 text-white border-kawa-600' : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700')}>
              Raw Base64
            </button>
            <button onClick={() => setShowUri(true)}
              className={cn('text-xs px-3 py-1.5 rounded-lg border font-medium transition-colors',
                showUri ? 'bg-kawa-600 text-white border-kawa-600' : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700')}>
              Data URI
            </button>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1">
              <label className={labelClass}>{showUri ? 'Data URI' : 'Base64'}</label>
              <CopyBtn text={showUri ? result.dataUri : result.encoded} id="b64-encoded" copied={copied} copy={copy} className="text-sm" />
            </div>
            <textarea
              readOnly
              value={showUri ? result.dataUri : result.encoded}
              rows={6}
              className="bg-slate-900 text-slate-100 rounded-xl p-3 text-xs font-mono w-full resize-none leading-relaxed outline-none"
            />
          </div>
        </div>
      )}
    </div>
  );
}

function DecodeSubTab() {
  const [encoded, setEncoded] = useState('');
  const [result, setResult] = useState<{ data: Uint8Array; mimeType: string; filename: string; size: number } | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const decode = useCallback(async () => {
    if (!encoded.trim()) { setResult(null); setError(''); return; }
    setLoading(true);
    try {
      const res = await client.base64ToFile({ encoded } as Parameters<typeof client.base64ToFile>[0]);
      if (res.error) { setError(res.error); setResult(null); setLoading(false); return; }
      setError('');
      setResult({ data: res.data, mimeType: res.mimeType, filename: res.filename, size: res.size });
    } catch (e: unknown) {
      setError(String(e));
    }
    setLoading(false);
  }, [encoded]);

  const download = () => {
    if (!result) return;
    downloadBytes(result.data, result.filename, result.mimeType);
  };

  return (
    <div className="space-y-4">
      <div>
        <label className={labelClass}>Base64 string or Data URI</label>
        <textarea
          value={encoded}
          onChange={e => setEncoded(e.target.value)}
          rows={7}
          placeholder="Paste base64 string or data:image/png;base64,... URI here"
          className={cn('bg-slate-900 text-slate-100 rounded-xl p-3 text-xs font-mono w-full resize-y leading-relaxed border-0 outline-none focus:ring-2 focus:ring-kawa-500')}
        />
      </div>

      <button onClick={decode} disabled={!encoded.trim()}
        className="px-4 py-2 bg-kawa-600 hover:bg-kawa-700 disabled:opacity-50 text-white rounded-lg text-sm font-semibold transition-colors flex items-center gap-2">
        {loading && <RefreshCw size={14} className="animate-spin" />}
        Decode
      </button>

      {error && <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2 text-sm text-red-700 dark:text-red-300">{error}</div>}

      {result && (
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-3">
            {[
              { label: 'MIME Type', value: result.mimeType },
              { label: 'Decoded Size', value: fmtBytes(result.size) },
              { label: 'Filename', value: result.filename },
            ].map((s, i) => (
              <div key={i} className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3 text-center">
                <div className="text-xs font-bold text-slate-400 uppercase tracking-wide mb-1">{s.label}</div>
                <div className="font-mono text-sm font-semibold text-slate-700 dark:text-slate-300 truncate">{s.value}</div>
              </div>
            ))}
          </div>

          {/* Image preview if it's an image */}
          {result.mimeType.startsWith('image/') && (
            <div className="rounded-xl border border-slate-200 dark:border-neutral-700 p-4 bg-[repeating-conic-gradient(#e2e8f0_0%_25%,transparent_0%_50%)] dark:bg-[repeating-conic-gradient(#404040_0%_25%,transparent_0%_50%)] bg-[length:20px_20px]">
              <img
                src={`data:${result.mimeType};base64,${btoa(String.fromCharCode(...result.data))}`}
                alt="decoded"
                className="mx-auto max-h-48 object-contain"
              />
            </div>
          )}

          <button onClick={download}
            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-semibold transition-colors">
            <Download size={15} />
            Download {result.filename}
          </button>
        </div>
      )}
    </div>
  );
}

function Base64FileTab() {
  const [sub, setSub] = useState<'encode' | 'decode'>('encode');
  return (
    <div className="space-y-3">
      <div className="flex gap-1">
        {(['encode', 'decode'] as const).map(s => (
          <button key={s} onClick={() => setSub(s)}
            className={cn('px-4 py-2 rounded-lg text-sm font-semibold border transition-colors',
              sub === s ? 'bg-kawa-600 text-white border-kawa-600' : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-neutral-700 hover:border-kawa-400')}>
            {s === 'encode' ? 'File → Base64' : 'Base64 → File'}
          </button>
        ))}
      </div>
      {sub === 'encode' ? <EncodeSubTab /> : <DecodeSubTab />}
    </div>
  );
}

// ─── Root component ────────────────────────────────────────────────────────────

export function MediaTool() {
  const [activeTab, setActiveTab] = useState<TabId>('svg');
  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Media Tools</h1>
        <p className="text-slate-500 dark:text-slate-400 text-sm">SVG optimizer, image metadata (EXIF), Base64 ↔ file converter</p>
      </div>

      <div className="flex gap-1 overflow-x-auto pb-1 scrollbar-none border-b border-slate-200 dark:border-neutral-700">
        {tabs.map(tab => {
          const Icon = tab.icon;
          return (
            <button key={tab.id} onClick={() => setActiveTab(tab.id)}
              className={cn('flex items-center gap-1.5 px-3 py-2 text-sm font-medium whitespace-nowrap transition-colors border-b-2 -mb-px',
                activeTab === tab.id
                  ? 'border-kawa-600 text-kawa-700 dark:text-kawa-400'
                  : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200')}>
              <Icon size={15} />
              {tab.label}
            </button>
          );
        })}
      </div>

      <div>
        {activeTab === 'svg'  && <SvgTab />}
        {activeTab === 'exif' && <ExifTab />}
        {activeTab === 'b64'  && <Base64FileTab />}
      </div>
    </div>
  );
}
