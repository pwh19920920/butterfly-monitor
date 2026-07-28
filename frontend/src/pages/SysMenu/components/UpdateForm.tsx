import { LoadingOutlined } from '@ant-design/icons';
import type { ProColumns } from '@ant-design/pro-components';
import {
  EditableProTable,
  ProForm,
  ProFormDigit,
  ProFormText,
} from '@ant-design/pro-components';
import { Col, Form, Row, Spin, TreeSelect } from 'antd';
import type { DataNode } from 'rc-tree-select/lib/interface';
import React, { useEffect, useState } from 'react';
import { sysMenuQuery } from '@/services/ant-design-pro/sys.menu';
import { HttpMethodEnum } from '@/services/ant-design-pro/enum';

const columns: ProColumns<API.SysMenuOption>[] = [
  {
    title: '操作名称',
    width: '140px',
    dataIndex: 'name',
  },
  {
    title: '操作权限',
    width: '200px',
    dataIndex: 'value',
  },
  {
    title: 'URL方法',
    dataIndex: 'method',
    width: '120px',
    valueEnum: HttpMethodEnum,
  },
  {
    title: 'URL地址',
    dataIndex: 'path',
  },
  {
    title: '操作',
    valueType: 'option',
    width: '50px',
  },
];

export type CreateOrUpdateFormProps = {
  initOptionsValues: API.SysMenuOption[];
};

const CreateOrUpdateForm: React.FC<CreateOrUpdateFormProps> = (props) => {
  const [treeData, setTreeData] = useState<DataNode[]>([]);

  const [editableKeys, setEditableRowKeys] = useState<React.Key[]>(() =>
    props.initOptionsValues && props.initOptionsValues.length > 0
      ? props.initOptionsValues.map((item) => item.id)
      : [],
  );

  const reloadTreeData = async () => {
    const resp = await sysMenuQuery();
    setTreeData(
      recursionAssignment([
        {
          id: 0,
          name: '顶级菜单',
          code: '',
          path: '/',
          options: [],
          children: resp.data,
          routes: [],
        },
      ]),
    );
  };

  // DataNode[]
  function recursionAssignment(resp: API.SysMenu[]): DataNode[] {
    if (resp.length === 0) {
      return [];
    }

    return resp.map((item) => {
      let children: DataNode[] = [];
      if (item.children) {
        children = recursionAssignment(item.children);
      }
      return {
        value: `${item.id}`,
        title: item.name + ': ' + item.path,
        children,
      };
    });
  }

  useEffect(() => {
    reloadTreeData().then(() => {});
  }, []);

  if (treeData.length === 0) {
    return (
      <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
    );
  }

  return (
    <>
      <Row gutter={24}>
        <Col span={12}>
          <ProFormText
            label="名称"
            rules={[
              {
                required: true,
                message: '名称不能为空',
              },
            ]}
            placeholder="请输入名称"
            name="name"
          />
        </Col>
        <Col span={12}>
          <ProFormText
            label="菜单代码"
            rules={[
              {
                required: true,
                message: '菜单代码不能为空',
              },
            ]}
            placeholder="请输入菜单代码"
            name="code"
          />
        </Col>
        <Col span={12}>
          <ProFormText
            label="路径"
            rules={[
              {
                required: true,
                message: '路径不能为空',
              },
            ]}
            placeholder="请输入路径"
            name="path"
          />
        </Col>
        <Col span={12}>
          <ProFormText
            label="图标"
            rules={[
              {
                required: true,
                message: '图标不能为空',
              },
            ]}
            placeholder="请输入图标"
            name="icon"
          />
        </Col>
        <Col span={12}>
          <ProFormDigit
            label="排序"
            rules={[
              {
                required: true,
                message: '排序不能为空',
              },
            ]}
            placeholder="请输入排序, 大的在前"
            name="sort"
          />
        </Col>
        <Col span={12}>
          <ProFormText label="组件" placeholder="请输入组件" name="component" />
        </Col>

        <Col span={24}>
          <Form.Item label="上级菜单" name="parent" style={{ width: '100%' }}>
            <TreeSelect
              allowClear
              key="id"
              treeDefaultExpandAll
              placeholder="请选择上级菜单"
              treeData={treeData}
            />
          </Form.Item>
        </Col>
      </Row>

      <ProForm.Item name="options" trigger="onValuesChange">
        <EditableProTable<API.SysMenuOption>
          rowKey="id"
          toolBarRender={false}
          columns={columns}
          recordCreatorProps={{
            newRecordType: 'dataSource',
            position: 'top',
            record: (record: any): API.SysMenuOption => {
              return { ...record, id: `${new Date().getMilliseconds()}` };
            },
          }}
          editable={{
            type: 'multiple',
            editableKeys,
            onChange: setEditableRowKeys,
            actionRender: (row, _, dom) => {
              return [dom.delete];
            },
          }}
        />
      </ProForm.Item>
    </>
  );
};

export default CreateOrUpdateForm;
