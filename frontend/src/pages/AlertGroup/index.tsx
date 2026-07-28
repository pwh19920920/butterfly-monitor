import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { ModalForm, PageContainer, ProFormSelect, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, message } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import {
  alertGroupCreate,
  alertGroupQuery,
  alertGroupUpdate,
  alertGroupUsers,
} from '@/services/ant-design-pro/alert.group';
import { sysUserQueryAll } from '@/services/ant-design-pro/sys.user';

const AlertGroupPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [modifyVisible, setModifyVisible] = useState(false);
  const [current, setCurrent] = useState<API.AlertGroup>();
  const [users, setUsers] = useState<{ label: string; value: string }[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);

  useEffect(() => {
    sysUserQueryAll().then((res) => {
      setUsers((res.data || []).map((u) => ({ label: `${u.name}(${u.username})`, value: String(u.id) })));
    });
  }, []);

  const columns: ProColumns<API.AlertGroup>[] = [
    { title: '名称', dataIndex: 'name' },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="edit"
          onClick={async () => {
            setCurrent(record);
            try {
              const res = await alertGroupUsers(record.id!);
              setSelectedUsers((res.data || []).map(String));
            } catch {
              setSelectedUsers([]);
            }
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
      <ProTable<API.AlertGroup>
        headerTitle="报警组"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        toolBarRender={() => [
          <Button type="primary" key="add" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await alertGroupQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
      />
      <ModalForm
        title="新建报警组"
        open={createVisible}
        modalProps={{ destroyOnClose: true, onCancel: () => setCreateVisible(false) }}
        onFinish={async (values) => {
          try {
            await alertGroupCreate({ name: values.name, userIds: values.userIds || [] });
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
        <ProFormSelect name="userIds" label="成员" mode="multiple" options={users} rules={[{ required: true }]} />
      </ModalForm>
      <ModalForm
        title="编辑报警组"
        open={modifyVisible}
        initialValues={{ name: current?.name, userIds: selectedUsers }}
        modalProps={{ destroyOnClose: true, onCancel: () => setModifyVisible(false) }}
        onFinish={async (values) => {
          try {
            await alertGroupUpdate({ id: current!.id!, name: values.name, userIds: values.userIds || [] });
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
        <ProFormSelect name="userIds" label="成员" mode="multiple" options={users} rules={[{ required: true }]} />
      </ModalForm>
    </PageContainer>
  );
};

export default AlertGroupPage;
