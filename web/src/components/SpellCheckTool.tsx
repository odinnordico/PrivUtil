import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { client } from '../lib/client';
import { Copy, Check, Wand2, Eye, X, CheckCircle2 } from 'lucide-react';
import { cn } from '../lib/utils';
import type { SpellIssue, SpellLanguage } from '../proto/proto/privutil';

const DEBOUNCE_MS = 500;

const SAMPLE: Record<string, string> = {
  en: 'i beleive this sentance has a errors.  she could of done it better,but she did not. The the end is near.',
  es: 'creo que esta oracion tiene un errores.  como estas? el perro es muy  grande.',
};

interface TypeStyle {
  label: string;
  underline: string; // wavy underline color
  badge: string;     // badge bg/text
  dot: string;       // small color dot
}

const typeStyles: Record<string, TypeStyle> = {
  spelling: {
    label: 'Spelling',
    underline: 'decoration-red-500 dark:decoration-red-400',
    badge: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    dot: 'bg-red-500',
  },
  grammar: {
    label: 'Grammar',
    underline: 'decoration-blue-500 dark:decoration-blue-400',
    badge: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    dot: 'bg-blue-500',
  },
  punctuation: {
    label: 'Punctuation',
    underline: 'decoration-amber-500 dark:decoration-amber-400',
    badge: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    dot: 'bg-amber-500',
  },
  style: {
    label: 'Style',
    underline: 'decoration-purple-500 dark:decoration-purple-400',
    badge: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    dot: 'bg-purple-500',
  },
};

function styleFor(type: string): TypeStyle {
  return typeStyles[type] ?? typeStyles.style;
}

// A unique-but-stable signature so an ignored issue stays hidden across
// re-checks (which renumber offset-based ids).
function signature(is: SpellIssue): string {
  return `${is.rule}::${is.text}::${is.replacements.join('|')}`;
}

// Splice a replacement into text using rune (code-point) offsets.
function applyReplacement(text: string, offset: number, length: number, replacement: string): string {
  const cp = Array.from(text);
  return cp.slice(0, offset).join('') + replacement + cp.slice(offset + length).join('');
}

interface Segment {
  text: string;
  issue: SpellIssue | null;
}

// Split text into renderable segments, wrapping each non-overlapping issue.
function buildSegments(text: string, issues: SpellIssue[]): Segment[] {
  const cp = Array.from(text);
  const sorted = [...issues].sort((a, b) => a.offset - b.offset || a.length - b.length);
  const segs: Segment[] = [];
  let cursor = 0;
  for (const is of sorted) {
    if (is.offset < cursor) continue; // overlaps a previous highlight; skip inline
    if (is.offset > cursor) segs.push({ text: cp.slice(cursor, is.offset).join(''), issue: null });
    segs.push({ text: cp.slice(is.offset, is.offset + is.length).join(''), issue: is });
    cursor = is.offset + is.length;
  }
  if (cursor < cp.length) segs.push({ text: cp.slice(cursor).join(''), issue: null });
  return segs;
}

interface PopoverState {
  id: string;
  top: number;
  left: number;
  pinned: boolean;
}

