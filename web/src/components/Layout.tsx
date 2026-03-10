import { Link, Outlet, useLocation, useNavigate, useSearchParams } from 'react-router-dom';
import { Search } from 'lucide-react';
import { cn } from '../lib/utils';
import { navItems } from '../lib/nav';
import { ThemeToggle } from './ThemeToggle';

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const term = searchParams.get('q') || '';

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    if (location.pathname !== '/') {
      navigate(`/?q=${encodeURIComponent(val)}`);
    } else {
      if (val) {
        setSearchParams({ q: val });
      } else {
        setSearchParams({});
      }
    }
  };

  return (
    <div className="flex min-h-screen bg-gray-100 dark:bg-neutral-900 text-neutral-900 dark:text-neutral-100 transition-colors">
      {/* Sidebar */}
      <aside className="w-64 border-r border-gray-200 dark:border-neutral-800 bg-white/50 dark:bg-neutral-900/50 backdrop-blur">
        <div className="flex h-16 items-center px-6 border-b border-slate-200 dark:border-neutral-800">
          <span className="text-xl font-bold bg-gradient-to-r from-kawa-500 to-kawa-400 bg-clip-text text-transparent">
            PrivUtil
          </span>
        </div>
        
        <nav className="p-4 space-y-2 overflow-y-auto max-h-[calc(100vh-4rem)]">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;
            
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  "flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-colors",
                  isActive 
                    ? "bg-kawa-500/10 text-kawa-600 dark:text-kawa-400" 
                    : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-900 dark:hover:text-slate-100"
                )}
              >
                <Icon className="h-5 w-5" />
                {item.name}
              </Link>
            );
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        <header className="h-16 border-b border-slate-200 dark:border-neutral-800 bg-white/50 dark:bg-slate-900/50 backdrop-blur sticky top-0 z-10 flex items-center justify-between px-6">
          <div className="flex-1 max-w-xl relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 w-4 h-4" />
            <input 
              type="text" 
              value={term}
              onChange={handleSearch}
              placeholder="Search tools..."
              className="w-full bg-slate-100 dark:bg-slate-800/50 border border-transparent focus:border-kawa-500/50 rounded-lg py-2 pl-10 pr-4 text-sm text-slate-900 dark:text-white focus:ring-1 focus:ring-kawa-500 focus:outline-none transition-all placeholder:text-slate-400"
            />
          </div>
          <div className="flex items-center gap-4 ml-4">
            <ThemeToggle />
          </div>
        </header>
        <div className="p-8 w-full mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
