import { useRef, useState } from 'react';
import { client } from '../lib/client';
import { cn } from '../lib/utils';
import { ArrowDownUp, Upload, Download, Copy, X } from 'lucide-react';

type Mode = 'encode' | 'decode';

function bytesToHuman(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

function isTextMime(mime: string) {
  return mime.startsWith('text/') || mime === 'application/json' || mime === 'application/xml';
}

function isImageMime(mime: string) {
  return mime.startsWith('image/');
}

export function Base64Tool() {
  const [mode, setMode] = useState<Mode>('encode');

  // encode state
  const [encodeText, setEncodeText]     = useState('');
  const [encodeFile, setEncodeFile]     = useState<File | null>(null);
  const [encodeOutput, setEncodeOutput] = useState('');

  // decode state
  const [decodeInput, setDecodeInput]   = useState('');
  const [decodeData, setDecodeData]     = useState<Uint8Array | null>(null);
  const [decodeMime, setDecodeMime]     = useState('');

  const [error, setError]     = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied]   = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // ── Encode ─────────────────────────────────────────────────────────────────

  const handleEncode = async () => {
    setLoading(true);
    setError(null);
    setEncodeOutput('');
    try {
      let resp;
      if (encodeFile) {
        const buf = await encodeFile.arrayBuffer();
        resp = await client.base64Encode({ raw: new Uint8Array(buf) } as Parameters<typeof client.base64Encode>[0]);
      } else {
        resp = await client.base64Encode({ text: encodeText } as Parameters<typeof client.base64Encode>[0]);
      }
      if (resp.error) { setError(resp.error); return; }
      setEncodeOutput(resp.text);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0] ?? null;
    setEncodeFile(f);
    setEncodeText('');
    setEncodeOutput('');
    setError(null);
  };

  const clearFile = () => {
    setEncodeFile(null);
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  // ── Decode ─────────────────────────────────────────────────────────────────

  const handleDecode = async () => {
    setLoading(true);
    setError(null);
    setDecodeData(null);
    setDecodeMime('');
    try {
      const resp = await client.base64Decode({ text: decodeInput } as Parameters<typeof client.base64Decode>[0]);
      if (resp.error) { setError(resp.error); return; }
      setDecodeData(resp.data);
      setDecodeMime(resp.mimeType);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  const downloadDecoded = () => {
    if (!decodeData) return;
    const blob = new Blob([new Uint8Array(decodeData)], { type: decodeMime || 'application/octet-stream' });
    const url  = URL.createObjectURL(blob);
    const a    = document.createElement('a');
    a.href = url;
    const ext  = decodeMime?.split('/')[1]?.split(';')[0] ?? 'bin';
    a.download = `decoded.${ext}`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const copy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  // ── Decoded output rendering ────────────────────────────────────────────────

  const renderDecodeOutput = () => {
    if (!decodeData) return null;

    if (isImageMime(decodeMime)) {
      const blob    = new Blob([new Uint8Array(decodeData)], { type: decodeMime });
      const dataUrl = URL.createObjectURL(blob);
      return (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-bold text-slate-600 dark:text-slate-400">
              Image — {decodeMime} · {bytesToHuman(decodeData.length)}
            </span>
            <button onClick={downloadDecoded}
              className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
              <Download className="w-3.5 h-3.5" /> Download
            </button>
          </div>
          <img src={dataUrl} alt="decoded"
            className="max-w-full rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm" />
        </div>
      );
    }

    if (isTextMime(decodeMime)) {
      const text = new TextDecoder().decode(decodeData);
      return (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <label className="text-sm font-bold text-slate-600 dark:text-slate-400">
              Output — {decodeMime}
            </label>
            <button onClick={() => copy(text)}
              className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
              <Copy className="w-3.5 h-3.5" /> {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
          <textarea readOnly
            className="w-full h-40 bg-slate-50 dark:bg-black/30 p-4 rounded-lg border border-slate-300 dark:border-gray-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none shadow-inner"
            value={text} />
        </div>
      );
    }

    // Binary / unknown
    return (
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm font-bold text-slate-600 dark:text-slate-400">
            Binary — {decodeMime || 'application/octet-stream'} · {bytesToHuman(decodeData.length)}
          </span>
          <button onClick={downloadDecoded}
            className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
            <Download className="w-3.5 h-3.5" /> Download
          </button>
        </div>
        <p className="text-sm text-slate-500 dark:text-slate-400">
          Binary content — use the Download button to save the file.
        </p>
      </div>
    );
  };

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6 max-w-4xl">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Base64 Encoder/Decoder</h2>

      {/* Mode tabs */}
      <div className="flex gap-2">
        {(['encode', 'decode'] as const).map(m => (
          <button key={m}
            onClick={() => { setMode(m); setError(null); }}
            className={cn(
              "px-5 py-2 rounded-lg font-medium text-sm transition-colors capitalize",
              mode === m
                ? "bg-kawa-500 text-slate-900"
                : "bg-slate-100 dark:bg-neutral-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-neutral-700"
            )}>
            {m}
          </button>
        ))}
      </div>

      {/* ── Encode panel ── */}
      {mode === 'encode' && (
        <div className="space-y-4">
          {encodeFile ? (
            <div className="flex items-center gap-3 p-4 bg-kawa-500/10 border border-kawa-500/30 rounded-lg">
              <Upload className="w-4 h-4 text-kawa-600 dark:text-kawa-400 shrink-0" />
              <span className="text-sm text-slate-700 dark:text-slate-300 flex-1 truncate">
                {encodeFile.name} <span className="text-slate-400">({bytesToHuman(encodeFile.size)})</span>
              </span>
              <button onClick={clearFile}
                className="text-slate-400 hover:text-red-500 transition-colors">
                <X className="w-4 h-4" />
              </button>
            </div>
          ) : (
            <div className="space-y-2">
              <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Input</label>
              <textarea
                className="w-full h-40 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
                value={encodeText}
                onChange={e => setEncodeText(e.target.value)}
                placeholder="Enter text to encode…"
              />
            </div>
          )}

          <div className="flex gap-3 flex-wrap">
            <button onClick={handleEncode}
              disabled={loading || (!encodeText && !encodeFile)}
              className={cn(
                "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
                "bg-kawa-500 hover:bg-kawa-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-900"
              )}>
              <ArrowDownUp className="w-4 h-4" />
              {loading ? 'Encoding…' : 'Encode'}
            </button>

            <button onClick={() => fileInputRef.current?.click()}
              className={cn(
                "flex items-center gap-2 px-5 py-2 rounded-lg font-medium text-sm transition-colors border",
                "border-slate-300 dark:border-neutral-600 bg-slate-100 dark:bg-neutral-800 text-slate-700 dark:text-slate-300",
                "hover:bg-slate-200 dark:hover:bg-neutral-700"
              )}>
              <Upload className="w-4 h-4" /> Upload file
            </button>
            <input ref={fileInputRef} type="file" className="hidden" onChange={handleFileSelect} />
          </div>

          {encodeOutput && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Base64 output</label>
                <button onClick={() => copy(encodeOutput)}
                  className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
                  <Copy className="w-3.5 h-3.5" /> {copied ? 'Copied!' : 'Copy'}
                </button>
              </div>
              <textarea readOnly
                className="w-full h-40 bg-slate-50 dark:bg-black/30 p-4 rounded-lg border border-slate-300 dark:border-gray-800 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none shadow-inner"
                value={encodeOutput} />
            </div>
          )}
        </div>
      )}

      {/* ── Decode panel ── */}
      {mode === 'decode' && (
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-bold text-slate-600 dark:text-slate-400">
              Base64 input <span className="font-normal text-slate-400">(raw or data URI)</span>
            </label>
            <textarea
              className="w-full h-40 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
              value={decodeInput}
              onChange={e => { setDecodeInput(e.target.value); setDecodeData(null); setDecodeMime(''); }}
              placeholder="Paste Base64 string or data URI here…"
            />
          </div>

          <button onClick={handleDecode}
            disabled={loading || !decodeInput.trim()}
            className={cn(
              "flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors",
              "bg-kawa-500 hover:bg-kawa-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-900"
            )}>
            <ArrowDownUp className="w-4 h-4" />
            {loading ? 'Decoding…' : 'Decode'}
          </button>

          {renderDecodeOutput()}
        </div>
      )}

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500 text-sm">
          {error}
        </div>
      )}
    </div>
  );
}