export function SpellCheckTool() {
  const [text, setText] = useState('');
  const [language, setLanguage] = useState('en');
  const [languages, setLanguages] = useState<SpellLanguage[]>([
    { code: 'en', label: 'English' },
    { code: 'es', label: 'Español (Latinoamérica)' },
  ]);
  const [issues, setIssues] = useState<SpellIssue[]>([]);
  const [ignored, setIgnored] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [popover, setPopover] = useState<PopoverState | null>(null);
  const [copied, setCopied] = useState(false);

  const spanRefs = useRef<Record<string, HTMLSpanElement | null>>({});
  const closeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Load supported languages from the backend (engine is the source of truth).
  useEffect(() => {
    client.spellLanguages({}).then(resp => {
      if (resp.languages.length > 0) setLanguages(resp.languages);
    }).catch(() => { /* keep static fallback */ });
  }, []);

  // Debounced check whenever text or language changes.
  useEffect(() => {
    const timer = setTimeout(async () => {
      if (!text.trim()) {
        setIssues([]);
        setError('');
        setPopover(null);
        return;
      }
      setLoading(true);
      try {
        const resp = await client.spellCheck({ text, language } as Parameters<typeof client.spellCheck>[0]);
        setIssues(resp.issues);
        setError(resp.error || '');
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Spell check failed');
      } finally {
        setLoading(false);
      }
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [text, language]);

  const visibleIssues = useMemo(
    () => issues.filter(is => !ignored.has(signature(is))),
    [issues, ignored],
  );

  const segments = useMemo(() => buildSegments(text, visibleIssues), [text, visibleIssues]);

  const issueById = useMemo(() => {
    const m: Record<string, SpellIssue> = {};
    for (const is of visibleIssues) m[is.id] = is;
    return m;
  }, [visibleIssues]);

  const cancelClose = useCallback(() => {
    if (closeTimer.current) { clearTimeout(closeTimer.current); closeTimer.current = null; }
  }, []);

  const scheduleClose = useCallback(() => {
    cancelClose();
    closeTimer.current = setTimeout(() => {
      setPopover(prev => (prev && prev.pinned ? prev : null));
    }, 200);
  }, [cancelClose]);

  const openPopover = useCallback((id: string, pinned: boolean) => {
    const el = spanRefs.current[id];
    if (!el) return;
    cancelClose();
    const r = el.getBoundingClientRect();
    const left = Math.min(r.left, window.innerWidth - 340);
    setPopover({ id, top: r.bottom + 6, left: Math.max(8, left), pinned });
  }, [cancelClose]);

  // Close pinned popovers on scroll/resize (positions would go stale).
  useEffect(() => {
    if (!popover) return;
    const close = () => setPopover(null);
    window.addEventListener('scroll', close, true);
    window.addEventListener('resize', close);
    return () => {
      window.removeEventListener('scroll', close, true);
      window.removeEventListener('resize', close);
    };
  }, [popover]);

  const applyFix = useCallback((is: SpellIssue, replacement: string) => {
    setText(prev => applyReplacement(prev, is.offset, is.length, replacement));
    setPopover(null);
  }, []);

  const ignoreIssue = useCallback((is: SpellIssue) => {
    setIgnored(prev => new Set(prev).add(signature(is)));
    setPopover(null);
  }, []);

  const focusIssue = useCallback((id: string) => {
    const el = spanRefs.current[id];
    if (el) {
      el.scrollIntoView({ block: 'center', behavior: 'smooth' });
      setTimeout(() => openPopover(id, true), 250);
    }
  }, [openPopover]);

  const copyText = useCallback(async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }, [text]);

  const activeIssue = popover ? issueById[popover.id] : null;
  const counts = useMemo(() => {
    const c: Record<string, number> = {};
    for (const is of visibleIssues) c[is.type] = (c[is.type] ?? 0) + 1;
    return c;
  }, [visibleIssues]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Spell &amp; Grammar Checker</h2>
        <div className="flex items-center gap-3">
          <select
            value={language}
            onChange={e => { setLanguage(e.target.value); setIgnored(new Set()); }}
            className="bg-white dark:bg-neutral-800 border border-slate-300 dark:border-neutral-700 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-neutral-100 focus:outline-none focus:ring-2 focus:ring-kawa-500/50"
          >
            {languages.map(l => <option key={l.code} value={l.code}>{l.label}</option>)}
          </select>
          <button
            onClick={() => setText(SAMPLE[language] ?? SAMPLE.en)}
            className="text-sm font-medium text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
          >
            Try sample
          </button>
        </div>
      </div>

      {error && (
        <div className="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Editor + review */}
        <div className="lg:col-span-2 space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-bold text-slate-600 dark:text-slate-300">Your Text</label>
              <div className="flex items-center gap-3 text-xs font-mono font-bold text-slate-400 dark:text-slate-500">
                {loading && <span className="text-kawa-500 animate-pulse">Checking…</span>}
                {text && (
                  <button onClick={copyText} className="flex items-center gap-1 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors">
                    {copied ? <><Check className="w-3.5 h-3.5" /> Copied</> : <><Copy className="w-3.5 h-3.5" /> Copy</>}
                  </button>
                )}
              </div>
            </div>
            <textarea
              className="w-full h-48 bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 text-sm leading-relaxed focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm resize-y"
              value={text}
              onChange={e => setText(e.target.value)}
              placeholder="Type or paste text to check spelling and grammar…"
              spellCheck={false}
            />
          </div>

          {/* Highlighted review pane */}
          <div className="space-y-2">
            <label className="flex items-center gap-1.5 text-sm font-bold text-slate-600 dark:text-slate-300">
              <Eye className="w-4 h-4" /> Review
              <span className="font-normal text-slate-400 dark:text-slate-500">— hover an underlined word to fix it</span>
            </label>
            <div className="min-h-[8rem] bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-200 dark:border-neutral-700 shadow-sm text-sm leading-relaxed text-slate-900 dark:text-neutral-100 whitespace-pre-wrap break-words">
              {text
                ? segments.map((seg, i) => {
                    if (!seg.issue) return <span key={i}>{seg.text}</span>;
                    const st = styleFor(seg.issue.type);
                    const isActive = popover?.id === seg.issue.id;
                    const id = seg.issue.id;
                    return (
                      <span
                        key={i}
                        ref={el => { spanRefs.current[id] = el; }}
                        onMouseEnter={() => openPopover(id, popover?.pinned && popover.id === id ? true : false)}
                        onMouseLeave={scheduleClose}
                        onClick={() => openPopover(id, true)}
                        className={cn(
                          'cursor-pointer underline decoration-wavy underline-offset-4 rounded-sm transition-colors',
                          st.underline,
                          isActive ? 'bg-kawa-100 dark:bg-kawa-900/30' : 'hover:bg-slate-100 dark:hover:bg-neutral-700/50',
                        )}
                      >
                        {seg.text}
                      </span>
                    );
                  })
                : <span className="text-slate-400 dark:text-slate-500">Your checked text will appear here.</span>}
            </div>
          </div>
        </div>

        {/* Suggestions side panel */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-bold text-slate-600 dark:text-slate-300">Suggestions</span>
            <div className="flex items-center gap-2">
              {Object.entries(counts).map(([type, n]) => (
                <span key={type} className={cn('flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold', styleFor(type).badge)}>
                  <span className={cn('w-1.5 h-1.5 rounded-full', styleFor(type).dot)} /> {n}
                </span>
              ))}
            </div>
          </div>

          <div className="space-y-2 lg:max-h-[34rem] lg:overflow-y-auto pr-1">
            {visibleIssues.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-center text-slate-400 dark:text-slate-500">
                {text.trim() && !loading
                  ? <><CheckCircle2 className="w-8 h-8 text-kawa-500" /><span className="text-sm font-medium text-slate-600 dark:text-slate-300">No issues found</span></>
                  : <span className="text-sm">Suggestions will appear here as you type.</span>}
              </div>
            ) : (
              visibleIssues.map(is => {
                const st = styleFor(is.type);
                return (
                  <div
                    key={is.id}
                    className={cn(
                      'group bg-white dark:bg-neutral-800 rounded-lg border shadow-sm p-3 space-y-2 transition-colors',
                      popover?.id === is.id ? 'border-kawa-400 dark:border-kawa-600' : 'border-slate-200 dark:border-neutral-700',
                    )}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <button
                        onClick={() => focusIssue(is.id)}
                        className="flex items-center gap-2 min-w-0 text-left"
                        title="Find in text"
                      >
                        <span className={cn('px-1.5 py-0.5 rounded text-[10px] font-bold shrink-0', st.badge)}>{st.label}</span>
                        <span className="text-sm font-mono font-medium text-slate-800 dark:text-slate-100 truncate">{is.text || '—'}</span>
                      </button>
                      <button
                        onClick={() => ignoreIssue(is)}
                        className="shrink-0 text-slate-300 hover:text-slate-500 dark:text-slate-600 dark:hover:text-slate-400 transition-colors"
                        title="Ignore"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400">{is.message}</p>
                    {is.replacements.length > 0 && (
                      <div className="flex flex-wrap gap-1.5">
                        {is.replacements.map(r => (
                          <button
                            key={r}
                            onClick={() => applyFix(is, r)}
                            className="px-2 py-1 rounded-md text-xs font-medium bg-kawa-50 text-kawa-800 border border-kawa-200 hover:bg-kawa-100 dark:bg-kawa-900/20 dark:text-kawa-300 dark:border-kawa-700/50 dark:hover:bg-kawa-900/40 transition-colors"
                          >
                            {r}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>

      {/* Inline hover/click popover */}
      {popover && activeIssue && (
        <div
          className="fixed z-50 w-80 max-w-[calc(100vw-1rem)] bg-white dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 shadow-xl p-3 space-y-2"
          style={{ top: popover.top, left: popover.left }}
          onMouseEnter={cancelClose}
          onMouseLeave={scheduleClose}
        >
          <div className="flex items-center justify-between">
            <span className={cn('px-1.5 py-0.5 rounded text-[10px] font-bold', styleFor(activeIssue.type).badge)}>
              {styleFor(activeIssue.type).label}
            </span>
            <button onClick={() => setPopover(null)} className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-300">
              <X className="w-4 h-4" />
            </button>
          </div>
          <p className="text-xs text-slate-600 dark:text-slate-300">{activeIssue.message}</p>
          {activeIssue.replacements.length > 0 ? (
            <div className="flex flex-wrap gap-1.5 pt-1">
              {activeIssue.replacements.map(r => (
                <button
                  key={r}
                  onClick={() => applyFix(activeIssue, r)}
                  className="flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium bg-kawa-500 text-black hover:bg-kawa-400 transition-colors"
                >
                  <Wand2 className="w-3 h-3" /> {r}
                </button>
              ))}
            </div>
          ) : (
            <p className="text-[11px] italic text-slate-400 dark:text-slate-500">No automatic suggestion available.</p>
          )}
          <button
            onClick={() => ignoreIssue(activeIssue)}
            className="w-full text-center text-xs font-medium text-slate-400 hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300 pt-1 transition-colors"
          >
            Ignore
          </button>
        </div>
      )}
    </div>
  );
}
