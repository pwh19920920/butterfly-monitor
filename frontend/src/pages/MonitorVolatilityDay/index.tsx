import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  ModalForm,
  PageContainer,
  ProFormDateTimeRangePicker,
  ProFormList,
  ProFormSelect,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { Button, message, Popconfirm, Space, Tag } from 'antd';
import dayjs from 'dayjs';
import React, { useRef, useState } from 'react';
import { VolatilityDayTypeEnum } from '@/services/ant-design-pro/enum';
import {
  volatilityDayBatchCreate,
  volatilityDayDelete,
  volatilityDayQueryAll,
  volatilityDayUpdate,
} from '@/services/ant-design-pro/monitorVolatilityDay';

const VolatilityDayTypeColor: Record<number, string> = {
  1: 'orange',
  2: 'blue',
};

const VolatilityDayPage: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [batchVisible, setBatchVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [current, setCurrent] = useState<API.MonitorVolatilityDay>();

  const columns: ProColumns<API.MonitorVolatilityDay>[] = [
    { title: '名称', dataIndex: 'name' , ellipsis: true },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      valueType: 'dateTime',
      search: false,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      valueType: 'dateTime',
      search: false,
    },
    {
      title: '类型',
      dataIndex: 'type',
      search: false,
      render: (_, r) => {
        const t = r.type != null ? VolatilityDayTypeEnum[r.type] : undefined;
        return t ? (
          <Tag color={VolatilityDayTypeColor[r.type ?? 0]}>{t}</Tag>
        ) : (
          '-'
        );
      },
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, record) => (
        <Space>
          <a
            onClick={() => {
              setCurrent(record);
              setEditVisible(true);
            }}
          >
            编辑
          </a>
          <Popconfirm
            title="确认删除？"
            onConfirm={async () => {
              await volatilityDayDelete(record.id);
              message.success('删除成功');
              actionRef.current?.reload();
            }}
          >
            <a>删除</a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.MonitorVolatilityDay>
        headerTitle="波动日管理"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        search={false}
        toolBarRender={() => [
          <Button
            type="primary"
            key="batch"
            onClick={() => setBatchVisible(true)}
          >
            <PlusOutlined /> 批量添加
          </Button>,
        ]}
        request={async () => {
          const res = await volatilityDayQueryAll();
          return { data: res.data || [], success: true };
        }}
      />

      <ModalForm
        title="批量添加波动日"
        open={batchVisible}
        modalProps={{
          destroyOnClose: true,
          onCancel: () => setBatchVisible(false),
        }}
        onFinish={async (values) => {
          try {
            // ProFormList 内 transform 不生效，这里显式把 dateRange 映射为
            // 后端约定的 items[].startTime / endTime（格式 2006-01-02 15:04:05）
            const items = (values.items || []).map((it: any) => {
              const [startTime, endTime] = it?.dateRange || [];
              return {
                startTime: startTime
                  ? dayjs(startTime).format('YYYY-MM-DD HH:mm:ss')
                  : undefined,
                endTime: endTime
                  ? dayjs(endTime).format('YYYY-MM-DD HH:mm:ss')
                  : undefined,
              };
            });
            await volatilityDayBatchCreate({
              name: values.name,
              type: values.type,
              items,
            } as API.MonitorVolatilityDayBatchCreateRequest);
            message.success('添加成功');
            setBatchVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('添加失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect
          name="type"
          label="类型"
          valueEnum={VolatilityDayTypeEnum as any}
          rules={[{ required: true }]}
        />

        <ProFormList name="items" label="日期区间" initialValue={[{}]} min={1}>
          <ProFormDateTimeRangePicker
            name="dateRange"
            rules={[{ required: true, message: '请选择起止时间' }]}
          />
        </ProFormList>
      </ModalForm>

      <ModalForm
        title="编辑波动日"
        open={editVisible}
        initialValues={
          current
            ? {
                ...current,
                // 后端返回字符串，RangePicker 需 dayjs 对象回显
                dateRange: [
                  current.startTime ? dayjs(current.startTime) : undefined,
                  current.endTime ? dayjs(current.endTime) : undefined,
                ],
              }
            : undefined
        }
        modalProps={{
          destroyOnClose: true,
          onCancel: () => setEditVisible(false),
        }}
        onFinish={async (values) => {
          try {
            const [startTime, endTime] = values.dateRange || [];
            await volatilityDayUpdate(current?.id, {
              ...current,
              name: values.name,
              type: values.type,
              // 后端 LocalTime 仅接受 2006-01-02 15:04:05 格式
              startTime: startTime
                ? dayjs(startTime).format('YYYY-MM-DD HH:mm:ss')
                : current?.startTime,
              endTime: endTime
                ? dayjs(endTime).format('YYYY-MM-DD HH:mm:ss')
                : current?.endTime,
            } as API.MonitorVolatilityDay);
            message.success('更新成功');
            setEditVisible(false);
            actionRef.current?.reload();
            return true;
          } catch {
            message.error('更新失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect
          name="type"
          label="类型"
          valueEnum={VolatilityDayTypeEnum as any}
          rules={[{ required: true }]}
        />
        <ProFormDateTimeRangePicker
          name="dateRange"
          label="日期区间"
          rules={[{ required: true, message: '请选择起止时间' }]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default VolatilityDayPage;
