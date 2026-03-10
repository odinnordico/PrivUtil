import { useState } from 'react';
import { client } from '../lib/client';
import { FileCode, ArrowRightLeft } from 'lucide-react';

type Mode = 'md2html' | 'html2md';

export function MarkdownTool() {
  const [mode, setMode] = useState<Mode>('md2html');
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');

  const handleConvert = async () => {
    try {
      const req = { text: input } as Parameters<typeof client.markdownToHtml>[0];
      const resp = mode === 'md2html'
        ? await client.markdownToHtml(req)
        : await client.htmlToMarkdown(req);
      setOutput(resp.text);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
        <FileCode className="w-6 h-6 text-kawa-500" /> Markdown &harr; HTML
      </h2>

      <div className="flex gap-4 border-b border-slate-300 dark:border-gray-700">
        <button
          onClick={() => { setMode('md2html'); setOutput(''); }}
          className={`pb-2 px-4 font-bold transition-colors ${mode === 'md2html' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          Markdown → HTML
        </button>
        <button
          onClick={() => { setMode('html2md'); setOutput(''); }}
          className={`pb-2 px-4 font-bold transition-colors ${mode === 'html2md' ? 'text-kawa-600 border-b-2 border-kawa-500' : 'text-slate-500 hover:text-slate-700 dark:text-gray-400 dark:hover:text-gray-200'}`}
        >
          HTML → Markdown
        </button>
      </div>

      <div className="p-6 bg-white dark:bg-neutral-800 rounded-lg border border-slate-300 dark:border-neutral-700 shadow-sm space-y-4">
        <div className="flex justify-between items-center">
          <h3 className="text-lg font-bold text-slate-800 dark:text-gray-200 flex items-center gap-2">
            <ArrowRightLeft className="w-5 h-5" />
            {mode === 'md2html' ? 'Markdown to HTML' : 'HTML to Markdown'}
          </h3>
          <button
            onClick={handleConvert}
            className="bg-kawa-500 hover:bg-kawa-600 text-slate-900 px-4 py-1.5 rounded text-sm font-medium transition-colors"
          >
            Convert
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">
              {mode === 'md2html' ? 'Markdown' : 'HTML'} Input
            </label>
            <textarea
              value={input}
              onChange={e => setInput(e.target.value)}
              placeholder={mode === 'md2html' ? '# Hello World\n\nSome **bold** text.' : '<h1>Hello World</h1>\n<p>Some <strong>bold</strong> text.</p>'}
              className="w-full h-80 bg-white dark:bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 focus:border-kawa-500 text-slate-900 dark:text-neutral-100 shadow-inner"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase">
              {mode === 'md2html' ? 'HTML' : 'Markdown'} Output
            </label>
            <textarea
              readOnly
              value={output}
              placeholder="Result will appear here..."
              className="w-full h-80 bg-slate-50 dark:bg-black/30 p-4 rounded font-mono text-sm resize-none focus:outline-none border border-slate-300 dark:border-neutral-700 text-slate-900 dark:text-neutral-100"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
