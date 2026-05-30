'use client';

import { useEffect, useState } from 'react';
import { fetchAPI, postAPI, patchAPI, putAPI, deleteAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState } from '../components/UI';
import { useNotification } from '../components/Notification';

export default function Alerts() {
  const [rules, setRules] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editRule, setEditRule] = useState<any>(null);
  const { addNotification } = useNotification();

  const loadRules = () => {
    fetchAPI('/alert-rules')
      .then(d => setRules(d.rules || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { loadRules(); }, []);

  const toggle = async (id: string) => {
    try {
      await patchAPI(`/alert-rules/${id}/toggle`);
      setRules(rules.map(r => r.id === id ? { ...r, enabled: !r.enabled } : r));
    } catch (e: any) { addNotification('error', '操作失败', e.message); }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除此告警规则？')) return;
    try {
      await deleteAPI(`/alert-rules/${id}`);
      setRules(rules.filter(r => r.id !== id));
      addNotification('success', '删除成功', '告警规则已删除');
    } catch (e: any) {
      addNotification('error', '删除失败', e.message);
    }
  };

  const handleSave = async (form: any) => {
    try {
      if (editRule) {
        await putAPI(`/alert-rules/${editRule.id}`, form);
        addNotification('success', '更新成功', '告警规则已更新');
      } else {
        await postAPI('/alert-rules', form);
        addNotification('success', '创建成功', '告警规则已创建');
      }
      setShowForm(false);
      setEditRule(null);
      loadRules();
    } catch (e: any) {
      addNotification('error', '保存失败', e.message);
    }
  };

  const enabled = rules.filter(r => r.enabled).length;
  const aiGenerated = rules.filter(r => r.is_ai).length;

  if (loading) return <LoadingState text="加载告警规则…" />;

  return (
    <>
      <div className="flex items-center justify-between">
        <div><h1 className="text-xl font-semibold text-white">告警规则</h1><p className="text-sm text-zinc-500 mt-0.5">配置阈值和通知渠道</p></div>
        <button onClick={() => { setEditRule(null); setShowForm(true); }}
          className="text-xs px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20 transition-colors">+ 新建规则</button>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">总规则数</span><p className="text-2xl font-bold text-white mt-1">{rules.length}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">已启用</span><p className="text-2xl font-bold text-success mt-1">{enabled}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">AI 生成</span><p className="text-2xl font-bold text-accent mt-1">{aiGenerated}</p></div>
        <div className="bg-surface-50 border border-white/5 rounded-xl p-4"><span className="font-mono text-xs text-zinc-500">今日触发</span><p className="text-2xl font-bold text-warn mt-1">-</p></div>
      </div>

      {rules.length === 0 ? <EmptyState title="暂无告警规则" description="点击上方按钮创建新规则" /> : (
        <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-left border-b border-white/5">
                {['规则', '条件', '阈值', '服务', '级别', '最后触发', '启用', '操作'].map(h => (
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
                      <button onClick={() => toggle(rule.id)} aria-label={rule.enabled ? '禁用规则' : '启用规则'} className={`w-9 h-5 rounded-full relative transition-colors focus-visible:ring-1 focus-visible:ring-accent/50 focus-visible:outline-none ${rule.enabled ? 'bg-[rgba(14,165,233,0.3)]' : 'bg-surface-300'}`}>
                        <div className={`w-4 h-4 rounded-full absolute top-0.5 transition-transform ${rule.enabled ? 'translate-x-4 bg-[#0ea5e9]' : 'translate-x-0.5 bg-zinc-500'}`} />
                      </button>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        <button onClick={() => { setEditRule(rule); setShowForm(true); }} className="text-xs text-zinc-500 hover:text-accent">编辑</button>
                        <button onClick={() => handleDelete(rule.id)} className="text-xs text-zinc-500 hover:text-danger">删除</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showForm && <RuleForm rule={editRule} onSave={handleSave} onClose={() => { setShowForm(false); setEditRule(null); }} />}
    </>
  );
}

function RuleForm({ rule, onSave, onClose }: { rule: any; onSave: (f: any) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    id: rule?.id || '',
    name: rule?.name || '',
    condition: rule?.condition || '',
    threshold: rule?.threshold || '',
    service: rule?.service || '',
    severity: rule?.severity || 'warning',
    enabled: rule?.enabled ?? true,
  });

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="bg-surface-50 border border-white/10 rounded-xl p-6 w-full max-w-md" onClick={e => e.stopPropagation()}>
        <h3 className="text-sm font-medium text-zinc-200 mb-4">{rule ? '编辑告警规则' : '新建告警规则'}</h3>
        <div className="space-y-3">
          {[
            { label: '名称', key: 'name', placeholder: 'CPU 使用率过高' },
            { label: '条件', key: 'condition', placeholder: 'cpu_usage > 85%' },
            { label: '阈值', key: 'threshold', placeholder: '85%' },
            { label: '服务', key: 'service', placeholder: 'api-gateway' },
          ].map(f => (
            <div key={f.key}>
              <label htmlFor={`rule-${f.key}`} className="font-mono text-[10px] text-zinc-500 mb-1 block">{f.label}</label>
              <input id={`rule-${f.key}`} value={(form as any)[f.key]} onChange={e => setForm({ ...form, [f.key]: e.target.value })} placeholder={f.placeholder + '…'}
                className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30" />
            </div>
          ))}
          <div>
            <label htmlFor="rule-severity" className="font-mono text-[10px] text-zinc-500 mb-1 block">级别</label>
            <select id="rule-severity" value={form.severity} onChange={e => setForm({ ...form, severity: e.target.value })}
              className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 rounded-lg px-3 py-2 outline-none focus:border-accent/30">
              <option value="critical">严重</option>
              <option value="warning">警告</option>
              <option value="info">信息</option>
            </select>
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-5">
          <button onClick={onClose} className="text-xs px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">取消</button>
          <button onClick={() => onSave(form)} className="text-xs px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20">保存</button>
        </div>
      </div>
    </div>
  );
}
