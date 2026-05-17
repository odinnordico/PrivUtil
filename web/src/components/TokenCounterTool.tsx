import { useState, useEffect, useCallback } from 'react';
import { client } from '../lib/client';
import { Copy, Check, ChevronDown, ChevronUp, Info } from 'lucide-react';
import { cn } from '../lib/utils';
import type { TokenStrategy } from '../proto/proto/privutil';

function useCopy() {
  const [copied, setCopied] = useState<string | null>(null);
  const copy = useCallback(async (text: string, key: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 1500);
  }, []);
  return { copied, copy };
}

interface StrategyOption {
  value: string;
  label: string;
  group: string;
  groupLabel: string;
}

const strategyOptions: StrategyOption[] = [
  { value: 'gpt-4o',     label: 'GPT-4o / GPT-4o-mini',    group: 'openai',    groupLabel: 'OpenAI' },
  { value: 'gpt-4',      label: 'GPT-4 / GPT-4-turbo',     group: 'openai',    groupLabel: 'OpenAI' },
  { value: 'gpt-3.5',    label: 'GPT-3.5-turbo',           group: 'openai',    groupLabel: 'OpenAI' },
  { value: 'claude',     label: 'Claude 3.5 / 4',          group: 'anthropic', groupLabel: 'Anthropic' },
  { value: 'llama-3',    label: 'Llama 3 / 3.1 / 3.2',    group: 'meta',      groupLabel: 'Meta' },
  { value: 'gemini',     label: 'Gemini 1.5 / 2',          group: 'google',    groupLabel: 'Google' },
  { value: 'mistral',    label: 'Mistral / Mixtral',       group: 'mistral',   groupLabel: 'Mistral AI' },
  { value: 'whitespace', label: 'Whitespace',              group: 'classic',   groupLabel: 'Classic' },
  { value: 'word',       label: 'Word',                    group: 'classic',   groupLabel: 'Classic' },
  { value: 'sentence',   label: 'Sentence',                group: 'classic',   groupLabel: 'Classic' },
  { value: 'character',  label: 'Character',               group: 'classic',   groupLabel: 'Classic' },
];

const DEFAULT_STRATEGY = 'gpt-4o';

const groupColors: Record<string, string> = {
  openai:    'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  anthropic: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  meta:      'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  google:    'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  mistral:   'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  classic:   'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-300',
};

