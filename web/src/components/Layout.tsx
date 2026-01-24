import { Link, Outlet, useLocation } from 'react-router-dom';
import { cn } from '../lib/utils';
import { navItems } from '../lib/nav';
import { ThemeToggle } from './ThemeToggle';

export function Layout() {
  const location = useLocation();

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
        <header className="h-16 border-b border-slate-200 dark:border-neutral-800 bg-white/50 dark:bg-slate-900/50 backdrop-blur sticky top-0 z-10 flex items-center justify-end px-6">
          <ThemeToggle />
        </header>
        <div className="p-8 w-full mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
