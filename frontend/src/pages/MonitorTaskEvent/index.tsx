import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProFormTextArea, ProTable } from '@ant-design/pro-components';
import { message, Modal, Space, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import {
  monitorTaskEventComplete,
  monitorTaskEventDeal,
  monitorTaskEventIgnore,
  monitorTaskEventQuery,
} from '@/services/ant-design-pro/monitor.event';
import { monitorTaskQuery } from '@/services/ant-design-pro/monitor.task';
import {
  CheckParamsLevelTypeEnum,
  MonitorTaskEventDealStatusEnum,
  MonitorTaskEventLevelEnum,
} from '@/services/ant-design-pro/enum';

const statusMap: Record<number, { text: string; color: string }> = {
  1: { text: MonitorTaskEventDealStatusEnum[1], color: 'orange' },
  2: { text: MonitorTaskEventDealStatusEnum[2], color: 'blue' },
  3: { text: MonitorTaskEventDealStatusEnum[3], color: 'green' },
  4: { text: MonitorTaskEventDealStatusEnum[4], color: 'default' },
};

const levelMap: Record<number, { text: string; color: string }> = {
  '-1': { text: MonitorTaskEventLevelEnum['-1'], color: 'default' },
  0: { text: MonitorTaskEventLevelEnum[0], color: 'red' },
  1: { text: MonitorTaskEventLevelEnum[1], color: 'orange' },
  2: { text: MonitorTaskEventLevelEnum[2], color: 'blue' },
  3: { text: MonitorTaskEventLevelEnum[3], color: 'gray' },
};

const MonitorTaskEventPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [dealVisible, setDealVisible] = useState(false);
  const [completeVisible, setCompleteVisible] = useState(false);
  const [current, setCurrent] = useState<API.MonitorTaskEvent>();

  const columns: ProColumns<API.MonitorTaskEvent>[] = [
    { title: '任务', dataIndex: 'taskName', search: false },
    {
      title: '任务',
      dataIndex: 'taskId',
      valueType: 'select',
      hideInTable: true,
      request: async () => {
        const res = await monitorTaskQuery({ current: 1, pageSize: 1000 });
        return (res.data || []).map((t) => ({ label: t.taskName, value: String(t.id) }));
      },
      fieldProps: { showSearch: true, allowClear: true, placeholder: '选择任务' },
    },
    { title: '告警信息', dataIndex: 'alertMsg', search: false, ellipsis: true },
    {
      title: '状态',
      dataIndex: 'dealStatus',
      valueType: 'select',
      valueEnum: MonitorTaskEventDealStatusEnum as any,
      render: (_, r) => {
        const s = statusMap[r.dealStatus || 0];
        return s ? <Tag color={s.color}>{s.text}</Tag> : '-';
      },
    },
    { title: '处理人', dataIndex: 'dealUserName', search: false },
    {
      title: '等级',
      dataIndex: 'eventLevel',
      valueType: 'select',
      valueEnum: CheckParamsLevelTypeEnum as any,
      fieldProps: { allowClear: true },
      render: (_, r) => {
        const l = levelMap[r.eventLevel as unknown as number];
        return l ? <Tag color={l.color}>{l.text}</Tag> : r.eventLevel != null ? String(r.eventLevel) : '-';
      },
    },
    { title: '告警次数', dataIndex: 'alertCount', search: false },
    { title: '下次告警', dataIndex: 'nextAlertTime', search: false, valueType: 'dateTime' },
    { title: '创建时间', dataIndex: 'createdAt', search: false, valueType: 'dateTime' },
    {
      title: '创建时间',
      dataIndex: 'dateTimeRange',
      valueType: 'dateTimeRange',
      hideInTable: true,
      search: { transform: (value) => ({ startTime: value[0], endTime: value[1] }) },
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => (
        <Space>
          {record.dealStatus === 1 && (
            <a
              onClick={() => {
                setCurrent(record);
                setDealVisible(true);
              }}
            >
              处理
            </a>
          )}
          {record.dealStatus === 1 && (
            <a
              onClick={() => {
                Modal.confirm({
                  title: '忽略告警',
                  content: '确认忽略该告警？忽略后将不再对该事件告警，同任务其他待处理事件不受影响。',
                  okText: '忽略',
                  okButtonProps: { danger: true },
                  cancelText: '取消',
                  onOk: async () => {
                    try {
                      await monitorTaskEventIgnore(record.id!);
                      message.success('已忽略');
                      actionRef.current?.reload();
                    } catch {
                      message.error('忽略失败');
                    }
                  },
                });
              }}
            >
              忽略
            </a>
          )}
          {record.dealStatus === 2 && (
            <a
              onClick={() => {
                setCurrent(record);
                setCompleteVisible(true);
              }}
            >
              完成
            </a>
          )}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorTaskEvent>
        headerTitle="异常事件"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async (params) => {
          const res = await monitorTaskEventQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />
      <ModalForm
        title="处理事件"
        open={dealVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setDealVisible(false) }}
        onFinish={async () => {
          try {
            await monitorTaskEventDeal(current!.id!);
            message.success('已处理');
            setDealVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('处理失败');
            return false;
          }
        }}
      >
        <div>确认处理事件？同任务其余待处理事件将标为误报。</div>
      </ModalForm>
      <ModalForm
        title="完成事件"
        open={completeVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCompleteVisible(false) }}
        onFinish={async (values) => {
          try {
            await monitorTaskEventComplete(current!.id!, { content: values.content });
            message.success('已完成');
            setCompleteVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('完成失败');
            return false;
          }
        }}
      >
        <ProFormTextArea name="content" label="事件经过" />
      </ModalForm>
    </PageContainer>
  );
};

export default MonitorTaskEventPage;
