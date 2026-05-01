import { useState, useCallback, useEffect } from 'react';
import { client } from '../lib/client';
import { Copy, Check, Plus, Trash2, Calculator, Percent, Thermometer, Ruler } from 'lucide-react';
import { cn } from '../lib/utils';
import { PercentMode, UnitCategory } from '../proto/proto/privutil';

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

function useDebounce<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

const tabs = [
  { id: 'math',    label: 'Math Evaluator',  icon: Calculator  },
  { id: 'percent', label: '% Calculator',    icon: Percent     },
  { id: 'temp',    label: 'Temperature',     icon: Thermometer },
  { id: 'unit',    label: 'Unit Converter',  icon: Ruler       },
] as const;
type TabId = typeof tabs[number]['id'];

const inputClass = "bg-white dark:bg-neutral-800 text-slate-900 dark:text-white rounded-lg px-3 py-2 border border-slate-300 dark:border-neutral-700 focus:ring-2 focus:ring-kawa-500 text-sm w-full";
const selectClass = "bg-slate-50 dark:bg-gray-700 text-slate-900 dark:text-white rounded px-3 py-2 border border-slate-300 dark:border-transparent focus:ring-2 focus:ring-kawa-500 text-sm";
const numInputClass = cn(inputClass, 'font-mono');

// ─── Math evaluator ───────────────────────────────────────────────────────────

const FUNCTION_REFERENCE = [
  { group: 'Arithmetic', items: ['sqrt(x)', 'cbrt(x)', 'abs(x)', 'pow(x,y)', 'mod(x,y)', 'factorial(x)'] },
  { group: 'Rounding',   items: ['floor(x)', 'ceil(x)', 'round(x)', 'trunc(x)'] },
  { group: 'Logarithms', items: ['log(x)', 'log(x,base)', 'log2(x)', 'log10(x)', 'exp(x)', 'exp2(x)'] },
  { group: 'Trig',       items: ['sin(x)', 'cos(x)', 'tan(x)', 'asin(x)', 'acos(x)', 'atan(x)', 'atan2(y,x)', 'sinh(x)', 'cosh(x)', 'tanh(x)'] },
  { group: 'Misc',       items: ['min(a,b,...)', 'max(a,b,...)', 'hypot(x,y)', 'gcd(a,b)', 'lcm(a,b)', 'clamp(v,lo,hi)', 'lerp(a,b,t)', 'sign(x)'] },
  { group: 'Constants',  items: ['pi', 'e', 'phi', 'tau', 'inf'] },
];

type Variable = { name: string; value: string };

