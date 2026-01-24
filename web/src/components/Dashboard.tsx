import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Search } from 'lucide-react';
import { navItems } from '../lib/nav';

export function Dashboard() {
  const [term, setTerm] = useState('');

  const filtered = navItems.filter(item => 
    item.path !== '/' && ( // Exclude dashboard itself
      item.name.toLowerCase().includes(term.toLowerCase()) || 
      item.description.toLowerCase().includes(term.toLowerCase())
    )
  );

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 max-w-2xl mx-auto text-center py-12">
        <h1 className="text-4xl font-bold bg-gradient-to-r from-kawa-400 to-kawa-600 bg-clip-text text-transparent">
          PrivUtil
        </h1>
        <p className="text-slate-500 dark:text-slate-400 text-lg">
          Offline-capable developer utility suite. Privacy-first, no server tracking.
        </p>
        
        <div className="relative max-w-lg mx-auto w-full mt-6">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 w-5 h-5" />
          <input 
            type="text" 
            value={term}
            onChange={e => setTerm(e.target.value)}
            placeholder="Search tools..."
            className="w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-full py-3 pl-12 pr-6 text-slate-900 dark:text-white focus:ring-2 focus:ring-kawa-500 focus:outline-none focus:border-transparent transition-all shadow-lg placeholder:text-slate-400"
            autoFocus
          />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-12">
        {filtered.map((item) => {
          const Icon = item.icon;
          return (
            <Link 
              key={item.path} 
              to={item.path}
              className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700/50 p-6 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 hover:border-kawa-500/50 transition-all group hover:-translate-y-1 hover:shadow-xl hover:shadow-kawa-500/10"
            >
              <div className="flex items-center gap-4 mb-3">
                <div className="p-2 bg-slate-100 dark:bg-slate-700/50 rounded-lg text-kawa-600 dark:text-kawa-400 group-hover:text-kawa-500 group-hover:bg-kawa-500/10 transition-colors">
                  <Icon className="w-6 h-6" />
                </div>
                <h3 className="font-semibold text-lg text-slate-800 dark:text-slate-200 group-hover:text-kawa-600 dark:group-hover:text-kawa-400">
                  {item.name}
                </h3>
              </div>
              <p className="text-sm text-slate-500 dark:text-slate-400 leading-relaxed">
                {item.description}
              </p>
            </Link>
          );
        })}
      </div>

      {filtered.length === 0 && (
        <div className="text-center text-slate-500 dark:text-slate-400 py-12">
          No tools found matching "{term}"
        </div>
      )}
    </div>
  );
}
