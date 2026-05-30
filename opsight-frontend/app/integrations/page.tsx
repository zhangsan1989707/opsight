'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';
import { LoadingState, EmptyState } from '../components/UI';

export default function Integrations() {
  const [integrations, setIntegrations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAPI('/integrations')
      .then(d => setIntegrations(d.integrations || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingState text="加载集成列表…" />;
  if (integrations.length === 0) return <EmptyState title="暂无集成" description="尚未配置任何第三方集成" />;

  return (
    <>
      <div className="flex items-center justify-between">
        <div><h1 className="text-xl font-semibold text-white">集成管理</h1><p className="text-sm text-zinc-500 mt-0.5">连接外部服务与数据源</p></div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {integrations.map(int => (
          <div key={int.id} className="bg-surface-50 border border-white/5 rounded-xl p-5">
            <div className="flex items-center justify-between mb-3">
              <div>
                <h3 className="text-sm font-medium text-zinc-200">{int.name}</h3>
                <span className="font-mono text-[10px] text-zinc-600">{int.category}</span>
              </div>
              <div className={`w-2 h-2 rounded-full ${int.status === 'connected' ? 'bg-success' : 'bg-zinc-600'}`} />
            </div>
            <div className="flex items-center justify-between mt-4">
              <span className="font-mono text-xs text-zinc-500">{int.type}</span>
              <span className="font-mono text-xs text-zinc-600">{int.event_count?.toLocaleString()} 事件</span>
            </div>
            <div className="mt-3 pt-3 border-t border-white/5 flex items-center justify-between">
              <span className={`font-mono text-[10px] ${int.status === 'connected' ? 'text-success' : 'text-zinc-600'}`}>{int.status === 'connected' ? '已连接' : '未连接'}</span>
              <span className={`font-mono text-[10px] ${int.enabled ? 'text-accent' : 'text-zinc-600'}`}>{int.enabled ? '已启用' : '已禁用'}</span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
