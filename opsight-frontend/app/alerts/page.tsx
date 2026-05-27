'use client';

import { useEffect, useState } from 'react';
import { fetchAPI, patchAPI } from '../lib/api';

function Badge({ children, variant }: { children: React.ReactNode; variant: string }) {
  const s: Record<string, string> = { critical: 'bg-[rgba(239,68,68,0.12)] text-[#f87171]', warning: 'bg-[rgba(245,158,11,0.12)] text-[#fbbf24]', info: 'bg-[rgba(14,165,233,0.12)] text-[#38bdf8]' };
  return <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium font-mono ${s[variant] || s.info}`}>{children}</span>;
}

export default function Alerts() {
  const [rules, setRules] = useState<any[]>([]);

  useEffect(() => { fetchAPI('/alert-rules').then(d => setRules(d.rules || [])).catch(console.error); }, []);

  const toggle = async (id: string) => {
    try {
      await patchAPI(`/alert-rules/${id}/toggle`);
      setRules(rules.map(r => r.id === id ? { ...r, enabled: !r.enabled } : r));
    } catch (e) { console.error(e); }
  };

  const enabled = rules.filter(r => r.enabled).length;
  const aiGenerated = rules.filter(r => r.is_ai).length;

  return (
    <>
      <div className="flex items-center justify-between">
        <div><h1 className="text-xl font-semibold text-white">Alert Rules</h1><p className="text-sm text-zinc-500 mt-0.5">Configure thresholds and notification channels</p></div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Total Rules</span><p className="text-2xl font-bold text-white mt-1">{rules.length}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Enabled</span><p className="text-2xl font-bold text-success mt-1">{enabled}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">AI-Generated</span><p className="text-2xl font-bold text-accent mt-1">{aiGenerated}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Triggered Today</span><p className="text-2xl font-bold text-warn mt-1">7</p></div>
      </div>

      <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead><tr className="text-left border-b border-white/5">
              {['Rule', 'Condition', 'Threshold', 'Service', 'Severity', 'Last Triggered', 'Enabled'].map(h => (
                <th key={h} className="font-mono text-[10px] uppercase tracking-wider text-zinc-600 px-4 py-2.5">{h}</th>
              ))}
            </tr></thead>
            <tbody>
              {rules.map(rule => (
                <tr key={rule.id} className="border-b border-white/[0.03] hover:bg-white/[0.02] transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-300">{rule.name}</span>
                      {rule.is_ai && <span className="font-mono text-[9px] px-1.5 py-0.5 rounded bg-accent/10 text-accent">AI</span>}
                    </div>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">{rule.condition}</td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">{rule.threshold}</td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">{rule.service}</td>
                  <td className="px-4 py-3"><Badge variant={rule.severity}>{rule.severity}</Badge></td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-600">{rule.last_triggered}</td>
                  <td className="px-4 py-3">
                    <button onClick={() => toggle(rule.id)} className={`w-9 h-5 rounded-full relative transition-colors ${rule.enabled ? 'bg-[rgba(14,165,233,0.3)]' : 'bg-surface-300'}`}>
                      <div className={`w-4 h-4 rounded-full absolute top-0.5 transition-transform ${rule.enabled ? 'translate-x-4 bg-[#0ea5e9]' : 'translate-x-0.5 bg-zinc-500'}`} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
