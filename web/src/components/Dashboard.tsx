import { Link, useSearchParams } from 'react-router-dom';
import { navItems } from '../lib/nav';
import mascot from '../assets/mascot.webp';

export function Dashboard() {
  const [searchParams] = useSearchParams();
  const term = searchParams.get('q') || '';

  const filtered = navItems.filter(item => 
    item.path !== '/' && ( // Exclude dashboard itself
      item.name.toLowerCase().includes(term.toLowerCase()) || 
      item.description.toLowerCase().includes(term.toLowerCase())
    )
  );

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row items-center justify-center gap-6 max-w-3xl mx-auto py-12">
        <img
          src={mascot}
          alt="PrivUtil mascot"
          width={160}
          height={166}
          className="w-32 md:w-40 shrink-0"
        />
        <div className="flex flex-col gap-4 text-center md:text-left">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-kawa-400 to-kawa-600 bg-clip-text text-transparent">
            PrivUtil
          </h1>
          <p className="text-slate-500 dark:text-slate-400 text-lg">
            Offline-capable developer utility suite. Privacy-first, no server tracking.
          </p>
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
        <div className="flex flex-col items-center gap-4 text-center text-slate-500 dark:text-slate-400 py-12">
          {/* Decorative: the message right below carries the meaning. */}
          <img src={mascot} alt="" width={96} height={100} className="w-24 opacity-60" />
          <p>No tools found matching "{term}"</p>
        </div>
      )}
    </div>
  );
}
