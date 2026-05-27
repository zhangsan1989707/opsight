'use client';

import { useEffect, useState } from 'react';
import { fetchAPI } from '../lib/api';

const POSITIONS: Record<string, { x: number; y: number }> = {
  'api-gateway': { x: 450, y: 40 },
  'auth-service': { x: 200, y: 190 },
  'cache-layer': { x: 450, y: 190 },
  'payment-gateway': { x: 700, y: 190 },
  'user-service': { x: 115, y: 350 },
  'notification-svc': { x: 275, y: 350 },
  'order-service': { x: 385, y: 350 },
  'search-service': { x: 530, y: 350 },
  'analytics-svc': { x: 640, y: 350 },
  'event-pipeline': { x: 795, y: 350 },
  'cdn-edge': { x: 310, y: 470 },
  'service-mesh': { x: 590, y: 470 },
};

const COLORS: Record<string, string> = {
  healthy: '#10b981', degraded: '#f59e0b', down: '#ef4444',
};

export default function Topology() {
  const [nodes, setNodes] = useState<any[]>([]);
  const [selected, setSelected] = useState<any>(null);

  useEffect(() => {
    fetchAPI('/topology').then(d => setNodes(d.nodes || [])).catch(console.error);
  }, []);

  return (
    <>
      <div><h1 className="text-xl font-semibold text-white">Service Topology</h1><p className="text-sm text-zinc-500 mt-0.5">Dependency graph and traffic flow</p></div>

      <div className="bg-surface-50 border border-white/5 rounded-xl p-6 overflow-auto" style={{ minHeight: 560 }}>
        <svg width="100%" height="520" viewBox="0 0 900 520" className="w-full h-auto">
          {/* Connection lines */}
          {nodes.map(node => {
            const pos = POSITIONS[node.id];
            if (!pos) return null;
            return node.deps.map((dep: string) => {
              const depPos = POSITIONS[dep];
              if (!depPos) return null;
              const depNode = nodes.find(n => n.id === dep);
              const strokeColor = depNode?.status === 'down' ? 'rgba(239,68,68,0.25)' : depNode?.status === 'degraded' ? 'rgba(245,158,11,0.2)' : 'rgba(255,255,255,0.08)';
              return <line key={`${node.id}-${dep}`} x1={pos.x} y1={pos.y + 28} x2={depPos.x} y2={depPos.y - 28} stroke={strokeColor} strokeWidth="1.5" strokeDasharray="4 4" />;
            });
          })}

          {/* Nodes */}
          {nodes.map(node => {
            const pos = POSITIONS[node.id];
            if (!pos) return null;
            const color = COLORS[node.status] || '#71717a';
            return (
              <g key={node.id} className="cursor-pointer" onClick={() => setSelected(node)}>
                <rect x={pos.x - 70} y={pos.y - 28} width="140" height="56" rx="8" fill="#111114" stroke={color + '40'} strokeWidth="1" />
                <circle cx={pos.x - 46} cy={pos.y - 6} r="4" fill={color} />
                <text x={pos.x - 34} y={pos.y - 2} fill="#e4e4e7" fontFamily="Space Grotesk" fontSize="13" fontWeight="500">{node.id.length > 16 ? node.id.slice(0, 14) + '..' : node.id}</text>
                <text x={pos.x - 34} y={pos.y + 12} fill="#52525b" fontFamily="JetBrains Mono" fontSize="9">{node.rps} rps</text>
              </g>
            );
          })}
        </svg>
      </div>

      {selected && (
        <div className="bg-surface-50 border border-white/5 rounded-xl p-5">
          <h3 className="text-sm font-medium text-zinc-200 mb-3">{selected.id}</h3>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div><p className="font-mono text-[10px] text-zinc-600">Status</p><p className={`text-sm mt-0.5 ${selected.status === 'healthy' ? 'text-success' : selected.status === 'degraded' ? 'text-warn' : 'text-danger'}`}>{selected.status}</p></div>
            <div><p className="font-mono text-[10px] text-zinc-600">RPS</p><p className="font-mono text-sm text-zinc-300 mt-0.5">{selected.rps}</p></div>
            <div><p className="font-mono text-[10px] text-zinc-600">p99 Latency</p><p className="font-mono text-sm text-zinc-300 mt-0.5">{selected.p99}</p></div>
            <div><p className="font-mono text-[10px] text-zinc-600">Dependencies</p><p className="font-mono text-xs text-zinc-400 mt-0.5">{selected.deps?.length ? selected.deps.join(', ') : 'None'}</p></div>
          </div>
        </div>
      )}
    </>
  );
}
