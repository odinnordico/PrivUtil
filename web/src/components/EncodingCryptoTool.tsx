import { useState, useCallback, useRef, useEffect } from 'react';
import { client } from '../lib/client';
import { Copy, Check, RefreshCw, ShieldCheck, Eye, EyeOff } from 'lucide-react';
import { cn } from '../lib/utils';

// ─── shared ───────────────────────────────────────────────────────────────────

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
        className,
      )}
    >
      {copied === id
        ? <><Check className="w-3.5 h-3.5" /> Copied</>
        : <><Copy className="w-3.5 h-3.5" /> Copy</>}
    </button>
  );
}

function ResultRow({ label, value, id, copied, copy, mono = true }: {
  label: string; value: string; id: string; copied: string | null;
  copy: (t: string, k: string) => void; mono?: boolean;
}) {
  return (
    <div className="bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-3">
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{label}</span>
        <CopyBtn text={value} id={id} copied={copied} copy={copy} />
      </div>
      <code className={cn('text-sm text-slate-800 dark:text-slate-200 break-all', mono && 'font-mono')}>
        {value || '—'}
      </code>
    </div>
  );
}

function ErrorBox({ msg }: { msg: string }) {
  return (
    <div className="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-500/30 rounded-lg px-4 py-2">
      {msg}
    </div>
  );
}

function useDebounce<T>(fn: (v: T) => void, delay = 300) {
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current); }, []);
  return useCallback((v: T) => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => fn(v), delay);
  }, [fn, delay]);
}

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const selectClass = "bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 text-sm";
const labelClass = "block text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5";

const tabs = [
  { id: 'hmac',      label: 'HMAC' },
  { id: 'otp',       label: 'OTP / TOTP' },
  { id: 'ulid',      label: 'ULID' },
  { id: 'caesar',    label: 'Caesar / ROT13' },
  { id: 'textencode',label: 'Text Encode' },
  { id: 'morse',     label: 'Morse Code' },
  { id: 'basicauth', label: 'Basic Auth' },
] as const;
type TabId = typeof tabs[number]['id'];

// ─── HMAC ─────────────────────────────────────────────────────────────────────

function HmacTab() {
  const [message, setMessage] = useState('');
  const [secret, setSecret] = useState('');
  const [algo, setAlgo] = useState('sha256');
  const [result, setResult] = useState<{ hex: string; base64: string } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (msg: string, sec: string, alg: string) => {
    if (!msg) { setResult(null); setError(''); return; }
    try {
      const resp = await client.hmacGenerate({ message: msg, secret: sec, algo: alg } as Parameters<typeof client.hmacGenerate>[0]);
      if (resp.error) { setError(resp.error); setResult(null); }
      else { setError(''); setResult({ hex: resp.hex, base64: resp.base64 }); }
    } catch { setError('Server error'); }
  }, []);

  const debounced = useDebounce(([msg, sec, alg]: [string, string, string]) => call(msg, sec, alg), 250);

  const handleChange = (msg = message, sec = secret, alg = algo) => {
    if (msg !== message) setMessage(msg);
    if (sec !== secret) setSecret(sec);
    if (alg !== algo) setAlgo(alg);
    debounced([msg, sec, alg]);
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className={labelClass}>Message</label>
          <textarea
            rows={3}
            value={message}
            onChange={e => handleChange(e.target.value, secret, algo)}
            placeholder="Enter your message"
            className={cn(inputClass, 'resize-y')}
          />
        </div>
        <div>
          <label className={labelClass}>Secret Key</label>
          <textarea
            rows={3}
            value={secret}
            onChange={e => handleChange(message, e.target.value, algo)}
            placeholder="Enter your secret"
            className={cn(inputClass, 'resize-y')}
          />
        </div>
      </div>
      <div className="flex items-center gap-3">
        <label className={cn(labelClass, 'mb-0 whitespace-nowrap')}>Algorithm</label>
        <select value={algo} onChange={e => handleChange(message, secret, e.target.value)} className={selectClass}>
          <option value="sha256">HMAC-SHA256</option>
          <option value="sha512">HMAC-SHA512</option>
          <option value="sha1">HMAC-SHA1</option>
          <option value="md5">HMAC-MD5</option>
        </select>
      </div>
      {error && <ErrorBox msg={error} />}
      {result && (
        <div className="space-y-2">
          <ResultRow label="Hex" value={result.hex} id="hmac-hex" copied={copied} copy={copy} />
          <ResultRow label="Base64" value={result.base64} id="hmac-b64" copied={copied} copy={copy} />
        </div>
      )}
    </div>
  );
}

