import { PlusOutlined } from '@ant-design/icons';
import type {
  ActionType,
  ProColumns,
  ProDescriptionsItemProps,
} from '@ant-design/pro-components';
import {
  ModalForm,
  PageContainer,
  ProDescriptions,
  ProTable,
} from '@ant-design/pro-components';
import { FormattedMessage } from '@umijs/max';
import { Button, Drawer, Image, message, Popconfirm } from 'antd';
import React, { useRef, useState } from 'react';
import CreateOrUpdateForm from '@/pages/SysUser/components/UpdateForm';
import {
  sysUserCreate,
  sysUserDelete,
  sysUserQuery,
  sysUserUpdate,
} from '@/services/ant-design-pro/sys.user';

const handleCreate = async (fields: API.SysUser) => {
  const hide = message.loading('正在添加');
  try {
    await sysUserCreate({ ...fields, roles: fields.roleList?.join(',') });
    hide();
    message.success('保存用户成功');
    return true;
  } catch (_error) {
    hide();
    message.error('保存用户失败!');
    return false;
  }
};

/**
 * @en-US Add node
 * @zh-CN 添加节点
 * @param fields
 */
const handleUpdate = async (fields: API.SysUser) => {
  const hide = message.loading('正在更新');
  try {
    await sysUserUpdate({ ...fields, roles: fields.roleList?.join(',') });
    hide();
    message.success('更新愈合成功');
    return true;
  } catch (_error) {
    hide();
    message.error('更新用户失败!');
    return false;
  }
};

/**
 * 删除
 * @param id
 */
const handleDelete = async (id: string) => {
  const hide = message.loading('正在删除');
  try {
    await sysUserDelete(id);
    hide();
    message.success('删除用户成功');
    return true;
  } catch (_error) {
    hide();
    message.error('删除用户失败!');
    return false;
  }
};

const TableList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [modifyModalVisible, handleModifyModalVisible] =
    useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);

  const actionRef = useRef<ActionType>(null);
  const [currentRow, setCurrentRow] = useState<API.SysUser>();

  const columns: ProColumns<API.SysUser>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: '用户名',
      dataIndex: 'username',
    },
    {
      title: '电话',
      dataIndex: 'mobile',
      search: false,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      search: false,
    },
    {
      title: '头像',
      dataIndex: 'avatar',
      search: false,
      render: (_, entity) => {
        return <Image width={20} src={entity.avatar} />;
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="config"
          onClick={() => {
            handleModifyModalVisible(true);
            setCurrentRow(record);
          }}
        >
          编辑
        </a>,

        <Popconfirm
          title="是否确认删除?"
          key="id"
          onConfirm={async () => {
            const success = await handleDelete(record.id);
            if (success && actionRef.current) {
              actionRef.current.reload();
            }
          }}
          onCancel={() => {}}
          okText="确认"
          cancelText="取消"
        >
          <a href="#">删除</a>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.SysUser>
        headerTitle="查询表格"
        actionRef={actionRef}
        rowKey="id"
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              handleModalVisible(true);
            }}
          >
            <PlusOutlined />{' '}
            <FormattedMessage id="pages.searchTable.new" defaultMessage="New" />
          </Button>,
        ]}
        request={sysUserQuery}
        columns={columns}
        rowSelection={
          {
            // onChange: (_, selectedRows) => {
            // setSelectedRows(selectedRows);
            // 批量选择
            // },
          }
        }
      />

      <Drawer
        width={600}
        open={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.name && (
          <ProDescriptions<API.SysUser>
            column={2}
            title={currentRow?.name}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.name,
            }}
            columns={columns as ProDescriptionsItemProps<API.SysUser>[]}
          />
        )}
      </Drawer>

      <ModalForm
        title="创建用户"
        width="740px"
        open={createModalVisible}
        onOpenChange={handleModalVisible}
        onFinish={async (value: API.SysUser) => {
          const success = await handleCreate({ ...value });
          if (success) {
            handleModalVisible(false);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
      >
        <CreateOrUpdateForm isCreate={true} />
      </ModalForm>

      {modifyModalVisible && currentRow ? (
        <ModalForm
          title={'更新用户'}
          width="740px"
          initialValues={currentRow}
          open={modifyModalVisible}
          onOpenChange={handleModifyModalVisible}
          onFinish={async (value: API.SysUser) => {
            const success = await handleUpdate({ ...currentRow, ...value });
            if (success) {
              handleModifyModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
        >
          <CreateOrUpdateForm isCreate={false} />
        </ModalForm>
      ) : null}
    </PageContainer>
  );
};

export default TableList;
