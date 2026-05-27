'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

export default function Services() {
  const [services, setServices] = useState<any[]>([]);

  useEffect(() => {
    fetchAPI('/services').then(d => setServices(d.services || [])).catch(console.error);
  }, []);

  const healthy = services.filter(s => s.status === 'healthy').length;
  const degraded = services.filter(s => s.status === 'degraded').length;
  const down = services.filter(s => s.status === 'down').length;

  return (
    <>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Total Services</span><p className="text-2xl font-bold text-white mt-1">{services.length}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Healthy</span><p className="text-2xl font-bold text-success mt-1">{healthy}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Degraded</span><p className="text-2xl font-bold text-warn mt-1">{degraded}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Down</span><p className="text-2xl font-bold text-danger mt-1">{down}</p></div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {services.map(svc => {
          const dot = ({ healthy: 'bg-success', degraded: 'bg-warn', down: 'bg-danger' } as Record<string, string>)[svc.status] || 'bg-zinc-500';
          const badge = ({ healthy: 'text-success', degraded: 'text-warn', down: 'text-danger' } as Record<string, string>)[svc.status] || 'text-zinc-500';
          const errNum = parseFloat(svc.err_rate) || 0;
          return (
            <div key={svc.name} className="bg-surface-50 border border-white/5 rounded-xl p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2"><div className={`w-2 h-2 rounded-full ${dot}`} /><span className="text-sm font-medium text-zinc-200">{svc.name}</span></div>
                <span className={`font-mono text-[11px] ${badge}`}>{svc.status}</span>
              </div>
              <div className="grid grid-cols-4 gap-2 mb-3">
                {[{ label: 'RPS', val: svc.rps }, { label: 'p50', val: svc.p50 }, { label: 'p99', val: svc.p99 }, { label: 'Errors', val: svc.err_rate }].map(m => (
                  <div key={m.label}><p className="font-mono text-[10px] text-zinc-600">{m.label}</p><p className="font-mono text-xs text-zinc-400">{m.val}</p></div>
                ))}
              </div>
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10px] text-zinc-600">{svc.uptime} uptime</span>
                <span className="font-mono text-[10px] text-zinc-600">{svc.team}</span>
              </div>
              {svc.deps?.length > 0 && (
                <div className="mt-3 pt-3 border-t border-white/5">
                  <p className="font-mono text-[10px] text-zinc-600 mb-1.5">Dependencies</p>
                  <div className="flex flex-wrap gap-1">{svc.deps.map((d: string) => <span key={d} className="font-mono text-[10px] px-1.5 py-0.5 rounded bg-surface-300 text-zinc-500">{d}</span>)}</div>
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Dependency Matrix */}
      <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
        <div className="p-4 border-b border-white/5"><h3 className="text-sm font-medium text-zinc-200">Dependency Matrix</h3></div>
        <div className="overflow-x-auto p-4">
          {services.length > 0 && <DepMatrix services={services} />}
        </div>
      </div>
    </>
  );
}

function DepMatrix({ services }: { services: any[] }) {
  const names = services.map(s => s.name.length > 10 ? s.name.slice(0, 8) + '..' : s.name);
  const fullNames = services.map(s => s.name);
  return (
    <>
      <table className="w-full text-xs">
        <thead><tr><th className="p-2 text-zinc-600 font-mono text-[10px]"></th>{names.map(n => <th key={n} className="p-2 text-zinc-600 font-mono text-[10px]">{n}</th>)}</tr></thead>
        <tbody>
          {services.map((row, ri) => (
            <tr key={row.name}>
              <td className="p-2 font-mono text-zinc-500">{names[ri]}</td>
              {services.map((col, ci) => {
                if (ri === ci) return <td key={ci} className="p-2 text-center text-zinc-700">-</td>;
                const depends = row.deps?.includes(col.name);
                if (!depends) return <td key={ci} className="p-2 text-center text-zinc-700" />;
                const color = col.status === 'healthy' ? 'bg-success' : col.status === 'degraded' ? 'bg-warn' : 'bg-danger';
                return <td key={ci} className="p-2 text-center"><div className={`w-3 h-3 rounded-full ${color} mx-auto`} /></td>;
              })}
            </tr>
          ))}
        </tbody>
      </table>
      <div className="flex items-center gap-4 mt-4 pt-3 border-t border-white/5">
        <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-success" /><span className="font-mono text-[10px] text-zinc-500">Healthy dep</span></div>
        <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-warn" /><span className="font-mono text-[10px] text-zinc-500">Degraded dep</span></div>
        <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-danger" /><span className="font-mono text-[10px] text-zinc-500">Down dep</span></div>
      </div>
    </>
  );
}