// ─── OTP ──────────────────────────────────────────────────────────────────────

function TimerDisplay({ initial, period }: { initial: number; period: number }) {
  const [timeLeft, setTimeLeft] = useState(initial);
  useEffect(() => {
    const id = setInterval(() => setTimeLeft(t => Math.max(0, t - 1)), 1000);
    return () => clearInterval(id);
  }, []);
  return (
    <>
      <div className="relative h-2 bg-slate-200 dark:bg-neutral-700 rounded-full overflow-hidden">
        <div
          className={cn(
            'absolute left-0 top-0 h-full rounded-full transition-all duration-1000',
            timeLeft > 10 ? 'bg-kawa-500' : timeLeft > 5 ? 'bg-amber-400' : 'bg-red-500',
          )}
          style={{ width: `${(timeLeft / period) * 100}%` }}
        />
      </div>
      <div className="text-2xl font-mono font-bold text-slate-700 dark:text-slate-200 mt-1">{timeLeft}s</div>
    </>
  );
}

function OtpTab() {
  const [secret, setSecret] = useState('');
  const [type, setType] = useState<'totp' | 'hotp'>('totp');
  const [digits, setDigits] = useState<6 | 8>(6);
  const [period, setPeriod] = useState(30);
  const [algo, setAlgo] = useState('sha1');
  const [counter, setCounter] = useState(0);
  const [mode, setMode] = useState<'generate' | 'validate'>('generate');
  const [validateCode, setValidateCode] = useState('');
  const [result, setResult] = useState<{
    code: string; secret: string; timeRemaining: number; uri: string;
  } | null>(null);
  const [validateResult, setValidateResult] = useState<boolean | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const generate = useCallback(async (opts: {
    secret: string; type: string; digits: number; period: number; algo: string;
    counter: number; generateSecret?: boolean;
  }) => {
    setError('');
    setResult(null);
    setValidateResult(null);
    try {
      const resp = await client.otpGenerate({
        secret: opts.secret,
        type: opts.type,
        digits: opts.digits,
        period: opts.period,
        algo: opts.algo,
        counter: opts.counter,
        generateSecret: opts.generateSecret ?? false,
      } as unknown as Parameters<typeof client.otpGenerate>[0]);
      if (resp.error) { setError(resp.error); }
      else {
        setResult({
          code: resp.code,
          secret: resp.secret,
          timeRemaining: Number(resp.timeRemaining),
          uri: resp.uri,
        });
        if (resp.secret && opts.generateSecret) setSecret(resp.secret);
      }
    } catch { setError('Server error'); }
  }, []);

  const validate = useCallback(async () => {
    setError('');
    try {
      const resp = await client.otpValidate({
        secret,
        code: validateCode,
        window: 1,
        period,
        algo,
      } as Parameters<typeof client.otpValidate>[0]);
      if (resp.error) { setError(resp.error); setValidateResult(null); }
      else { setValidateResult(resp.valid); }
    } catch { setError('Server error'); }
  }, [secret, validateCode, period, algo]);

  const triggerGenerate = () => generate({ secret, type, digits, period, algo, counter });

  return (
    <div className="space-y-4">
      {/* Settings */}
      <div className="flex flex-wrap gap-3 items-end bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex-1 min-w-48">
          <label className={labelClass}>Secret (Base32)</label>
          <input
            type="text"
            value={secret}
            onChange={e => setSecret(e.target.value.toUpperCase())}
            placeholder="JBSWY3DPEHPK3PXP"
            className={inputClass}
            spellCheck={false}
          />
        </div>
        <div>
          <label className={labelClass}>Type</label>
          <select value={type} onChange={e => setType(e.target.value as 'totp' | 'hotp')} className={selectClass}>
            <option value="totp">TOTP (time-based)</option>
            <option value="hotp">HOTP (counter-based)</option>
          </select>
        </div>
        <div>
          <label className={labelClass}>Digits</label>
          <select value={digits} onChange={e => setDigits(+e.target.value as 6 | 8)} className={selectClass}>
            <option value={6}>6</option>
            <option value={8}>8</option>
          </select>
        </div>
        {type === 'totp' && (
          <div className="w-24">
            <label className={labelClass}>Period (s)</label>
            <select value={period} onChange={e => setPeriod(+e.target.value)} className={selectClass}>
              <option value={30}>30</option>
              <option value={60}>60</option>
            </select>
          </div>
        )}
        {type === 'hotp' && (
          <div className="w-28">
            <label className={labelClass}>Counter</label>
            <input type="number" min={0} value={counter} onChange={e => setCounter(+e.target.value)} className={inputClass} />
          </div>
        )}
        <div>
          <label className={labelClass}>Algorithm</label>
          <select value={algo} onChange={e => setAlgo(e.target.value)} className={selectClass}>
            <option value="sha1">SHA-1</option>
            <option value="sha256">SHA-256</option>
            <option value="sha512">SHA-512</option>
          </select>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={triggerGenerate}
          className="flex items-center gap-2 px-4 py-2 bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold rounded-lg shadow-md transition-all active:scale-95"
        >
          <RefreshCw className="w-4 h-4" />
          {type === 'totp' ? 'Get Current Code' : 'Generate Code'}
        </button>
        <button
          onClick={() => generate({ secret: '', type, digits, period, algo, counter, generateSecret: true })}
          className="flex items-center gap-2 px-4 py-2 bg-slate-100 dark:bg-slate-700 hover:bg-kawa-500/20 text-slate-700 dark:text-slate-300 font-medium rounded-lg transition-all active:scale-95"
        >
          New Secret
        </button>
        <button
          onClick={() => setMode(mode === 'generate' ? 'validate' : 'generate')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all active:scale-95',
            mode === 'validate'
              ? 'bg-kawa-500/20 text-kawa-600 dark:text-kawa-400'
              : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-kawa-500/10',
          )}
        >
          <ShieldCheck className="w-4 h-4" />
          Validate
        </button>
      </div>

      {error && <ErrorBox msg={error} />}

      {/* Validate mode */}
      {mode === 'validate' && (
        <div className="flex items-end gap-3 bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700">
          <div className="flex-1">
            <label className={labelClass}>Enter Code to Validate</label>
            <input
              type="text"
              value={validateCode}
              onChange={e => setValidateCode(e.target.value)}
              placeholder="123456"
              maxLength={8}
              className={inputClass}
            />
          </div>
          <button
            onClick={validate}
            className="px-4 py-2 bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold rounded-lg shadow-md transition-all active:scale-95"
          >
            Check
          </button>
          {validateResult !== null && (
            <div className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg font-semibold text-sm',
              validateResult
                ? 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400'
                : 'bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400',
            )}>
              {validateResult ? <><Check className="w-4 h-4" /> Valid</> : '✗ Invalid'}
            </div>
          )}
        </div>
      )}

      {/* Result */}
      {result && mode === 'generate' && (
        <div className="space-y-3">
          <div className="flex items-center gap-6 bg-kawa-500/10 dark:bg-kawa-500/20 rounded-xl border border-kawa-400/30 p-5">
            <div className="text-center">
              <div className="text-5xl font-mono font-black text-kawa-600 dark:text-kawa-400 tracking-widest">
                {result.code}
              </div>
              <CopyBtn text={result.code} id="otp-code" copied={copied} copy={copy} className="justify-center mt-2" />
            </div>
            {type === 'totp' && (
              <div className="flex-1">
                <div className="text-xs text-slate-500 dark:text-slate-400 mb-1.5">Expires in</div>
                <TimerDisplay key={result.timeRemaining} initial={result.timeRemaining} period={period} />
              </div>
            )}
          </div>
          {result.secret && (
            <ResultRow label="Secret" value={result.secret} id="otp-secret" copied={copied} copy={copy} />
          )}
          <ResultRow label="otpauth:// URI" value={result.uri} id="otp-uri" copied={copied} copy={copy} />
        </div>
      )}
    </div>
  );
}

