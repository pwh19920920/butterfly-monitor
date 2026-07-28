import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProDescriptions, ProTable } from '@ant-design/pro-components';
import { Button, Drawer, message, Switch, Tag } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import dayjs from 'dayjs';
import CreateOrUpdateForm from '@/pages/MonitorTask/components/UpdateForm';
import {
  monitorTaskCreate,
  monitorTaskGetById,
  monitorTaskModifyAlertStatus,
  monitorTaskModifySampled,
  monitorTaskModifyTaskStatus,
  monitorTaskQuery,
  monitorTaskUpdate,
} from '@/services/ant-design-pro/monitor.task';
import { MonitorTaskAlertStatusEnum, TaskStatusEnum, TaskTypeEnum } from '@/services/ant-design-pro/enum';

// 生效时间统一格式化为 HH:mm:ss；兼容 dayjs / 字符串 / Date
const formatOneEffectTime = (v: any, fallback: string): string => {
  if (v == null || v === '') {
    return fallback;
  }
  if (dayjs.isDayjs(v)) {
    return v.isValid() ? v.format('HH:mm:ss') : fallback;
  }
  if (typeof v === 'string') {
    // 已是 HH:mm:ss / HH:mm
    if (/^\d{2}:\d{2}(:\d{2})?$/.test(v)) {
      return v.length === 5 ? `${v}:00` : v;
    }
    const d = dayjs(v, 'HH:mm:ss', true);
    if (d.isValid()) {
      return d.format('HH:mm:ss');
    }
    const d2 = dayjs(v);
    return d2.isValid() ? d2.format('HH:mm:ss') : fallback;
  }
  const d = dayjs(v);
  return d.isValid() ? d.format('HH:mm:ss') : fallback;
};

const formatEffectTimes = (checkParams: any[]) => {
  checkParams.forEach((cp) => {
    if (cp.effectTimes && cp.effectTimes.length === 2) {
      cp.effectTimes = [
        formatOneEffectTime(cp.effectTimes[0], '00:00:00'),
        formatOneEffectTime(cp.effectTimes[1], '23:59:59'),
      ];
    } else {
      // 未填时补全天，避免后端拿到空数组
      cp.effectTimes = ['00:00:00', '23:59:59'];
    }
  });
};

// 报警检查配置与异常检测规则均为非必填；
// 若填了任意一侧，则另一侧也必须完整填写（对称校验）
// 所选通道全为 Webhook 时不需要报警分组
// 返回错误提示，通过则返回 null
const validateAlertConfig = (values: any): string | null => {
  const alert = values?.taskAlert || {};
  const checkParams = alert.checkParams || [];

  const hasChannels = Array.isArray(alert.alertChannels) && alert.alertChannels.length > 0;
  const hasGroups = Array.isArray(alert.alertGroups) && alert.alertGroups.length > 0;
  const hasTimeSpan = Number(alert.timeSpan) > 0;
  const hasDuration = Number(alert.duration) > 0;
  // UpdateForm 根据通道类型写入；缺省按需要分组处理
  const needAlertGroups = alert.needAlertGroups !== false;
  // 报警检查配置：任一字段非空即视为「填了」
  const hasAnyCheckConfig = hasChannels || hasGroups || hasTimeSpan || hasDuration;
  // 完整：通道 + 间隔 + 持续；非 Webhook 场景还要分组
  const hasCompleteCheckConfig =
    hasChannels && hasTimeSpan && hasDuration && (!needAlertGroups || hasGroups);
  const hasCheckParams = checkParams.length > 0;

  // 两侧都空：允许（纯采集任务）
  if (!hasAnyCheckConfig && !hasCheckParams) {
    return null;
  }
  // 只填了报警检查配置
  if (hasAnyCheckConfig && !hasCheckParams) {
    return '已填写报警检查配置，必须补充异常检测规则';
  }
  // 只填了异常检测规则
  if (hasCheckParams && !hasAnyCheckConfig) {
    return needAlertGroups
      ? '已填写异常检测规则，必须补充报警检查配置（通道/分组/间隔/持续时间）'
      : '已填写异常检测规则，必须补充报警检查配置（通道/间隔/持续时间）';
  }
  // 两侧都有内容：报警检查配置必须齐全
  if (!hasCompleteCheckConfig) {
    return needAlertGroups
      ? '报警检查配置需完整填写：报警通道、报警分组、检查间隔、持续时间'
      : '报警检查配置需完整填写：报警通道、检查间隔、持续时间（Webhook 无需分组）';
  }
  // 每组规则至少一条
  for (const cp of checkParams) {
    if (!cp.rules || cp.rules.length === 0) {
      return '每个规则条件组至少需要一条规则';
    }
  }
  return null;
};

