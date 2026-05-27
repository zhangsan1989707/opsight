'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

const tabs = [
  { key: 'root-cause', label: 'Root Causes' },
  { key: 'predictions', label: 'Predictions' },
  { key: 'remediation', label: 'Remediations' },
  { key: 'patterns', label: 'Patterns' },
];

function Badge({ children, variant }: { children: React.ReactNode; variant: string }) {
  const s: Record<string, string> = { critical: 'bg-[rgba(239,68,68,0.12)] text-[#f87171]', warning: 'bg-[rgba(245,158,11,0.12)] text-[#fbbf24]', resolved: 'bg-[rgba(16,185,129,0.12)] text-[#34d399]', info: 'bg-[rgba(14,165,233,0.12)] text-[#38bdf8]' };
  return <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium font-mono ${s[variant] || s.info}`}>{children}</span>;
}

export default function Insights() {
  const [activeTab, setActiveTab] = useState('root-cause');
  const [data, setData] = useState<any[]>([]);

  useEffect(() => {
    fetchAPI(`/insights?type=${activeTab}`).then(d => setData(d.insights || [])).catch(console.error);
  }, [activeTab]);

  return (
    <>
      <div><h1 className="text-xl font-semibold text-white">AI Insights</h1><p className="text-sm text-zinc-500 mt-0.5">AI-powered analysis, predictions, and automated remediation</p></div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Root Causes Found</span><p className="text-2xl font-bold text-accent mt-1">147</p><p className="font-mono text-[10px] text-zinc-600 mt-1">last 30d</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Auto-Remediated</span><p className="text-2xl font-bold text-success mt-1">89</p><p className="font-mono text-[10px] text-zinc-600 mt-1">61% of total</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">Predictions Made</span><p className="text-2xl font-bold text-warn mt-1">34</p><p className="font-mono text-[10px] text-zinc-600 mt-1">28 confirmed</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">MTTR Reduction</span><p className="text-2xl font-bold text-success mt-1">-62%</p><p className="font-mono text-[10px] text-zinc-600 mt-1">vs pre-AI baseline</p></div>
      </div>

      <div className="flex gap-6 border-b border-white/5">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setActiveTab(t.key)}
            className={`text-sm py-2 px-1 border-b-2 transition-colors ${activeTab === t.key ? 'text-[#0ea5e9] border-[#0ea5e9]' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}>
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-3">
        {data.map((ins, i) => (
          <div key={i} className="bg-surface-50 border border-white/5 rounded-xl p-5">
            <div className="flex items-center gap-2 mb-2 flex-wrap">
              <Badge variant={ins.type}>{ins.type?.charAt(0).toUpperCase() + ins.type?.slice(1)}</Badge>
              <span className="font-mono text-[10px] text-zinc-600">{ins.service}</span>
              <span className="font-mono text-[10px] text-zinc-600">Confidence: {ins.confidence}</span>
            </div>
            <h3 className="text-sm text-zinc-200 mb-2">{ins.title}</h3>
            <p className="text-xs text-zinc-400 leading-relaxed mb-3">{ins.body}</p>
            <div className="flex items-center gap-3">
              <span className="font-mono text-[10px] text-zinc-600">{ins.time}</span>
              {ins.related && <a href="/incidents" className="font-mono text-[10px] text-accent hover:underline">{ins.related}</a>}
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
