import { useRef, useState } from 'react';
import { client } from '../lib/client';
import { cn } from '../lib/utils';
import { ArrowRightLeft, Upload, X, CheckCircle2, XCircle, FileWarning } from 'lucide-react';

const MAX_FILE_BYTES = 10 * 1024 * 1024; // must match the server's maxDiffFileBytes

function bytesToHuman(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

interface ChecksumResult {
  algo: string;
  sum1: string;
  sum2: string;
  match: boolean;
  message: string;
  size1: number;
  size2: number;
}

interface SideProps {
  label: string;
  placeholder: string;
  text: string;
  onText: (v: string) => void;
  file: File | null;
  onFile: (f: File | null) => void;
}

function Side({ label, placeholder, text, onText, file, onFile }: SideProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-bold text-slate-600 dark:text-slate-400">{label}</label>
        <button
          type="button"
          onClick={() => inputRef.current?.click()}
          className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
        >
          <Upload className="w-3.5 h-3.5" /> Upload file
        </button>
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          onChange={(e) => { onFile(e.target.files?.[0] ?? null); e.target.value = ''; }}
        />
      </div>
      {file ? (
        <div className="flex items-center gap-3 h-64 p-4 bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700">
          <Upload className="w-4 h-4 text-kawa-600 dark:text-kawa-400 shrink-0" />
          <span className="text-sm text-slate-700 dark:text-slate-300 flex-1 truncate">
            {file.name} <span className="text-slate-400">({bytesToHuman(file.size)})</span>
          </span>
          <button
            onClick={() => onFile(null)}
            title="Remove file"
            className="text-slate-400 hover:text-red-500 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ) : (
        <textarea
          className="w-full h-64 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm"
          value={text}
          onChange={(e) => onText(e.target.value)}
          placeholder={placeholder}
        />
      )}
    </div>
  );
}

export function DiffTool() {
  const [text1, setText1] = useState('');
  const [text2, setText2] = useState('');
  const [file1, setFile1] = useState<File | null>(null);
  const [file2, setFile2] = useState<File | null>(null);
  const [diffHtml, setDiffHtml] = useState<string | null>(null);
  const [checksum, setChecksum] = useState<ChecksumResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const pickFile = (setter: (f: File | null) => void) => (f: File | null) => {
    setError(null);
    if (f && f.size > MAX_FILE_BYTES) {
      setError(`"${f.name}" is ${bytesToHuman(f.size)} — files must be under ${bytesToHuman(MAX_FILE_BYTES)}.`);
      return;
    }
    setter(f);
  };

  const handleCompare = async () => {
    setLoading(true);
    setError(null);
    setDiffHtml(null);
    setChecksum(null);
    try {
      if (file1 || file2) {
        // At least one side is a file: compare as files. A text side is encoded
        // to bytes so the server treats it as (readable) content.
        const b1 = file1 ? new Uint8Array(await file1.arrayBuffer()) : new TextEncoder().encode(text1);
        const b2 = file2 ? new Uint8Array(await file2.arrayBuffer()) : new TextEncoder().encode(text2);
        if (b1.length > MAX_FILE_BYTES || b2.length > MAX_FILE_BYTES) {
          setError(`Each side must be under ${bytesToHuman(MAX_FILE_BYTES)}.`);
          return;
        }
        const resp = await client.diffFiles({ file1: b1, file2: b2 } as Parameters<typeof client.diffFiles>[0]);
        if (resp.error) { setError(resp.error); return; }
        if (resp.isText) {
          setDiffHtml(resp.diffHtml);
        } else {
          setChecksum({
            algo: resp.checksumAlgo,
            sum1: resp.checksum1,
            sum2: resp.checksum2,
            match: resp.checksumsMatch,
            message: resp.message,
            size1: b1.length,
            size2: b2.length,
          });
        }
      } else {
        const resp = await client.diff({ text1, text2 } as Parameters<typeof client.diff>[0]);
        setDiffHtml(resp.diffHtml);
      }
    } catch (err) {
      console.error('Error comparing:', err);
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Diff Viewer</h2>
        <button
          onClick={handleCompare}
          disabled={loading}
          className={cn(
            'flex items-center gap-2 px-6 py-2 rounded-lg font-medium transition-colors',
            'bg-kawa-500 hover:bg-kawa-600 disabled:opacity-50 disabled:cursor-not-allowed text-slate-900',
          )}
        >
          <ArrowRightLeft className="w-4 h-4" />
          {loading ? 'Comparing...' : 'Compare'}
        </button>
      </div>

      <p className="text-sm text-slate-500 dark:text-slate-400">
        Paste text or upload a file on each side. Readable text is diffed inline;
        binary files are compared by SHA-256 checksum. Max {bytesToHuman(MAX_FILE_BYTES)} per file.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Side label="Original" placeholder="Paste original text here..." text={text1} onText={setText1} file={file1} onFile={pickFile(setFile1)} />
        <Side label="Modified" placeholder="Paste modified text here..." text={text2} onText={setText2} file={file2} onFile={pickFile(setFile2)} />
      </div>

      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500 text-sm">
          {error}
        </div>
      )}

      {diffHtml && (
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Result</label>
          <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 overflow-hidden text-slate-900 dark:text-neutral-100 shadow-sm">
            <div
              data-testid="diff-output"
              className="p-6 font-mono text-sm overflow-auto"
              dangerouslySetInnerHTML={{ __html: diffHtml }}
            />
          </div>
        </div>
      )}

      {checksum && (
        <div className="space-y-3">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-400">Result — checksum comparison</label>

          <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-300 text-sm">
            <FileWarning className="w-4 h-4 mt-0.5 shrink-0" />
            <span>{checksum.message}</span>
          </div>

          <div className={cn(
            'flex items-center gap-2 p-3 rounded-lg text-sm font-semibold',
            checksum.match
              ? 'bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300'
              : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300',
          )}>
            {checksum.match
              ? <><CheckCircle2 className="w-5 h-5" /> Files are identical — {checksum.algo} checksums match.</>
              : <><XCircle className="w-5 h-5" /> Files differ — {checksum.algo} checksums do not match.</>}
          </div>

          <div className="rounded-lg border border-slate-200 dark:border-neutral-700 divide-y divide-slate-200 dark:divide-neutral-700 overflow-hidden">
            {[
              { label: 'Original', sum: checksum.sum1, size: checksum.size1 },
              { label: 'Modified', sum: checksum.sum2, size: checksum.size2 },
            ].map((f) => (
              <div key={f.label} className="p-3 bg-white dark:bg-neutral-800">
                <div className="flex items-center justify-between text-xs font-bold text-slate-500 dark:text-slate-400 mb-1">
                  <span>{f.label} · {bytesToHuman(f.size)}</span>
                  <span className="uppercase">{checksum.algo}</span>
                </div>
                <code className="block text-xs font-mono break-all text-slate-800 dark:text-slate-200">{f.sum}</code>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
