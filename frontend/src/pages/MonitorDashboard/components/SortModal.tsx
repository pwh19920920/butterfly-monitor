import { HolderOutlined } from '@ant-design/icons';
import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Button, List, Modal, message, Spin } from 'antd';
import React, { useEffect, useState } from 'react';
import {
  monitorDashboardTask,
  monitorDashboardTaskSort,
} from '@/services/ant-design-pro/monitor.dashboard';

interface SortModalProps {
  // 面板 id；为空时弹窗关闭
  dashboardId?: string;
  dashboardName?: string;
  onClose: () => void;
}

// 单行可拖拽项
const SortableRow: React.FC<{
  item: API.MonitorDashboardTask;
  index: number;
}> = ({ item, index }) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: String(item.id ?? index),
  });
  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    cursor: 'move',
    opacity: isDragging ? 0.5 : 1,
    background: isDragging ? '#f0f5ff' : undefined,
  };
  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <List.Item>
        <HolderOutlined style={{ color: '#999', marginRight: 8 }} />
        <span style={{ color: '#999', marginRight: 8 }}>{index + 1}.</span>
        {item.taskName || `任务#${item.taskId}`}
      </List.Item>
    </div>
  );
};

const SortModal: React.FC<SortModalProps> = ({
  dashboardId,
  dashboardName,
  onClose,
}) => {
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [rows, setRows] = useState<API.MonitorDashboardTask[]>([]);

  // 拖拽传感器：鼠标 + 键盘（无障碍），鼠标移动阈值避免误触
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  useEffect(() => {
    if (!dashboardId) {
      return;
    }
    setLoading(true);
    monitorDashboardTask(dashboardId)
      .then((res) => {
        // 后端已按 sort desc 返回，直接作为展示顺序
        setRows(res.data || []);
      })
      .catch(() => {
        message.error('加载面板任务失败');
      })
      .finally(() => setLoading(false));
  }, [dashboardId]);

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }
    setRows((prev) => {
      const oldIndex = prev.findIndex(
        (r) => String(r.id) === String(active.id),
      );
      const newIndex = prev.findIndex((r) => String(r.id) === String(over.id));
      if (oldIndex < 0 || newIndex < 0) {
        return prev;
      }
      return arrayMove(prev, oldIndex, newIndex);
    });
  };

  const handleSave = async () => {
    // 当前 rows 从上到下为期望展示顺序；后端按 sort desc 展示
    // → 越靠前 sort 越大：第一项 sort = rows.length，逐项递减
    const items: { id: string; sort: number }[] = [];
    for (let i = 0; i < rows.length; i++) {
      const r = rows[i];
      if (r.id == null) {
        message.warning('存在无效的排序项');
        return;
      }
      items.push({ id: r.id, sort: rows.length - i });
    }
    setSaving(true);
    try {
      await monitorDashboardTaskSort(items);
      message.success('排序已保存');
      onClose();
    } catch {
      message.error('保存排序失败');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal
      title={`排序面板${dashboardName ? ` - ${dashboardName}` : ''}`}
      open={!!dashboardId}
      onCancel={onClose}
      width={520}
      footer={[
        <Button key="cancel" onClick={onClose}>
          取消
        </Button>,
        <Button key="save" type="primary" loading={saving} onClick={handleSave}>
          保存
        </Button>,
      ]}
    >
      <Spin spinning={loading}>
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={rows.map((r, idx) => String(r.id ?? idx))}
            strategy={verticalListSortingStrategy}
          >
            <List
              size="small"
              dataSource={rows}
              locale={{ emptyText: '该面板暂无任务' }}
              renderItem={(item, idx) => (
                <SortableRow
                  key={String(item.id ?? idx)}
                  item={item}
                  index={idx}
                />
              )}
            />
          </SortableContext>
        </DndContext>
      </Spin>
    </Modal>
  );
};

export default SortModal;
