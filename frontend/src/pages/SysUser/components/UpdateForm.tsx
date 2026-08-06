import { LoadingOutlined } from '@ant-design/icons';
import {
  ProForm,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import { Spin } from 'antd';
import React, { useEffect, useState } from 'react';
import { sysRoleQueryAll } from '@/services/ant-design-pro/sys.role';

type SysRoleItem = {
  label: string;
  value: string;
};

type CreateOrUpdateFormProps = {
  isCreate: boolean;
};

const CreateOrUpdateForm: React.FC<CreateOrUpdateFormProps> = (
  props: CreateOrUpdateFormProps,
) => {
  const [roleData, setRoleData] = useState<SysRoleItem[]>([]);

  const reloadRoleData = async () => {
    const resp: API.Resp<API.SysRole[]> = await sysRoleQueryAll();
    if (resp.data) {
      setRoleData(
        resp.data.map((item: API.SysRole): SysRoleItem => {
          return { label: item.name, value: `${item.id}` };
        }),
      );
    }
  };

  useEffect(() => {
    reloadRoleData().then(() => {});
  }, []);

  if (roleData.length === 0) {
    return (
      <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
    );
  }

  return (
    <ProForm.Group>
      <ProFormText
        label="名称"
        rules={[
          {
            required: true,
            message: '名称不能为空',
          },
        ]}
        width="md"
        placeholder="请输入名称"
        name="name"
      />

      <ProFormText
        label="电话"
        rules={[
          {
            required: true,
            message: '电话不能为空',
          },
        ]}
        width="md"
        placeholder="请输入电话"
        name="mobile"
      />

      <ProFormText
        label="邮箱"
        rules={[
          {
            required: true,
            message: '邮箱不能为空',
          },
        ]}
        width="md"
        placeholder="请输入邮箱"
        name="email"
      />

      <ProFormText
        label="用户名"
        width="md"
        rules={[
          {
            required: true,
            message: '名称不能为空',
          },
        ]}
        placeholder="请输入用户名"
        name="username"
      />

      <ProFormText
        label="密码"
        width="md"
        rules={[
          {
            required: props.isCreate,
            message: '密码不能为空',
          },
        ]}
        placeholder="新增时必须输入, 修改过程中不修改无需输入"
        name="password"
      />

      <ProFormSelect
        name="roleList"
        label="角色"
        width="md"
        mode="multiple"
        options={roleData}
        placeholder="请选择用户角色"
        rules={[{ required: true, message: '请选择用户角色!' }]}
      />
    </ProForm.Group>
  );
};

export default CreateOrUpdateForm;
