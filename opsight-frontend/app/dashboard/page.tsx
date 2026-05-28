'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState, Card } from '../components/UI';
import { useNotification } from '../components/Notification';

export default function Dashboard() {
  const [summary, setSummary] = useState<any>(null);
  const [incidents, setIncidents] = useState<any[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [topErrors, setTopErrors] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { addNotification } = useNotification();

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetchAPI('/dashboard/summary').catch(e => { addNotification('error', 'Failed to load summary', e.message); return null; }),
      fetchAPI('/incidents?limit=8').catch(e => { addNotification('error', 'Failed to load incidents', e.message); return { incidents: [] }; }),
      fetchAPI('/services').catch(e => { addNotification('error', 'Failed to load services', e.message); return { services: [] }; }),
      fetchAPI('/dashboard/top-errors').catch(e => { addNotification('error', 'Failed to load errors', e.message); return { errors: [] }; }),
    ]).then(([sum, inc, svc, err]) => {
      if (sum) setSummary(sum);
      setIncidents(inc.incidents || []);
      setServices(svc.services || []);
      setTopErrors(err.errors || []);
      setLoading(false);
    });
  }, []);

  const aiInsights = [
    { type: 'root-cause', title: 'Root cause identified', body: 'auth-svc v2.4.1 disabled session-cache eviction. Memory grows 12 MB/min.', time: '2 min ago', color: '#ef4444' },
    { type: 'pattern', title: 'Correlated pattern', body: 'Payment 5xx errors correlate with Redis latency spike.', time: '8 min ago', color: '#f59e0b' },
    { type: 'prediction', title: 'Capacity forecast', body: 'us-east-1 node 7 disk will reach 95% in ~6 hours.', time: '1 hr ago', color: '#0ea5e9' },
    { type: 'resolved', title: 'Auto-remediated', body: 'Kafka consumer lag resolved by scaling consumer group.', time: '2 hr ago', color: '#10b981' },
  ];

  const healthyCount = services.filter(s => s.status === 'healthy').length;
  const degradedCount = services.filter(s => s.status === 'degraded').length;
  const downCount = services.filter(s => s.status === 'down').length;

  if (loading) {
    return <LoadingState text="Loading dashboard..." />;
  }

  return (
    <>
      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">Active Incidents</span>
            <Badge variant="critical"><span className="w-1.5 h-1.5 rounded-full bg-danger inline-block" />Live</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.active_incidents ?? '-'}</p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-danger">+1</span> from last hour</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">Avg MTTR</span>
            <Badge variant="resolved">-62%</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.mttr_minutes ?? '-'}<span className="text-base font-normal text-zinc-500 ml-1">min</span></p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-success">-2.8 min</span> vs last 30d</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">Services Healthy</span>
            <Badge variant="info">{summary ? ((summary.services_healthy / summary.services_total) * 100).toFixed(1) + '%' : '-'}</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.services_healthy ?? '-'}<span className="text-base font-normal text-zinc-500">/{summary?.services_total ?? '-'}</span></p>
          <p className="text-xs text-zinc-600 mt-2">across 12 clusters</p>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <span className="font-mono text-xs text-zinc-500">AI Alerts Today</span>
            <Badge variant="muted">Auto-resolved 73%</Badge>
          </div>
          <p className="text-3xl font-bold text-white tabular-nums">{summary?.ai_alerts_today ?? '-'}</p>
          <p className="text-xs text-zinc-600 mt-2"><span className="text-success">{summary?.ai_auto_resolved ?? '-'} auto</span> / <span className="text-warn">13 manual</span></p>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 bg-surface-50 border border-white/5 rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <div><h3 className="text-sm font-medium text-zinc-200">Error Rate</h3><p className="font-mono text-xs text-zinc-600 mt-0.5">Last 24 hours</p></div>
            <div className="flex gap-1">
              <button className="font-mono text-xs px-2.5 py-1 rounded-md bg-accent/10 text-accent">24h</button>
              <button className="font-mono text-xs px-2.5 py-1 rounded-md text-zinc-500">7d</button>
              <button className="font-mono text-xs px-2.5 py-1 rounded-md text-zinc-500">30d</button>
            </div>
          </div>
          <div className="relative h-52">
            <ErrorRateChart />
          </div>
        </div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
          <h3 className="text-sm font-medium text-zinc-200 mb-4">Service Health</h3>
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
            <h3 className="text-sm font-medium text-zinc-200">Recent Incidents</h3>
            <a href="/incidents" className="text-xs text-accent hover:text-accent/80">View all</a>
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
            <h3 className="text-sm font-medium text-zinc-200">AI Insights</h3>
          </div>
          <div className="space-y-4">
            {aiInsights.map((ins, i) => (
              <div key={i} className="border-l-2 pl-3" style={{ borderColor: ins.color + '30' }}>
                <div className="flex items-center gap-2 mb-1">
                  <div className="w-2 h-2 rounded-full" style={{ backgroundColor: ins.color }} />
                  <span className="text-xs font-medium" style={{ color: ins.color }}>{ins.title}</span>
                </div>
                <p className="text-xs text-zinc-400 leading-relaxed mb-1">{ins.body}</p>
                <span className="font-mono text-[10px] text-zinc-600">{ins.time}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Service Grid */}
      <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-medium text-zinc-200">Service Overview</h3>
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
        <h3 className="text-sm font-medium text-zinc-200 mb-1">Top Errors</h3>
        <p className="font-mono text-xs text-zinc-600 mb-4">Most frequent in last 24h</p>
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
function ErrorRateChart() {
  const [ChartComp, setChartComp] = useState<any>(null);
  useEffect(() => {
    Promise.all([import('chart.js'), import('react-chartjs-2')]).then(([chartjs, reactChart]) => {
      chartjs.Chart.register(chartjs.CategoryScale, chartjs.LinearScale, chartjs.PointElement, chartjs.LineElement, chartjs.Filler, chartjs.Tooltip);
      setChartComp(() => reactChart.Line);
    });
  }, []);

  if (!ChartComp) return <div className="flex items-center justify-center h-full text-zinc-600 text-xs">Loading chart...</div>;

  const labels = Array.from({ length: 24 }, (_, i) => `${String(i).padStart(2, '0')}:00`);
  const data = [0.12, 0.10, 0.08, 0.09, 0.11, 0.15, 0.18, 0.22, 0.31, 0.45, 0.38, 0.29, 0.24, 0.20, 0.18, 0.15, 0.21, 0.33, 0.41, 0.52, 0.38, 0.28, 0.19, 0.14];

  return (
    <ChartComp
      data={{
        labels,
        datasets: [{
          data,
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
