'use client';

import './globals.css';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState } from 'react';
import { NotificationProvider } from './components/Notification';
import { ErrorBoundary } from './components/ErrorBoundary';

const navItems = [
  { group: 'Overview', items: [
    { href: '/dashboard', label: 'Dashboard', icon: 'rect' },
    { href: '/metrics', label: 'Metrics', icon: 'chart' },
    { href: '/incidents', label: 'Incidents', icon: 'chat' },
    { href: '/services', label: 'Services', icon: 'circle' },
  ]},
  { group: 'Intelligence', items: [
    { href: '/insights', label: 'AI Insights', icon: 'ai' },
    { href: '/topology', label: 'Topology', icon: 'grid' },
    { href: '/alerts', label: 'Alert Rules', icon: 'bell' },
  ]},
  { group: 'Settings', items: [
    { href: '/integrations', label: 'Integrations', icon: 'link' },
    { href: '/team', label: 'Team', icon: 'users' },
  ]},
];

function NavIcon({ type }: { type: string }) {
  const cls = "w-4 h-4";
  const s = { strokeWidth: 1.5, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const, fill: 'none', stroke: 'currentColor' };
  switch (type) {
    case 'rect': return <svg className={cls} {...s} viewBox="0 0 16 16"><rect x="2" y="2" width="5" height="5" rx="1"/><rect x="9" y="2" width="5" height="5" rx="1"/><rect x="2" y="9" width="5" height="5" rx="1"/><rect x="9" y="9" width="5" height="5" rx="1"/></svg>;
    case 'chart': return <svg className={cls} {...s} viewBox="0 0 16 16"><path d="M2 12h4l3-8 4 16 3-8h4"/></svg>;
    case 'chat': return <svg className={cls} {...s} viewBox="0 0 16 16"><path d="M14 2H2v12l3-3h9V2z"/><path d="M6 6h4M6 9h2"/></svg>;
    case 'circle': return <svg className={cls} {...s} viewBox="0 0 16 16"><circle cx="8" cy="8" r="6"/><path d="M8 5v3l2 2"/></svg>;
    case 'ai': return <svg className={cls} {...s} viewBox="0 0 16 16"><path d="M9 1L1 9l4 4 8-8-4-4z"/><path d="M5 11l-2 5 5-2"/></svg>;
    case 'grid': return <svg className={cls} {...s} viewBox="0 0 16 16"><path d="M14 2H2v12h12V2z"/><path d="M2 6h12"/></svg>;
    case 'bell': return <svg className={cls} {...s} viewBox="0 0 18 18"><path d="M13.73 21a2 2 0 0 1-3.46 0"/><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/></svg>;
    case 'link': return <svg className={cls} {...s} viewBox="0 0 16 16"><circle cx="8" cy="8" r="2"/><path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.05 3.05l1.41 1.41M11.54 11.54l1.41 1.41M3.05 12.95l1.41-1.41M11.54 4.46l1.41-1.41"/></svg>;
    case 'users': return <svg className={cls} {...s} viewBox="0 0 16 16"><circle cx="8" cy="5" r="3"/><path d="M2 15c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>;
    default: return null;
  }
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pageName = navItems.flatMap(g => g.items).find(i => pathname.startsWith(i.href))?.label || 'Dashboard';

  return (
    <html lang="zh-CN" className="dark">
      <head><title>Opsight - {pageName}</title><meta name="viewport" content="width=device-width, initial-scale=1.0" /></head>
      <body className="dark">
        <NotificationProvider>
          <ErrorBoundary>
        {/* Sidebar */}
        <aside className={`sidebar fixed top-0 left-0 bottom-0 w-60 bg-surface-50 border-r border-white/5 z-50 flex flex-col transition-transform duration-250 ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'} lg:translate-x-0`}>
          <div className="flex items-center gap-2.5 px-5 h-14 border-b border-white/5 flex-shrink-0">
            <svg width="24" height="24" viewBox="0 0 28 28" fill="none">
              <rect width="28" height="28" rx="6" fill="#0ea5e9" fillOpacity="0.15"/>
              <path d="M8 14L12 10L16 14L20 10" stroke="#0ea5e9" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M8 19L12 15L16 19L20 15" stroke="#0ea5e9" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" opacity="0.5"/>
            </svg>
            <span className="text-white font-semibold tracking-tight">Opsight</span>
          </div>
          <nav className="flex-1 py-3 overflow-y-auto">
            {navItems.map(group => (
              <div key={group.group}>
                <div className="px-4 mb-2 mt-4 first:mt-0">
                  <span className="font-mono text-[10px] uppercase tracking-widest text-zinc-600">{group.group}</span>
                </div>
                {group.items.map(item => {
                  const active = pathname.startsWith(item.href);
                  return (
                    <Link key={item.href} href={item.href}
                      className={`flex items-center gap-3 px-5 py-2 text-sm border-l-2 transition-colors ${active ? 'bg-[rgba(14,165,233,0.1)] text-[#0ea5e9] border-[#0ea5e9]' : 'border-transparent text-zinc-500 hover:bg-white/[0.03] hover:text-zinc-300'}`}
                      onClick={() => setSidebarOpen(false)}>
                      <NavIcon type={item.icon} />
                      {item.label}
                    </Link>
                  );
                })}
              </div>
            ))}
          </nav>
          <div className="px-4 py-3 border-t border-white/5 flex-shrink-0">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-accent/15 flex items-center justify-center text-accent text-xs font-semibold">LH</div>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-zinc-300 truncate">Leo Hang</p>
                <p className="font-mono text-xs text-zinc-600 truncate">Admin</p>
              </div>
            </div>
          </div>
        </aside>

        {/* Overlay */}
        {sidebarOpen && <div className="fixed inset-0 bg-black/50 z-40 lg:hidden" onClick={() => setSidebarOpen(false)} />}

        {/* Main */}
        <div className="lg:ml-60 min-h-screen flex flex-col">
          <header className="sticky top-0 z-40 h-14 bg-surface/80 backdrop-blur-xl border-b border-white/5 flex items-center px-4 sm:px-6 gap-4">
            <button className="lg:hidden text-zinc-400 hover:text-white" onClick={() => setSidebarOpen(true)}>
              <svg width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <line x1="3" y1="5" x2="17" y2="5"/><line x1="3" y1="10" x2="17" y2="10"/><line x1="3" y1="15" x2="17" y2="15"/>
              </svg>
            </button>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-zinc-500">Opsight</span>
              <span className="text-zinc-700">/</span>
              <span className="text-zinc-300">{pageName}</span>
            </div>
            <div className="flex-1" />
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-success" style={{ animation: 'pulse-live 2s ease-in-out infinite' }} />
              <span className="font-mono text-xs text-zinc-500">Live</span>
            </div>
          </header>
          <main className="flex-1 p-4 sm:p-6 space-y-6">{children}</main>
        </div>
          </ErrorBoundary>
        </NotificationProvider>
      </body>
    </html>
  );
}
