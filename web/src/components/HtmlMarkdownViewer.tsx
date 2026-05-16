import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type DragEvent,
} from 'react';
import {
  Eye,
  FileCode,
  FileText,
  Upload,
  Trash2,
  Download,
  ShieldAlert,
  ShieldCheck,
  RefreshCw,
  Code2,
  AlertTriangle,
  Maximize2,
  X,
} from 'lucide-react';
import { client } from '../lib/client';
import { cn } from '../lib/utils';

type Mode = 'html' | 'markdown';

const MAX_INPUT_BYTES = 1024 * 1024;
const ACCEPT_ATTR = '.html,.htm,.md,.markdown,.txt,text/html,text/markdown,text/plain';
const ALLOWED_EXTENSIONS = ['html', 'htm', 'md', 'markdown', 'txt'] as const;

const SAMPLE_HTML = `<h1>Hello, PrivUtil</h1>
<p>This is rendered inside a <strong>sandboxed iframe</strong>.</p>
<ul>
  <li>Scripts are blocked by the sandbox.</li>
  <li>External resources are blocked by CSP.</li>
  <li>Toggle "Allow images" to permit image loading.</li>
</ul>
<blockquote>Edit the source on the left and the preview updates automatically.</blockquote>`;

const SAMPLE_MD = `# Hello, PrivUtil

Render Markdown safely with **server-side** conversion plus an
isolated iframe sandbox.

- No scripts execute
- No tracking pixels by default
- Drop a \`.md\` file to load it instantly

> Use the *Allow images* toggle if you trust the source.`;

function fileExtension(name: string): string {
  const idx = name.lastIndexOf('.');
  return idx >= 0 ? name.slice(idx + 1).toLowerCase() : '';
}

function fmtBytes(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(2)} MB`;
}

function inferModeFromFile(name: string): Mode | null {
  const ext = fileExtension(name);
  if (ext === 'html' || ext === 'htm') return 'html';
  if (ext === 'md' || ext === 'markdown') return 'markdown';
  return null;
}

function buildSrcDoc(bodyHtml: string, allowImages: boolean, enableMermaid: boolean): string {
  const imgSrc = allowImages ? "img-src data: https: http:;" : "img-src 'none';";
  const cspParts = [
    "default-src 'none'",
    "style-src 'unsafe-inline'",
    imgSrc,
    "font-src data:",
    "base-uri 'none'",
    "form-action 'none'",
    "frame-ancestors 'none'",
  ];
  if (enableMermaid) {
    cspParts.push("script-src https://cdn.jsdelivr.net 'unsafe-inline'");
  }
  const csp = cspParts.join('; ');

  const mermaidScripts = enableMermaid ? `
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>
  const dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  mermaid.initialize({ startOnLoad: true, theme: dark ? 'dark' : 'default' });
</script>` : '';

  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="${csp}">
<title>Preview</title>
<style>
  html, body { margin: 0; padding: 16px; background: #fff; color: #111; font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif; line-height: 1.55; }
  @media (prefers-color-scheme: dark) { html, body { background: #0b0b0c; color: #e5e5e5; } a { color: #93c5fd; } }
  img, video, table { max-width: 100%; }
  pre { white-space: pre-wrap; word-break: break-word; background: rgba(127,127,127,0.12); padding: 12px; border-radius: 6px; overflow-x: auto; }
  code { background: rgba(127,127,127,0.15); padding: 1px 4px; border-radius: 4px; }
  blockquote { border-left: 4px solid rgba(127,127,127,0.4); margin: 0; padding: 4px 12px; color: inherit; opacity: 0.85; }
  table { border-collapse: collapse; width: auto; margin: 12px 0; }
  th, td { border: 1px solid rgba(127,127,127,0.35); padding: 6px 12px; text-align: left; }
  th { background: rgba(127,127,127,0.12); font-weight: 600; }
  tr:nth-child(even) td { background: rgba(127,127,127,0.05); }
  pre.mermaid { background: transparent; padding: 0; }
</style>
</head>
<body>
${bodyHtml}${mermaidScripts}
</body>
</html>`;
}

