import { LoadingOutlined } from '@ant-design/icons';
import {
  ProCard,
  ProForm,
  ProFormDigit,
  ProFormList,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProFormTimePicker,
} from '@ant-design/pro-components';
import { Button, Col, Divider, Form, Modal, message, Row, Spin } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { alertChannelQueryAll } from '@/services/ant-design-pro/alert.channel';
import { alertGroupQueryAll } from '@/services/ant-design-pro/alert.group';
import {
  CheckParamCompareTypeEnum,
  CheckParamRelationEnum,
  CheckParamsLevelTypeEnum,
  CheckParamValueTypeEnum,
  DataTypeEnum,
  TaskTypeEnum,
} from '@/services/ant-design-pro/enum';
import { monitorDashboardQueryAll } from '@/services/ant-design-pro/monitor.dashboard';
import { monitorDatabaseQueryAll } from '@/services/ant-design-pro/monitor.database';
import { monitorGroupQueryAll } from '@/services/ant-design-pro/monitor.group';
import {
  monitorTaskGetById,
  monitorTaskPreviewAggregate,
  monitorTaskQuery,
} from '@/services/ant-design-pro/monitor.task';

// 与后端 entity.AlertChannelTypeWebhook 一致
const ALERT_CHANNEL_TYPE_WEBHOOK = 2;
// 任务类型常量
const TASK_TYPE_DRILLDOWN = 4;
const TASK_TYPE_DATABASE = 1;
const TASK_TYPE_URL = 2;
const TASK_TYPE_PUSH = 3;
// 数据类型常量
const DATA_TYPE_NORMAL = 1;
const DATA_TYPE_AGGREGATE = 2;

type SelectItem = { label: string; value: number | string };

type CreateOrUpdateFormProps = {
  // -1 表示新建态
  taskType: number;
  taskId?: string;
  // 显式覆盖「创建视图」（可编辑 taskKey）。默认 taskType===-1。
  // 拷贝任务时 taskType 传被拷贝对象真实类型、isCreateView 传 true，
  // 保证 URL/下钻/聚合字段区正确显示且 taskKey 可改。
  isCreateView?: boolean;
};

const taskTypes: SelectItem[] = Object.keys(TaskTypeEnum).map((item) => ({
  value: Number(item),
  label: TaskTypeEnum[Number(item)],
}));
const dataTypes: SelectItem[] = Object.keys(DataTypeEnum).map((item) => ({
  value: Number(item),
  label: DataTypeEnum[Number(item)],
}));
const relations: SelectItem[] = Object.keys(CheckParamRelationEnum).map(
  (item) => ({
    value: Number(item),
    label: CheckParamRelationEnum[Number(item)],
  }),
);
const levelTypes: SelectItem[] = Object.keys(CheckParamsLevelTypeEnum).map(
  (item) => ({
    value: Number(item),
    label: CheckParamsLevelTypeEnum[Number(item)],
  }),
);
const compareTypes: SelectItem[] = Object.keys(CheckParamCompareTypeEnum).map(
  (item) => ({
    value: Number(item),
    label: CheckParamCompareTypeEnum[Number(item)],
  }),
);
const valueTypes: SelectItem[] = Object.keys(CheckParamValueTypeEnum).map(
  (item) => ({
    value: Number(item),
    label: CheckParamValueTypeEnum[Number(item)],
  }),
);
// 无值默认值下拉：仅正常查询 / 系统下钻有意义
const defaultValueOptions: SelectItem[] = [
  { value: 0, label: '0（数值类无值）' },
  { value: 100, label: '100（百分比类无值）' },
];

