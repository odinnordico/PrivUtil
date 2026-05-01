import { useState, useCallback, useEffect } from 'react';
import { client } from '../lib/client';
import {
  Copy, Check, Link2, EyeOff, Search, Hash, Radio, List, Shuffle
} from 'lucide-react';
import { cn } from '../lib/utils';
import { ListAction } from '../proto/proto/privutil';

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
      {copied === id
        ? <><Check className="w-3.5 h-3.5" /> Copied</>
        : <><Copy className="w-3.5 h-3.5" /> Copy</>
      }
    </button>
  );
}

function ResultBlock({ label, value, id, copied, copy, mono = true }: {
  label?: string; value: string; id: string; copied: string | null;
  copy: (t: string, k: string) => void; mono?: boolean;
}) {
  return (
    <div className="bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-3">
      {label && (
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{label}</span>
          <CopyBtn text={value} id={id} copied={copied} copy={copy} />
        </div>
      )}
      {!label && (
        <div className="flex justify-end mb-1">
          <CopyBtn text={value} id={id} copied={copied} copy={copy} />
        </div>
      )}
      {mono
        ? <code className="text-sm font-mono text-slate-800 dark:text-slate-200 break-all whitespace-pre-wrap">{value || '—'}</code>
        : <p className="text-sm text-slate-800 dark:text-slate-200 break-all whitespace-pre-wrap">{value || '—'}</p>
      }
    </div>
  );
}

function ErrorMsg({ msg }: { msg: string }) {
  return msg ? (
    <p className="text-sm text-red-500 dark:text-red-400 mt-2">{msg}</p>
  ) : null;
}

function useDebounce<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

const tabs = [
  { id: 'slugify',   label: 'Slugify',         icon: Link2    },
  { id: 'hidden',    label: 'Hidden Chars',     icon: EyeOff   },
  { id: 'replace',   label: 'Text Replacer',    icon: Search   },
  { id: 'obfuscate', label: 'Obfuscator',       icon: EyeOff   },
  { id: 'numeronym', label: 'Numeronym',         icon: Hash     },
  { id: 'nato',      label: 'NATO Alphabet',    icon: Radio    },
  { id: 'list',      label: 'List Tools',       icon: List     },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const selectClass = "bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 text-sm";
const textareaClass = `${inputClass} font-mono resize-y min-h-[120px]`;

// ─── Slugify ──────────────────────────────────────────────────────────────────

function SlugifyTab() {
  const [text, setText] = useState('');
  const [separator, setSeparator] = useState('-');
  const [uppercase, setUppercase] = useState(false);
  const [maxLen, setMaxLen] = useState(0);
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced.trim()) { setResult(''); setError(''); return; }
    client.slugify({ text: debounced, separator: separator || '-', uppercase, maxLen } as Parameters<typeof client.slugify>[0])
      .then(r => { setResult(r.result); setError(r.error); })
      .catch(e => setError(String(e)));
  }, [debounced, separator, uppercase, maxLen]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Convert any string to a URL-safe slug — unicode-aware, diacritics stripped.
      </p>
      <textarea
        className={textareaClass}
        placeholder="Hello Wörld! This is a test..."
        value={text}
        onChange={e => setText(e.target.value)}
        rows={3}
      />
      <div className="flex flex-wrap gap-4 items-end">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Separator</label>
          <select className={selectClass} value={separator} onChange={e => setSeparator(e.target.value)}>
            <option value="-">Hyphen  (-)</option>
            <option value="_">Underscore (_)</option>
            <option value=".">Dot (.)</option>
            <option value="none">None (no separator)</option>
          </select>
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Max length (0 = unlimited)</label>
          <input
            type="number" min={0} max={500}
            className={cn(selectClass, 'w-32')}
            value={maxLen}
            onChange={e => setMaxLen(Number(e.target.value))}
          />
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" checked={uppercase} onChange={e => setUppercase(e.target.checked)} className="accent-kawa-500" />
          Uppercase
        </label>
      </div>
      {error && <ErrorMsg msg={error} />}
      {result && <ResultBlock value={result} id="slug" copied={copied} copy={copy} />}
    </div>
  );
}

