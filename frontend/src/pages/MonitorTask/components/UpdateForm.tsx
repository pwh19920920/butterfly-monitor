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
import { Col, Divider, Form, Modal, Row, Spin } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useState } from 'react';
import { monitorDatabaseQueryAll } from '@/services/ant-design-pro/monitor.database';
import { monitorDashboardQueryAll } from '@/services/ant-design-pro/monitor.dashboard';
import { alertChannelQueryAll } from '@/services/ant-design-pro/alert.channel';
import { alertGroupQueryAll } from '@/services/ant-design-pro/alert.group';
import { monitorGroupQueryAll } from '@/services/ant-design-pro/monitor.group';
import {
  CheckParamCompareTypeEnum,
  CheckParamRelationEnum,
  CheckParamsLevelTypeEnum,
  CheckParamValueTypeEnum,
  TaskTypeEnum,
} from '@/services/ant-design-pro/enum';

// 与后端 entity.AlertChannelTypeWebhook 一致
const ALERT_CHANNEL_TYPE_WEBHOOK = 2;

type SelectItem = { label: string; value: number | string };

type CreateOrUpdateFormProps = {
  // -1 表示新建态
  taskType: number;
  taskId?: string;
};

const taskTypes: SelectItem[] = Object.keys(TaskTypeEnum).map((item) => ({
  value: Number(item),
  label: TaskTypeEnum[Number(item)],
}));
const relations: SelectItem[] = Object.keys(CheckParamRelationEnum).map((item) => ({
  value: Number(item),
  label: CheckParamRelationEnum[Number(item)],
}));
const levelTypes: SelectItem[] = Object.keys(CheckParamsLevelTypeEnum).map((item) => ({
  value: Number(item),
  label: CheckParamsLevelTypeEnum[Number(item)],
}));
const compareTypes: SelectItem[] = Object.keys(CheckParamCompareTypeEnum).map((item) => ({
  value: Number(item),
  label: CheckParamCompareTypeEnum[Number(item)],
}));
const valueTypes: SelectItem[] = Object.keys(CheckParamValueTypeEnum).map((item) => ({
  value: Number(item),
  label: CheckParamValueTypeEnum[Number(item)],
}));