// labelParams(标签数组) → labels JSON；monitorGroups(多选) → monitorGroup 逗号串
const buildPayload = (values: any): API.MonitorTaskCreate => {
  const checkParams = values?.taskAlert?.checkParams || [];
  formatEffectTimes(checkParams);

  const labelParams = values.labelParams || [];
  const labels = labelParams.length
    ? JSON.stringify(labelParams.reduce((acc: any, cur: any) => ({ ...acc, [cur.label]: cur.value }), {}))
    : '';

  const monitorGroup = (values.monitorGroups || []).join(',');

  return {
    taskKey: values.taskKey,
    taskName: values.taskName,
    taskType: values.taskType,
    timeSpan: values.timeSpan,
    stepSpan: values.stepSpan,
    command: values.command,
    monitorGroup,
    labels,
    taskExecParams: values.taskExecParams || {},
    dashboards: values.dashboards || [],
    taskAlert: {
      duration: Number(values?.taskAlert?.duration) || 0,
      timeSpan: Number(values?.taskAlert?.timeSpan) || 0,
      alertChannels: values?.taskAlert?.alertChannels || [],
      // 纯 Webhook 通道不需要分组，提交时空数组
      alertGroups:
        values?.taskAlert?.needAlertGroups === false
          ? []
          : values?.taskAlert?.alertGroups || [],
      checkParams,
    },
  } as API.MonitorTaskCreate;
};

// labels JSON → labelParams 数组；monitorGroup 串 → monitorGroups 数组；effectTimes 串 → dayjs
const toFormValues = (row: API.MonitorTaskQueryResponse): any => {
  let labelParams: { label: string; value: string }[] = [];
  if (row.labels) {
    try {
      const obj = JSON.parse(row.labels);
      labelParams = Object.entries(obj).map(([label, value]) => ({ label, value: String(value) }));
    } catch {
      labelParams = [];
    }
  }
  const monitorGroups = row.monitorGroup ? row.monitorGroup.split(',').filter(Boolean) : [];
  const checkParams = (row.taskAlert?.checkParams || []).map((item) => {
    let effectTimes = [dayjs().startOf('day'), dayjs().endOf('day')];
    if (item.effectTimes && item.effectTimes.length === 2) {
      const start = dayjs(item.effectTimes[0], 'HH:mm:ss', true);
      const end = dayjs(item.effectTimes[1], 'HH:mm:ss', true);
      effectTimes = [
        start.isValid() ? start : dayjs().startOf('day'),
        end.isValid() ? end : dayjs().endOf('day'),
      ];
    }
    return {
      ...item,
      effectTimes,
    };
  });
  return {
    ...row,
    labelParams,
    monitorGroups,
    dashboards: row.dashboards || [],
    taskExecParams: row.taskExecParams || {},
    taskAlert: {
      ...(row.taskAlert || {}),
      alertChannels: row.taskAlert?.alertChannels || [],
      alertGroups: row.taskAlert?.alertGroups || [],
      checkParams,
    },
  };
};