function MathTab() {
  const [expr, setExpr] = useState('');
  const [variables, setVariables] = useState<Variable[]>([]);
  const [precision, setPrecision] = useState(10);
  const [degrees, setDegrees] = useState(false);
  const [result, setResult] = useState('');
  const [rawValue, setRawValue] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [showRef, setShowRef] = useState(false);
  const { copied, copy } = useCopy();
  const debounced = useDebounce(expr);
  const debouncedVars = useDebounce(variables);

  useEffect(() => {
    const trimmed = debounced.trim();
    if (!trimmed) { setResult(''); setError(''); setRawValue(null); return; }

    const validVars = debouncedVars.filter(v => v.name.trim() && v.value.trim() !== '');
    const protoVars = validVars.map(v => ({ name: v.name.trim(), value: parseFloat(v.value) }))
      .filter(v => !isNaN(v.value));

    client.mathEval({
      expression: trimmed,
      variables: protoVars,
      precision,
      degrees,
    } as Parameters<typeof client.mathEval>[0])
      .then(r => { setResult(r.result); setRawValue(r.rawValue); setError(r.error); })
      .catch(e => setError(String(e)));
  }, [debounced, debouncedVars, precision, degrees]);

  const addVariable = () => setVariables(v => [...v, { name: '', value: '' }]);
  const removeVariable = (i: number) => setVariables(v => v.filter((_, j) => j !== i));
  const updateVariable = (i: number, field: keyof Variable, val: string) =>
    setVariables(v => v.map((item, j) => j === i ? { ...item, [field]: val } : item));

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Evaluate math expressions with variables, constants (<code className="text-kawa-600 dark:text-kawa-400">pi</code>, <code className="text-kawa-600 dark:text-kawa-400">e</code>, <code className="text-kawa-600 dark:text-kawa-400">phi</code>), and 30+ functions.
      </p>

      {/* Expression input */}
      <div>
        <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1">Expression</label>
        <input
          className={cn(inputClass, 'font-mono text-base')}
          placeholder="pi * r^2  or  sin(45) + cos(45)"
          value={expr}
          onChange={e => setExpr(e.target.value)}
        />
      </div>

      {/* Options row */}
      <div className="flex flex-wrap gap-4 items-center">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Decimal places</label>
          <input
            type="number" min={1} max={15}
            className={cn(selectClass, 'w-24')}
            value={precision}
            onChange={e => setPrecision(Math.min(15, Math.max(1, Number(e.target.value))))}
          />
        </div>
        <label className="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer mt-4">
          <input type="checkbox" checked={degrees} onChange={e => setDegrees(e.target.checked)} className="accent-kawa-500" />
          Degrees mode (trig functions use °)
        </label>
      </div>

      {/* Variables */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <label className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Variables</label>
          <button
            onClick={addVariable}
            className="flex items-center gap-1 text-xs text-kawa-600 dark:text-kawa-400 hover:underline"
          >
            <Plus className="w-3 h-3" /> Add variable
          </button>
        </div>
        {variables.length === 0 && (
          <p className="text-xs text-slate-400 dark:text-slate-500 italic">No variables — type names directly in the expression (e.g. <code>r</code>, <code>x</code>)</p>
        )}
        <div className="space-y-2">
          {variables.map((v, i) => (
            <div key={i} className="flex items-center gap-2">
              <input
                className={cn(inputClass, 'font-mono w-28')}
                placeholder="name"
                value={v.name}
                onChange={e => updateVariable(i, 'name', e.target.value)}
              />
              <span className="text-slate-400">=</span>
              <input
                type="number"
                className={cn(numInputClass, 'flex-1')}
                placeholder="0"
                value={v.value}
                onChange={e => updateVariable(i, 'value', e.target.value)}
              />
              <button onClick={() => removeVariable(i)} className="text-slate-400 hover:text-red-500 dark:hover:text-red-400 transition-colors">
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Result */}
      {error && <ErrorMsg msg={error} />}
      {result && !error && (
        <div className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-4">
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs font-bold text-slate-500 uppercase tracking-wide">Result</span>
            <CopyBtn text={result} id="math-result" copied={copied} copy={copy} />
          </div>
          <div className="font-mono text-2xl font-semibold text-kawa-600 dark:text-kawa-400 break-all">
            {result}
          </div>
          {rawValue !== null && Math.abs(rawValue) >= 1e6 && (
            <p className="text-xs text-slate-400 mt-1 font-mono">
              {rawValue.toExponential(6)}
            </p>
          )}
        </div>
      )}

      {/* Function reference */}
      <button
        onClick={() => setShowRef(r => !r)}
        className="text-xs text-kawa-600 dark:text-kawa-400 hover:underline"
      >
        {showRef ? '▲ Hide' : '▼ Show'} function reference
      </button>
      {showRef && (
        <div className="rounded-lg border border-slate-200 dark:border-neutral-700 p-3 space-y-2">
          {FUNCTION_REFERENCE.map(group => (
            <div key={group.group}>
              <span className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wide">{group.group}</span>
              <div className="flex flex-wrap gap-1 mt-1">
                {group.items.map(fn => (
                  <code
                    key={fn}
                    className="text-xs bg-slate-100 dark:bg-neutral-800 text-kawa-700 dark:text-kawa-300 px-1.5 py-0.5 rounded cursor-pointer hover:bg-kawa-100 dark:hover:bg-kawa-900/20"
                    onClick={() => setExpr(e => e + (e.endsWith('(') || e === '' ? '' : ' ') + fn)}
                  >
                    {fn}
                  </code>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Percentage calculator ────────────────────────────────────────────────────

const PCT_MODES = [
  { value: PercentMode.PCT_X_OF_Y,  label: 'What is A% of B?',        placeholder: ['Percent (A)', 'Value (B)']  },
  { value: PercentMode.PCT_WHAT,    label: 'A is what % of B?',        placeholder: ['Value (A)',   'Total (B)']  },
  { value: PercentMode.PCT_CHANGE,  label: '% change from A to B',     placeholder: ['From (A)',    'To (B)']     },
  { value: PercentMode.PCT_REVERSE, label: 'A is B% of what?',         placeholder: ['Value (A)',   'Percent (B)'] },
] as const;

function PercentTab() {
  const [mode, setMode] = useState<PercentMode>(PercentMode.PCT_X_OF_Y);
  const [a, setA] = useState('');
  const [b, setB] = useState('');
  const [result, setResult] = useState<{ result: number; formatted: string; formula: string } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const dA = useDebounce(a);
  const dB = useDebounce(b);

  useEffect(() => {
    const av = parseFloat(dA);
    const bv = parseFloat(dB);
    if (dA === '' || dB === '' || isNaN(av) || isNaN(bv)) {
      setResult(null); setError(''); return;
    }
    client.percentageCalc({ mode, a: av, b: bv } as Parameters<typeof client.percentageCalc>[0])
      .then(r => {
        setError(r.error);
        if (!r.error) setResult({ result: r.result, formatted: r.formatted, formula: r.formula });
      })
      .catch(e => setError(String(e)));
  }, [mode, dA, dB]);

  const current = PCT_MODES.find(m => m.value === mode)!;

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Four percentage calculation modes — enter values and get the result instantly.
      </p>

      {/* Mode picker */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {PCT_MODES.map(m => (
          <button
            key={m.value}
            onClick={() => setMode(m.value)}
            className={cn(
              'text-left px-4 py-3 rounded-lg border text-sm transition-colors',
              mode === m.value
                ? 'border-kawa-500 bg-kawa-50 dark:bg-kawa-900/20 text-kawa-700 dark:text-kawa-300 font-medium'
                : 'border-slate-200 dark:border-neutral-700 text-slate-600 dark:text-slate-400 hover:border-kawa-300 dark:hover:border-kawa-700'
            )}
          >
            {m.label}
          </button>
        ))}
      </div>

      {/* Inputs */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div>
          <label className="block text-xs text-slate-500 mb-1">{current.placeholder[0]}</label>
          <input
            type="number"
            className={numInputClass}
            placeholder="0"
            value={a}
            onChange={e => setA(e.target.value)}
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">{current.placeholder[1]}</label>
          <input
            type="number"
            className={numInputClass}
            placeholder="0"
            value={b}
            onChange={e => setB(e.target.value)}
          />
        </div>
      </div>

      {error && <ErrorMsg msg={error} />}

      {result && !error && (
        <div className="bg-slate-50 dark:bg-neutral-800 rounded-xl border border-slate-200 dark:border-neutral-700 p-4 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-500 uppercase tracking-wide">Result</span>
            <CopyBtn text={String(result.result)} id="pct-result" copied={copied} copy={copy} />
          </div>
          <div className="font-mono text-3xl font-bold text-kawa-600 dark:text-kawa-400">
            {result.result % 1 === 0 ? result.result : result.result.toFixed(6).replace(/\.?0+$/, '')}
            {(mode === PercentMode.PCT_WHAT || mode === PercentMode.PCT_CHANGE) && (
              <span className="text-xl ml-1 font-semibold">%</span>
            )}
          </div>
          <p className="text-sm text-slate-600 dark:text-slate-300">{result.formatted}</p>
          <p className="text-xs text-slate-400 dark:text-slate-500 font-mono">Formula: {result.formula}</p>
        </div>
      )}
    </div>
  );
}

// ─── Temperature converter ────────────────────────────────────────────────────

const TEMP_UNITS = [
  { value: 'c', label: 'Celsius (°C)',    symbol: '°C' },
  { value: 'f', label: 'Fahrenheit (°F)', symbol: '°F' },
  { value: 'k', label: 'Kelvin (K)',      symbol: 'K'  },
] as const;

function TempCard({ label, symbol, value, id, copied, copy }: {
  label: string; symbol: string; value: number | null; id: string;
  copied: string | null; copy: (t: string, k: string) => void;
}) {
  const fmt = value === null ? '—'
    : (Math.abs(value) < 1e6 && Math.abs(value) >= 1e-4 || value === 0)
      ? value.toFixed(4).replace(/\.?0+$/, '')
      : value.toExponential(4);
  return (
    <div className="flex-1 bg-slate-50 dark:bg-neutral-800 rounded-lg border border-slate-200 dark:border-neutral-700 p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-bold text-slate-500 uppercase tracking-wide">{label}</span>
        <CopyBtn text={fmt} id={id} copied={copied} copy={copy} />
      </div>
      <div className="flex items-end gap-1">
        <span className="font-mono text-2xl font-bold text-slate-800 dark:text-white">{fmt}</span>
        <span className="text-sm text-slate-500 dark:text-slate-400 mb-0.5">{symbol}</span>
      </div>
    </div>
  );
}

function TempTab() {
  const [value, setValue] = useState('');
  const [fromUnit, setFromUnit] = useState<'c' | 'f' | 'k'>('c');
  const [temps, setTemps] = useState<{ celsius: number; fahrenheit: number; kelvin: number } | null>(null);
  const [error, setError] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(value);

  useEffect(() => {
    const v = parseFloat(debounced);
    if (debounced === '' || isNaN(v)) { setTemps(null); setError(''); return; }
    client.tempConvert({ value: v, fromUnit } as Parameters<typeof client.tempConvert>[0])
      .then(r => {
        setError(r.error);
        if (!r.error) setTemps({ celsius: r.celsius, fahrenheit: r.fahrenheit, kelvin: r.kelvin });
      })
      .catch(e => setError(String(e)));
  }, [debounced, fromUnit]);

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Convert between Celsius, Fahrenheit, and Kelvin instantly.
      </p>

      <div className="flex gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">Value</label>
          <input
            type="number"
            className={numInputClass}
            placeholder="0"
            value={value}
            onChange={e => setValue(e.target.value)}
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Unit</label>
          <select className={selectClass} value={fromUnit} onChange={e => setFromUnit(e.target.value as 'c' | 'f' | 'k')}>
            {TEMP_UNITS.map(u => (
              <option key={u.value} value={u.value}>{u.label}</option>
            ))}
          </select>
        </div>
      </div>

      {error && <ErrorMsg msg={error} />}

      {temps && !error && (
        <div className="flex flex-col sm:flex-row gap-3">
          <TempCard label="Celsius" symbol="°C" value={temps.celsius} id="temp-c" copied={copied} copy={copy} />
          <TempCard label="Fahrenheit" symbol="°F" value={temps.fahrenheit} id="temp-f" copied={copied} copy={copy} />
          <TempCard label="Kelvin" symbol="K" value={temps.kelvin} id="temp-k" copied={copied} copy={copy} />
        </div>
      )}

      {/* Reference card */}
      <div className="rounded-lg border border-slate-200 dark:border-neutral-700 p-3 text-xs space-y-1">
        <p className="font-semibold text-slate-500 uppercase tracking-wide mb-2">Formulas</p>
        <p className="font-mono text-slate-600 dark:text-slate-300">°F = (°C × 9/5) + 32</p>
        <p className="font-mono text-slate-600 dark:text-slate-300">°C = (°F − 32) × 5/9</p>
        <p className="font-mono text-slate-600 dark:text-slate-300">K = °C + 273.15</p>
      </div>
    </div>
  );
}

// ─── Unit converter ───────────────────────────────────────────────────────────

type CategoryConfig = {
  value: UnitCategory;
  label: string;
  defaultUnit: string;
  units: string[];
};

const CATEGORIES: CategoryConfig[] = [
  {
    value: UnitCategory.UNIT_BYTES,
    label: 'Storage (Bytes)',
    defaultUnit: 'MB',
    units: ['B','KB','MB','GB','TB','PB','EB','KiB','MiB','GiB','TiB','PiB','EiB'],
  },
  {
    value: UnitCategory.UNIT_LENGTH,
    label: 'Length',
    defaultUnit: 'm',
    units: ['nm','µm','mm','cm','dm','m','km','in','ft','yd','mi','nmi','ly'],
  },
  {
    value: UnitCategory.UNIT_MASS,
    label: 'Mass / Weight',
    defaultUnit: 'kg',
    units: ['µg','mg','g','kg','t','oz','lb','st','ton','lt'],
  },
  {
    value: UnitCategory.UNIT_AREA,
    label: 'Area',
    defaultUnit: 'm²',
    units: ['mm²','cm²','m²','km²','ha','in²','ft²','yd²','ac','mi²'],
  },
  {
    value: UnitCategory.UNIT_VOLUME,
    label: 'Volume',
    defaultUnit: 'l',
    units: ['ml','cl','dl','l','m³','in³','ft³','tsp','tbsp','fl oz','cup','pt','qt','gal','imp gal'],
  },
  {
    value: UnitCategory.UNIT_SPEED,
    label: 'Speed',
    defaultUnit: 'km/h',
    units: ['m/s','km/h','mph','ft/s','kn','mach','c'],
  },
];

function UnitTab() {
  const [category, setCategory] = useState<UnitCategory>(UnitCategory.UNIT_LENGTH);
  const [value, setValue] = useState('');
  const [fromUnit, setFromUnit] = useState('m');
  const [results, setResults] = useState<{ unit: string; label: string; value: number; formatted: string }[]>([]);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const { copied, copy } = useCopy();
  const debounced = useDebounce(value);

  const catConfig = CATEGORIES.find(c => c.value === category)!;

  // Reset fromUnit when category changes
  const handleCategoryChange = (cat: UnitCategory) => {
    setCategory(cat);
    const cfg = CATEGORIES.find(c => c.value === cat)!;
    setFromUnit(cfg.defaultUnit);
    setResults([]);
    setError('');
  };

  useEffect(() => {
    const v = parseFloat(debounced);
    if (debounced === '' || isNaN(v)) { setResults([]); setError(''); return; }
    client.unitConvert({
      value: v,
      fromUnit,
      category,
    } as Parameters<typeof client.unitConvert>[0])
      .then(r => {
        setError(r.error);
        if (!r.error) setResults(r.results);
      })
      .catch(e => setError(String(e)));
  }, [debounced, fromUnit, category]);

  const filteredResults = results.filter(r =>
    !search || r.unit.toLowerCase().includes(search.toLowerCase()) ||
    r.label.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">
        Convert a value across all units in a category — bytes, length, mass, area, volume, and speed.
      </p>

      {/* Category tabs */}
      <div className="flex flex-wrap gap-1">
        {CATEGORIES.map(c => (
          <button
            key={c.value}
            onClick={() => handleCategoryChange(c.value)}
            className={cn(
              'px-3 py-1.5 rounded-full text-xs font-medium border transition-colors',
              category === c.value
                ? 'bg-kawa-600 text-white border-kawa-600'
                : 'bg-white dark:bg-neutral-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-neutral-600 hover:border-kawa-400'
            )}
          >
            {c.label}
          </button>
        ))}
      </div>

      {/* Input row */}
      <div className="flex gap-3 items-end">
        <div className="flex-1">
          <label className="block text-xs text-slate-500 mb-1">Value</label>
          <input
            type="number"
            className={numInputClass}
            placeholder="1"
            value={value}
            onChange={e => setValue(e.target.value)}
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">From unit</label>
          <select
            className={selectClass}
            value={fromUnit}
            onChange={e => setFromUnit(e.target.value)}
          >
            {catConfig.units.map(u => (
              <option key={u} value={u}>{u}</option>
            ))}
          </select>
        </div>
      </div>

      {error && <ErrorMsg msg={error} />}

      {results.length > 0 && !error && (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <input
              className={cn(inputClass, 'max-w-xs')}
              placeholder="Filter units…"
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
            <span className="text-xs text-slate-400">{filteredResults.length} unit{filteredResults.length !== 1 ? 's' : ''}</span>
          </div>
          <div className="rounded-lg border border-slate-200 dark:border-neutral-700 overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-slate-100 dark:bg-neutral-800">
                <tr>
                  <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Unit</th>
                  <th className="text-left px-3 py-2 text-xs font-semibold text-slate-500 uppercase hidden sm:table-cell">Name</th>
                  <th className="text-right px-3 py-2 text-xs font-semibold text-slate-500 uppercase">Value</th>
                  <th className="w-8 px-2 py-2"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-neutral-700">
                {filteredResults.map(r => (
                  <tr
                    key={r.unit}
                    className={cn(
                      'transition-colors',
                      r.unit === fromUnit
                        ? 'bg-kawa-50 dark:bg-kawa-900/20'
                        : 'bg-white dark:bg-neutral-900 hover:bg-slate-50 dark:hover:bg-neutral-800'
                    )}
                  >
                    <td className="px-3 py-2 font-mono font-semibold text-kawa-700 dark:text-kawa-300">{r.unit}</td>
                    <td className="px-3 py-2 text-slate-500 dark:text-slate-400 hidden sm:table-cell">{r.label}</td>
                    <td className="px-3 py-2 text-right font-mono text-slate-800 dark:text-slate-200">{r.formatted}</td>
                    <td className="px-2 py-2">
                      <CopyBtn text={r.formatted} id={`unit-${r.unit}`} copied={copied} copy={copy} />
                    </td>
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

// ─── Main component ───────────────────────────────────────────────────────────

export function MathUnitTool() {
  const [activeTab, setActiveTab] = useState<TabId>('math');

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 dark:text-white flex items-center gap-2">
          <Calculator className="w-6 h-6 text-kawa-500" />
          Math &amp; Unit Tools
        </h1>
        <p className="text-slate-500 dark:text-slate-400 mt-1">
          Expression evaluator, percentage calculator, temperature converter, and unit converter with 6 categories.
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
        {activeTab === 'math'    && <MathTab />}
        {activeTab === 'percent' && <PercentTab />}
        {activeTab === 'temp'    && <TempTab />}
        {activeTab === 'unit'    && <UnitTab />}
      </div>
    </div>
  );
}
