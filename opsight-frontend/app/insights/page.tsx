'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState } from '../components/UI';

const tabs = [
  { key: 'root-cause', label: '根因分析' },
  { key: 'predictions', label: '预测' },
  { key: 'remediation', label: '自动修复' },
  { key: 'patterns', label: '模式识别' },
];

export default function Insights() {
  const [activeTab, setActiveTab] = useState('root-cause');
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchAPI(`/insights?type=${activeTab}`)
      .then(d => setData(d.insights || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [activeTab]);

  return (
    <>
      <div><h1 className="text-xl font-semibold text-white">AI 洞察</h1><p className="text-sm text-zinc-500 mt-0.5">AI 驱动的分析、预测与自动修复</p></div>

      <div className="flex gap-6 border-b border-white/5">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setActiveTab(t.key)}
            className={`text-sm py-2 px-1 border-b-2 transition-colors ${activeTab === t.key ? 'text-[#0ea5e9] border-[#0ea5e9]' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {loading ? <LoadingState text="加载洞察数据…" /> : data.length === 0 ? <EmptyState title="暂无洞察" description="当前类型下没有数据" /> : (
        <div className="space-y-3">
          {data.map((ins, i) => (
            <div key={i} className="bg-surface-50 border border-white/5 rounded-xl p-5">
              <div className="flex items-center gap-2 mb-2 flex-wrap">
                <Badge variant={ins.type}>{ins.type}</Badge>
                <span className="font-mono text-[10px] text-zinc-600">{ins.service}</span>
                <span className="font-mono text-[10px] text-zinc-600">置信度: {ins.confidence}</span>
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
      )}
    </>
  );
}
