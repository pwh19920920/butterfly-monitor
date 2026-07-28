import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProFormSelect, ProFormText, ProFormTextArea, ProTable } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import React, { useRef, useState } from 'react';
import { alertConfCreate, alertConfQuery, alertConfUpdate } from '@/services/ant-design-pro/alert.conf';
import { AlertConfTypeEnum } from '@/services/ant-design-pro/enum';

const AlertConfPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [current, setCurrent] = useState<API.AlertConf>();

  const columns: ProColumns<API.AlertConf>[] = [
    { title: 'Key', dataIndex: 'confKey' },
    { title: 'Value', dataIndex: 'confVal', search: false, ellipsis: true },
    { title: '描述', dataIndex: 'confDesc', search: false },
    {
      title: '类型',
      dataIndex: 'confType',
      search: false,
      valueEnum: AlertConfTypeEnum as any,
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
      <ProTable<API.AlertConf>
        headerTitle="报警配置"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await alertConfQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />
      <ModalForm
        title="新建配置"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        onFinish={async (values) => {
          try {
            await alertConfCreate(values as API.AlertConf);
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
        <ProFormText name="confKey" label="Key" rules={[{ required: true }]} />
        <ProFormTextArea name="confVal" label="Value" rules={[{ required: true }]} />
        <ProFormText name="confDesc" label="描述" />
        <ProFormSelect
          name="confType"
          label="类型"
          options={[
            { label: AlertConfTypeEnum[1], value: 1 },
            { label: AlertConfTypeEnum[2], value: 2 },
          ]}
          rules={[{ required: true }]}
        />
      </ModalForm>
      <ModalForm
        title="编辑配置"
        open={modifyVisible}
        initialValues={current}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        onFinish={async (values) => {
          try {
            await alertConfUpdate({ ...current, ...values } as API.AlertConf);
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
        <ProFormText name="confKey" label="Key" disabled />
        <ProFormTextArea name="confVal" label="Value" rules={[{ required: true }]} />
        <ProFormText name="confDesc" label="描述" />
      </ModalForm>
    </PageContainer>
  );
};

export default AlertConfPage;
