import { Column } from '@ant-design/plots';
import { PageContainer, StatisticCard } from '@ant-design/pro-components';
import { Card, Col, Row, Table, Tag, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useEffect, useMemo, useState } from 'react';
import { monitorHomeCount } from '@/services/ant-design-pro/monitor.event';

const DEAL_STATUS_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '待处理', color: 'red' },
  2: { text: '处理中', color: 'orange' },
  3: { text: '已完成', color: 'green' },
  4: { text: '已忽略', color: 'gray' },
};

const LEVEL_MAP: Record<number, { text: string; color: string }> = {
  '-1': { text: '正常', color: 'default' },
  0: { text: '严重', color: 'red' },
  1: { text: '高', color: 'orange' },
  2: { text: '中', color: 'blue' },
  3: { text: '低', color: 'gray' },
};

const Welcome: React.FC = () => {
  const [count, setCount] = useState<API.MonitorHomeCount>({});
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    monitorHomeCount()
      .then((res) => {
        setCount(res.data || {});
      })
      .catch(() => setCount({}))
      .finally(() => setLoading(false));
  }, []);

  // ── 告警级别分布（柱状图） ──
  const alertLevelData: Array<{ level: string; value: number }> =
    useMemo(() => {
      const dist = count.alertLevelDistribution || {};
      return Object.entries(dist).map(([key, val]) => ({
        level: LEVEL_MAP[parseInt(key, 10)]?.text || `级别${key}`,
        value: val,
      }));
    }, [count.alertLevelDistribution]);

  const barConfig = {
    data: alertLevelData,
    title: '告警级别分布（近一月）',
    xField: 'level',
    yField: 'value',
    color: ['#52c41a', '#faad14', '#ff4d4f'],
    label: { position: 'inside' },
    xAxis: { label: { autoRotate: true } },
  };

  // ── 最近告警事件表格 ──
  type RecentEventItem = NonNullable<
    API.MonitorHomeCount['recentEvents']
  >[number];
  const recentEventColumns: ColumnsType<RecentEventItem> = [
    {
      title: '任务名称',
      dataIndex: 'taskName',
      key: 'taskName',
      ellipsis: true,
      width: 150,
    },
    {
      title: '时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 160,
    },
    {
      title: '告警信息',
      dataIndex: 'alertMsg',
      key: 'alertMsg',
      ellipsis: true,
    },
    {
      title: '处理状态',
      dataIndex: 'dealStatus',
      key: 'dealStatus',
      width: 90,
      render: (status: number) => {
        const info = DEAL_STATUS_MAP[status] || {
          text: '未知',
          color: 'default',
        };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '告警等级',
      dataIndex: 'eventLevel',
      key: 'eventLevel',
      width: 80,
      render: (level: number) => {
        const info =
          LEVEL_MAP[level] ??
          (LEVEL_MAP[`${level}`] as (typeof LEVEL_MAP)[number]);
        if (!info) return '-';
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
  ];

  const recentEvents = count.recentEvents || [];

  return (
    <PageContainer>
      {/* 平台介绍 */}
      <Card style={{ marginBottom: 16 }}>
        <Typography.Title level={3}>
          系统工作流程
        </Typography.Title>
        <Typography.Paragraph>
          指标采集 → 样本基线 → 规则检测 → 事件通知， 请从左侧菜单进入业务模块。
        </Typography.Paragraph>
      </Card>

      {/* ── 核心指标 ── */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '监控任务', value: count.taskCount || 0 }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '监控面板', value: count.dashboardCount || 0 }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '数据源', value: count.databaseCount || 0 }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{
              title: '告警通道',
              value: count.alertChannelCount || 0,
            }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '告警分组', value: count.alertGroupCount || 0 }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '告警事件（近一月）', value: count.eventCount || 0 }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{
              title: '待处理（近一月）',
              value: count.pendingEvents || 0,
            }}
          />
        </Col>
        <Col xs={12} sm={8} md={6} lg={6} xl={6}>
          <StatisticCard
            statistic={{ title: '处理中（近一月）', value: count.processingEvents || 0 }}
          />
        </Col>
      </Row>

      {/* ── 图表行：告警级别分布 + 最近告警事件 ── */}
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Card title="告警级别分布（近一月）" styles={{ body: { height: 304 } }}>
            {alertLevelData.length > 0 ? (
              <Column {...barConfig} height={260} />
            ) : (
              <div
                style={{
                  textAlign: 'center',
                  padding: '40px 0',
                  color: '#999',
                }}
              >
                暂无告警数据
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card
            title="最近告警事件"
            styles={{
              body: { height: 304, display: 'flex', flexDirection: 'column' },
            }}
          >
            <Table
              columns={recentEventColumns}
              dataSource={recentEvents}
              rowKey="id"
              loading={loading}
              pagination={false}
              size="small"
              style={{ flex: 1, minHeight: 0 }}
              locale={{ emptyText: '暂无事件' }}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Welcome;