export function TokenCounterTool() {
  const [input, setInput] = useState('');
  const [selected, setSelected] = useState(DEFAULT_STRATEGY);
  const [allStrategies, setAllStrategies] = useState<TokenStrategy[]>([]);
  const [charCount, setCharCount] = useState(0);
  const [byteCount, setByteCount] = useState(0);
  const [showTokens, setShowTokens] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  useEffect(() => {
    const timer = setTimeout(async () => {
      if (!input.trim()) {
        setAllStrategies([]);
        setCharCount(0);
        setByteCount(0);
        setError('');
        return;
      }
      setLoading(true);
      try {
        const resp = await client.tokenCount({ text: input } as Parameters<typeof client.tokenCount>[0]);
        setAllStrategies(resp.strategies);
        setCharCount(resp.charCount);
        setByteCount(resp.byteCount);
        setError(resp.error || '');
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Token counting failed');
      } finally {
        setLoading(false);
      }
    }, 400);
    return () => clearTimeout(timer);
  }, [input]);

  const primary = allStrategies.find(s => s.name === selected);
  const others = allStrategies.filter(s => s.name !== selected);

  const groups = strategyOptions.reduce<Record<string, StrategyOption[]>>((acc, opt) => {
    (acc[opt.groupLabel] ??= []).push(opt);
    return acc;
  }, {});

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Token Counter</h2>

      {/* Model selector */}
      <div className="flex flex-wrap items-end gap-4">
        <div className="space-y-1.5">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-300">Tokenizer</label>
          <select
            value={selected}
            onChange={e => { setSelected(e.target.value); setShowTokens(false); }}
            className="block w-64 bg-white dark:bg-neutral-800 border border-slate-300 dark:border-neutral-700 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-neutral-100 focus:outline-none focus:ring-2 focus:ring-kawa-500/50"
          >
            {Object.entries(groups).map(([groupLabel, opts]) => (
              <optgroup key={groupLabel} label={groupLabel}>
                {opts.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </optgroup>
            ))}
          </select>
        </div>

        {input && (
          <div className="flex flex-wrap gap-4 text-xs font-mono font-bold text-slate-500 dark:text-slate-400 pb-2">
            <span>Chars: {charCount.toLocaleString()}</span>
            <span>Bytes: {byteCount.toLocaleString()}</span>
            {loading && <span className="text-kawa-500 animate-pulse">Counting...</span>}
          </div>
        )}
      </div>

      {error && (
        <div className="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input */}
        <div className="space-y-2">
          <label className="text-sm font-bold text-slate-600 dark:text-slate-300">Input Text</label>
          <textarea
            className="w-full h-[500px] bg-white dark:bg-neutral-800 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-kawa-500/50 shadow-sm resize-y"
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder="Paste or type text to count tokens..."
          />
        </div>

        {/* Results */}
        <div className="space-y-4">
          {/* Primary count */}
          {primary ? (
            <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 shadow-sm p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className={cn('px-2 py-0.5 rounded text-xs font-bold', groupColors[primary.group] || groupColors.classic)}>
                    {strategyOptions.find(o => o.value === selected)?.groupLabel}
                  </span>
                  <span className="text-sm font-medium text-slate-600 dark:text-slate-300">{primary.label}</span>
                </div>
                {!primary.exact && (
                  <span className="flex items-center gap-1 text-xs text-amber-600 dark:text-amber-400" title="Heuristic estimate — exact tokenizer not available offline">
                    <Info className="w-3.5 h-3.5" /> Estimate
                  </span>
                )}
              </div>

              <div className="text-center py-4">
                <div className="text-5xl font-bold text-slate-900 dark:text-white tabular-nums">
                  {primary.count.toLocaleString()}
                </div>
                <div className="text-sm text-slate-500 dark:text-slate-400 mt-1">tokens</div>
              </div>

              {primary.encoding && (
                <div className="text-xs text-slate-400 dark:text-slate-500 text-center font-mono">
                  encoding: {primary.encoding}
                </div>
              )}

              {/* Token preview toggle */}
              {primary.sample.length > 0 && (
                <div>
                  <button
                    onClick={() => setShowTokens(!showTokens)}
                    className="w-full flex items-center justify-center gap-1.5 text-xs font-bold text-kawa-600 dark:text-kawa-400 hover:text-kawa-700 dark:hover:text-kawa-300 transition-colors py-2"
                  >
                    {showTokens ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                    {showTokens ? 'Hide' : 'Show'} Token Preview
                  </button>

                  {showTokens && (
                    <div className="border-t border-slate-200 dark:border-neutral-700 pt-3 mt-2 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                          {primary.sample.length < primary.count
                            ? `First ${primary.sample.length} of ${primary.count.toLocaleString()} tokens`
                            : `${primary.count.toLocaleString()} tokens`}
                        </span>
                        <button
                          onClick={() => copy(primary.sample.join(' | '), 'primary-sample')}
                          className="flex items-center gap-1 text-xs text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors"
                        >
                          {copied === 'primary-sample'
                            ? <><Check className="w-3.5 h-3.5" /> Copied</>
                            : <><Copy className="w-3.5 h-3.5" /> Copy</>}
                        </button>
                      </div>
                      <div className="flex flex-wrap gap-1.5 max-h-48 overflow-y-auto">
                        {primary.sample.map((token, i) => (
                          <span
                            key={i}
                            className={cn(
                              'inline-block px-1.5 py-0.5 rounded text-xs font-mono border',
                              i % 2 === 0
                                ? 'bg-kawa-50 border-kawa-200 text-kawa-800 dark:bg-kawa-900/20 dark:border-kawa-700/50 dark:text-kawa-300'
                                : 'bg-slate-50 border-slate-200 text-slate-700 dark:bg-neutral-700/50 dark:border-neutral-600 dark:text-slate-300',
                            )}
                            title={`Token #${i + 1}`}
                          >
                            {token.replace(/ /g, '·').replace(/\n/g, '\\n').replace(/\t/g, '\\t')}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center justify-center h-40 bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 text-slate-400 dark:text-slate-500 text-sm">
              Enter text to see token count
            </div>
          )}

          {/* Comparison table */}
          {others.length > 0 && (
            <div className="bg-white dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 shadow-sm overflow-hidden">
              <div className="px-4 py-3 border-b border-slate-200 dark:border-neutral-700">
                <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                  Compare with other tokenizers
                </span>
              </div>
              <div className="divide-y divide-slate-100 dark:divide-neutral-700/50 max-h-[300px] overflow-y-auto">
                {others.map(s => {
                  const opt = strategyOptions.find(o => o.value === s.name);
                  return (
                    <button
                      key={s.name}
                      onClick={() => { setSelected(s.name); setShowTokens(false); }}
                      className="w-full flex items-center justify-between px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-neutral-700/30 transition-colors text-left"
                    >
                      <div className="flex items-center gap-2">
                        <span className={cn('px-1.5 py-0.5 rounded text-[10px] font-bold', groupColors[s.group] || groupColors.classic)}>
                          {opt?.groupLabel}
                        </span>
                        <span className="text-sm text-slate-700 dark:text-slate-300">{s.label}</span>
                        {!s.exact && (
                          <span className="text-[10px] text-amber-500 dark:text-amber-400">~</span>
                        )}
                      </div>
                      <span className="text-sm font-bold text-slate-900 dark:text-white tabular-nums">
                        {s.count.toLocaleString()}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
