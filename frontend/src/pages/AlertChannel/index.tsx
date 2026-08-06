import { PlusOutlined } from '@ant-design/icons';
import type {
  ActionType,
  ProColumns,
  ProFormInstance,
} from '@ant-design/pro-components';
import {
  ModalForm,
  PageContainer,
  ProDescriptions,
  ProTable,
} from '@ant-design/pro-components';
import { Button, Drawer, message } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import CreateOrUpdateForm from '@/pages/AlertChannel/components/UpdateForm';
import {
  alertChannelCreate,
  alertChannelHandlers,
  alertChannelQuery,
  alertChannelUpdate,
} from '@/services/ant-design-pro/alert.channel';
import {
  AlertChannelFailRouteEnum,
  AlertChannelTypeEnum,
} from '@/services/ant-design-pro/enum';

const handleCreate = async (fields: API.AlertChannelSaveRequest) => {
  const hide = message.loading('正在测试并保存');
  try {
    await alertChannelCreate(fields);
    hide();
    message.success('保存通道成功（已测试发送）');
    return true;
  } catch (error: any) {
    hide();
    message.error(`测试发送失败: ${error?.message || '请检查参数与连通性'}`);
    return false;
  }
};

const handleUpdate = async (fields: API.AlertChannelSaveRequest) => {
  const hide = message.loading('正在测试并更新');
  try {
    await alertChannelUpdate(fields);
    hide();
    message.success('更新通道成功（已测试发送）');
    return true;
  } catch (error: any) {
    hide();
    message.error(`测试发送失败: ${error?.message || '请检查参数与连通性'}`);
    return false;
  }
};

// paramsObj → params 字符串，sslValue(0/1) → ssl(bool)
const stringifyParams = (paramsObj: any): string => {
  if (!paramsObj) return '';
  const { sslValue, ...rest } = paramsObj;
  const obj = { ...rest };
  if (sslValue !== undefined) {
    obj.ssl = sslValue === 1;
  }
  return JSON.stringify(obj);
};

// params 字符串 → paramsObj，ssl(bool) → sslValue(0/1)
const parseParams = (params?: string): any => {
  if (!params) return undefined;
  try {
    const obj = JSON.parse(params);
    if (obj.ssl !== undefined) {
      obj.sslValue = obj.ssl ? 1 : 0;
      delete obj.ssl;
    }
    return obj;
  } catch {
    return undefined;
  }
};

const TableList: React.FC = () => {
  const [createModalVisible, handleCreateModalVisible] =
    useState<boolean>(false);
  const [modifyModalVisible, handleModifyModalVisible] =
    useState<boolean>(false);

  const actionRef = useRef<ActionType>(undefined);
  const drawRef = useRef<ActionType>(undefined);
  const createFormRef = useRef<ProFormInstance<API.AlertChannel>>(undefined);
  const updateFormRef = useRef<ProFormInstance<API.AlertChannel>>(undefined);

  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.AlertChannel>();
  const [channelHandlers, setChannelHandlers] = useState<
    API.AlertChannelHandler[]
  >([]);

  const loadHandlers = async () => {
    const resp = await alertChannelHandlers();
    if (resp.data) {
      setChannelHandlers(resp.data);
    }
  };

  useEffect(() => {
    loadHandlers();
  }, []);

  const columns: ProColumns<API.AlertChannel>[] = [
    {
      title: '渠道名称',
      dataIndex: 'name',
      render: (dom, entity) => (
        <a
          onClick={() => {
            setCurrentRow(entity);
            setShowDetail(true);
          }}
        >
          {dom}
        </a>
      ),
    },
    {
      title: '渠道类型',
      dataIndex: 'type',
      valueEnum: AlertChannelTypeEnum as any,
    },
    { title: '渠道处理器', dataIndex: 'handler' },
    {
      title: '失败路由',
      dataIndex: 'failRoute',
      valueEnum: AlertChannelFailRouteEnum as any,
    },
    {
      title: '告警模板',
      dataIndex: 'template',
      ellipsis: true,
      search: false,
      render: (dom) => (dom && dom !== '-' ? dom : '默认'),
    },
    { title: '参数', dataIndex: 'params', ellipsis: true },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="config"
          onClick={() => {
            setCurrentRow(record);
            handleModifyModalVisible(true);
          }}
        >
          修改
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.AlertChannel>
        headerTitle="报警通道"
        actionRef={actionRef}
        rowKey="id"
        search={false}
        columns={columns}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => handleCreateModalVisible(true)}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params) => {
          const res = await alertChannelQuery(params);
          return { data: res.data || [], success: true, total: res.total || 0 };
        }}
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
          <ProDescriptions<API.AlertChannel>
            actionRef={drawRef}
            column={2}
            title={currentRow?.name}
            request={async () => ({ data: currentRow || {} })}
            params={{ id: currentRow?.name }}
            columns={columns as any}
          />
        )}
      </Drawer>

      {createModalVisible && (
        <ModalForm
          title="创建渠道"
          width="740px"
          open={createModalVisible}
          formRef={updateFormRef}
          modalProps={{
            destroyOnClose: true,
            onCancel: () => handleCreateModalVisible(false),
          }}
          submitter={{ searchConfig: { submitText: '测试并保存' } }}
          onFinish={async (value: any) => {
            const payload: API.AlertChannelSaveRequest = {
              ...value,
              params: stringifyParams(value.paramsObj),
            };
            delete (payload as any).paramsObj;
            const success = await handleCreate(payload);
            if (success) {
              handleCreateModalVisible(false);
              actionRef.current?.reload();
            }
            return success;
          }}
        >
          <CreateOrUpdateForm
            formRef={updateFormRef}
            channelHandlers={channelHandlers}
            channelType={-1}
          />
        </ModalForm>
      )}

      {modifyModalVisible && currentRow ? (
        <ModalForm
          title="更新渠道"
          width="740px"
          open={modifyModalVisible}
          formRef={createFormRef}
          initialValues={{
            ...currentRow,
            paramsObj: parseParams(currentRow.params),
          }}
          modalProps={{
            destroyOnClose: true,
            onCancel: () => handleModifyModalVisible(false),
          }}
          submitter={{ searchConfig: { submitText: '测试并更新' } }}
          onFinish={async (value: any) => {
            const payload: API.AlertChannelSaveRequest = {
              ...currentRow,
              ...value,
              params: stringifyParams(value.paramsObj),
            };
            delete (payload as any).paramsObj;
            const success = await handleUpdate(payload);
            if (success) {
              handleModifyModalVisible(false);
              actionRef.current?.reload();
            }
            return success;
          }}
        >
          <CreateOrUpdateForm
            formRef={createFormRef}
            channelHandlers={channelHandlers}
            channelType={currentRow.type ?? -1}
          />
        </ModalForm>
      ) : null}
    </PageContainer>
  );
};

export default TableList;
