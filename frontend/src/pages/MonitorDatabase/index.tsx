import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  ModalForm,
  PageContainer,
  ProDescriptions,
  ProForm,
  ProFormDependency,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { Button, Divider, Drawer, message } from 'antd';
import React, { useRef, useState } from 'react';
import {
  monitorDatabaseCreate,
  monitorDatabaseQuery,
  monitorDatabaseUpdate,
} from '@/services/ant-design-pro/monitor.database';
import { DatabaseTypeEnum } from '@/services/ant-design-pro/enum';

const typeOptions = Object.keys(DatabaseTypeEnum).map((item) => ({
  label: DatabaseTypeEnum[Number(item)],
  value: Number(item),
}));

const handleCreate = async (fields: API.MonitorDatabase) => {
  const hide = message.loading('正在测试连接并保存');
  try {
    await monitorDatabaseCreate(fields);
    hide();
    message.success('保存成功（已测试连接）');
    return true;
  } catch (error: any) {
    hide();
    message.error('测试连接失败: ' + (error?.message || '请检查连接信息与后端服务'));
    return false;
  }
};

const handleUpdate = async (fields: API.MonitorDatabase) => {
  const hide = message.loading('正在测试连接并更新');
  try {
    await monitorDatabaseUpdate(fields);
    hide();
    message.success('更新成功（已测试连接）');
    return true;
  } catch (error: any) {
    hide();
    message.error('测试连接失败: ' + (error?.message || '请检查连接信息与后端服务'));
    return false;
  }
};

const MonitorDatabasePage: React.FC = () => {
  const actionRef = useRef<ActionType>(undefined);
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const [current, setCurrent] = useState<API.MonitorDatabase>();

  const columns: ProColumns<API.MonitorDatabase>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (dom, entity) => (
        <a
          onClick={() => {
            setCurrent(entity);
            setShowDetail(true);
          }}
        >
          {dom}
        </a>
      ),
    },
    { title: '库名', dataIndex: 'database', search: false },
    {
      title: '类型',
      dataIndex: 'type',
      valueType: 'select',
      fieldProps: { options: typeOptions },
      valueEnum: DatabaseTypeEnum as any,
    },
    { title: '地址', dataIndex: 'url', search: false, ellipsis: true },
    { title: '账号', dataIndex: 'username', search: false },
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

  // 结构化表单字段；isEdit 时密码可留空以保留原密码
  const formFields = (isEdit: boolean) => (
    <>
      <Divider>基础信息</Divider>
      <ProForm.Group>
        <ProFormText
          name="name"
          label="名称"
          width="md"
          placeholder="请输入名称"
          rules={[{ required: true, message: '名称不能为空' }]}
        />
        <ProFormSelect
          name="type"
          label="数据库类型"
          width="md"
          options={typeOptions}
          rules={[{ required: true, message: '数据库类型不能为空' }]}
        />
      </ProForm.Group>

      <Divider>连接信息</Divider>
      <ProForm.Group>
        <ProFormText
          name="url"
          label="连接地址"
          width="md"
          placeholder="例如 127.0.0.1:3306"
          rules={[{ required: true, message: '连接地址不能为空' }]}
        />
        <ProFormText name="database" width="md" label="库名/DB索引" placeholder="请输入库名" />
      </ProForm.Group>

      <ProForm.Group>
        <ProFormText name="username" width="md" label="账号" placeholder="请输入账号" />
        <ProFormText.Password
          name="password"
          width="md"
          label="密码"
          placeholder={isEdit ? '留空则保留原密码' : '请输入密码'}
          extra={isEdit ? '不修改密码请留空，将继续使用原密码' : undefined}
          rules={
            isEdit
              ? undefined
              : [{ required: true, message: '密码不能为空' }]
          }
          fieldProps={{
            autoComplete: 'new-password',
          }}
        />
      </ProForm.Group>

      <Divider>附加参数</Divider>
      {/* 不放进 ProForm.Group，单独整行展示 */}
      <ProFormDependency name={['type']}>
        {({ type }) => {
          const isMongo = Number(type) === 1;
          return (
            <ProFormTextArea
              name="params"
              label="附加参数"
              // 对齐上方两列 md（328*2 + gap），占满一整行
              width={688}
              placeholder={
                isMongo
                  ? 'mongo URI 查询串，例如 collection=log&connectTimeoutMS=5000'
                  : 'mysql DSN 查询串，例如 charset=utf8mb4&parseTime=True&loc=Local'
              }
              tooltip={
                isMongo
                  ? '拼接为 mongo URI 的 ? 后查询参数；可含 collection=xxx'
                  : '拼接为 mysql DSN 的 ? 后查询参数；留空时默认 charset=utf8mb4&parseTime=True&loc=Local'
              }
              fieldProps={{
                autoSize: { minRows: 2, maxRows: 6 },
              }}
            />
          );
        }}
      </ProFormDependency>
    </>
  );

  return (
    <PageContainer>
      <ProTable<API.MonitorDatabase>
        headerTitle="数据源"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await monitorDatabaseQuery({
            current: params.current,
            pageSize: params.pageSize,
            name: params.name,
            type: params.type,
          });
          return {
            data: res.data || [],
            success: res.status === 200 || res.status === 0,
            total: res.total || 0,
          };
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
        {current?.name && (
          <ProDescriptions<API.MonitorDatabase>
            column={2}
            title={current?.name}
            request={async () => ({ data: current || {} })}
            params={{ id: current?.name }}
            columns={columns as any}
          />
        )}
      </Drawer>

      <ModalForm
        title="新建数据源"
        width="740px"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        submitter={{
          searchConfig: {
            submitText: '测试并保存',
          },
        }}
        onFinish={async (values) => {
          const ok = await handleCreate(values as API.MonitorDatabase);
          if (ok) {
            setCreateVisible(false);
            actionRef.current?.reload();
          }
          return ok;
        }}
      >
        {formFields(false)}
      </ModalForm>

      <ModalForm
        title="编辑数据源"
        width="740px"
        open={modifyVisible}
        initialValues={{ ...current, password: undefined }}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        submitter={{
          searchConfig: {
            submitText: '测试并保存',
          },
        }}
        onFinish={async (values) => {
          // 密码留空不传明文，后端沿用原密码
          const payload = {
            ...current,
            ...values,
            password: values.password || '',
          } as API.MonitorDatabase;
          const ok = await handleUpdate(payload);
          if (ok) {
            setModifyVisible(false);
            actionRef.current?.reload();
          }
          return ok;
        }}
      >
        {formFields(true)}
      </ModalForm>
    </PageContainer>
  );
};

export default MonitorDatabasePage;
