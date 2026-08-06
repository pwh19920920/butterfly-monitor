import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  ModalForm,
  PageContainer,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { Button, Modal, message, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useRef, useState } from 'react';
import {
  monitorDashboardCreate,
  monitorDashboardQuery,
  monitorDashboardTask,
  monitorDashboardUpdate,
} from '@/services/ant-design-pro/monitor.dashboard';
import SortModal from './components/SortModal';

const MonitorDashboardPage: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [current, setCurrent] = useState<API.MonitorDashboard>();
  const [sortDashboard, setSortDashboard] = useState<API.MonitorDashboard>();
  const [taskVisible, setTaskVisible] = useState(false);
  const [taskDashboard, setTaskDashboard] = useState<API.MonitorDashboard>();
  const [taskList, setTaskList] = useState<API.MonitorDashboardTask[]>([]);
  const [taskLoading, setTaskLoading] = useState(false);

  const openTasks = async (record: API.MonitorDashboard) => {
    if (record.id == null) return;
    setTaskDashboard(record);
    setTaskVisible(true);
    setTaskLoading(true);
    try {
      const res = await monitorDashboardTask(record.id);
      setTaskList(res.data || []);
    } catch {
      message.error('获取任务列表失败');
    } finally {
      setTaskLoading(false);
    }
  };

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
          key="task"
          onClick={() => {
            openTasks(record);
          }}
        >
          查看任务
        </a>,
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

  const taskColumns: ColumnsType<API.MonitorDashboardTask> = [
    { title: '序号', width: 60, render: (_text, _record, index) => index + 1 },
    { title: '任务名称', dataIndex: 'taskName' },
    { title: '排序', dataIndex: 'sort', width: 80 },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorDashboard>
        headerTitle="监控面板"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button
            type="primary"
            key="add"
            onClick={() => setCreateVisible(true)}
          >
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
        modalProps={{
          destroyOnClose: true,
          onCancel: () => setCreateVisible(false),
        }}
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
        modalProps={{
          destroyOnClose: true,
          onCancel: () => setModifyVisible(false),
        }}
        onFinish={async (values) => {
          try {
            await monitorDashboardUpdate({
              ...current,
              ...values,
            } as API.MonitorDashboard);
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
      <Modal
        title={`面板「${taskDashboard?.name}」关联任务`}
        open={taskVisible}
        footer={null}
        width={600}
        onCancel={() => setTaskVisible(false)}
      >
        <Table<API.MonitorDashboardTask>
          rowKey="id"
          loading={taskLoading}
          dataSource={taskList}
          columns={taskColumns}
          pagination={false}
        />
      </Modal>
    </PageContainer>
  );
};

export default MonitorDashboardPage;
