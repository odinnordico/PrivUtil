import { useState } from 'react';
import { Key, Copy, RefreshCw, Check } from 'lucide-react';
import { PasswordRequest } from '../proto/proto/privutil';
import { client } from '../lib/client';

export function PasswordTool() {
  const [length, setLength] = useState(16);
  const [count, setCount] = useState(1);
  const [uppercase, setUppercase] = useState(true);
  const [lowercase, setLowercase] = useState(true);
  const [numbers, setNumbers] = useState(true);
  const [symbols, setSymbols] = useState(true);
  const [customChars, setCustomChars] = useState('');
  const [passwords, setPasswords] = useState<string[]>([]);
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);

  const generate = async () => {
    setLoading(true);
    try {
      const resp = await client.generatePassword(
        PasswordRequest.create({
          length,
          count,
          uppercase,
          lowercase,
          numbers,
          symbols,
          customChars,
        }) as any
      );
      setPasswords(resp.passwords);
    } catch {
      setPasswords(['Error generating password']);
    }
    setLoading(false);
  };

  const copyToClipboard = async (text: string, index: number) => {
    await navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold flex items-center gap-2 text-slate-800 dark:text-slate-100">
        <Key className="w-6 h-6 text-kawa-500" />
        Password Generator
      </h2>
      
      <p className="text-slate-500 dark:text-slate-400">
        Generate secure random passwords with customizable character sets.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              Length: {length}
            </label>
            <input
              type="range"
              min="4"
              max="64"
              value={length}
              onChange={e => setLength(parseInt(e.target.value))}
              className="w-full accent-kawa-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              Count: {count}
            </label>
            <input
              type="range"
              min="1"
              max="10"
              value={count}
              onChange={e => setCount(parseInt(e.target.value))}
              className="w-full accent-kawa-500"
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
              Character Sets
            </label>
            <div className="flex flex-wrap gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={uppercase}
                  onChange={e => setUppercase(e.target.checked)}
                  className="accent-kawa-500"
                />
                <span className="text-sm text-slate-600 dark:text-slate-400">A-Z</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={lowercase}
                  onChange={e => setLowercase(e.target.checked)}
                  className="accent-kawa-500"
                />
                <span className="text-sm text-slate-600 dark:text-slate-400">a-z</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={numbers}
                  onChange={e => setNumbers(e.target.checked)}
                  className="accent-kawa-500"
                />
                <span className="text-sm text-slate-600 dark:text-slate-400">0-9</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={symbols}
                  onChange={e => setSymbols(e.target.checked)}
                  className="accent-kawa-500"
                />
                <span className="text-sm text-slate-600 dark:text-slate-400">!@#$%</span>
              </label>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              Custom Characters (optional)
            </label>
            <input
              type="text"
              value={customChars}
              onChange={e => setCustomChars(e.target.value)}
              placeholder="Leave empty to use selected sets"
              className="w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg px-4 py-2 text-slate-900 dark:text-white"
            />
          </div>

          <button
            onClick={generate}
            disabled={loading}
            className="flex items-center gap-2 px-6 py-3 bg-kawa-500 hover:bg-kawa-600 text-black font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Generate
          </button>
        </div>

        <div className="space-y-3">
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
            Generated Passwords
          </label>
          {passwords.length === 0 ? (
            <div className="bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-4 text-slate-500 dark:text-slate-400 text-center">
              Click "Generate" to create passwords
            </div>
          ) : (
            <div className="space-y-2">
              {passwords.map((pw, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-3"
                >
                  <code className="text-sm font-mono text-slate-800 dark:text-slate-200 break-all">
                    {pw}
                  </code>
                  <button
                    onClick={() => copyToClipboard(pw, idx)}
                    className="ml-3 p-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-colors"
                    title="Copy to clipboard"
                  >
                    {copiedIndex === idx ? (
                      <Check className="w-4 h-4 text-kawa-500" />
                    ) : (
                      <Copy className="w-4 h-4 text-slate-500" />
                    )}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
