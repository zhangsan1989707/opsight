'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState, Card } from '../components/UI';
import { useNotification } from '../components/Notification';
import { useWS } from '../context/WSContext';

export default function Dashboard() {
  const [summary, setSummary] = useState<any>(null);
  const [incidents, setIncidents] = useState<any[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [topErrors, setTopErrors] = useState<any[]>([]);
  const [aiInsights, setAiInsights] = useState<any[]>([]);
  const [errorRate, setErrorRate] = useState<{ labels: string[]; values: number[] }>({ labels: [], values: [] });
  const [loading, setLoading] = useState(true);
  const { lastEvent } = useWS();
  const { addNotification } = useNotification();

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetchAPI('/dashboard/summary').catch(e => { addNotification('error', '加载失败', e.message); return null; }),
      fetchAPI('/incidents?limit=8').catch(e => { addNotification('error', '加载失败', e.message); return { incidents: [] }; }),
      fetchAPI('/services').catch(e => { addNotification('error', '加载失败', e.message); return { services: [] }; }),
      fetchAPI('/dashboard/top-errors').catch(e => { addNotification('error', '加载失败', e.message); return { errors: [] }; }),
      fetchAPI('/insights?type=root-cause').catch(() => ({ insights: [] })),
      fetchAPI('/dashboard/error-rate').catch(() => ({ labels: [], values: [] })),
    ]).then(([sum, inc, svc, err, ins, rate]) => {
      if (sum) setSummary(sum);
      setIncidents(inc.incidents || []);
      setServices(svc.services || []);
      setTopErrors(err.errors || []);
      setAiInsights(ins.insights || []);
      if (rate.labels && rate.values) setErrorRate(rate);
      setLoading(false);
    });
  }, []);

  // React to WebSocket events
  useEffect(() => {
    if (!lastEvent) return;
    if (lastEvent.type === 'alert_firing') {
      addNotification('error', `告警触发: ${lastEvent.data?.name}`, `${lastEvent.data?.hostname} - ${lastEvent.data?.severity}`);
    }
    if (lastEvent.type === 'service_status') {
      fetchAPI('/services').then(d => setServices(d.services || [])).catch(() => {});
    }
  }, [lastEvent]);

  const insightColors: Record<string, string> = {
    'root-cause': '#ef4444',
    patterns: '#f59e0b',
    predictions: '#0ea5e9',
    remediation: '#10b981',
  };

  const healthyCount = services.filter(s => s.status === 'healthy').length;
  const degradedCount = services.filter(s => s.status === 'degraded').length;
  const downCount = services.filter(s => s.status === 'down').length;

  if (loading) {
    return <LoadingState text="加载仪表盘…" />;
  }

  return (
    <>
      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">活跃事件</span>
            <Badge variant="critical"><span className="w-1.5 h-1.5 rounded-full bg-danger inline-block" />Live</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.active_incidents ?? '-'}</p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-danger">+1</span> 较上小时</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">平均恢复时间</span>
            <Badge variant="resolved">-62%</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.mttr_minutes ?? '-'}<span className="text-base font-normal text-zinc-500 ml-1">min</span></p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-success">-2.8 min</span> vs last 30d</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">服务健康</span>
            <Badge variant="info">{summary ? ((summary.services_healthy / summary.services_total) * 100).toFixed(1) + '%' : '-'}</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.services_healthy ?? '-'}<span className="text-base font-normal text-zinc-500">/{summary?.services_total ?? '-'}</span></p>
          <p className="text-xs text-zinc-600 mt-2">集群总览</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">今日 AI 告警</span>
            <Badge variant="muted">自动修复</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.ai_alerts_today ?? '-'}</p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-success">{summary?.ai_auto_resolved ?? '-'} 自动</span> / <span className="text-warn">手动</span></p>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 bg-surface-50 border border-white/5 rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <div><h3 className="text-sm font-medium text-zinc-200">错误率</h3><p className="font-mono text-xs text-zinc-600 mt-0.5">最近 24 小时</p></div>
            <div className="flex gap-1">
              <button className="font-mono text-xs px-2.5 py-1 rounded-md bg-accent/10 text-accent focus-visible:ring-1 focus-visible:ring-accent/50 focus-visible:outline-none">24h</button>
              <button className="font-mono text-xs px-2.5 py-1 rounded-md text-zinc-500 focus-visible:ring-1 focus-visible:ring-accent/50 focus-visible:outline-none">7d</button>
              <button className="font-mono text-xs px-2.5 py-1 rounded-md text-zinc-500 focus-visible:ring-1 focus-visible:ring-accent/50 focus-visible:outline-none">30d</button>
            </div>
          </div>
          <div className="relative h-52">
            <ErrorRateChart labels={errorRate.labels} values={errorRate.values} />
          </div>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
          <h3 className="text-sm font-medium text-zinc-200 mb-4">服务健康度</h3>
          <div className="flex justify-center mb-4">
            <div className="relative w-36 h-36">
              <HealthDonut healthy={healthyCount} degraded={degradedCount} down={downCount} />
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-2xl font-bold text-white">{summary ? ((summary.services_healthy / summary.services_total) * 100).toFixed(1) : '-'}%</span>
                <span className="font-mono text-xs text-zinc-500">healthy</span>
              </div>
            </div>
          </div>
          <div className="space-y-2.5">
            {[{ label: 'Healthy', count: healthyCount, color: 'bg-success' }, { label: 'Degraded', count: degradedCount, color: 'bg-warn' }, { label: 'Down', count: downCount, color: 'bg-danger' }].map(i => (
              <div key={i.label} className="flex items-center justify-between">
                <div className="flex items-center gap-2"><div className={`w-2 h-2 rounded-full ${i.color}`} /><span className="text-sm text-zinc-400">{i.label}</span></div>
                <span className="font-mono text-sm text-zinc-300">{i.count}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Incidents + AI Insights */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
          <div className="flex items-center justify-between p-4 border-b border-white/5">
            <h3 className="text-sm font-medium text-zinc-200">最近事件</h3>
            <a href="/incidents" className="text-xs text-accent hover:text-accent/80">查看全部</a>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-left border-b border-white/5">
                {['ID', 'Summary', 'Service', 'Status', 'Duration', 'Time'].map(h => (
                  <th key={h} className="font-mono text-[10px] uppercase tracking-wider text-zinc-600 px-4 py-2.5">{h}</th>
                ))}
              </tr></thead>
              <tbody>
                {incidents.slice(0, 8).map(inc => (
                  <tr key={inc.id} className="border-b border-white/[0.03] hover:bg-white/[0.02] transition-colors">
                    <td className="px-4 py-3 font-mono text-xs text-zinc-500">{inc.id}</td>
                    <td className="px-4 py-3 text-zinc-300 max-w-[280px] truncate">{inc.summary}</td>
                    <td className="px-4 py-3 font-mono text-xs text-zinc-400">{inc.service}</td>
                    <td className="px-4 py-3"><Badge variant={inc.status}>{inc.status.charAt(0).toUpperCase() + inc.status.slice(1)}</Badge></td>
                    <td className="px-4 py-3 font-mono text-xs text-zinc-500">{inc.duration}</td>
                    <td className="px-4 py-3 font-mono text-xs text-zinc-600">{inc.time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
          <div className="flex items-center gap-2 mb-4">
            <svg width="16" height="16" fill="none" stroke="#0ea5e9" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M9 1L1 9l4 4 8-8-4-4z"/><path d="M5 11l-2 5 5-2"/></svg>
            <h3 className="text-sm font-medium text-zinc-200">AI 洞察</h3>
          </div>
          <div className="space-y-4">
            {aiInsights.length === 0 && <p className="text-xs text-zinc-600">No insights available</p>}
            {aiInsights.map((ins: any, i: number) => {
              const color = insightColors[ins.type] || '#0ea5e9';
              return (
              <div key={i} className="border-l-2 pl-3" style={{ borderColor: color + '30' }}>
                <div className="flex items-center gap-2 mb-1">
                  <div className="w-2 h-2 rounded-full" style={{ backgroundColor: color }} />
                  <span className="text-xs font-medium" style={{ color: color }}>{ins.title}</span>
                </div>
                <p className="text-xs text-zinc-400 leading-relaxed mb-1">{ins.body}</p>
                {ins.service && <span className="font-mono text-[10px] text-zinc-600">{ins.service}</span>}
              </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Service Grid */}
      <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-medium text-zinc-200">服务总览</h3>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-success" /><span className="font-mono text-xs text-zinc-500">OK</span></div>
            <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-warn" /><span className="font-mono text-xs text-zinc-500">Degraded</span></div>
            <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-danger" /><span className="font-mono text-xs text-zinc-500">Down</span></div>
          </div>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
          {services.map(svc => {
            const dotColor = ({ healthy: 'bg-success', degraded: 'bg-warn', down: 'bg-danger' } as Record<string, string>)[svc.status] || 'bg-zinc-500';
            const borderColor = ({ healthy: 'border-white/5', degraded: 'border-[#f59e0b]/20', down: 'border-[#ef4444]/30' } as Record<string, string>)[svc.status] || 'border-white/5';
            return (
              <div key={svc.name} className={`bg-surface border ${borderColor} rounded-lg p-3 hover:bg-surface-100 transition-colors cursor-pointer`}>
                <div className="flex items-center gap-2 mb-2">
                  <div className={`w-1.5 h-1.5 rounded-full ${dotColor}`} />
                  <span className="text-xs font-medium text-zinc-300 truncate">{svc.name}</span>
                </div>
                <div className="flex justify-between font-mono text-[10px] text-zinc-500">
                  <span>{svc.rps} rps</span>
                  <span>p99 {svc.p99}</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Top Errors */}
      <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
        <h3 className="text-sm font-medium text-zinc-200 mb-1">高频错误</h3>
        <p className="font-mono text-xs text-zinc-600 mb-4">最近 24 小时最频繁错误</p>
        <div className="space-y-3">
          {topErrors.map((err: any, i: number) => {
            const maxCount = topErrors[0]?.count || 1;
            const width = Math.round((err.count / maxCount) * 100);
            return (
              <div key={i}>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-zinc-300 truncate max-w-[240px]">{err.error}</span>
                  <div className="flex items-center gap-1.5 flex-shrink-0 ml-2">
                    <span className="font-mono text-xs text-zinc-500">{err.count?.toLocaleString()}</span>
                    <span className={`text-[10px] ${err.trend === 'up' ? 'text-danger' : err.trend === 'down' ? 'text-success' : 'text-zinc-600'}`}>
                      {err.trend === 'up' ? '↑' : err.trend === 'down' ? '↓' : '—'}
                    </span>
                  </div>
                </div>
                <div className="h-1 bg-surface-300 rounded-full overflow-hidden">
                  <div className="h-full bg-accent/30 rounded-full" style={{ width: `${width}%` }} />
                </div>
                <span className="font-mono text-[10px] text-zinc-600">{err.service}</span>
              </div>
            );
          })}
        </div>
      </div>
    </>
  );
}

// Chart components (lazy loaded to avoid SSR issues)
function ErrorRateChart({ labels, values }: { labels: string[]; values: number[] }) {
  const [ChartComp, setChartComp] = useState<any>(null);
  useEffect(() => {
    Promise.all([import('chart.js'), import('react-chartjs-2')]).then(([chartjs, reactChart]) => {
      chartjs.Chart.register(chartjs.CategoryScale, chartjs.LinearScale, chartjs.PointElement, chartjs.LineElement, chartjs.Filler, chartjs.Tooltip);
      setChartComp(() => reactChart.Line);
    });
  }, []);

  if (!ChartComp) return <div className="flex items-center justify-center h-full text-zinc-600 text-xs">加载图表中…</div>;

  return (
    <ChartComp
      data={{
        labels,
        datasets: [{
          data: values,
          borderColor: '#0ea5e9',
          borderWidth: 2,
          backgroundColor: 'rgba(14, 165, 233, 0.1)',
          fill: true,
          tension: 0.4,
          pointRadius: 0,
          pointHoverRadius: 4,
        }],
      }}
      options={{
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { display: false }, tooltip: { backgroundColor: '#18181b', borderColor: 'rgba(255,255,255,0.08)', borderWidth: 1, bodyColor: '#e4e4e7', cornerRadius: 8, displayColors: false } },
        scales: {
          x: { grid: { display: false }, ticks: { maxTicksLimit: 8, color: '#52525b', font: { family: 'JetBrains Mono', size: 10 } } },
          y: { grid: { color: 'rgba(255,255,255,0.03)' }, ticks: { callback: (v: any) => v.toFixed(1) + '%', color: '#52525b', font: { family: 'JetBrains Mono', size: 10 } }, min: 0 },
        },
      }}
    />
  );
}

function HealthDonut({ healthy, degraded, down }: { healthy: number; degraded: number; down: number }) {
  const [ChartComp, setChartComp] = useState<any>(null);
  useEffect(() => {
    Promise.all([import('chart.js'), import('react-chartjs-2')]).then(([chartjs, reactChart]) => {
      chartjs.Chart.register(chartjs.ArcElement, chartjs.Tooltip);
      setChartComp(() => reactChart.Doughnut);
    });
  }, []);

  if (!ChartComp) return <div className="w-full h-full bg-surface-300 rounded-full animate-pulse" />;

  return (
    <ChartComp
      data={{ labels: ['Healthy', 'Degraded', 'Down'], datasets: [{ data: [healthy, degraded, down], backgroundColor: ['#10b981', '#f59e0b', '#ef4444'], borderWidth: 0 }] }}
      options={{ responsive: true, maintainAspectRatio: true, cutout: '78%', plugins: { legend: { display: false } } }}
    />
  );
}
