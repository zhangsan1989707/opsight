'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

export default function Metrics() {
  const [names, setNames] = useState<string[]>([]);
  const [selected, setSelected] = useState('cpu_usage');
  const [data, setData] = useState<any>(null);

  useEffect(() => { fetchAPI('/metrics/names').then(d => setNames(d.metrics || [])).catch(console.error); }, []);
  useEffect(() => { fetchAPI(`/metrics/query?metric=${selected}`).then(setData).catch(console.error); }, [selected]);

  return (
    <>
      <div><h1 className="text-xl font-semibold text-white">Metrics</h1><p className="text-sm text-zinc-500 mt-0.5">System and service metrics</p></div>
      <div className="flex flex-wrap gap-2">
        {names.map(n => (
          <button key={n} onClick={() => setSelected(n)}
            className={`font-mono text-xs px-3 py-1.5 rounded-lg transition-colors ${selected === n ? 'bg-accent/10 text-accent' : 'bg-surface-50 border border-white/5 text-zinc-500 hover:text-zinc-300'}`}>
            {n}
          </button>
        ))}
      </div>
      <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <div><h3 className="text-sm font-medium text-zinc-200">{selected}</h3><p className="font-mono text-xs text-zinc-600 mt-0.5">Last 24 hours</p></div>
        </div>
        <div className="relative h-64">
          <MetricChart data={data} />
        </div>
      </div>
      {data?.points && (
        <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
          <h3 className="text-sm font-medium text-zinc-200 mb-3">Data Points</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-left border-b border-white/5">
                {['Time', 'Value', 'Avg', 'P95', 'P99'].map(h => <th key={h} className="font-mono text-[10px] uppercase tracking-wider text-zinc-600 px-4 py-2.5">{h}</th>)}
              </tr></thead>
              <tbody>
                {data.points.slice(-12).map((p: any, i: number) => (
                  <tr key={i} className="border-b border-white/[0.03]">
                    <td className="px-4 py-2 font-mono text-xs text-zinc-500">{p.timestamp}</td>
                    <td className="px-4 py-2 font-mono text-xs text-zinc-300">{p.value}</td>
                    <td className="px-4 py-2 font-mono text-xs text-zinc-400">{p.avg}</td>
                    <td className="px-4 py-2 font-mono text-xs text-warn">{p.p95}</td>
                    <td className="px-4 py-2 font-mono text-xs text-danger">{p.p99}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </>
  );
}

function MetricChart({ data }: { data: any }) {
  const [ChartComp, setChartComp] = useState<any>(null);
  useEffect(() => {
    if (!data) return;
    Promise.all([import('chart.js'), import('react-chartjs-2')]).then(([chartjs, reactChart]) => {
      chartjs.Chart.register(chartjs.CategoryScale, chartjs.LinearScale, chartjs.PointElement, chartjs.LineElement, chartjs.Filler, chartjs.Tooltip);
      setChartComp(() => reactChart.Line);
    });
  }, [data]);

  if (!data) return <div className="flex items-center justify-center h-full text-zinc-600 text-xs">Select a metric</div>;
  if (!ChartComp) return <div className="flex items-center justify-center h-full text-zinc-600 text-xs">Loading chart...</div>;

  const labels = data.points.map((p: any) => p.timestamp);
  return (
    <ChartComp
      data={{
        labels,
        datasets: [
          { label: 'Avg', data: data.points.map((p: any) => p.avg), borderColor: '#0ea5e9', borderWidth: 2, fill: false, tension: 0.4, pointRadius: 0 },
          { label: 'P95', data: data.points.map((p: any) => p.p95), borderColor: '#f59e0b', borderWidth: 1.5, fill: false, tension: 0.4, pointRadius: 0, borderDash: [4, 4] },
          { label: 'P99', data: data.points.map((p: any) => p.p99), borderColor: '#ef4444', borderWidth: 1.5, fill: false, tension: 0.4, pointRadius: 0, borderDash: [4, 4] },
        ],
      }}
      options={{
        responsive: true, maintainAspectRatio: false,
        plugins: { legend: { position: 'top', align: 'end', labels: { boxWidth: 8, usePointStyle: true, color: '#a1a1aa', font: { family: 'JetBrains Mono', size: 10 } } } },
        scales: {
          x: { grid: { display: false }, ticks: { maxTicksLimit: 12, color: '#52525b', font: { family: 'JetBrains Mono', size: 10 } } },
          y: { grid: { color: 'rgba(255,255,255,0.03)' }, ticks: { color: '#52525b', font: { family: 'JetBrains Mono', size: 10 } } },
        },
      }}
    />
  );
}
