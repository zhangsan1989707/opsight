'use client';

import { useEffect, useState } from 'react';
import { fetchAPI, postAPI, putAPI, deleteAPI } from '../lib/api';
import { Badge, LoadingState, EmptyState } from '../components/UI';
import { useNotification } from '../components/Notification';

export default function Notifications() {
  const [channels, setChannels] = useState<any[]>([]);
  const [history, setHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editChannel, setEditChannel] = useState<any>(null);
  const { addNotification } = useNotification();

  const loadChannels = () => {
    fetchAPI('/notifications/channels')
      .then(d => setChannels(d.channels || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  const loadHistory = () => {
    fetchAPI('/notifications/history?page_size=20')
      .then(d => setHistory(d.data || []))
      .catch(() => {});
  };

  useEffect(() => { loadChannels(); loadHistory(); }, []);

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此通知渠道？')) return;
    try {
      await deleteAPI(`/notifications/channels/${id}`);
      setChannels(channels.filter(c => c.id !== id));
      addNotification('success', '删除成功', '通知渠道已删除');
    } catch (e: any) {
      addNotification('error', '删除失败', e.message);
    }
  };

  const handleToggle = async (ch: any) => {
    try {
      await putAPI(`/notifications/channels/${ch.id}`, { ...ch, enabled: !ch.enabled });
      setChannels(channels.map(c => c.id === ch.id ? { ...c, enabled: !c.enabled } : c));
    } catch (e: any) {
      addNotification('error', '操作失败', e.message);
    }
  };

  const handleTest = async (id: number) => {
    try {
      await postAPI(`/notifications/test/${id}`);
      addNotification('success', '测试成功', '测试通知已发送');
    } catch (e: any) {
      addNotification('error', '测试失败', e.message);
    }
  };

  const handleSave = async (form: any) => {
    try {
      if (editChannel) {
        await putAPI(`/notifications/channels/${editChannel.id}`, form);
        addNotification('success', '更新成功', '通知渠道已更新');
      } else {
        await postAPI('/notifications/channels', form);
        addNotification('success', '创建成功', '通知渠道已创建');
      }
      setShowForm(false);
      setEditChannel(null);
      loadChannels();
    } catch (e: any) {
      addNotification('error', '保存失败', e.message);
    }
  };

  if (loading) return <LoadingState text="加载通知渠道…" />;

  return (
    <>
      <div className="flex items-center justify-between">
        <div><h1 className="text-xl font-semibold text-white">通知管理</h1><p className="text-sm text-zinc-500 mt-0.5">配置告警通知渠道和查看发送历史</p></div>
        <button onClick={() => { setEditChannel(null); setShowForm(true); }}
          className="text-xs px-3 py-1.5 rounded-lg bg-accent/10 text-accent hover:bg-accent/20 transition-colors">+ 新建渠道</button>
      </div>

      {channels.length === 0 ? <EmptyState title="暂无通知渠道" description="点击上方按钮创建通知渠道" /> : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {channels.map(ch => (
            <div key={ch.id} className="bg-surface-50 border border-white/5 rounded-xl p-5">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <h3 className="text-sm font-medium text-zinc-200">{ch.name}</h3>
                  <span className="font-mono text-[10px] text-zinc-600">
                    {ch.type === 'email' ? '邮件通知' : ch.type === 'wechat_work' ? '企业微信' : ch.type}
                  </span>
                </div>
                <button onClick={() => handleToggle(ch)} aria-label={ch.enabled ? '禁用渠道' : '启用渠道'} className={`w-9 h-5 rounded-full relative transition-colors focus-visible:ring-1 focus-visible:ring-accent/50 focus-visible:outline-none ${ch.enabled ? 'bg-[rgba(14,165,233,0.3)]' : 'bg-surface-300'}`}>
                  <div className={`w-4 h-4 rounded-full absolute top-0.5 transition-transform ${ch.enabled ? 'translate-x-4 bg-[#0ea5e9]' : 'translate-x-0.5 bg-zinc-500'}`} />
                </button>
              </div>
              <div className="flex items-center gap-2 mt-4 pt-3 border-t border-white/5">
                <button onClick={() => handleTest(ch.id)} className="text-xs text-zinc-500 hover:text-accent">测试发送</button>
                <button onClick={() => { setEditChannel(ch); setShowForm(true); }} className="text-xs text-zinc-500 hover:text-accent">编辑</button>
                <button onClick={() => handleDelete(ch.id)} className="text-xs text-zinc-500 hover:text-danger">删除</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Send History */}
      {history.length > 0 && (
        <div className="bg-surface-50 border border-white/5 rounded-xl overflow-hidden">
          <div className="p-4 border-b border-white/5"><h3 className="text-sm font-medium text-zinc-200">发送历史</h3></div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="text-left border-b border-white/5">
                {['渠道', '规则', '级别', '标题', '状态', '时间'].map(h => (
                  <th key={h} className="font-mono text-[10px] uppercase tracking-wider text-zinc-600 px-4 py-2.5">{h}</th>
                ))}
              </tr></thead>
              <tbody>
                {history.map((item: any) => (
                  <tr key={item.id} className="border-b border-white/[0.03]">
                    <td className="px-4 py-2 text-xs text-zinc-400">{item.channel_name}</td>
                    <td className="px-4 py-2 font-mono text-xs text-zinc-500">{item.alert_rule_id}</td>
                    <td className="px-4 py-2"><Badge variant={item.severity}>{item.severity}</Badge></td>
                    <td className="px-4 py-2 text-xs text-zinc-300 max-w-[200px] truncate">{item.title}</td>
                    <td className="px-4 py-2"><Badge variant={item.status === 'success' ? 'resolved' : 'critical'}>{item.status === 'success' ? '成功' : '失败'}</Badge></td>
                    <td className="px-4 py-2 font-mono text-[10px] text-zinc-600">{item.created_at}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showForm && <ChannelForm channel={editChannel} onSave={handleSave} onClose={() => { setShowForm(false); setEditChannel(null); }} />}
    </>
  );
}

function ChannelForm({ channel, onSave, onClose }: { channel: any; onSave: (f: any) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    name: channel?.name || '',
    type: channel?.type || 'email',
    config: channel?.config || '{}',
    enabled: channel?.enabled ?? true,
  });

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="bg-surface-50 border border-white/10 rounded-xl p-6 w-full max-w-md" onClick={e => e.stopPropagation()}>
        <h3 className="text-sm font-medium text-zinc-200 mb-4">{channel ? '编辑通知渠道' : '新建通知渠道'}</h3>
        <div className="space-y-3">
          <div>
            <label htmlFor="channel-name" className="font-mono text-[10px] text-zinc-500 mb-1 block">名称</label>
            <input id="channel-name" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="邮件告警通知…"
              className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30" />
          </div>
          <div>
            <label htmlFor="channel-type" className="font-mono text-[10px] text-zinc-500 mb-1 block">类型</label>
            <select id="channel-type" value={form.type} onChange={e => setForm({ ...form, type: e.target.value })}
              className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 rounded-lg px-3 py-2 outline-none focus:border-accent/30">
              <option value="email">邮件</option>
              <option value="wechat_work">企业微信</option>
            </select>
          </div>
          <div>
            <label htmlFor="channel-config" className="font-mono text-[10px] text-zinc-500 mb-1 block">
              配置 ({form.type === 'email' ? 'JSON: {"recipients":["a@b.com"]}' : 'JSON: {"webhook_url":"…"}'})
            </label>
            <textarea id="channel-config" value={form.config} onChange={e => setForm({ ...form, config: e.target.value })} rows={3}
              className="w-full bg-surface-100 border border-white/5 text-sm text-zinc-300 placeholder-zinc-600 rounded-lg px-3 py-2 outline-none focus:border-accent/30 font-mono text-xs" />
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