// ─── Hidden character detector ────────────────────────────────────────────────

function HiddenCharsTab() {
  const [text, setText] = useState('');
  const [result, setResult] = useState<{
    hasHidden: boolean;
    annotated: string;
    cleaned: string;
    chars: { name: string; codepoint: string; count: number }[];
  } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced) { setResult(null); setError(''); return; }
    client.hiddenChars({ text: debounced } as Parameters<typeof client.hiddenChars>[0])
      .then(r => {
        setError(r.error || '');
        setResult({ hasHidden: r.hasHidden, annotated: r.annotated, cleaned: r.cleaned, chars: r.chars });
      })
      .catch(e => setError(String(e)));
  }, [debounced]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Reveal zero-width spaces, directional marks, BOM, and other invisible Unicode characters.
      </p>
      <textarea
        className={textareaClass}
        placeholder="Paste text here to inspect for hidden characters..."
        value={text}
        onChange={e => setText(e.target.value)}
        rows={4}
      />
      {error && <ErrorMsg msg={error} />}
      {result && !result.hasHidden && (
        <div className="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400 font-medium">
          <Check className="w-4 h-4" /> No hidden characters found.
        </div>
      )}
      {result?.hasHidden && (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400 font-medium">
            <EyeOff className="w-4 h-4" />
            {result.chars.length} type{result.chars.length !== 1 ? 's' : ''} of hidden characters found
          </div>
          <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-slate-100 dark:bg-neutral-800">
                <tr>
                  <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase">Codepoint</th>
                  <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase">Name</th>
                  <th className="text-right px-3 py-2 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase">Count</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                {result.chars.map(c => (
                  <tr key={c.codepoint} className="bg-white dark:bg-neutral-900">
                    <td className="px-3 py-2 font-mono text-kawa-600 dark:text-kawa-400">{c.codepoint}</td>
                    <td className="px-3 py-2 text-slate-700 dark:text-slate-300">{c.name}</td>
                    <td className="px-3 py-2 text-right font-mono">{c.count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <ResultBlock label="Annotated (with markers)" value={result.annotated} id="annotated" copied={copied} copy={copy} />
          <ResultBlock label="Cleaned (hidden chars removed)" value={result.cleaned} id="cleaned" copied={copied} copy={copy} />
        </div>
      )}
    </div>
  );
}

// ─── Text replacer ────────────────────────────────────────────────────────────

function TextReplacerTab() {
  const [text, setText] = useState('');
  const [find, setFind] = useState('');
  const [replaceWith, setReplaceWith] = useState('');
  const [useRegex, setUseRegex] = useState(false);
  const [caseInsensitive, setCaseInsensitive] = useState(false);
  const [result, setResult] = useState('');
  const [count, setCount] = useState<number | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(find);

  useEffect(() => {
    if (!text || !debounced) { setResult(''); setCount(null); setError(''); return; }
    client.textReplace({ text, find: debounced, replaceWith, useRegex, caseInsensitive } as Parameters<typeof client.textReplace>[0])
      .then(r => { setResult(r.result); setCount(r.count); setError(r.error); })
      .catch(e => setError(String(e)));
  }, [text, debounced, replaceWith, useRegex, caseInsensitive]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Find and replace text using plain-text or regular expression patterns.
      </p>
      <textarea
        className={textareaClass}
        placeholder="Paste your text here..."
        value={text}
        onChange={e => setText(e.target.value)}
        rows={5}
      />
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Find{useRegex && ' (regex)'}</label>
          <input className={inputClass} placeholder={useRegex ? '^hello.+' : 'foo'} value={find} onChange={e => setFind(e.target.value)} />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Replace with</label>
          <input className={inputClass} placeholder="bar" value={replaceWith} onChange={e => setReplaceWith(e.target.value)} />
        </div>
      </div>
      <div className="flex gap-4">
        <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" checked={useRegex} onChange={e => setUseRegex(e.target.checked)} className="accent-kawa-500" />
          Use regex
        </label>
        <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" checked={caseInsensitive} onChange={e => setCaseInsensitive(e.target.checked)} className="accent-kawa-500" />
          Case insensitive
        </label>
      </div>
      {error && <ErrorMsg msg={error} />}
      {count !== null && !error && (
        <p className="text-xs text-slate-500 dark:text-slate-400">
          {count} replacement{count !== 1 ? 's' : ''} made
        </p>
      )}
      {result && !error && (
        <div>
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Result</span>
            <CopyBtn text={result} id="replace" copied={copied} copy={copy} />
          </div>
          <textarea className={textareaClass} readOnly value={result} rows={5} />
        </div>
      )}
    </div>
  );
}

// ─── String obfuscator ────────────────────────────────────────────────────────

function ObfuscatorTab() {
  const [text, setText] = useState('');
  const [keepStart, setKeepStart] = useState(4);
  const [keepEnd, setKeepEnd] = useState(4);
  const [maskChar, setMaskChar] = useState('*');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced.trim()) { setResult(''); setError(''); return; }
    client.stringObfuscate({ text: debounced, keepStart, keepEnd, maskChar: maskChar || '*' } as Parameters<typeof client.stringObfuscate>[0])
      .then(r => { setResult(r.result); setError(r.error); })
      .catch(e => setError(String(e)));
  }, [debounced, keepStart, keepEnd, maskChar]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Partially mask a string — ideal for displaying API keys, passwords, or tokens safely.
      </p>
      <textarea
        className={textareaClass}
        placeholder="sk-AbCdEfGhIjKlMnOpQrStUvWxYz..."
        value={text}
        onChange={e => setText(e.target.value)}
        rows={3}
      />
      <div className="flex flex-wrap gap-4 items-end">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Keep start (chars)</label>
          <input
            type="number" min={0} max={50}
            className={cn(selectClass, 'w-24')}
            value={keepStart}
            onChange={e => setKeepStart(Number(e.target.value))}
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Keep end (chars)</label>
          <input
            type="number" min={0} max={50}
            className={cn(selectClass, 'w-24')}
            value={keepEnd}
            onChange={e => setKeepEnd(Number(e.target.value))}
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Mask character</label>
          <input
            type="text" maxLength={1}
            className={cn(selectClass, 'w-16 text-center font-mono')}
            value={maskChar}
            onChange={e => setMaskChar(e.target.value)}
          />
        </div>
      </div>
      {error && <ErrorMsg msg={error} />}
      {result && !error && <ResultBlock value={result} id="obfuscate" copied={copied} copy={copy} />}
    </div>
  );
}

