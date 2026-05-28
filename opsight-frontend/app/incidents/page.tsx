'use client';

import { useEffect, useState } from 'react';
import { fetchAPI, patchAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState } from '../components/UI';
import { useNotification } from '../components/Notification';

export default function Incidents() {
  const [incidents, setIncidents] = useState<any[]>([]);
  const [filter, setFilter] = useState({ status: 'all', search: '' });
  const [expanded, setExpanded] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const { addNotification } = useNotification();

  useEffect(() => {
    setLoading(true);
    const params = new URLSearchParams();
    if (filter.status !== 'all') params.set('status', filter.status);
    if (filter.search) params.set('search', filter.search);
    fetchAPI(`/incidents?${params}`)
      .then(d => setIncidents(d.incidents || []))
      .catch(e => {
        addNotification('error', 'Failed to load incidents', e.message);
        setIncidents([]);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  const handleResolve = async (id: string) => {
    try {
      await patchAPI(`/incidents/${id}/resolve`);
      setIncidents(prev => prev.map(inc => inc.id === id ? { ...inc, status: 'resolved' } : inc));
      addNotification('success', 'Incident Resolved', `Incident ${id} has been marked as resolved.`);
    } catch (e: any) {
      addNotification('error', 'Failed to resolve', e.message);
    }
  };

  const criticalCount = incidents.filter(i => i.status === 'critical').length;
  const warningCount = incidents.filter(i => i.status === 'warning').length;
  const resolvedCount = incidents.filter(i => i.status === 'resolved').length;

  if (loading) {
    return <LoadingState text="Loading incidents..." />;
  }

  if (incidents.length === 0) {
    return <EmptyState title="No incidents found" description="Try adjusting your search or filter criteria." />;
  }

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
                    <button
                      onClick={() => handleResolve(inc.id)}
                      className="text-xs text-accent hover:text-accent/80 transition-colors"
                    >
                      Mark Resolved
                    </button>
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