export function HtmlMarkdownViewer() {
  const [mode, setMode] = useState<Mode>('markdown');
  const [input, setInput] = useState<string>(SAMPLE_MD);
  const [renderedHtml, setRenderedHtml] = useState<string>('');
  const [allowImages, setAllowImages] = useState(false);
  const [enableMermaid, setEnableMermaid] = useState(false);
  const [showSource, setShowSource] = useState(false);
  const [fileName, setFileName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [maximized, setMaximized] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const reqIdRef = useRef(0);

  const inputBytes = useMemo(() => new TextEncoder().encode(input).length, [input]);
  const overLimit = inputBytes > MAX_INPUT_BYTES;

  const convert = useCallback(async (text: string, currentMode: Mode) => {
    if (!text) {
      setRenderedHtml('');
      setError(null);
      return;
    }
    if (new TextEncoder().encode(text).length > MAX_INPUT_BYTES) {
      setRenderedHtml('');
      setError(`Input exceeds ${fmtBytes(MAX_INPUT_BYTES)} limit.`);
      return;
    }
    setError(null);
    if (currentMode === 'html') {
      setRenderedHtml(text);
      return;
    }
    const reqId = ++reqIdRef.current;
    setBusy(true);
    try {
      const req = { text } as Parameters<typeof client.markdownToHtml>[0];
      const resp = await client.markdownToHtml(req);
      if (reqId !== reqIdRef.current) return;
      setRenderedHtml(resp.text);
    } catch (e) {
      if (reqId !== reqIdRef.current) return;
      setError(e instanceof Error ? e.message : 'Failed to render Markdown.');
      setRenderedHtml('');
    } finally {
      if (reqId === reqIdRef.current) setBusy(false);
    }
  }, []);

  useEffect(() => {
    const t = window.setTimeout(() => { void convert(input, mode); }, 250);
    return () => window.clearTimeout(t);
  }, [input, mode, convert]);

  useEffect(() => {
    if (!maximized) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setMaximized(false); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [maximized]);

  const srcDoc = useMemo(
    () => buildSrcDoc(renderedHtml, allowImages, enableMermaid),
    [renderedHtml, allowImages, enableMermaid],
  );

  const handleFile = useCallback(async (file: File) => {
    setError(null);
    if (file.size > MAX_INPUT_BYTES) {
      setError(`File "${file.name}" is ${fmtBytes(file.size)} (max ${fmtBytes(MAX_INPUT_BYTES)}).`);
      return;
    }
    const ext = fileExtension(file.name);
    if (!ALLOWED_EXTENSIONS.includes(ext as typeof ALLOWED_EXTENSIONS[number])) {
      setError(`Unsupported file type ".${ext}". Allowed: ${ALLOWED_EXTENSIONS.join(', ')}.`);
      return;
    }
    try {
      const text = await file.text();
      const detected = inferModeFromFile(file.name);
      if (detected) setMode(detected);
      setInput(text);
      setFileName(file.name);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to read file.');
    }
  }, []);

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    const f = e.dataTransfer.files[0];
    if (f) void handleFile(f);
  };

  const clearAll = () => {
    setInput('');
    setRenderedHtml('');
    setFileName(null);
    setError(null);
  };

  const loadSample = () => {
    setFileName(null);
    setError(null);
    setInput(mode === 'html' ? SAMPLE_HTML : SAMPLE_MD);
  };

  const downloadHtml = () => {
    const blob = new Blob([srcDoc], { type: 'text/html;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = (fileName?.replace(/\.(md|markdown|html|htm|txt)$/i, '') || 'preview') + '.html';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <Eye className="w-6 h-6 text-kawa-500" /> HTML &amp; Markdown Viewer
        </h2>
        <div className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400">
          <ShieldCheck className="w-4 h-4 text-emerald-500" />
          Sandboxed iframe • CSP enforced • {allowImages ? 'images allowed' : 'images blocked'} • {enableMermaid ? 'scripts allowed (mermaid)' : 'scripts blocked'}
        </div>
      </div>

      <div className="flex flex-wrap gap-2 border-b border-slate-300 dark:border-gray-700">
        <button
          onClick={() => setMode('markdown')}
          className={cn(
            'pb-2 px-4 font-bold transition-colors flex items-center gap-2',
            mode === 'markdown'
              ? 'text-kawa-600 border-b-2 border-kawa-500'
              : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200',
          )}>
          <FileText className="w-4 h-4" /> Markdown
        </button>
        <button
          onClick={() => setMode('html')}
          className={cn(
            'pb-2 px-4 font-bold transition-colors flex items-center gap-2',
            mode === 'html'
              ? 'text-kawa-600 border-b-2 border-kawa-500'
              : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200',
          )}>
          <FileCode className="w-4 h-4" /> HTML
        </button>
      </div>

      <div
        onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
        className={cn(
          'rounded-xl border p-4 transition-all',
          dragging
            ? 'border-kawa-500 bg-kawa-50 dark:bg-kawa-900/20'
            : 'border-slate-200 dark:border-neutral-700 bg-white dark:bg-neutral-800',
        )}>
        <div className="flex flex-wrap items-center gap-2 mb-3">
          <button
            onClick={() => fileInputRef.current?.click()}
            className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-3 py-1.5 rounded text-sm font-medium transition-colors flex items-center gap-1.5">
            <Upload className="w-4 h-4" /> Upload file
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept={ACCEPT_ATTR}
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) void handleFile(f);
              e.target.value = '';
            }}
          />
          <button
            onClick={loadSample}
            className="text-sm px-3 py-1.5 rounded bg-slate-100 dark:bg-neutral-700 hover:bg-slate-200 dark:hover:bg-neutral-600 text-slate-700 dark:text-slate-200 transition-colors">
            Load sample
          </button>
          <button
            onClick={clearAll}
            disabled={!input}
            className="text-sm px-3 py-1.5 rounded bg-slate-100 dark:bg-neutral-700 hover:bg-slate-200 dark:hover:bg-neutral-600 text-slate-700 dark:text-slate-200 transition-colors disabled:opacity-40 flex items-center gap-1.5">
            <Trash2 className="w-4 h-4" /> Clear
          </button>
          <button
            onClick={() => void convert(input, mode)}
            className="text-sm px-3 py-1.5 rounded bg-slate-100 dark:bg-neutral-700 hover:bg-slate-200 dark:hover:bg-neutral-600 text-slate-700 dark:text-slate-200 transition-colors flex items-center gap-1.5">
            <RefreshCw className={cn('w-4 h-4', busy && 'animate-spin')} /> Refresh
          </button>
          <button
            onClick={downloadHtml}
            disabled={!renderedHtml}
            className="text-sm px-3 py-1.5 rounded bg-slate-100 dark:bg-neutral-700 hover:bg-slate-200 dark:hover:bg-neutral-600 text-slate-700 dark:text-slate-200 transition-colors disabled:opacity-40 flex items-center gap-1.5">
            <Download className="w-4 h-4" /> Download HTML
          </button>

          <div className="ml-auto flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
            <label className="flex items-center gap-1.5 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={allowImages}
                onChange={(e) => setAllowImages(e.target.checked)}
                className="accent-kawa-500"
              />
              Allow images
            </label>
            <label className="flex items-center gap-1.5 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={enableMermaid}
                onChange={(e) => setEnableMermaid(e.target.checked)}
                className="accent-kawa-500"
              />
              Mermaid diagrams
            </label>
            <label className="flex items-center gap-1.5 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={showSource}
                onChange={(e) => setShowSource(e.target.checked)}
                className="accent-kawa-500"
              />
              <Code2 className="w-3.5 h-3.5" /> Show rendered HTML
            </label>
          </div>
        </div>

        <div className="text-xs text-slate-500 dark:text-slate-400 mb-2 flex items-center gap-3 flex-wrap">
          {fileName ? <span className="font-mono">{fileName}</span> : <span>Drop a .{mode === 'markdown' ? 'md' : 'html'} file or paste below</span>}
          <span>•</span>
          <span className={cn(overLimit && 'text-red-500 font-semibold')}>{fmtBytes(inputBytes)} / {fmtBytes(MAX_INPUT_BYTES)}</span>
        </div>

        {error && (
          <div className="mb-3 flex items-start gap-2 rounded-md border border-red-300 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-2 text-sm text-red-700 dark:text-red-300">
            <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <div className="space-y-1">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">
              {mode === 'markdown' ? 'Markdown source' : 'HTML source'}
            </label>
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder={mode === 'markdown' ? '# Heading\n\nSome **markdown** text…' : '<h1>Hello</h1>\n<p>Some HTML…</p>'}
              spellCheck={false}
              className="w-full h-[28rem] bg-white dark:bg-black/30 p-3 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
            />
          </div>
          <div className="space-y-1">
            <div className="flex items-center justify-between">
              <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase flex items-center gap-2">
                Preview
                <span className="inline-flex items-center gap-1 text-[10px] font-medium normal-case text-emerald-600 dark:text-emerald-400">
                  <ShieldAlert className="w-3 h-3" /> sandboxed
                </span>
              </label>
              <button
                onClick={() => setMaximized(true)}
                disabled={!renderedHtml}
                title="Maximise preview"
                className="text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 disabled:opacity-30 transition-colors">
                <Maximize2 className="w-4 h-4" />
              </button>
            </div>
            <iframe
              key={`${allowImages}-${enableMermaid}`}
              title="Rendered preview"
              sandbox={enableMermaid ? 'allow-scripts' : ''}
              srcDoc={srcDoc}
              referrerPolicy="no-referrer"
              className="w-full h-[28rem] rounded border border-slate-300 dark:border-neutral-700 bg-white"
            />
          </div>
        </div>

        {showSource && (
          <div className="mt-4 space-y-1">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">
              Rendered HTML (sent to iframe)
            </label>
            <textarea
              readOnly
              value={renderedHtml}
              className="w-full h-48 bg-slate-50 dark:bg-black/30 p-3 rounded font-mono text-xs resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100"
            />
          </div>
        )}
      </div>

      {maximized && (
        <div
          className="fixed inset-0 z-50 flex flex-col bg-black/80 backdrop-blur-sm"
          onClick={(e) => { if (e.target === e.currentTarget) setMaximized(false); }}>
          <div className="flex items-center justify-between px-4 py-2 bg-neutral-900 border-b border-neutral-700 shrink-0">
            <span className="text-sm font-semibold text-white flex items-center gap-2">
              <Eye className="w-4 h-4 text-kawa-400" /> Preview
              <span className="text-[10px] font-medium text-emerald-400 flex items-center gap-1">
                <ShieldAlert className="w-3 h-3" /> sandboxed
              </span>
            </span>
            <button
              onClick={() => setMaximized(false)}
              title="Close (Esc)"
              className="text-neutral-400 hover:text-white transition-colors">
              <X className="w-5 h-5" />
            </button>
          </div>
          <iframe
            key={`${allowImages}-${enableMermaid}-max`}
            title="Rendered preview (maximised)"
            sandbox={enableMermaid ? 'allow-scripts' : ''}
            srcDoc={srcDoc}
            referrerPolicy="no-referrer"
            className="flex-1 w-full bg-white"
          />
        </div>
      )}

      <div className="rounded-lg border border-slate-200 dark:border-neutral-700 bg-slate-50 dark:bg-neutral-800/50 p-4 text-xs text-slate-600 dark:text-slate-400 space-y-1">
        <p className="font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
          <ShieldCheck className="w-4 h-4 text-emerald-500" /> Security guardrails
        </p>
        <ul className="list-disc list-inside space-y-0.5">
          <li>Preview runs inside an <code>iframe</code> with a strict <code>sandbox</code> attribute — forms, popups, plugins and top-level navigation are always disabled. Scripts are blocked unless <em>Mermaid diagrams</em> is enabled.</li>
          <li>A strict Content-Security-Policy <code>meta</code> tag blocks scripts, objects, frames and external fetches; only inline styles and (optionally) images are permitted.</li>
          <li>File uploads are limited to {fmtBytes(MAX_INPUT_BYTES)} and to <code>.html</code>, <code>.htm</code>, <code>.md</code>, <code>.markdown</code>, <code>.txt</code>.</li>
          <li>Markdown is rendered server-side with goldmark, then handed to the same sandboxed iframe — never injected into the parent page.</li>
        </ul>
      </div>
    </div>
  );
}