// ─── Numeronym generator ──────────────────────────────────────────────────────

function NumeronymTab() {
  const [text, setText] = useState('');
  const [result, setResult] = useState('');
  const [words, setWords] = useState<{ original: string; numeronym: string }[]>([]);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced.trim()) { setResult(''); setWords([]); setError(''); return; }
    client.numeronymGenerate({ text: debounced } as Parameters<typeof client.numeronymGenerate>[0])
      .then(r => {
        setError(r.error);
        setResult(r.result);
        const srcWords = debounced.trim().split(/\s+/);
        setWords(srcWords.map((w, i) => ({ original: w, numeronym: r.words[i] ?? w })));
      })
      .catch(e => setError(String(e)));
  }, [debounced]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Convert words to numeronyms: <code className="text-kawa-600 dark:text-kawa-400">internationalization</code> → <code className="text-kawa-600 dark:text-kawa-400">i18n</code>
      </p>
      <textarea
        className={textareaClass}
        placeholder="internationalization kubernetes accessibility"
        value={text}
        onChange={e => setText(e.target.value)}
        rows={3}
      />
      {error && <ErrorMsg msg={error} />}
      {result && !error && (
        <div className="space-y-3">
          <ResultBlock label="Result" value={result} id="numeronym" copied={copied} copy={copy} />
          {words.length > 1 && (
            <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-slate-100 dark:bg-neutral-800">
                  <tr>
                    <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Original</th>
                    <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Numeronym</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                  {words.map((w, i) => (
                    <tr key={i} className="bg-white dark:bg-neutral-900">
                      <td className="px-3 py-2 font-mono text-slate-700 dark:text-slate-300">{w.original}</td>
                      <td className="px-3 py-2 font-mono text-kawa-600 dark:text-kawa-400">{w.numeronym}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ─── NATO alphabet ────────────────────────────────────────────────────────────

function NatoTab() {
  const [text, setText] = useState('');
  const [action, setAction] = useState<'encode' | 'decode'>('encode');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced.trim()) { setResult(''); setError(''); return; }
    client.natoAlphabet({ text: debounced, action } as Parameters<typeof client.natoAlphabet>[0])
      .then(r => { setResult(r.result); setError(r.error); })
      .catch(e => setError(String(e)));
  }, [debounced, action]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Encode text to NATO phonetic words (<em>Alfa Bravo Charlie…</em>) or decode back to text.
      </p>
      <div className="flex gap-3">
        {(['encode', 'decode'] as const).map(a => (
          <button
            key={a}
            onClick={() => setAction(a)}
            className={cn(
              'px-4 py-1.5 rounded-full text-sm font-medium border transition-colors',
              action === a
                ? 'bg-kawa-600 text-white border-kawa-600'
                : 'bg-white dark:bg-neutral-800 text-slate-700 dark:text-slate-300 border-slate-300 dark:border-neutral-600 hover:border-kawa-500'
            )}
          >
            {a === 'encode' ? 'Text → NATO' : 'NATO → Text'}
          </button>
        ))}
      </div>
      <textarea
        className={textareaClass}
        placeholder={action === 'encode' ? 'SOS' : 'Sierra Oscar Sierra'}
        value={text}
        onChange={e => setText(e.target.value)}
        rows={3}
      />
      {error && <ErrorMsg msg={error} />}
      {result && !error && <ResultBlock value={result} id="nato" copied={copied} copy={copy} />}
    </div>
  );
}

// ─── List tools ────────────────────────────────────────────────────────────────

const LIST_ACTIONS = [
  { value: ListAction.LIST_SORT_AZ,      label: 'Sort A → Z'        },
  { value: ListAction.LIST_SORT_ZA,      label: 'Sort Z → A'        },
  { value: ListAction.LIST_SORT_NUMERIC, label: 'Sort numeric'      },
  { value: ListAction.LIST_SHUFFLE,      label: 'Shuffle'           },
  { value: ListAction.LIST_DEDUPE,       label: 'Remove duplicates' },
  { value: ListAction.LIST_UNIQUE_ONLY,  label: 'Unique only'       },
  { value: ListAction.LIST_DUPLICATES,   label: 'Duplicates only'   },
  { value: ListAction.LIST_FREQUENCY,    label: 'Frequency count'   },
  { value: ListAction.LIST_REVERSE,      label: 'Reverse'           },
  { value: ListAction.LIST_TRIM,         label: 'Trim whitespace'   },
  { value: ListAction.LIST_REMOVE_EMPTY, label: 'Remove empty lines'},
] as const;

function ListToolsTab() {
  const [text, setText] = useState('');
  const [action, setAction] = useState<ListAction>(ListAction.LIST_SORT_AZ);
  const [caseInsensitive, setCaseInsensitive] = useState(false);
  const [result, setResult] = useState('');
  const [inputCount, setInputCount] = useState(0);
  const [outputCount, setOutputCount] = useState(0);
  const [freq, setFreq] = useState<{ line: string; count: number }[]>([]);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(text);

  useEffect(() => {
    if (!debounced.trim()) { setResult(''); setFreq([]); setError(''); return; }
    client.listProcess({ text: debounced, action, caseInsensitive } as Parameters<typeof client.listProcess>[0])
      .then(r => {
        setError(r.error);
        setResult(r.result);
        setInputCount(r.inputCount);
        setOutputCount(r.outputCount);
        setFreq(r.frequency);
      })
      .catch(e => setError(String(e)));
  }, [debounced, action, caseInsensitive]);

  const isFrequency = action === ListAction.LIST_FREQUENCY;

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Process newline-separated lists: sort, deduplicate, shuffle, analyze frequency, and more.
      </p>
      <div className="flex flex-wrap gap-4 items-end">
        <div className="flex-1 min-w-[180px]">
          <label className="block text-xs text-slate-500 mb-1">Action</label>
          <select
            className={selectClass}
            value={action}
            onChange={e => setAction(Number(e.target.value) as ListAction)}
          >
            {LIST_ACTIONS.map(a => (
              <option key={a.value} value={a.value}>{a.label}</option>
            ))}
          </select>
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" checked={caseInsensitive} onChange={e => setCaseInsensitive(e.target.checked)} className="accent-kawa-500" />
          Case insensitive
        </label>
      </div>
      <textarea
        className={textareaClass}
        placeholder={'apple\nbanana\ncherry\napple'}
        value={text}
        onChange={e => setText(e.target.value)}
        rows={6}
      />
      {error && <ErrorMsg msg={error} />}
      {result && !error && (
        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
            <span>Input: {inputCount} line{inputCount !== 1 ? 's' : ''} → Output: {outputCount} line{outputCount !== 1 ? 's' : ''}</span>
            <CopyBtn text={result} id="list" copied={copied} copy={copy} />
          </div>
          {isFrequency && freq.length > 0 ? (
            <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-slate-100 dark:bg-neutral-800">
                  <tr>
                    <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Line</th>
                    <th className="text-right px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Count</th>
                    <th className="text-right px-3 py-2 text-xs font-semibold text-slate-500 uppercase">%</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                  {freq.map((f, i) => (
                    <tr key={i} className="bg-white dark:bg-neutral-900">
                      <td className="px-3 py-2 font-mono text-slate-700 dark:text-slate-300">{f.line || '(empty)'}</td>
                      <td className="px-3 py-2 text-right font-mono">{f.count}</td>
                      <td className="px-3 py-2 text-right font-mono text-slate-400">
                        {inputCount > 0 ? ((f.count / inputCount) * 100).toFixed(1) : 0}%
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <textarea className={textareaClass} readOnly value={result} rows={6} />
          )}
        </div>
      )}
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

export function TextStringTool() {
  const [activeTab, setActiveTab] = useState<TabId>('slugify');

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 dark:text-white flex items-center gap-2">
          <Shuffle className="w-6 h-6 text-kawa-500" />
          Text &amp; String Tools
        </h1>
        <p className="text-slate-500 dark:text-slate-400 mt-1">
          Slugify, hidden-char detection, find/replace, obfuscation, numeronyms, NATO alphabet, and list utilities.
        </p>
      </div>

      {/* Tab bar */}
      <div className="flex flex-wrap gap-1 border-b border-slate-200 dark:border-neutral-700">
        {tabs.map(tab => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-t-lg border-b-2 transition-colors',
                activeTab === tab.id
                  ? 'border-kawa-500 text-kawa-600 dark:text-kawa-400'
                  : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'
              )}
            >
              <Icon className="w-3.5 h-3.5" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab panels */}
      <div className="bg-white dark:bg-neutral-900 rounded-xl border border-slate-200 dark:border-neutral-700 p-5">
        {activeTab === 'slugify'   && <SlugifyTab />}
        {activeTab === 'hidden'    && <HiddenCharsTab />}
        {activeTab === 'replace'   && <TextReplacerTab />}
        {activeTab === 'obfuscate' && <ObfuscatorTab />}
        {activeTab === 'numeronym' && <NumeronymTab />}
        {activeTab === 'nato'      && <NatoTab />}
        {activeTab === 'list'      && <ListToolsTab />}
      </div>
    </div>
  );
}
