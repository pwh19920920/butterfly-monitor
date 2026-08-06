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
import { Button, Drawer, message, Popconfirm } from 'antd';
import React, { useRef, useState } from 'react';
import CreateOrUpdateForm from '@/pages/SysMenu/components/UpdateForm';
import {
  sysMenuCreate,
  sysMenuDelete,
  sysMenuOptionQuery,
  sysMenuQuery,
  sysMenuUpdate,
} from '@/services/ant-design-pro/sys.menu';

const handleCreate = async (fields: API.SysMenu) => {
  const hide = message.loading('正在添加');
  try {
    await sysMenuCreate({ ...fields });
    hide();
    message.success('保存菜单成功');
    return true;
  } catch (_error) {
    hide();
    message.error('保存菜单失败!');
    return false;
  }
};

/**
 * @en-US Add node
 * @zh-CN 添加节点
 * @param fields
 */
const handleUpdate = async (fields: API.SysMenu) => {
  const hide = message.loading('正在更新');
  try {
    await sysMenuUpdate({ ...fields });
    hide();
    message.success('更新菜单成功');
    return true;
  } catch (_error) {
    hide();
    message.error('更新菜单失败!');
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
    await sysMenuDelete(id);
    hide();
    message.success('删除菜单成功');
    return true;
  } catch (_error) {
    hide();
    message.error('删除菜单失败!');
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
  const [currentRow, setCurrentRow] = useState<API.SysMenu>();

  const columns: ProColumns<API.SysMenu>[] = [
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
      title: '菜单代码',
      dataIndex: 'code',
    },
    {
      title: '路径',
      dataIndex: 'path',
    },
    {
      title: '图标',
      dataIndex: 'icon',
    },
    {
      title: '组件',
      dataIndex: 'component',
      search: false,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      search: false,
    },
    {
      title: '菜单操作',
      dataIndex: 'options',
      search: false,
      hideInTable: true,
      render: (_, record) =>
        record.options?.map((item) => item.name).join(', '),
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="config"
          onClick={async () => {
            // 远程get相关option
            const resp = await sysMenuOptionQuery(record.id);
            if (!resp.data) {
              message.error('获取菜单操作失败!');
              return;
            }
            setCurrentRow({ ...record, options: resp.data });
            handleModifyModalVisible(true);
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
      <ProTable<API.SysMenu>
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
        request={() => sysMenuQuery()}
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
          <ProDescriptions<API.SysMenu>
            column={2}
            title={currentRow?.name}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.name,
            }}
            columns={columns as ProDescriptionsItemProps<API.SysMenu>[]}
          />
        )}
      </Drawer>

      {createModalVisible && (
        <ModalForm
          title="创建菜单"
          width="850px"
          open={createModalVisible}
          onOpenChange={handleModalVisible}
          onFinish={async (value: API.SysMenu) => {
            const success = await handleCreate({ ...value });
            if (success) {
              handleModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
        >
          <CreateOrUpdateForm initOptionsValues={[]} />
        </ModalForm>
      )}

      {modifyModalVisible && currentRow ? (
        <ModalForm
          title={'更新数据源'}
          width="850px"
          initialValues={currentRow}
          open={modifyModalVisible}
          onOpenChange={handleModifyModalVisible}
          onFinish={async (value: API.SysMenu) => {
            if (value.parent === value.id) {
              message.error('上级菜单不能是自己!');
              return;
            }

            // 检查自己是不是塞到儿子下面了
            if (currentRow.children) {
              for (let i = 0; i < currentRow.children.length; i++) {
                if (currentRow.children[i].id === value.parent) {
                  message.error('上级菜单不能是本菜单下的子菜单!');
                  return;
                }
              }
            }

            const success = await handleUpdate({ ...currentRow, ...value });
            if (success) {
              handleModifyModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
        >
          <CreateOrUpdateForm initOptionsValues={currentRow.options || []} />
        </ModalForm>
      ) : null}
    </PageContainer>
  );
};

export default TableList;