const CreateOrUpdateForm: React.FC<CreateOrUpdateFormProps> = (props) => {
  const isCreateView = props.isCreateView ?? props.taskType === -1;
  const form = Form.useFormInstance();
  const [selectTaskType, setSelectTaskType] = useState<number>(props.taskType);
  // 用 Form.useWatch 驱动，编辑态 initialValues 的 dataType 自动同步
  const dataType: number = Form.useWatch('dataType', form) ?? DATA_TYPE_NORMAL;
  const [databases, setDatabases] = useState<SelectItem[]>([]);
  const [databaseTypeMap, setDatabaseTypeMap] = useState<Map<string, number>>(
    new Map(),
  );
  const [databaseType, setDatabaseType] = useState<number>();
  const [dashboards, setDashboards] = useState<SelectItem[]>([]);
  const [alertChannels, setAlertChannels] = useState<SelectItem[]>([]);
  const [channelTypeMap, setChannelTypeMap] = useState<Map<string, number>>(
    new Map(),
  );
  const [alertGroups, setAlertGroups] = useState<SelectItem[]>([]);
  const [monitorGroups, setMonitorGroups] = useState<SelectItem[]>([]);
  const [needAlertGroups, setNeedAlertGroups] = useState(true);
  const selectedChannelIds = Form.useWatch(
    ['taskAlert', 'alertChannels'],
    form,
  );

  // 聚合预览
  const [columns, setColumns] = useState<string[]>([]);
  const [previewLoading, setPreviewLoading] = useState(false);
  const commandValue: string = Form.useWatch('command', form) || '';

  // 系统下钻
  const [aggregateTasks, setAggregateTasks] = useState<SelectItem[]>([]);
  const [_sourceLabelColumns, setSourceLabelColumns] = useState<string[]>([]);
  const [sourceValueColumns, setSourceValueColumns] = useState<string[]>([]);
  const [sourceLoading, setSourceLoading] = useState(false);

  // 关联任务：把其它任务的实时/样本曲线叠加到本任务面板
  const [relatedTasks, setRelatedTasks] = useState<SelectItem[]>([]);

  // 聚合任务不显示告警块
  const showAlertConfig =
    selectTaskType !== TASK_TYPE_DRILLDOWN
      ? !(
          selectTaskType === TASK_TYPE_DATABASE ||
          selectTaskType === TASK_TYPE_URL
        ) || dataType !== DATA_TYPE_AGGREGATE
      : true;

  useEffect(() => {
    monitorDatabaseQueryAll().then((resp) => {
      const data = (resp.data || []).map((item) => ({
        label: item.name || '',
        value: String(item.id),
      }));
      setDatabases(data);
      const map = new Map<string, number>();
      (resp.data || []).forEach((item) => {
        if (item.id != null && item.type != null)
          map.set(String(item.id), item.type);
      });
      setDatabaseTypeMap(map);
    });

    monitorDashboardQueryAll()
      .then((resp) => {
        setDashboards(
          (resp.data || []).map((item) => ({
            label: item.name || '',
            value: String(item.id),
          })),
        );
      })
      .catch(() => {
        Modal.info({
          title: '操作提示',
          content: '当前还没有添加任何面板，请先添加面板再进行添加任务',
        });
      });

    alertChannelQueryAll().then((resp) => {
      setAlertChannels(
        (resp.data || []).map((item) => ({
          label: item.name || '',
          value: String(item.id),
        })),
      );
      const typeMap = new Map<string, number>();
      (resp.data || []).forEach((item) => {
        if (item.id != null && item.type != null) {
          typeMap.set(String(item.id), item.type);
        }
      });
      setChannelTypeMap(typeMap);
    });

    alertGroupQueryAll().then((resp) => {
      setAlertGroups(
        (resp.data || []).map((item) => ({
          label: item.name || '',
          value: String(item.id),
        })),
      );
    });

    monitorGroupQueryAll().then((resp) => {
      const nameMap: Record<string, string> = {};
      (resp.data || []).forEach((item) => {
        nameMap[String(item.id)] = item.name || '';
      });
      setMonitorGroups(
        (resp.data || []).map((item) => {
          const ids = (item.route || '').split('/').filter(Boolean);
          const route = ids.map((id) => nameMap[id] || id).join('/');
          return {
            label: route ? `${item.name} - ${route}` : item.name || '',
            value: String(item.id),
          };
        }),
      );
    });

    // 下钻候选聚合任务
    monitorTaskQuery({ pageSize: 1000, taskType: TASK_TYPE_DATABASE } as any)
      .then((resp) => {
        const list = (resp.data || []).filter(
          (item: any) => item.dataType === DATA_TYPE_AGGREGATE,
        );
        setAggregateTasks(
          list.map((item: any) => ({
            label: item.taskName || item.taskKey,
            value: String(item.id),
          })),
        );
      })
      .catch(() => {});

    // 关联任务候选：排除聚合任务（聚合任务无单一实时 panel，不适合叠加曲线）及自身
    monitorTaskQuery({ pageSize: 1000 } as any)
      .then((resp) => {
        const list = (resp.data || []).filter(
          (item: any) => String(item.id) !== String(props.taskId) && item.dataType !== DATA_TYPE_AGGREGATE,
        );
        setRelatedTasks(
          list.map((item: any) => ({
            label: item.taskName || item.taskKey,
            value: String(item.id),
          })),
        );
      })
      .catch(() => {});
  }, []);

  // 已选通道全为 Webhook 时不需报警分组
  useEffect(() => {
    if (channelTypeMap.size === 0) return;
    const ids: string[] = (selectedChannelIds || []).map(
      (id: string | number) => String(id),
    );
    const need =
      ids.length === 0 ||
      ids.some((id) => channelTypeMap.get(id) !== ALERT_CHANNEL_TYPE_WEBHOOK);
    setNeedAlertGroups(need);
    form.setFieldValue(['taskAlert', 'needAlertGroups'], need);
    if (!need) form.setFieldValue(['taskAlert', 'alertGroups'], undefined);
  }, [selectedChannelIds, channelTypeMap, form]);

  // 任务类型变化：非 Database/URL → 强制 dataType=Normal
  useEffect(() => {
    if (
      selectTaskType === TASK_TYPE_DATABASE ||
      selectTaskType === TASK_TYPE_URL
    )
      return;
    form.setFieldValue('dataType', DATA_TYPE_NORMAL);
  }, [selectTaskType, form]);

  // 编辑态初始化 databaseType：initialValues 不触发下拉 onChange，需按已选 databaseId 从 map 反查（D-019）。
  // 否则编辑 Database 任务时 mongo 额外参数/提取字段不渲染，且预览 dataSource 为 undefined。
  const databaseId = Form.useWatch(['taskExecParams', 'databaseId'], form);
  useEffect(() => {
    if (databaseType === undefined && databaseId && databaseTypeMap.size > 0) {
      setDatabaseType(databaseTypeMap.get(String(databaseId)));
    }
  }, [databaseId, databaseTypeMap, databaseType]);

  // 编辑下钻任务：按已存的 sourceTaskId 加载源任务指标列，填充取数指标下拉（不改动 filters）
  const drilldownInitRef = useRef(false);
  useEffect(() => {
    if (drilldownInitRef.current) return;
    if (selectTaskType !== TASK_TYPE_DRILLDOWN || isCreateView) return;
    const sourceId = form.getFieldValue(['taskExecParams', 'sourceTaskId']);
    if (!sourceId) return;
    drilldownInitRef.current = true;
    monitorTaskGetById(String(sourceId))
      .then((resp) => {
        const ep = (resp.data as any)?.taskExecParams || {};
        setSourceLabelColumns(ep.labelColumns || []);
        setSourceValueColumns(ep.valueColumns || []);
      })
      .catch(() => {});
  }, [selectTaskType, isCreateView, form]);

  // 预览
  const handlePreview = async () => {
    const values = form.getFieldsValue(true);
    const execParams = values.taskExecParams || {};
    setPreviewLoading(true);
    try {
      const resp = await monitorTaskPreviewAggregate({
        taskType: selectTaskType,
        dataSource: databaseType,
        databaseId: execParams.databaseId,
        command: values.command,
        execParams,
      });
      const cols = resp.data?.columns || [];
      setColumns(cols);
      if (cols.length === 0)
        message.info('查询未返回任何列，请检查 SQL 或数据源');
    } catch (e: any) {
      message.error(
        `预览失败：${e?.info?.errorMessage || e?.message || '请稍后重试'}`,
      );
    } finally {
      setPreviewLoading(false);
    }
  };

  // 选源聚合任务 → 自动带出维度
  const handleSourceTaskChange = async (sourceId: string) => {
    if (!sourceId) {
      setSourceLabelColumns([]);
      setSourceValueColumns([]);
      return;
    }
    setSourceLoading(true);
    try {
      const resp = await monitorTaskGetById(sourceId);
      const src: any = resp.data || {};
      const ep = src.taskExecParams || {};
      const labels: string[] = ep.labelColumns || [];
      const values: string[] = ep.valueColumns || [];
      setSourceLabelColumns(labels);
      setSourceValueColumns(values);
      if (labels.length > 0) {
        // 仅当维度结构（fieldName 集合）变化时才重置 filters，保留用户已填的过滤值（D-021）
        const currentFilters: any[] =
          form.getFieldValue(['taskExecParams', 'filters']) || [];
        const sameStructure =
          currentFilters.length === labels.length &&
          labels.every((field, i) => currentFilters[i]?.fieldName === field);
        if (!sameStructure) {
          form.setFieldValue(
            ['taskExecParams', 'filters'],
            labels.map((field: string) => ({
              fieldName: field,
              operator: 'eq',
              value: '',
            })),
          );
        }
      }
      if (values.length > 0) {
        form.setFieldValue(['taskExecParams', 'queryMetric'], values[0]);
      }
    } catch {
      setSourceLabelColumns([]);
      setSourceValueColumns([]);
    } finally {
      setSourceLoading(false);
    }
  };

  const selectedLabelColumns: string[] =
    Form.useWatch(['taskExecParams', 'labelColumns'], form) || [];
  const selectedValueColumns: string[] =
    Form.useWatch(['taskExecParams', 'valueColumns'], form) || [];
  const allColumns: string[] = useMemo(() => {
    const set = new Set<string>(columns);
    selectedLabelColumns.forEach((c) => {
      set.add(c);
    });
    selectedValueColumns.forEach((c) => {
      set.add(c);
    });
    return Array.from(set);
  }, [columns, selectedLabelColumns, selectedValueColumns]);

  // 聚合任务回显：label/value 已有值但未预览（columns 空）时，用其初始化预览列集合，
  // 否则编辑/拷贝聚合任务会被「已选列却未预览」校验误拦截（该守卫只针对新建绕过预览）
  const aggregateColsKey = `${selectedLabelColumns.join(',')}|${selectedValueColumns.join(',')}`;
  useEffect(() => {
    if (columns.length === 0 && aggregateColsKey !== '|') {
      const labels: string[] =
        form.getFieldValue(['taskExecParams', 'labelColumns']) || [];
      const values: string[] =
        form.getFieldValue(['taskExecParams', 'valueColumns']) || [];
      setColumns(Array.from(new Set([...labels, ...values])));
    }
  }, [columns.length, aggregateColsKey, form]);

  if (dashboards.length === 0 && !isCreateView) {
    return (
      <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
    );
  }

  // ---------- render ----------

  return (
    <>
      <Divider>任务基础信息</Divider>
      <ProForm.Group>
        <ProFormText
          name="taskName"
          label="任务名称"
          width="md"
          placeholder="请输入任务名称"
          rules={[{ required: true, message: '任务名称不能为空' }]}
        />
        <ProFormText
          name="taskKey"
          label="任务key"
          width="md"
          placeholder="请输入任务key"
          disabled={!isCreateView}
          rules={[{ required: true, message: '任务key不能为空' }]}
        />
        {/* 聚合任务不需要归属面板 */}
        {dataType !== DATA_TYPE_AGGREGATE && (
          <ProFormSelect
            name="dashboards"
            label="归属面板"
            width="md"
            mode="multiple"
            options={dashboards}
            rules={[{ required: true, message: '归属面板不能为空' }]}
          />
        )}
                {/* 聚合任务不需要监控分组 */}
        {dataType !== DATA_TYPE_AGGREGATE && (
          <ProFormSelect
            name="monitorGroups"
            label="监控分组"
            width="md"
            mode="multiple"
            options={monitorGroups}
          />
        )}
        {/* 聚合任务不支持关联任务，不展示该字段 */}
        {dataType !== DATA_TYPE_AGGREGATE && (
          <ProFormSelect
            name="relatedTaskIds"
            label="关联任务"
            width="md"
            mode="multiple"
            showSearch
            options={relatedTasks}
            tooltip="把关联任务的实时/样本曲线叠加到本任务面板展示，图例为 任务名_实时 / 任务名_样本"
          />
        )}
        {/* 聚合任务无大促敏感概念，不展示；提交侧后端强制为否(1) */}
        {dataType !== DATA_TYPE_AGGREGATE && (
          <ProFormSelect
            name="promoSensitive"
            label="波动敏感"
            width="md"
            options={[
              { label: '否', value: 1 },
              { label: '是', value: 2 },
            ]}
            initialValue={1}
            tooltip="仅量级指标（QPS/订单量/支付笔数/UV等）适合开启。质量指标（成功率/错误率/核心RT等）严禁开启——大促不应成为成功率下降的借口。开启后：特殊日历史原料不进日常基线，格子写冻结基线，告警侧按 promoPeakRatio/promoTroughRatio 配置倍数放宽阈值。"
          />
        )}
        <ProFormDigit
          name="timeSpan"
          label="任务执行周期(秒)"
          width="md"
          min={30}
          initialValue={30}
          fieldProps={{ step: 30 }}
          placeholder="多久收集一次，30s 的倍数"
          rules={[
            { required: true, message: '任务执行周期不能为空' },
            {
              validator: (_: any, value: number) =>
                value % 30 === 0
                  ? Promise.resolve()
                  : Promise.reject(new Error('必须是 30s 的倍数')),
            },
          ]}
        />
        <ProFormDigit
          name="stepSpan"
          label="跨步间隔(秒)"
          width="md"
          min={30}
          initialValue={30}
          fieldProps={{ step: 30 }}
          placeholder="开始时间与当前时间间隔，30s 的倍数"
          rules={[
            { required: true, message: '跨步间隔不能为空' },
            {
              validator: (_: any, value: number) =>
                value % 30 === 0
                  ? Promise.resolve()
                  : Promise.reject(new Error('必须是 30s 的倍数')),
            },
          ]}
        />
        <ProFormSelect
          name="taskType"
          label="任务类型"
          width="md"
          options={taskTypes}
          rules={[{ required: true, message: '任务类型不能为空' }]}
          fieldProps={{
            onChange: (value: number) => {
              if (value === TASK_TYPE_DATABASE && databases.length === 0) {
                Modal.info({
                  title: '操作提示',
                  content:
                    '当前还没有添加任何数据源，请先添加数据源再进行添加任务',
                  onOk() {
                    location.href = '/monitor/database';
                  },
                });
              }
              setSelectTaskType(value);
            },
          }}
        />

        {/* Database 任务数据源选择 */}
        {selectTaskType === TASK_TYPE_DATABASE && (
          <ProFormSelect
            name={['taskExecParams', 'databaseId']}
            label="数据库"
            width="md"
            options={databases}
            fieldProps={{
              showSearch: true,
              onChange: (value: string) => {
                setDatabaseType(value ? databaseTypeMap.get(value) : -1);
              },
            }}
            rules={[{ required: true, message: '数据库不能为空' }]}
          />
        )}

        {/* 提取字段（http / mongo） */}
        {(selectTaskType === TASK_TYPE_URL ||
          (selectTaskType === TASK_TYPE_DATABASE && databaseType === 1)) && (
          <ProFormText
            name={['taskExecParams', 'resultFieldPath']}
            label="提取字段"
            width="md"
            placeholder="结果字段，支持 对象.属性"
            rules={[{ required: true, message: '提取字段不能为空' }]}
          />
        )}

        {/* mongo 额外参数 */}
        {selectTaskType === TASK_TYPE_DATABASE && databaseType === 1 && (
          <>
            <ProFormText
              name={['taskExecParams', 'collectName']}
              label="mongo集合名称"
              width="md"
              placeholder="mongo集合名称"
              rules={[{ required: true, message: 'mongo集合名称不能为空' }]}
            />
            <ProFormSelect
              name={['taskExecParams', 'defaultValue']}
              label="无结果默认值"
              width="md"
              options={defaultValueOptions}
              initialValue={0}
              rules={[{ required: true, message: '无结果默认值不能为空' }]}
              tooltip="查询无数据/无结果时回落该值：0=数值类 100=百分比类"
            />
          </>
        )}

        {/* DB/URL 任务：数据类型 */}
        {(selectTaskType === TASK_TYPE_DATABASE ||
          selectTaskType === TASK_TYPE_URL) && (
          <ProFormSelect
            name="dataType"
            label="数据类型"
            width="md"
            options={dataTypes}
            initialValue={DATA_TYPE_NORMAL}
            tooltip="正常查询：单值采集 + 采样 + 告警；聚合查询：分组多行写入时序库，只收集不告警"
          />
        )}

        {/* DB(非mongo)/URL 任务：无结果默认值（必填），聚合任务不适用 */}
        {(selectTaskType === TASK_TYPE_URL ||
          (selectTaskType === TASK_TYPE_DATABASE && databaseType !== 1)) &&
          dataType !== DATA_TYPE_AGGREGATE && (
            <ProFormSelect
              name={['taskExecParams', 'defaultValue']}
              label="无结果默认值"
              width="md"
              options={defaultValueOptions}
              initialValue={0}
              rules={[{ required: true, message: '无结果默认值不能为空' }]}
              tooltip="查询无数据/无结果时回落该值：0=数值类 100=百分比类"
            />
          )}
      </ProForm.Group>

      {/* 系统下钻 */}
      {selectTaskType === TASK_TYPE_DRILLDOWN && (
        <>
          <Divider>下钻配置</Divider>
          <ProForm.Group>
            <ProFormSelect
              name={['taskExecParams', 'sourceTaskId']}
              label="依赖的聚合任务"
              width="md"
              options={aggregateTasks}
              showSearch
              rules={[{ required: true, message: '请选择依赖的聚合任务' }]}
              fieldProps={{
                loading: sourceLoading,
                onChange: (value: string) => handleSourceTaskChange(value),
              }}
            />
            <ProFormSelect
              name={['taskExecParams', 'queryMetric']}
              label="取数指标(valueColumn)"
              width="md"
              options={sourceValueColumns.map((c) => ({ label: c, value: c }))}
              rules={[{ required: true, message: '请选择取数指标' }]}
            />
            <ProFormSelect
              name={['taskExecParams', 'defaultValue']}
              label="无值默认值"
              width="md"
              options={defaultValueOptions}
              initialValue={0}
              tooltip="点位不存在时返回该值：0=数值类 100=百分比类"
            />
          </ProForm.Group>
          <ProCard
            title="过滤维度（必须全部填写）"
            style={{ marginBottom: 8 }}
            boxShadow
          >
            <ProFormList
              name={['taskExecParams', 'filters']}
              copyIconProps={false}
              deleteIconProps={false}
              creatorButtonProps={false}
            >
              <Row gutter={23}>
                <Col span={8}>
                  <ProFormText
                    name="fieldName"
                    label="维度名"
                    width="md"
                    disabled
                  />
                </Col>
                <Col span={8}>
                  <ProFormText
                    name="operator"
                    label="匹配方式"
                    width="md"
                    disabled
                    initialValue="eq"
                  />
                </Col>
                <Col span={8}>
                  <ProFormText
                    name="value"
                    label="过滤值"
                    width="md"
                    placeholder="请输入过滤值"
                    rules={[{ required: true, message: '过滤值不能为空' }]}
                  />
                </Col>
              </Row>
            </ProFormList>
          </ProCard>
        </>
      )}

      {/* 执行指令：Push / 下钻 不需要 */}
      {selectTaskType !== TASK_TYPE_PUSH &&
        selectTaskType !== TASK_TYPE_DRILLDOWN && (
          <ProFormTextArea
            name="command"
            label="执行指令"
            rules={[{ required: true, message: '执行指令不能为空' }]}
            placeholder="示例: createTime >= '{{.startTime}}' and createTime < '{{.endTime}}'"
            tooltip={
              <>
                <p>startTime：endTime - 跨步间隔</p>
                <p>endTime：任务开始执行的时间</p>
                <p>startTimeMilli / endTimeMilli：时间戳格式</p>
              </>
            }
          />
        )}

      {/* 聚合查询：预览按钮在同一行右侧 */}
      {dataType === DATA_TYPE_AGGREGATE && (
        <>
          <Divider>聚合配置</Divider>
          <ProCard style={{ marginBottom: 8 }} boxShadow>
            <Row gutter={16}>
              <Col span={20}>
                <ProFormSelect
                  name={['taskExecParams', 'labelColumns']}
                  label="labelColumns"
                  fieldProps={{
                    mode: 'multiple',
                    style: { width: '100%' },
                    onChange: () =>
                      form.validateFields([['taskExecParams', 'valueColumns']]),
                  }}
                  options={allColumns
                    .filter((c) => !selectedValueColumns.includes(c))
                    .map((c) => ({ label: c, value: c }))}
                  dependencies={[['taskExecParams', 'valueColumns']]}
                  rules={[
                    { required: true, message: '请至少选择一个 label' },
                    {
                      validator: () => {
                        const labels: string[] =
                          form.getFieldValue([
                            'taskExecParams',
                            'labelColumns',
                          ]) || [];
                        const values: string[] =
                          form.getFieldValue([
                            'taskExecParams',
                            'valueColumns',
                          ]) || [];
                        // 已选列却未预览（columns 为空）时拦截，避免绕过互斥校验提交重叠列（D-020）
                        if (
                          labels.length + values.length > 0 &&
                          columns.length === 0
                        ) {
                          return Promise.reject(
                            new Error(
                              '请先点击「预览选列」获取列，再分配 label/value',
                            ),
                          );
                        }
                        if (
                          allColumns.length > 0 &&
                          labels.length + values.length !== allColumns.length
                        ) {
                          return Promise.reject(
                            new Error(
                              '预览列必须全部分配到 label/value，每列只能归属其一',
                            ),
                          );
                        }
                        return Promise.resolve();
                      },
                    },
                  ]}
                />
                <ProFormSelect
                  name={['taskExecParams', 'valueColumns']}
                  label="valueColumns"
                  fieldProps={{
                    mode: 'multiple',
                    style: { width: '100%' },
                    onChange: () =>
                      form.validateFields([['taskExecParams', 'labelColumns']]),
                  }}
                  options={allColumns
                    .filter((c) => !selectedLabelColumns.includes(c))
                    .map((c) => ({ label: c, value: c }))}
                  dependencies={[['taskExecParams', 'labelColumns']]}
                  rules={[
                    { required: true, message: '请至少选择一个 value' },
                    {
                      validator: () => {
                        const labels: string[] =
                          form.getFieldValue([
                            'taskExecParams',
                            'labelColumns',
                          ]) || [];
                        const values: string[] =
                          form.getFieldValue([
                            'taskExecParams',
                            'valueColumns',
                          ]) || [];
                        // 已选列却未预览（columns 为空）时拦截，避免绕过互斥校验提交重叠列（D-020）
                        if (
                          labels.length + values.length > 0 &&
                          columns.length === 0
                        ) {
                          return Promise.reject(
                            new Error(
                              '请先点击「预览选列」获取列，再分配 label/value',
                            ),
                          );
                        }
                        if (
                          allColumns.length > 0 &&
                          labels.length + values.length !== allColumns.length
                        ) {
                          return Promise.reject(
                            new Error(
                              '预览列必须全部分配到 label/value，每列只能归属其一',
                            ),
                          );
                        }
                        return Promise.resolve();
                      },
                    },
                  ]}
                />
              </Col>
              <Col span={4} style={{ paddingTop: 30, textAlign: 'right' }}>
                <Button
                  type="primary"
                  loading={previewLoading}
                  disabled={!commandValue.trim()}
                  onClick={handlePreview}
                >
                  预览选列
                </Button>
                <div style={{ color: '#999', fontSize: 12, marginTop: 8 }}>
                  先预览取列，再勾选
                </div>
              </Col>
            </Row>
          </ProCard>
        </>
      )}

      <Divider>任务标签</Divider>
      <ProCard title="任务标签" style={{ marginBottom: 8 }}>
        <ProFormList
          name="labelParams"
          copyIconProps={false}
          deleteIconProps={{ tooltipText: '不需要这行了' }}
        >
          <Row gutter={23}>
            <Col span={12}>
              <ProFormText
                name="label"
                label="标签名"
                width="md"
                placeholder="请输入标签名"
                rules={[{ required: true, message: '标签名不能为空' }]}
              />
            </Col>
            <Col span={12}>
              <ProFormText
                name="value"
                label="标签值"
                width="md"
                placeholder="请输入标签值"
                rules={[{ required: true, message: '标签值不能为空' }]}
              />
            </Col>
          </Row>
        </ProFormList>
      </ProCard>

      {/* 报警配置：聚合任务不展示 */}
      {showAlertConfig && (
        <>
          <Divider>报警配置（选填）</Divider>
          <ProCard title="报警检查配置" style={{ marginBottom: 8 }} boxShadow>
            <Row gutter={23}>
              <Col span={8}>
                <ProFormSelect
                  name={['taskAlert', 'alertChannels']}
                  label="报警通道"
                  width="md"
                  showSearch
                  options={alertChannels}
                  fieldProps={{ mode: 'multiple' }}
                  tooltip="若所选通道均为 Webhook 则无需再选报警分组"
                />
              </Col>
              {needAlertGroups && (
                <Col span={8}>
                  <ProFormSelect
                    name={['taskAlert', 'alertGroups']}
                    label="报警分组"
                    width="md"
                    showSearch
                    options={alertGroups}
                    fieldProps={{ mode: 'multiple' }}
                  />
                </Col>
              )}
              <Col span={8}>
                <ProFormDigit
                  name={['taskAlert', 'timeSpan']}
                  label="检查间隔(秒)"
                  width="md"
                  placeholder="请输入检查间隔"
                  min={1}
                />
              </Col>
              <Col span={8}>
                <ProFormDigit
                  name={['taskAlert', 'duration']}
                  label="持续时间(秒)"
                  width="md"
                  placeholder="请输入持续时间"
                  min={1}
                />
              </Col>
            </Row>
          </ProCard>
          <ProCard title="异常检测规则" boxShadow>
            <ProFormList
              name={['taskAlert', 'checkParams']}
              creatorRecord={{
                relation: 1,
                effectTimes: [dayjs().startOf('day'), dayjs().endOf('day')],
                level: 2,
                rules: [],
              }}
              itemRender={({ listDom, action }) => (
                <ProCard
                  extra={action}
                  title="规则条件组"
                  style={{ marginBottom: 8 }}
                  type="inner"
                >
                  {listDom}
                </ProCard>
              )}
            >
              <Row gutter={23} style={{ marginRight: 14 }}>
                <Col span={8}>
                  <ProFormSelect
                    name="relation"
                    label="组内条件关系"
                    width="md"
                    showSearch
                    options={relations}
                    initialValue={1}
                    extra="组与组之间固定为「或者」"
                    rules={[{ required: true, message: '条件关系不能为空' }]}
                  />
                </Col>
                <Col span={8}>
                  <ProFormTimePicker.RangePicker
                    name="effectTimes"
                    label="生效时间"
                    width="md"
                    initialValue={[
                      dayjs().startOf('day'),
                      dayjs().endOf('day'),
                    ]}
                    rules={[{ required: true, message: '生效时间不能为空' }]}
                    fieldProps={{ format: 'HH:mm:ss' }}
                  />
                </Col>
                <Col span={8}>
                  <ProFormSelect
                    name="level"
                    label="告警等级"
                    width="md"
                    showSearch
                    options={levelTypes}
                    initialValue={2}
                    rules={[{ required: true, message: '告警等级不能为空' }]}
                  />
                </Col>
              </Row>
              <ProFormList
                name="rules"
                copyIconProps={false}
                deleteIconProps={{ tooltipText: '不需要这行了' }}
              >
                <Row gutter={23}>
                  <Col span={8}>
                    <ProFormSelect
                      name="compareType"
                      label="比较类型"
                      width="md"
                      showSearch
                      options={compareTypes}
                      rules={[{ required: true, message: '比较类型不能为空' }]}
                    />
                  </Col>
                  <Col span={8}>
                    <ProFormDigit
                      name="value"
                      label="比较值"
                      width="md"
                      placeholder="请输入比较值"
                      rules={[{ required: true, message: '比较值不能为空' }]}
                    />
                  </Col>
                  <Col span={8}>
                    <ProFormSelect
                      name="valueType"
                      label="值类型"
                      width="md"
                      showSearch
                      options={valueTypes}
                      rules={[{ required: true, message: '值类型不能为空' }]}
                    />
                  </Col>
                </Row>
              </ProFormList>
            </ProFormList>
          </ProCard>
        </>
      )}
    </>
  );
};

export default CreateOrUpdateForm;
