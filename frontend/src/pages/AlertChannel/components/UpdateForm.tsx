import { LoadingOutlined } from '@ant-design/icons';
import {
  ProForm,
  ProFormDigit,
  type ProFormInstance,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { Spin } from 'antd';
import React, { useEffect, useState } from 'react';
import {
  AlertChannelFailRouteEnum,
  AlertChannelTypeEnum,
} from '@/services/ant-design-pro/enum';

type SelectItem = {
  label: string;
  value: number;
};

type HandlerSelectItem = {
  label: string;
  value: string;
};

type CreateOrUpdateFormProps = {
  channelHandlers: API.AlertChannelHandler[];
  // -1 表示新建态，尚未选择类型；编辑态传入当前通道类型
  channelType: number;
  formRef: React.MutableRefObject<
    ProFormInstance<API.AlertChannel> | undefined
  >;
};

const failRoutes: SelectItem[] = Object.keys(AlertChannelFailRouteEnum).map(
  (item) => ({
    value: Number(item),
    label: AlertChannelFailRouteEnum[Number(item)],
  }),
);

// 是否 SSL：前端用 0/1 下拉，提交时转 bool 写入 paramsObj.ssl
const sslSelects: SelectItem[] = [
  { value: 1, label: '是' },
  { value: 0, label: '否' },
];

const CreateOrUpdateForm: React.FC<CreateOrUpdateFormProps> = (prop) => {
  const [selectChannelType, setSelectChannelType] = useState<number>(
    prop.channelType,
  );
  const [selectChannelHandlers, setSelectChannelHandlers] =
    useState<HandlerSelectItem[]>();
  const [channelHandlerMaps, setChannelHandlerMaps] = useState<
    Map<number, HandlerSelectItem[]>
  >(new Map());
  // 当前选中的 handler 类名，用于按 handler 渲染参数表单
  const [selectedHandler, setSelectedHandler] = useState<string>('');

  const channelTypes: SelectItem[] = prop.channelHandlers.map(
    (item): SelectItem => ({
      value: Number(item.channelType),
      label:
        AlertChannelTypeEnum[item.channelType] || `类型${item.channelType}`,
    }),
  );

  useEffect(() => {
    const handlersMap = new Map<number, HandlerSelectItem[]>();
    prop.channelHandlers.forEach((item) => {
      handlersMap.set(
        Number(item.channelType),
        item.handlers.map(
          (handle): HandlerSelectItem => ({ value: handle, label: handle }),
        ),
      );
    });

    setChannelHandlerMaps(handlersMap);

    // 设置当前类型下的处理器选项
    const handlers = handlersMap.get(prop.channelType);
    setSelectChannelHandlers(handlers);
    // 初始化选中 handler：编辑态取已保存值，否则取该类型首个
    const initialHandler = prop.formRef.current?.getFieldValue('handler');
    setSelectedHandler(
      initialHandler ||
        (handlers && handlers.length > 0 ? handlers[0].value : ''),
    );
  }, []);

  if (channelHandlerMaps.size === 0) {
    return (
      <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
    );
  }

  return (
    <>
      <ProForm.Group title="通道基础信息">
        <ProFormText
          label="通道名称"
          rules={[{ required: true, message: '通道名称不能为空' }]}
          width="md"
          placeholder="请输入通道名称"
          name="name"
        />

        <ProFormSelect
          showSearch
          options={failRoutes}
          rules={[{ required: true, message: '失败路由不能为空' }]}
          width="md"
          name="failRoute"
          label="失败路由"
        />

        <ProFormSelect
          showSearch
          options={channelTypes}
          rules={[{ required: true, message: '通道类型不能为空' }]}
          fieldProps={{
            onChange: (value: number) => {
              setSelectChannelType(value);
              const current = channelHandlerMaps.get(value);
              setSelectChannelHandlers(current);
              const firstHandler =
                current && current.length > 0 ? current[0].value : '';
              prop.formRef.current?.setFieldsValue({ handler: firstHandler });
              setSelectedHandler(firstHandler);
            },
          }}
          width="md"
          name="type"
          label="通道类型"
        />

        {selectChannelType !== -1 && (
          <ProFormSelect
            showSearch
            options={selectChannelHandlers}
            rules={[{ required: true, message: '通道处理器不能为空' }]}
            width="md"
            name="handler"
            label="通道处理器"
            fieldProps={{
              onChange: (value: string) => setSelectedHandler(value),
            }}
          />
        )}
      </ProForm.Group>

      {selectChannelType === 1 && (
        <ProForm.Group title="报警通道参数">
          <ProFormText
            label="smtp地址"
            rules={[{ required: true, message: 'smtp地址不能为空' }]}
            width="md"
            placeholder="smtp地址，例如 smtp.exmail.qq.com"
            name={['paramsObj', 'host']}
          />

          <ProFormDigit
            label="smtp端口"
            rules={[{ required: true, message: 'smtp端口不能为空' }]}
            width="md"
            placeholder="smtp端口"
            name={['paramsObj', 'port']}
          />

          <ProFormText
            label="smtp用户名"
            rules={[{ required: true, message: 'smtp用户名不能为空' }]}
            width="md"
            placeholder="smtp用户名"
            name={['paramsObj', 'username']}
          />

          <ProFormText.Password
            label="smtp密码"
            rules={[{ required: true, message: 'smtp密码不能为空' }]}
            width="md"
            placeholder="smtp密码"
            name={['paramsObj', 'password']}
          />

          <ProFormSelect
            options={sslSelects}
            rules={[{ required: true, message: '是否SSL不能为空' }]}
            width="md"
            name={['paramsObj', 'sslValue']}
            label="是否是SSL"
          />
        </ProForm.Group>
      )}

      {selectChannelType === 2 && selectedHandler === 'ChannelWechatHandler' && (
        <ProForm.Group title="报警通道参数">
          <ProFormText
            label="webhook地址"
            rules={[{ required: true, message: 'webhook地址不能为空' }]}
            width="md"
            placeholder="企业微信 webhook 地址"
            name={['paramsObj', 'addr']}
          />
        </ProForm.Group>
      )}

      {selectChannelType === 2 &&
        selectedHandler === 'ChannelDingtalkHandler' && (
          <ProForm.Group title="报警通道参数">
            <ProFormText
              label="webhook地址"
              rules={[{ required: true, message: 'webhook地址不能为空' }]}
              width="md"
              placeholder="钉钉机器人 webhook 地址"
              name={['paramsObj', 'addr']}
            />
            <ProFormText
              label="加签密钥"
              width="md"
              placeholder="可选，启用了加签的机器人填写"
              name={['paramsObj', 'secret']}
            />
          </ProForm.Group>
        )}

      {selectChannelType === 2 &&
        selectedHandler === 'ChannelFeishuHandler' && (
          <ProForm.Group title="报警通道参数">
            <ProFormText
              label="webhook地址"
              rules={[{ required: true, message: 'webhook地址不能为空' }]}
              width="md"
              placeholder="飞书机器人 webhook 地址"
              name={['paramsObj', 'addr']}
            />
            <ProFormText
              label="加签密钥"
              width="md"
              placeholder="可选，启用了加签的机器人填写"
              name={['paramsObj', 'secret']}
            />
          </ProForm.Group>
        )}

      <ProForm.Group title="告警模板" />
      <ProFormTextArea
        label="通道告警模板"
        placeholder="留空则使用 alertConf 中该处理器的默认模板"
        name="template"
        fieldProps={{ rows: 6 }}
        extra="支持 Go text/template，可用字段：items[].TaskName / items[].HitRule / relationTaskNames。邮件按 HTML 发送（模板可写 HTML）；企微/钉钉为 markdown，飞书为 text。保存时将用假参数渲染并测试发送。"
      />

      {selectChannelType === 1 && (
        <>
          <ProForm.Group title="通道测试参数" />
          <ProFormText
            label="测试接收人邮箱"
            rules={[
              { required: true, message: '测试接收人邮箱不能为空' },
              { type: 'email', message: '邮箱格式不正确' },
            ]}
            placeholder="请输入测试接收人邮箱"
            name={['testParams', 'email']}
          />
        </>
      )}
    </>
  );
};

export default CreateOrUpdateForm;
