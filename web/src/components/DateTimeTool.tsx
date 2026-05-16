import { useState, useCallback, useEffect } from 'react';
import { client } from '../lib/client';
import { Copy, Check, Calendar, Clock, ArrowLeftRight, Plus, Minus, Info, RotateCcw } from 'lucide-react';
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
      {copied === id
        ? <><Check className="w-3.5 h-3.5" /> Copied</>
        : <><Copy className="w-3.5 h-3.5" /> Copy</>}
    </button>
  );
}

function ErrorMsg({ msg }: { msg: string }) {
  return msg
    ? <p className="text-sm text-red-500 dark:text-red-400 mt-1">{msg}</p>
    : null;
}

function useDebounce<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

const tabs = [
  { id: 'diff',    label: 'Date Diff',    icon: ArrowLeftRight },
  { id: 'leap',    label: 'Leap Year',    icon: Calendar       },
  { id: 'add',     label: 'Date Add/Sub', icon: Plus           },
  { id: 'format',  label: 'Formats',      icon: Clock          },
  { id: 'info',    label: 'Date Info',    icon: Info           },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const numClass = cn(inputClass, 'font-mono text-center w-20');

function InfoCard({ label, value, sub, id, copied, copy }: {
  label: string; value: string | number; sub?: string; id: string;
  copied: string | null; copy: (t: string, k: string) => void;
}) {
  const txt = String(value);
  return (
    <div className="bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-3">
      <div className="flex items-center justify-between mb-0.5">
        <span className="text-xs font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wide">{label}</span>
        <CopyBtn text={txt} id={id} copied={copied} copy={copy} />
      </div>
      <div className="font-mono text-sm font-semibold text-slate-800 dark:text-slate-200 break-all">{txt}</div>
      {sub && <div className="text-xs text-slate-400 mt-0.5">{sub}</div>}
    </div>
  );
}

// ─── Date Diff ────────────────────────────────────────────────────────────────

type DiffResult = {
  years: number; months: number; weeks: number; days: number;
  hours: number; minutes: number; seconds: number;
  totalDays: number; totalHours: number; totalSecs: number;
  negative: boolean; human: string;
};

function DiffBlock({ value, unit }: { value: number; unit: string }) {
  return (
    <div className="flex flex-col items-center bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-3 min-w-[70px]">
      <span className="text-2xl font-bold font-mono text-kawa-600 dark:text-kawa-400">{value}</span>
      <span className="text-xs text-slate-500 mt-0.5">{unit}</span>
    </div>
  );
}

function DateDiffTab() {
  const [from, setFrom] = useState(todayISO());
  const [to, setTo] = useState(todayISO());
  const [result, setResult] = useState<DiffResult | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const dFrom = useDebounce(from);
  const dTo = useDebounce(to);

  useEffect(() => {
    if (!dFrom || !dTo) return;
    client.dateDiff({ fromDate: dFrom, toDate: dTo } as Parameters<typeof client.dateDiff>[0])
      .then(r => { setError(r.error); if (!r.error) setResult(r); })
      .catch(e => setError(String(e)));
  }, [dFrom, dTo]);

  const swap = () => { setFrom(to); setTo(from); };

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Calculate the duration between two dates — calendar-accurate breakdown plus totals.
      </p>

      <div className="flex flex-col sm:flex-row gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">From</label>
          <input type="date" className={inputClass} value={from} onChange={e => setFrom(e.target.value)} />
        </div>
        <button
          onClick={swap}
          className="flex items-center gap-1 text-sm text-slate-500 hover:text-kawa-600 dark:hover:text-kawa-400 transition-colors pb-2"
          title="Swap dates"
        >
          <ArrowLeftRight className="w-4 h-4" />
        </button>
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">To</label>
          <input type="date" className={inputClass} value={to} onChange={e => setTo(e.target.value)} />
        </div>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => { setFrom(todayISO()); }}
          className="text-xs text-kawa-600 dark:text-kawa-400 hover:underline"
        >
          Set From → today
        </button>
        <span className="text-slate-300 dark:text-neutral-600">·</span>
        <button
          onClick={() => { setTo(todayISO()); }}
          className="text-xs text-kawa-600 dark:text-kawa-400 hover:underline"
        >
          Set To → today
        </button>
      </div>

      <ErrorMsg msg={error} />

      {dFrom && dTo && result && !error && (
        <div className="space-y-4">
          {result.negative && (
            <div className="text-xs text-amber-600 dark:text-amber-400 font-medium">
              ⚠ From is after To — showing absolute difference
            </div>
          )}

          {/* Human summary */}
          <div className="bg-kawa-50 dark:bg-kawa-900/20 rounded-xl border border-kawa-200 dark:border-kawa-800 p-4">
            <div className="flex items-start justify-between">
              <p className="text-base font-semibold text-kawa-700 dark:text-kawa-300">{result.human || '0 seconds'}</p>
              <CopyBtn text={result.human} id="diff-human" copied={copied} copy={copy} />
            </div>
          </div>

          {/* Calendar breakdown */}
          <div>
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Calendar breakdown</p>
            <div className="flex flex-wrap gap-2">
              <DiffBlock value={result.years}   unit="years"   />
              <DiffBlock value={result.months}  unit="months"  />
              <DiffBlock value={result.days}    unit="days"    />
              <DiffBlock value={result.hours}   unit="hours"   />
              <DiffBlock value={result.minutes} unit="minutes" />
              <DiffBlock value={result.seconds} unit="seconds" />
            </div>
          </div>

          {/* Totals */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <InfoCard label="Total days"  value={result.totalDays}  id="td" copied={copied} copy={copy} />
            <InfoCard label="Total hours" value={result.totalHours} id="th" copied={copied} copy={copy} />
            <InfoCard label="Total seconds" value={result.totalSecs} id="ts" copied={copied} copy={copy} />
          </div>
          <div className="text-xs text-slate-400 dark:text-slate-500">
            ≈ {result.weeks} week{result.weeks !== 1 ? 's' : ''}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Leap Year ────────────────────────────────────────────────────────────────

function LeapYearTab() {
  const [input, setInput] = useState(String(new Date().getFullYear()));
  const [results, setResults] = useState<{ year: number; isLeap: boolean }[]>([]);
  const [leapCount, setLeapCount] = useState(0);
  const [error, setError] = useState('');
  const debounced = useDebounce(input);

  useEffect(() => {
    const trimmed = debounced.trim();
    if (!trimmed) return;
    client.leapYear({ input: trimmed } as Parameters<typeof client.leapYear>[0])
      .then(r => { setError(r.error); if (!r.error) { setResults(r.results); setLeapCount(r.leapCount); } })
      .catch(e => setError(String(e)));
  }, [debounced]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Check whether years are leap years. Enter a single year, a comma-separated list, or a range like <code className="text-kawa-600 dark:text-kawa-400">2020-2030</code>.
      </p>
      <div>
        <label className="block text-xs text-slate-500 mb-1">Year(s)</label>
        <input
          className={inputClass}
          placeholder="2024  or  2020,2021,2022  or  2000-2010"
          value={input}
          onChange={e => setInput(e.target.value)}
        />
      </div>
      <ErrorMsg msg={error} />
      {debounced.trim() && results.length > 0 && !error && (
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400">
            <span className="font-semibold text-kawa-600 dark:text-kawa-400">{leapCount}</span> leap year{leapCount !== 1 ? 's' : ''} out of {results.length}
          </div>
          {results.length === 1 ? (
            <div className={cn(
              'rounded-xl border p-5 text-center',
              results[0].isLeap
                ? 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800'
                : 'bg-slate-50 dark:bg-neutral-800 border-slate-200 dark:border-neutral-700'
            )}>
              <div className="text-4xl font-bold font-mono mb-1">{results[0].year}</div>
              <div className={cn(
                'text-lg font-semibold',
                results[0].isLeap ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'
              )}>
                {results[0].isLeap ? '✓ Leap year (366 days)' : '✗ Common year (365 days)'}
              </div>
            </div>
          ) : (
            <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden max-h-80 overflow-y-auto">
              <table className="w-full text-sm">
                <thead className="bg-slate-100 dark:bg-neutral-800 sticky top-0">
                  <tr>
                    <th className="text-left px-4 py-2 text-xs font-semibold text-slate-500 uppercase">Year</th>
                    <th className="text-left px-4 py-2 text-xs font-semibold text-slate-500 uppercase">Type</th>
                    <th className="text-right px-4 py-2 text-xs font-semibold text-slate-500 uppercase">Days</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                  {results.map(r => (
                    <tr
                      key={r.year}
                      className={cn(
                        r.isLeap
                          ? 'bg-emerald-50 dark:bg-emerald-900/10'
                          : 'bg-white dark:bg-neutral-900'
                      )}
                    >
                      <td className="px-4 py-2 font-mono font-semibold">{r.year}</td>
                      <td className={cn(
                        'px-4 py-2 font-medium',
                        r.isLeap ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400'
                      )}>
                        {r.isLeap ? '✓ Leap' : '✗ Common'}
                      </td>
                      <td className="px-4 py-2 text-right font-mono text-slate-500">{r.isLeap ? 366 : 365}</td>
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

// ─── Date Add/Sub ─────────────────────────────────────────────────────────────

type AddFields = { years: number; months: number; weeks: number; days: number; hours: number; minutes: number; seconds: number };

function SpinField({ label, value, onChange }: { label: string; value: number; onChange: (v: number) => void }) {
  return (
    <div className="flex flex-col items-center gap-1">
      <label className="text-xs text-slate-500 dark:text-slate-400">{label}</label>
      <div className="flex items-center border border-slate-300 dark:border-neutral-700 rounded-lg overflow-hidden">
        <button
          onClick={() => onChange(value - 1)}
          className="px-2 py-1.5 text-slate-500 hover:bg-slate-100 dark:hover:bg-neutral-700 transition-colors"
        ><Minus className="w-3 h-3" /></button>
        <input
          type="number"
          className={cn(numClass, 'border-0 border-x border-slate-300 dark:border-neutral-700 rounded-none')}
          value={value}
          onChange={e => onChange(Number(e.target.value))}
        />
        <button
          onClick={() => onChange(value + 1)}
          className="px-2 py-1.5 text-slate-500 hover:bg-slate-100 dark:hover:bg-neutral-700 transition-colors"
        ><Plus className="w-3 h-3" /></button>
      </div>
    </div>
  );
}

function DateAddTab() {
  const [date, setDate] = useState(todayISO());
  const [fields, setFields] = useState<AddFields>({ years: 0, months: 0, weeks: 0, days: 0, hours: 0, minutes: 0, seconds: 0 });
  const [result, setResult] = useState<{ iso: string; isoFull: string; unix: string; weekday: string; dayOfYear: number; isoWeek: number; formatted: string } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const dDate = useDebounce(date);
  const dFields = useDebounce(fields);

  useEffect(() => {
    if (!dDate) return;
    client.dateAdd({
      date: dDate,
      ...dFields,
    } as Parameters<typeof client.dateAdd>[0])
      .then(r => { setError(r.error); if (!r.error) setResult(r); })
      .catch(e => setError(String(e)));
  }, [dDate, dFields]);

  const reset = () => setFields({ years: 0, months: 0, weeks: 0, days: 0, hours: 0, minutes: 0, seconds: 0 });
  const set = (k: keyof AddFields) => (v: number) => setFields(f => ({ ...f, [k]: v }));

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Add or subtract any combination of years, months, weeks, days, hours, minutes, and seconds from a date.
      </p>
      <div className="flex gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">Base date</label>
          <input type="date" className={inputClass} value={date} onChange={e => setDate(e.target.value)} />
        </div>
        <button
          onClick={() => setDate(todayISO())}
          className="flex items-center gap-1 text-xs text-kawa-600 dark:text-kawa-400 hover:underline pb-2"
        >
          <RotateCcw className="w-3 h-3" /> Today
        </button>
      </div>

      <div className="flex flex-wrap gap-3">
        <SpinField label="Years"   value={fields.years}   onChange={set('years')}   />
        <SpinField label="Months"  value={fields.months}  onChange={set('months')}  />
        <SpinField label="Weeks"   value={fields.weeks}   onChange={set('weeks')}   />
        <SpinField label="Days"    value={fields.days}    onChange={set('days')}    />
        <SpinField label="Hours"   value={fields.hours}   onChange={set('hours')}   />
        <SpinField label="Minutes" value={fields.minutes} onChange={set('minutes')} />
        <SpinField label="Seconds" value={fields.seconds} onChange={set('seconds')} />
      </div>

      <button onClick={reset} className="text-xs text-slate-400 hover:text-red-500 dark:hover:text-red-400 flex items-center gap-1 transition-colors">
        <RotateCcw className="w-3 h-3" /> Reset offsets
      </button>

      <ErrorMsg msg={error} />

      {dDate && result && !error && (
        <div className="space-y-3">
          {/* Big result */}
          <div className="bg-kawa-50 dark:bg-kawa-900/20 rounded-xl border border-kawa-200 dark:border-kawa-800 p-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-mono text-2xl font-bold text-kawa-700 dark:text-kawa-300">{result.iso}</div>
                <div className="text-sm text-slate-500 mt-0.5">{result.formatted}</div>
              </div>
              <CopyBtn text={result.iso} id="add-result" copied={copied} copy={copy} />
            </div>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <InfoCard label="RFC 3339"     value={result.isoFull}  id="add-full"    copied={copied} copy={copy} />
            <InfoCard label="Unix (s)"     value={result.unix}     id="add-unix"    copied={copied} copy={copy} />
            <InfoCard label="Weekday"      value={result.weekday}  id="add-weekday" copied={copied} copy={copy} />
            <InfoCard label="Day of year"  value={result.dayOfYear} id="add-doy"   copied={copied} copy={copy} />
            <InfoCard label="ISO week"     value={result.isoWeek}  id="add-week"    copied={copied} copy={copy} />
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Date Formatter ───────────────────────────────────────────────────────────

const COMMON_TIMEZONES = [
  'UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
  'America/Sao_Paulo', 'America/Toronto', 'Europe/London', 'Europe/Paris', 'Europe/Berlin',
  'Europe/Moscow', 'Africa/Cairo', 'Asia/Dubai', 'Asia/Kolkata', 'Asia/Bangkok',
  'Asia/Singapore', 'Asia/Shanghai', 'Asia/Tokyo', 'Australia/Sydney', 'Pacific/Auckland',
];

function DateFormatTab() {
  const [dateStr, setDateStr] = useState(todayISO());
  const [timezone, setTimezone] = useState('UTC');
  const [formats, setFormats] = useState<{ label: string; result: string }[]>([]);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const dDate = useDebounce(dateStr);
  const dTz = useDebounce(timezone, 500);

  useEffect(() => {
    if (!dDate) return;
    client.dateFormat({ dateStr: dDate, timezone: dTz || 'UTC' } as Parameters<typeof client.dateFormat>[0])
      .then(r => { setError(r.error); if (!r.error) setFormats(r.formats); })
      .catch(e => setError(String(e)));
  }, [dDate, dTz]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Show a date in every common format — ISO 8601, RFC, Unix timestamps, locale-friendly, and more.
      </p>
      <div className="flex flex-col sm:flex-row gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">Date or datetime</label>
          <input
            className={inputClass}
            placeholder="2024-01-15  or  now  or  1705276800"
            value={dateStr}
            onChange={e => setDateStr(e.target.value)}
          />
        </div>
        <div className="sm:w-56">
          <label className="block text-xs text-slate-500 mb-1">Timezone</label>
          <input
            className={inputClass}
            list="tz-list"
            placeholder="UTC"
            value={timezone}
            onChange={e => setTimezone(e.target.value)}
          />
          <datalist id="tz-list">
            {COMMON_TIMEZONES.map(tz => <option key={tz} value={tz} />)}
          </datalist>
        </div>
        <button
          onClick={() => setDateStr('now')}
          className="text-xs text-kawa-600 dark:text-kawa-400 hover:underline pb-2"
        >
          Use now
        </button>
      </div>

      <ErrorMsg msg={error} />

      {dDate && formats.length > 0 && !error && (
        <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-slate-100 dark:bg-neutral-800">
              <tr>
                <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Format</th>
                <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Result</th>
                <th className="w-8 px-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
              {formats.map((f, i) => (
                <tr key={i} className="bg-white dark:bg-neutral-900 hover:bg-slate-50 dark:hover:bg-neutral-800 transition-colors">
                  <td className="px-3 py-2 text-slate-500 dark:text-slate-400 whitespace-nowrap">{f.label}</td>
                  <td className="px-3 py-2 font-mono text-slate-800 dark:text-slate-200 break-all">{f.result}</td>
                  <td className="px-2 py-2">
                    <CopyBtn text={f.result} id={`fmt-${i}`} copied={copied} copy={copy} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// ─── Date Info ────────────────────────────────────────────────────────────────

function DateInfoTab() {
  const [date, setDate] = useState(todayISO());
  const [info, setInfo] = useState<{
    weekday: string; isWeekend: boolean; dayOfYear: number; daysInYear: number;
    daysLeftYear: number; isoWeek: number; isoYear: number; quarter: string;
    daysInMonth: number; daysLeftMonth: number; unixSec: number;
    zodiac: string; season: string; daysSinceEpoch: number;
  } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(date);

  useEffect(() => {
    if (!debounced) return;
    client.dateInfo({ date: debounced } as Parameters<typeof client.dateInfo>[0])
      .then(r => { setError(r.error); if (!r.error) setInfo(r); })
      .catch(e => setError(String(e)));
  }, [debounced]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Full calendar metadata for any date — week number, day of year, quarter, zodiac, season, and more.
      </p>
      <div className="flex gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">Date</label>
          <input type="date" className={inputClass} value={date} onChange={e => setDate(e.target.value)} />
        </div>
        <button
          onClick={() => setDate(todayISO())}
          className="flex items-center gap-1 text-xs text-kawa-600 dark:text-kawa-400 hover:underline pb-2"
        >
          <RotateCcw className="w-3 h-3" /> Today
        </button>
      </div>

      <ErrorMsg msg={error} />

      {debounced && info && !error && (
        <div className="space-y-4">
          {/* Weekday highlight */}
          <div className={cn(
            'rounded-xl border p-4 flex items-center justify-between',
            info.isWeekend
              ? 'bg-sky-50 dark:bg-sky-900/20 border-sky-200 dark:border-sky-800'
              : 'bg-slate-50 dark:bg-neutral-800 border-slate-200 dark:border-neutral-700'
          )}>
            <div>
              <div className="text-xl font-bold text-slate-800 dark:text-white">{info.weekday}</div>
              <div className="text-sm text-slate-500">
                {info.isWeekend ? '🏖 Weekend' : '💼 Weekday'} · {info.season} · {info.zodiac}
              </div>
            </div>
            <div className={cn(
              'text-right text-sm font-medium',
              info.isWeekend ? 'text-sky-600 dark:text-sky-400' : 'text-slate-500'
            )}>
              {info.quarter}
            </div>
          </div>

          {/* Grid of info cards */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <InfoCard label="ISO week"       value={`W${info.isoWeek} (${info.isoYear})`} id="i-week"  copied={copied} copy={copy} />
            <InfoCard label="Day of year"    value={info.dayOfYear}    id="i-doy"    copied={copied} copy={copy}
              sub={`${info.daysLeftYear} days left in year`} />
            <InfoCard label="Days in year"   value={info.daysInYear}   id="i-diy"    copied={copied} copy={copy}
              sub={info.daysInYear === 366 ? 'Leap year' : 'Common year'} />
            <InfoCard label="Days in month"  value={info.daysInMonth}  id="i-dim"    copied={copied} copy={copy}
              sub={`${info.daysLeftMonth} days left`} />
            <InfoCard label="Unix (seconds)" value={info.unixSec}      id="i-unix"   copied={copied} copy={copy} />
            <InfoCard label="Days since epoch" value={info.daysSinceEpoch} id="i-epoch" copied={copied} copy={copy}
              sub="since 1970-01-01" />
          </div>

          {/* Progress bars */}
          <div className="space-y-2">
            <ProgressBar
              label="Year progress"
              value={info.dayOfYear}
              total={info.daysInYear}
            />
            <ProgressBar
              label="Month progress"
              value={info.daysInMonth - info.daysLeftMonth}
              total={info.daysInMonth}
            />
          </div>
        </div>
      )}
    </div>
  );
}

function ProgressBar({ label, value, total }: { label: string; value: number; total: number }) {
  const pct = total > 0 ? Math.round((value / total) * 100) : 0;
  return (
    <div>
      <div className="flex justify-between text-xs text-slate-500 mb-1">
        <span>{label}</span>
        <span>{pct}% ({value}/{total})</span>
      </div>
      <div className="h-2 bg-slate-100 dark:bg-neutral-700 rounded-full overflow-hidden">
        <div
          className="h-full bg-kawa-500 dark:bg-kawa-400 rounded-full transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

export function DateTimeTool() {
  const [activeTab, setActiveTab] = useState<TabId>('diff');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 dark:text-white flex items-center gap-2">
          <Calendar className="w-6 h-6 text-kawa-500" />
          Date &amp; Time Tools
        </h1>
        <p className="text-slate-500 dark:text-slate-400 mt-1">
          Date difference, leap year checker, date arithmetic, multi-format output, and full calendar info.
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
        {activeTab === 'diff'   && <DateDiffTab />}
        {activeTab === 'leap'   && <LeapYearTab />}
        {activeTab === 'add'    && <DateAddTab />}
        {activeTab === 'format' && <DateFormatTab />}
        {activeTab === 'info'   && <DateInfoTab />}
      </div>
    </div>
  );
}
