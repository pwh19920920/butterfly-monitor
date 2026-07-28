import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import React, { useRef, useState } from 'react';
import {
  monitorDashboardCreate,
  monitorDashboardQuery,
  monitorDashboardUpdate,
} from '@/services/ant-design-pro/monitor.dashboard';
import SortModal from './components/SortModal';

const MonitorDashboardPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [current, setCurrent] = useState<API.MonitorDashboard>();
  const [sortDashboard, setSortDashboard] = useState<API.MonitorDashboard>();

  const columns: ProColumns<API.MonitorDashboard>[] = [
    { title: '名称', dataIndex: 'name' },
    { title: 'Slug', dataIndex: 'slug', search: false },
    {
      title: '地址',
      dataIndex: 'url',
      search: false,
      render: (_, r) =>
        r.url ? (
          <a href={r.url} target="_blank" rel="noreferrer">
            打开
          </a>
        ) : (
          '-'
        ),
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
        <a
          key="sort"
          onClick={() => {
            setSortDashboard(record);
          }}
        >
          排序
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorDashboard>
        headerTitle="监控面板"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await monitorDashboardQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />
      <ModalForm
        title="新建面板"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        onFinish={async (values) => {
          try {
            await monitorDashboardCreate(values as API.MonitorDashboard);
            message.success('创建成功');
            setCreateVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('创建失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
      </ModalForm>
      <ModalForm
        title="编辑面板"
        open={modifyVisible}
        initialValues={current}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        onFinish={async (values) => {
          try {
            await monitorDashboardUpdate({ ...current, ...values } as API.MonitorDashboard);
            message.success('更新成功');
            setModifyVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('更新失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
      </ModalForm>
      <SortModal
        dashboardId={sortDashboard?.id}
        dashboardName={sortDashboard?.name}
        onClose={() => setSortDashboard(undefined)}
      />
    </PageContainer>
  );
};

export default MonitorDashboardPage;
