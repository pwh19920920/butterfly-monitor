import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProFormSelect, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import {
  monitorGroupCreate,
  monitorGroupQuery,
  monitorGroupQueryAll,
  monitorGroupUpdate,
} from '@/services/ant-design-pro/monitor.group';

const MonitorGroupPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [current, setCurrent] = useState<API.MonitorGroup>();
  const [parents, setParents] = useState<{ label: string; value: string }[]>([]);
  const [groupMap, setGroupMap] = useState<Record<string, string>>({});

  const loadParents = () => {
    monitorGroupQueryAll().then((res) => {
      const list = res.data || [];
      const map: Record<string, string> = { '0': '根节点' };
      list.forEach((g) => {
        map[String(g.id)] = g.name || '';
      });
      setGroupMap(map);
      setParents([
        { label: '根节点', value: '0' },
        ...list.map((g) => ({ label: g.name || '', value: String(g.id) })),
      ]);
    });
  };

  useEffect(() => {
    loadParents();
  }, []);

  const columns: ProColumns<API.MonitorGroup>[] = [
    { title: '名称', dataIndex: 'name' },
    {
      title: '路由路径',
      dataIndex: 'route',
      search: false,
      render: (_, record) => {
        const route = record.route || '';
        // route 形如 /1/2/3/，按 / 拆分后逐段把 id 替换为名称
        const names = route
          .split('/')
          .filter(Boolean)
          .map((seg) => groupMap[seg] || seg);
        return names.length ? names.join(' / ') : '-';
      },
    },
    {
      title: '上级分组',
      dataIndex: 'parent',
      search: false,
      render: (_, record) => {
        if (record.parent === undefined || record.parent === null || record.parent === '') return '-';
        return groupMap[String(record.parent)] || String(record.parent);
      },
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="edit"
          onClick={() => {
            setCurrent(record);
            setModifyVisible(true);
          }}
        >
          编辑
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorGroup>
        headerTitle="监控分组"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await monitorGroupQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />
      <ModalForm
        title="新建分组"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        onFinish={async (values) => {
          try {
            await monitorGroupCreate(values as API.MonitorGroup);
            message.success('创建成功');
            setCreateVisible(false);
            loadParents();
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('创建失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect name="parent" label="上级分组" options={parents} initialValue="0" />
      </ModalForm>
      <ModalForm
        title="编辑分组"
        open={modifyVisible}
        initialValues={current}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        onFinish={async (values) => {
          try {
            await monitorGroupUpdate({ ...current, ...values } as API.MonitorGroup);
            message.success('更新成功');
            setModifyVisible(false);
            loadParents();
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('更新失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect name="parent" label="上级分组" options={parents} />
      </ModalForm>
    </PageContainer>
  );
};

export default MonitorGroupPage;
