import { PlusOutlined } from '@ant-design/icons';
import type {
  ActionType,
  ProColumns,
  ProDescriptionsItemProps,
} from '@ant-design/pro-components';
import {
  PageContainer,
  ProDescriptions,
  ProTable,
} from '@ant-design/pro-components';
import { FormattedMessage } from '@umijs/max';
import { Button, Drawer, message, Popconfirm } from 'antd';
import React, { useRef, useState } from 'react';
import CreateOrUpdateForm from '@/pages/SysRole/components/UpdateForm';
import {
  sysRoleCreate,
  sysRoleDelete,
  sysRoleQuery,
  sysRoleUpdate,
} from '@/services/ant-design-pro/sys.role';

const handleCreate = async (fields: API.SysRole) => {
  const hide = message.loading('正在添加');
  try {
    await sysRoleCreate({ ...fields });
    hide();
    message.success('保存角色成功');
    return true;
  } catch (_error) {
    hide();
    message.error('保存角色失败!');
    return false;
  }
};

/**
 * @en-US Add node
 * @zh-CN 添加节点
 * @param fields
 */
const handleUpdate = async (fields: API.SysRole) => {
  const hide = message.loading('正在更新');
  try {
    await sysRoleUpdate({ ...fields });
    hide();
    message.success('更新角色成功');
    return true;
  } catch (_error) {
    hide();
    message.error('更新角色失败!');
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
    await sysRoleDelete(id);
    hide();
    message.success('删除角色成功');
    return true;
  } catch (_error) {
    hide();
    message.error('删除角色失败!');
    return false;
  }
};

const TableList: React.FC = () => {
  /**
   * @en-US Pop-up window of new window
   * @zh-CN 新建窗口的弹窗
   *  */
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [modifyModalVisible, handleModifyModalVisible] =
    useState<boolean>(false);

  const [showDetail, setShowDetail] = useState<boolean>(false);

  const actionRef = useRef<ActionType>(null);
  const [currentRow, setCurrentRow] = useState<API.SysRole>();

  const columns: ProColumns<API.SysRole>[] = [
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
      <ProTable<API.SysRole>
        search={false}
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
        request={sysRoleQuery}
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
          <ProDescriptions<API.SysRole>
            column={2}
            title={currentRow?.name}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.name,
            }}
            columns={columns as ProDescriptionsItemProps<API.SysRole>[]}
          />
        )}
      </Drawer>

      {createModalVisible ? (
        <CreateOrUpdateForm
          title="创建角色信息"
          width="740px"
          open={createModalVisible}
          onOpenChange={handleModalVisible}
          onSubmit={async (
            value: { name: string },
            permissions: API.SysPermission[],
          ) => {
            const success = await handleCreate({
              id: '0',
              ...value,
              permissions,
            });
            if (success) {
              handleModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
        />
      ) : null}

      {modifyModalVisible && currentRow ? (
        <CreateOrUpdateForm
          title="更新角色信息"
          width="740px"
          initValues={currentRow}
          open={modifyModalVisible}
          onOpenChange={handleModifyModalVisible}
          onSubmit={async (
            value: { name: string },
            permissions: API.SysPermission[],
          ) => {
            const success = await handleUpdate({
              ...currentRow,
              ...value,
              permissions,
            });
            if (success) {
              handleModifyModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
        />
      ) : null}
    </PageContainer>
  );
};

export default TableList;