// ─── ULID ─────────────────────────────────────────────────────────────────────

function UlidTab() {
  const [count, setCount] = useState(5);
  const [monotonic, setMonotonic] = useState(true);
  const [ulids, setUlids] = useState<string[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { copied, copy } = useCopy();

  const generate = async () => {
    setLoading(true); setError('');
    try {
      const resp = await client.ulidGenerate({ count, monotonic } as Parameters<typeof client.ulidGenerate>[0]);
      if (resp.error) { setError(resp.error); setUlids([]); }
      else { setUlids(resp.ulids as string[]); }
    } catch { setError('Server error'); }
    finally { setLoading(false); }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-end bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="w-24">
          <label className={labelClass}>Count</label>
          <input type="number" min={1} max={100} value={count}
            onChange={e => setCount(Math.min(100, Math.max(1, +e.target.value)))}
            className={inputClass} />
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer select-none pb-2">
          <input type="checkbox" checked={monotonic} onChange={e => setMonotonic(e.target.checked)} className="w-4 h-4 accent-kawa-500" />
          Monotonic (ordered within same ms)
        </label>
        <button
          onClick={generate} disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-kawa-500 hover:bg-kawa-600 text-slate-900 font-bold rounded-lg shadow-md transition-all active:scale-95 disabled:opacity-50 ml-auto"
        >
          <RefreshCw className={cn('w-4 h-4', loading && 'animate-spin')} />
          Generate
        </button>
      </div>

      {error && <ErrorBox msg={error} />}

      {ulids.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{ulids.length} ULIDs</span>
            <CopyBtn text={ulids.join('\n')} id="ulids-all" copied={copied} copy={copy} />
          </div>
          <div className="space-y-1.5">
            {ulids.map((u, i) => (
              <div key={i} className="flex items-center justify-between bg-white dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 px-4 py-2">
                <code className="font-mono text-sm text-slate-800 dark:text-slate-200 tracking-wider">{u}</code>
                <CopyBtn text={u} id={`ulid-${i}`} copied={copied} copy={copy} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Caesar / ROT13 ───────────────────────────────────────────────────────────

function CaesarTab() {
  const [input, setInput] = useState('');
  const [shift, setShift] = useState(13);
  const [action, setAction] = useState<'encode' | 'decode'>('encode');
  const [output, setOutput] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (text: string, sh: number, act: string) => {
    if (!text) { setOutput(''); setError(''); return; }
    try {
      const resp = await client.caesarCipher({
        text, shift: sh, action: act,
      } as Parameters<typeof client.caesarCipher>[0]);
      if (resp.error) { setError(resp.error); setOutput(''); }
      else { setError(''); setOutput(resp.result); }
    } catch { setError('Server error'); }
  }, []);

  const debounced = useDebounce(([text, sh, act]: [string, number, string]) => call(text, sh, act), 200);
  const update = (text = input, sh = shift, act = action) => {
    if (text !== input) setInput(text);
    if (sh !== shift) setShift(sh);
    if (act !== action) setAction(act as 'encode' | 'decode');
    debounced([text, sh, act]);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-3 items-center bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex items-center gap-2">
          <label className={cn(labelClass, 'mb-0')}>Shift</label>
          <input
            type="range" min={0} max={25} value={shift}
            onChange={e => update(input, +e.target.value, action)}
            className="w-32 accent-kawa-500"
          />
          <span className="w-6 text-center font-mono font-bold text-slate-700 dark:text-slate-200">{shift}</span>
        </div>

        <div className="flex rounded-lg overflow-hidden border border-slate-300 dark:border-neutral-600">
          {(['encode', 'decode'] as const).map(a => (
            <button
              key={a}
              onClick={() => update(input, shift, a)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium capitalize transition-colors',
                action === a
                  ? 'bg-kawa-500 text-slate-900'
                  : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-400 hover:bg-kawa-500/10',
              )}
            >{a}</button>
          ))}
        </div>

        <div className="flex gap-1.5 ml-auto">
          {[13, 1, 3, 7, 25].map(s => (
            <button
              key={s}
              onClick={() => update(input, s, action)}
              className={cn(
                'px-2.5 py-1 text-xs font-mono rounded border transition-colors',
                shift === s
                  ? 'bg-kawa-500 text-slate-900 border-kawa-500'
                  : 'bg-white dark:bg-neutral-800 border-slate-300 dark:border-neutral-600 hover:border-kawa-400 text-slate-600 dark:text-slate-400',
              )}
            >{s === 13 ? 'ROT13' : `+${s}`}</button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className={labelClass}>Input</label>
          <textarea
            rows={8}
            value={input}
            onChange={e => update(e.target.value, shift, action)}
            placeholder="Enter text to cipher…"
            className={cn(inputClass, 'resize-none h-48 font-mono')}
          />
        </div>
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className={cn(labelClass, 'mb-0')}>Output</label>
            <CopyBtn text={output} id="caesar-out" copied={copied} copy={copy} />
          </div>
          {error
            ? <ErrorBox msg={error} />
            : <textarea
                readOnly rows={8}
                value={output}
                placeholder="Output appears here…"
                className={cn(inputClass, 'resize-none h-48 font-mono bg-slate-50 dark:bg-black/30 cursor-default focus:ring-0')}
              />
          }
        </div>
      </div>
    </div>
  );
}

// ─── Text Encode ──────────────────────────────────────────────────────────────

const TEXT_FORMATS = ['binary', 'hex', 'octal', 'decimal'] as const;
type TextFmt = typeof TEXT_FORMATS[number];

function TextEncodeTab() {
  const [input, setInput] = useState('');
  const [format, setFormat] = useState<TextFmt>('hex');
  const [action, setAction] = useState<'encode' | 'decode'>('encode');
  const [output, setOutput] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (text: string, fmt: string, act: string) => {
    if (!text) { setOutput(''); setError(''); return; }
    try {
      const resp = await client.textEncode({
        text, format: fmt, action: act,
      } as Parameters<typeof client.textEncode>[0]);
      if (resp.error) { setError(resp.error); setOutput(''); }
      else { setError(''); setOutput(resp.result); }
    } catch { setError('Server error'); }
  }, []);

  const debounced = useDebounce(([text, fmt, act]: [string, string, string]) => call(text, fmt, act), 200);
  const update = (text = input, fmt: string = format, act = action) => {
    if (text !== input) setInput(text);
    if (fmt !== format) setFormat(fmt as TextFmt);
    if (act !== action) setAction(act as 'encode' | 'decode');
    debounced([text, fmt, act]);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-3 items-center bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex gap-1 rounded-lg overflow-hidden border border-slate-300 dark:border-neutral-600">
          {TEXT_FORMATS.map(f => (
            <button
              key={f}
              onClick={() => update(input, f, action)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium capitalize transition-colors',
                format === f
                  ? 'bg-kawa-500 text-slate-900'
                  : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-400 hover:bg-kawa-500/10',
              )}
            >{f}</button>
          ))}
        </div>

        <div className="flex rounded-lg overflow-hidden border border-slate-300 dark:border-neutral-600">
          {(['encode', 'decode'] as const).map(a => (
            <button
              key={a}
              onClick={() => update(input, format, a)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium capitalize transition-colors',
                action === a
                  ? 'bg-kawa-500 text-slate-900'
                  : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-400 hover:bg-kawa-500/10',
              )}
            >{a}</button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className={labelClass}>{action === 'encode' ? 'Text' : format.charAt(0).toUpperCase() + format.slice(1)}</label>
          <textarea
            rows={8}
            value={input}
            onChange={e => update(e.target.value)}
            placeholder={action === 'encode' ? 'Enter plain text…' : `Enter ${format}-encoded values separated by spaces…`}
            className={cn(inputClass, 'resize-none h-48 font-mono')}
          />
        </div>
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className={cn(labelClass, 'mb-0')}>
              {action === 'encode' ? format.charAt(0).toUpperCase() + format.slice(1) : 'Decoded Text'}
            </label>
            <CopyBtn text={output} id="textencode-out" copied={copied} copy={copy} />
          </div>
          {error
            ? <ErrorBox msg={error} />
            : <textarea
                readOnly
                value={output}
                placeholder="Output appears here…"
                className={cn(inputClass, 'resize-none h-48 font-mono bg-slate-50 dark:bg-black/30 cursor-default focus:ring-0')}
              />
          }
        </div>
      </div>
    </div>
  );
}

// ─── Morse Code ───────────────────────────────────────────────────────────────

function MorseTab() {
  const [input, setInput] = useState('');
  const [action, setAction] = useState<'encode' | 'decode'>('encode');
  const [output, setOutput] = useState('');
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (text: string, act: string) => {
    if (!text) { setOutput(''); setError(''); return; }
    try {
      const resp = await client.morseCode({ text, action: act } as Parameters<typeof client.morseCode>[0]);
      if (resp.error) { setError(resp.error); setOutput(''); }
      else { setError(''); setOutput(resp.result); }
    } catch { setError('Server error'); }
  }, []);

  const debounced = useDebounce(([text, act]: [string, string]) => call(text, act), 200);
  const update = (text = input, act = action) => {
    if (text !== input) setInput(text);
    if (act !== action) setAction(act as 'encode' | 'decode');
    debounced([text, act]);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div className="flex rounded-lg overflow-hidden border border-slate-300 dark:border-neutral-600">
          {(['encode', 'decode'] as const).map(a => (
            <button
              key={a}
              onClick={() => update(input, a)}
              className={cn(
                'px-3 py-1.5 text-sm font-medium capitalize transition-colors',
                action === a
                  ? 'bg-kawa-500 text-slate-900'
                  : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-400 hover:bg-kawa-500/10',
              )}
            >{a}</button>
          ))}
        </div>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          {action === 'encode'
            ? 'Supports A–Z, 0–9, and common punctuation. Spaces become " / ".'
            : 'Use dots and dashes separated by spaces. " / " separates words.'}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className={labelClass}>{action === 'encode' ? 'Plain Text' : 'Morse Code'}</label>
          <textarea
            rows={8}
            value={input}
            onChange={e => update(e.target.value)}
            placeholder={action === 'encode' ? 'SOS' : '... --- ...'}
            className={cn(inputClass, 'resize-none h-48 font-mono')}
          />
        </div>
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className={cn(labelClass, 'mb-0')}>{action === 'encode' ? 'Morse Code' : 'Plain Text'}</label>
            <CopyBtn text={output} id="morse-out" copied={copied} copy={copy} />
          </div>
          {error
            ? <ErrorBox msg={error} />
            : <textarea
                readOnly
                value={output}
                placeholder="Output appears here…"
                className={cn(inputClass, 'resize-none h-48 font-mono bg-slate-50 dark:bg-black/30 cursor-default focus:ring-0 leading-relaxed')}
              />
          }
        </div>
      </div>
    </div>
  );
}

// ─── Basic Auth ───────────────────────────────────────────────────────────────

function BasicAuthTab() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPass, setShowPass] = useState(false);
  const [result, setResult] = useState<{ header: string; token: string; decoded: string } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();

  const call = useCallback(async (u: string, p: string) => {
    if (!u) { setResult(null); setError(''); return; }
    try {
      const resp = await client.basicAuthGenerate({ username: u, password: p } as Parameters<typeof client.basicAuthGenerate>[0]);
      if (resp.error) { setError(resp.error); setResult(null); }
      else { setError(''); setResult({ header: resp.header, token: resp.token, decoded: resp.decoded }); }
    } catch { setError('Server error'); }
  }, []);

  const debounced = useDebounce(([u, p]: [string, string]) => call(u, p), 200);
  const update = (u = username, p = password) => {
    if (u !== username) setUsername(u);
    if (p !== password) setPassword(p);
    debounced([u, p]);
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 bg-white dark:bg-neutral-800/50 p-4 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        <div>
          <label className={labelClass}>Username</label>
          <input
            type="text"
            value={username}
            onChange={e => update(e.target.value, password)}
            placeholder="admin"
            className={inputClass}
            autoComplete="off"
          />
        </div>
        <div>
          <label className={labelClass}>Password</label>
          <div className="relative">
            <input
              type={showPass ? 'text' : 'password'}
              value={password}
              onChange={e => update(username, e.target.value)}
              placeholder="••••••••"
              className={cn(inputClass, 'pr-10')}
              autoComplete="off"
            />
            <button
              onClick={() => setShowPass(v => !v)}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
            >
              {showPass ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>
      </div>

      {error && <ErrorBox msg={error} />}

      {result && (
        <div className="space-y-2">
          <ResultRow label="Authorization Header" value={result.header} id="auth-header" copied={copied} copy={copy} />
          <ResultRow label="Token (Base64)" value={result.token} id="auth-token" copied={copied} copy={copy} />
          <ResultRow label="Decoded (user:password)" value={result.decoded} id="auth-decoded" copied={copied} copy={copy} />
        </div>
      )}
    </div>
  );
}

// ─── root ─────────────────────────────────────────────────────────────────────

export function EncodingCryptoTool() {
  const [activeTab, setActiveTab] = useState<TabId>('hmac');

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Encoding & Crypto</h2>

      <div className="flex flex-wrap gap-1 bg-white dark:bg-neutral-800/50 p-1 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'px-3 py-1.5 rounded text-sm font-medium transition-colors',
              activeTab === tab.id
                ? 'bg-kawa-500 text-slate-900'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-700',
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="bg-white dark:bg-neutral-900/30 rounded-lg border border-slate-200 dark:border-neutral-700 p-5 shadow-sm">
        {activeTab === 'hmac'       && <HmacTab />}
        {activeTab === 'otp'        && <OtpTab />}
        {activeTab === 'ulid'       && <UlidTab />}
        {activeTab === 'caesar'     && <CaesarTab />}
        {activeTab === 'textencode' && <TextEncodeTab />}
        {activeTab === 'morse'      && <MorseTab />}
        {activeTab === 'basicauth'  && <BasicAuthTab />}
      </div>
    </div>
  );
}