const MonitorTaskPage: React.FC = () => {
  const actionRef = useRef<ActionType>(undefined);
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const [current, setCurrent] = useState<API.MonitorTaskQueryResponse>();
  const [editLoading, setEditLoading] = useState(false);

  const columns: ProColumns<API.MonitorTask>[] = [
    {
      title: '任务名称',
      dataIndex: 'taskName',
      ellipsis: true,
      width: 240,
      render: (dom, entity) => (
        <a
          onClick={() => {
            setCurrent(entity as any);
            setShowDetail(true);
          }}
        >
          {dom}
        </a>
      ),
    },
    { title: 'TaskKey', dataIndex: 'taskKey', copyable: true, ellipsis: true },
    {
      title: '任务类型',
      dataIndex: 'taskType',
      valueType: 'select',
      responsive: ['md'],
      valueEnum: TaskTypeEnum as any,
    },
    {
      title: '任务状态',
      dataIndex: 'taskStatus',
      search: false,
      valueEnum: TaskStatusEnum as any,
      render: (_, r) => (
        <Switch
          checked={r.taskStatus === 1}
          onChange={async (checked) => {
            await monitorTaskModifyTaskStatus(r.id!, checked ? 1 : 0);
            actionRef.current?.reload();
          }}
        />
      ),
    },
    {
      title: '告警开关',
      dataIndex: 'alertStatus',
      search: false,
      responsive: ['md'],
      render: (_, r) => (
        <Switch
          checked={r.alertStatus === 1}
          onChange={async (checked) => {
            try {
              await monitorTaskModifyAlertStatus(r.id!, checked ? 1 : 0);
              message.success(checked ? '告警已开启' : '告警已关闭');
              actionRef.current?.reload();
            } catch (e: any) {
              message.error(e?.info?.errorMessage || e?.message || '更新告警开关失败');
              actionRef.current?.reload();
            }
          }}
        />
      ),
    },
    {
      title: '告警状态',
      dataIndex: 'taskAlertStatus',
      search: false,
      width: 100,
      // 无 taskAlert 时后端固定返回 1(正常)
      valueEnum: {
        1: { text: MonitorTaskAlertStatusEnum[1], status: 'Success' },
        2: { text: MonitorTaskAlertStatusEnum[2], status: 'Warning' },
        3: { text: MonitorTaskAlertStatusEnum[3], status: 'Error' },
      },
      render: (_, r) => {
        const status = r.taskAlertStatus ?? 1;
        const color = status === 3 ? 'error' : status === 2 ? 'warning' : 'success';
        const text = MonitorTaskAlertStatusEnum[status] || '正常';
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '样本展示',
      dataIndex: 'sampled',
      search: false,
      responsive: ['md'],
      render: (_, r) => (
        <Switch
          checked={r.sampled === 1}
          onChange={async (checked) => {
            await monitorTaskModifySampled(r.id!, checked ? 1 : 0);
            actionRef.current?.reload();
          }}
        />
      ),
    },
    { title: '上次执行', dataIndex: 'preExecuteTime', search: false, width: 160, responsive: ['xxl'] },
    { title: '首次异常', dataIndex: 'firstFlagTime', search: false, width: 160, valueType: 'dateTime', responsive: ['xxl'] },
    { title: '采集错误', dataIndex: 'collectErrMsg', search: false, ellipsis: true, responsive: ['lg'] },
    {
      title: '操作',
      valueType: 'option',
      width: 80,
      render: (_, record) => (
        <a
          onClick={async () => {
            setEditLoading(true);
            try {
              const resp = await monitorTaskGetById(record.id!);
              if (resp.data) {
                setCurrent(resp.data);
                setModifyVisible(true);
              } else {
                message.error('任务详情加载失败');
              }
            } catch {
              message.error('任务详情加载失败');
            }
            setEditLoading(false);
          }}
        >
          {editLoading ? '加载中...' : '编辑'}
        </a>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorTask>
        headerTitle="监控任务"
        actionRef={actionRef}
        rowKey="id"
        scroll={{ x: 1400 }}
        columns={columns}
        toolBarRender={() => [
          <Button
            type="primary"
            key="add"
            onClick={() => {
              setCurrent(undefined);
              setCreateVisible(true);
            }}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await monitorTaskQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />

      <Drawer
        width={600}
        open={showDetail}
        onClose={() => {
          setCurrent(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {current?.taskName && (
          <ProDescriptions<API.MonitorTask>
            column={2}
            title={current?.taskName}
            request={async () => ({ data: current || {} })}
            params={{ id: current?.taskName }}
            columns={columns as any}
          />
        )}
      </Drawer>

      <ModalForm
        title="新建任务"
        width="1100px"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        onFinish={async (values: any) => {
          const err = validateAlertConfig(values);
          if (err) {
            message.warning(err);
            return false;
          }
          try {
            await monitorTaskCreate(buildPayload(values));
            message.success('创建成功');
            setCreateVisible(false);
            actionRef.current?.reload();
            return true;
          } catch (e: any) {
            message.error(e?.message || '创建失败');
            return false;
          }
        }}
      >
        <CreateOrUpdateForm taskType={-1} />
      </ModalForm>

      <ModalForm
        title="编辑任务"
        width="1100px"
        open={modifyVisible}
        initialValues={current ? toFormValues(current) : undefined}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        onFinish={async (values: any) => {
          const err = validateAlertConfig(values);
          if (err) {
            message.warning(err);
            return false;
          }
          try {
            await monitorTaskUpdate({ id: current?.id, ...buildPayload(values) } as any);
            message.success('更新成功');
            setModifyVisible(false);
            actionRef.current?.reload();
            return true;
          } catch (e: any) {
            message.error(e?.message || '更新失败');
            return false;
          }
        }}
      >
        <CreateOrUpdateForm taskType={current?.taskType ?? -1} taskId={current?.id ? String(current.id) : undefined} />
      </ModalForm>
    </PageContainer>
  );
};

export default MonitorTaskPage;
