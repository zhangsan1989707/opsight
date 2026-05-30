export function Badge({ children, variant }: { children: React.ReactNode; variant: string }) {
  const styles: Record<string, string> = {
    critical: 'bg-[rgba(239,68,68,0.12)] text-[#f87171]',
    warning: 'bg-[rgba(245,158,11,0.12)] text-[#fbbf24]',
    resolved: 'bg-[rgba(16,185,129,0.12)] text-[#34d399]',
    info: 'bg-[rgba(14,165,233,0.12)] text-[#38bdf8]',
    muted: 'bg-[rgba(113,113,122,0.12)] text-[#a1a1aa]',
    healthy: 'bg-[rgba(16,185,129,0.12)] text-[#34d399]',
    degraded: 'bg-[rgba(245,158,11,0.12)] text-[#fbbf24]',
    down: 'bg-[rgba(239,68,68,0.12)] text-[#f87171]',
    connected: 'bg-[rgba(16,185,129,0.12)] text-[#34d399]',
    disconnected: 'bg-[rgba(113,113,122,0.12)] text-[#a1a1aa]',
  };
  return <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium font-mono ${styles[variant] || styles.muted}`}>{children}</span>;
}

export function LoadingSpinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizes = { sm: 'w-4 h-4', md: 'w-8 h-8', lg: 'w-12 h-12' };
  return (
    <div className={`${sizes[size]} animate-spin`}>
      <svg className="w-full h-full" viewBox="0 0 24 24" fill="none">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
      </svg>
    </div>
  );
}

export function LoadingState({ text = 'Loading…' }: { text?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-zinc-600">
      <LoadingSpinner />
      <p className="mt-3 text-sm font-mono">{text}</p>
    </div>
  );
}

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <svg width="48" height="48" fill="none" stroke="#52525b" strokeWidth="1.5" viewBox="0 0 24 24">
        <path d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
      <h3 className="mt-3 text-sm font-medium text-zinc-400">{title}</h3>
      <p className="mt-1 text-xs text-zinc-600">{description}</p>
    </div>
  );
}

export function Card({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`bg-surface-50 border border-white/5 rounded-xl p-5 ${className}`}>
      {children}
    </div>
  );
}
