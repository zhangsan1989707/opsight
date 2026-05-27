'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

function Badge({ children, variant }: { children: React.ReactNode; variant: string }) {
  const s: Record<string, string> = { critical: 'bg-[rgba(239,68,68,0.12)] text-[#f87171]', warning: 'bg-[rgba(245,158,11,0.12)] text-[#fbbf24]', resolved: 'bg-[rgba(16,185,129,0.12)] text-[#34d399]', info: 'bg-[rgba(14,165,233,0.12)] text-[#38bdf8]', muted: 'bg-[rgba(113,113,122,0.12)] text-[#a1a1aa]' };
  return <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium font-mono ${s[variant] || s.muted}`}>{children}</span>;
}

export default function Incidents() {
  const [incidents, setIncidents] = useState<any[]>([]);
  const [filter, setFilter] = useState({ status: 'all', search: '' });
  const [expanded, setExpanded] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams();
    if (filter.status !== 'all') params.set('status', filter.status);
    if (filter.search) params.set('search', filter.search);
    fetchAPI(`/incidents?${params}`).then(d => setIncidents(d.incidents || [])).catch(console.error);
  }, [filter]);

  const criticalCount = incidents.filter(i => i.status === 'critical').length;
  const warningCount = incidents.filter(i => i.status === 'warning').length;
  const resolvedCount = incidents.filter(i => i.status === 'resolved').length;

  return (
    <>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div><h1 className="text-xl font-semibold text-white">Incidents</h1><p className="text-sm text-zinc-500 mt-0.5">Track and resolve active incidents</p></div>
        <div className="flex items-center gap-2">
          <Badge variant="critical">{criticalCount} Critical</Badge>
          <Badge variant="warning">{warningCount} Warning</Badge>
          <Badge variant="resolved">{resolvedCount} Resolved</Badge>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <input type="text" placeholder="Search incidents..." value={filter.search} onChange={e => setFilter({ ...filter, search: e.target.value })}
          className="bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30 w-64" />
        <select value={filter.status} onChange={e => setFilter({ ...filter, status: e.target.value })}
          className="bg-surface-100 border border-white/5 text-sm text-zinc-300 rounded-lg px-3 py-2 pr-8 outline-none focus:border-accent/30">
          <option value="all">All Status</option><option value="critical">Critical</option><option value="warning">Warning</option><option value="resolved">Resolved</option>
        </select>
      </div>

      <div className="space-y-3">
        {incidents.map(inc => (
          <div key={inc.id} className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
            <button className="w-full text-left p-4 hover:bg-white/[0.02] transition-colors" onClick={() => setExpanded(expanded === inc.id ? null : inc.id)}>
              <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <span className="font-mono text-xs text-zinc-500 flex-shrink-0">{inc.id}</span>
                  <span className="text-sm text-zinc-300 truncate">{inc.summary}</span>
                </div>
                <div className="flex items-center gap-3 flex-shrink-0">
                  <span className="font-mono text-xs text-zinc-400">{inc.service}</span>
                  <Badge variant={inc.status}>{inc.status.charAt(0).toUpperCase() + inc.status.slice(1)}</Badge>
                  <span className="font-mono text-xs text-zinc-600">{inc.time}</span>
                  <svg className={`w-4 h-4 text-zinc-600 transition-transform ${expanded === inc.id ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 16 16"><path d="M4 6l4 4 4-4" /></svg>
                </div>
              </div>
            </button>
            {expanded === inc.id && (
              <div className="px-4 pb-4 border-t border-white/5 pt-3">
                <p className="text-sm text-zinc-400 leading-relaxed">{inc.detail || 'No additional details available.'}</p>
                <div className="flex items-center gap-4 mt-3">
                  <span className="font-mono text-[10px] text-zinc-600">Duration: {inc.duration}</span>
                  {inc.status !== 'resolved' && (
                    <button className="text-xs text-accent hover:text-accent/80 transition-colors">Mark Resolved</button>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="flex items-center justify-between">
        <span className="font-mono text-xs text-zinc-600">Showing {incidents.length} incidents</span>
      </div>
    </>
  );
}