const CreateOrUpdateForm: React.FC<CreateOrUpdateFormProps> = (props) => {
  const isCreateView = props.taskType === -1;
  const form = Form.useFormInstance();
  const [selectTaskType, setSelectTaskType] = useState<number>(props.taskType);
  const [databases, setDatabases] = useState<SelectItem[]>([]);
  const [databaseTypeMap, setDatabaseTypeMap] = useState<Map<string, number>>(new Map());
  const [databaseType, setDatabaseType] = useState<number>();
  const [dashboards, setDashboards] = useState<SelectItem[]>([]);
  const [alertChannels, setAlertChannels] = useState<SelectItem[]>([]);
  // 通道 id → 类型（1邮件 2Webhook 3短信）
  const [channelTypeMap, setChannelTypeMap] = useState<Map<string, number>>(new Map());
  const [alertGroups, setAlertGroups] = useState<SelectItem[]>([]);
  const [monitorGroups, setMonitorGroups] = useState<SelectItem[]>([]);
  // 已选通道是否需要报警分组：全 Webhook 时不需要
  const [needAlertGroups, setNeedAlertGroups] = useState(true);
  const selectedChannelIds = Form.useWatch(['taskAlert', 'alertChannels'], form);

  useEffect(() => {
    // 数据源
    monitorDatabaseQueryAll().then((resp) => {
      const data = (resp.data || []).map((item) => ({
        label: item.name || '',
        value: String(item.id),
      }));
      setDatabases(data);
      const map = new Map<string, number>();
      (resp.data || []).forEach((item) => {
        if (item.id != null && item.type != null) map.set(String(item.id), item.type);
      });
      setDatabaseTypeMap(map);
    });

    // 面板
    monitorDashboardQueryAll()
      .then((resp) => {
        setDashboards(
          (resp.data || []).map((item) => ({ label: item.name || '', value: String(item.id) })),
        );
      })
      .catch(() => {
        Modal.info({
          title: '操作提示',
          content: '当前还没有添加任何面板，请先添加面板再进行添加任务',
        });
      });

    // 报警通道（同时缓存类型，用于判断是否需要报警分组）
    alertChannelQueryAll().then((resp) => {
      setAlertChannels(
        (resp.data || []).map((item) => ({ label: item.name || '', value: String(item.id) })),
      );
      const typeMap = new Map<string, number>();
      (resp.data || []).forEach((item) => {
        if (item.id != null && item.type != null) {
          typeMap.set(String(item.id), item.type);
        }
      });
      setChannelTypeMap(typeMap);
    });

    // 报警组
    alertGroupQueryAll().then((resp) => {
      setAlertGroups(
        (resp.data || []).map((item) => ({ label: item.name || '', value: String(item.id) })),
      );
    });

    // 监控分组
    monitorGroupQueryAll().then((resp) => {
      const nameMap: Record<string, string> = {};
      (resp.data || []).forEach((item) => {
        nameMap[String(item.id)] = item.name || '';
      });
      setMonitorGroups(
        (resp.data || []).map((item) => {
          const ids = (item.route || '').split('/').filter(Boolean);
          const route = ids.map((id) => nameMap[id] || id).join('/');
          return { label: route ? `${item.name} - ${route}` : item.name || '', value: String(item.id) };
        }),
      );
    });
  }, []);

  // 已选通道全为 Webhook 时不需要报警分组，并清空已选分组
  useEffect(() => {
    if (channelTypeMap.size === 0) {
      return;
    }
    const ids: string[] = (selectedChannelIds || []).map((id: string | number) => String(id));
    // 未选通道时默认仍展示分组字段；全 Webhook 才隐藏
    const need =
      ids.length === 0 ||
      ids.some((id) => channelTypeMap.get(id) !== ALERT_CHANNEL_TYPE_WEBHOOK);
    setNeedAlertGroups(need);
    // 写入表单供提交校验读取（非接口字段，buildPayload 会剥离）
    form.setFieldValue(['taskAlert', 'needAlertGroups'], need);
    if (!need) {
      form.setFieldValue(['taskAlert', 'alertGroups'], undefined);
    }
  }, [selectedChannelIds, channelTypeMap, form]);

  if (dashboards.length === 0 && !isCreateView) {
    return <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />;
  }

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
        <ProFormSelect
          name="dashboards"
          label="归属面板"
          width="md"
          mode="multiple"
          options={dashboards}
          rules={[{ required: true, message: '归属面板不能为空' }]}
        />
        <ProFormSelect
          name="monitorGroups"
          label="监控分组"
          width="md"
          mode="multiple"
          options={monitorGroups}
        />
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
                value % 30 === 0 ? Promise.resolve() : Promise.reject(new Error('必须是 30s 的倍数')),
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
                value % 30 === 0 ? Promise.resolve() : Promise.reject(new Error('必须是 30s 的倍数')),
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
              if (value === 1 && databases.length === 0) {
                Modal.info({
                  title: '操作提示',
                  content: '当前还没有添加任何数据源，请先添加数据源再进行添加任务',
                  onOk() {
                    location.href = '/monitor/database';
                  },
                });
              }
              setSelectTaskType(value);
            },
          }}
        />

        {/* 数据库任务：选择数据源 */}
        {selectTaskType === 1 && (
          <ProFormSelect
            name={['taskExecParams', 'databaseId']}
            label="数据库"
            width="md"
            options={databases}
            fieldProps={{
              showSearch: true,
              onChange: (value: string) => {
                if (!value) {
                  setDatabaseType(-1);
                  return;
                }
                setDatabaseType(databaseTypeMap.get(value));
              },
            }}
            rules={[{ required: true, message: '数据库不能为空' }]}
          />
        )}

        {/* http(2) 或 mongo(1) 需要提取字段 */}
        {(selectTaskType === 2 || (selectTaskType === 1 && databaseType === 1)) && (
          <ProFormText
            name={['taskExecParams', 'resultFieldPath']}
            label="提取字段"
            width="md"
            placeholder="结果字段，支持 对象.属性"
            rules={[{ required: true, message: '提取字段不能为空' }]}
          />
        )}

        {/* mongo 额外参数 */}
        {selectTaskType === 1 && databaseType === 1 && (
          <>
            <ProFormText
              name={['taskExecParams', 'collectName']}
              label="mongo集合名称"
              width="md"
              placeholder="mongo集合名称"
              rules={[{ required: true, message: 'mongo集合名称不能为空' }]}
            />
            <ProFormDigit
              name={['taskExecParams', 'defaultValue']}
              label="无结果默认值"
              width="md"
              min={0}
              placeholder="查询结果为比例时填100，数值填0"
              rules={[{ required: true, message: '无结果默认值不能为空' }]}
            />
          </>
        )}
      </ProForm.Group>

      {selectTaskType !== -1 && selectTaskType !== 3 && (
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

      {/* 报警检查配置与异常检测规则合为一组（均选填；填其一则另一侧也必须填） */}
      <Divider>报警配置（选填）</Divider>
      <ProCard title="报警检查配置" style={{ marginBottom: 8 }} bordered>
        {/* 三列一行，与异常检测规则布局一致 */}
        <Row gutter={23}>
          <Col span={8}>
            <ProFormSelect
              name={['taskAlert', 'alertChannels']}
              label="报警通道"
              width="md"
              showSearch
              options={alertChannels}
              fieldProps={{ mode: 'multiple' }}
              tooltip="若所选通道均为 Webhook，无需再选报警分组"
            />
          </Col>
          {/* 邮件/短信等需要接收人；纯 Webhook 通道不依赖报警分组 */}
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
      <ProCard title="异常检测规则" bordered>
        <ProFormList
          name={['taskAlert', 'checkParams']}
          // 新增规则组时写入真实表单值（defaultValue 只影响展示，不进 Form）
          creatorRecord={{
            relation: 1,
            effectTimes: [dayjs().startOf('day'), dayjs().endOf('day')],
            // 与 CheckParamsLevelTypeEnum 一致：0严重 1高 2中 3低（无 -1）
            level: 2,
            rules: [],
          }}
          itemRender={({ listDom, action }) => (
            <ProCard extra={action} title="规则条件组" style={{ marginBottom: 8 }} type="inner">
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
                extra="组与组之间固定为「或者」；此处仅控制本组内规则的或/且"
                rules={[{ required: true, message: '条件关系不能为空' }]}
              />
            </Col>
            <Col span={8}>
              <ProFormTimePicker.RangePicker
                name="effectTimes"
                label="生效时间"
                width="md"
                // initialValue 才会写入 Form；fieldProps.defaultValue 只是 UI 占位
                initialValue={[dayjs().startOf('day'), dayjs().endOf('day')]}
                rules={[{ required: true, message: '生效时间不能为空' }]}
                fieldProps={{
                  format: 'HH:mm:ss',
                }}
              />
            </Col>
            <Col span={8}>
              <ProFormSelect
                name="level"
                label="告警等级"
                width="md"
                showSearch
                options={levelTypes}
                // 默认中级，对应 enum: 0严重 1高 2中 3低
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
  );
};

export default CreateOrUpdateForm;
