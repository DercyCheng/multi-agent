import React from 'react';
import { Table, Card, Button, Tag, Space, Typography, Avatar, Tooltip } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, MoreOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import type { Agent } from '../../types';

const { Title, Text } = Typography;

const AgentList: React.FC = () => {
  const mockAgents: Agent[] = [
    {
      id: '1',
      name: 'GPT-4 Agent',
      type: 'LLM',
      status: 'active',
      capabilities: ['text-generation', 'reasoning'],
      configuration: {},
      metrics: {
        requestCount: 1250,
        successRate: 98.5,
        averageResponseTime: 1200,
        lastActive: '2024-01-20T10:30:00Z'
      },
      createdAt: '2024-01-15T08:00:00Z',
      updatedAt: '2024-01-20T10:30:00Z'
    },
    {
      id: '2',
      name: 'Claude Assistant',
      type: 'LLM',
      status: 'active',
      capabilities: ['text-generation', 'analysis'],
      configuration: {},
      metrics: {
        requestCount: 856,
        successRate: 97.2,
        averageResponseTime: 980,
        lastActive: '2024-01-20T09:15:00Z'
      },
      createdAt: '2024-01-18T14:30:00Z',
      updatedAt: '2024-01-20T09:15:00Z'
    },
    {
      id: '3',
      name: 'Code Generator',
      type: 'Tool',
      status: 'inactive',
      capabilities: ['code-generation', 'debugging'],
      configuration: {},
      metrics: {
        requestCount: 423,
        successRate: 95.8,
        averageResponseTime: 1500,
        lastActive: '2024-01-19T16:45:00Z'
      },
      createdAt: '2024-01-16T11:20:00Z',
      updatedAt: '2024-01-19T16:45:00Z'
    }
  ];

  const getStatusColor = (status: string) => {
    return status === 'active' ? '#00B341' : '#676767';
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'LLM':
        return '#000000';
      case 'Tool':
        return '#F5A623';
      default:
        return '#676767';
    }
  };

  const columns = [
    {
      title: 'Agent 信息',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Agent) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <Avatar
            size={48}
            style={{
              backgroundColor: getTypeColor(record.type) + '15',
              color: getTypeColor(record.type),
              fontWeight: '600',
              fontSize: '16px'
            }}
          >
            {name.charAt(0)}
          </Avatar>
          <div>
            <Text style={{ fontSize: '16px', fontWeight: '600', color: '#000000', display: 'block' }}>
              {name}
            </Text>
            <Text style={{ fontSize: '14px', color: '#676767' }}>
              ID: {record.id} • 创建于 {new Date(record.createdAt).toLocaleDateString()}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag
          style={{
            backgroundColor: getTypeColor(type) + '15',
            color: getTypeColor(type),
            border: 'none',
            borderRadius: '6px',
            padding: '4px 12px',
            fontWeight: '500'
          }}
        >
          {type}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              backgroundColor: getStatusColor(status)
            }}
          />
          <Text style={{
            fontSize: '14px',
            color: '#000000',
            fontWeight: '500',
            textTransform: 'capitalize'
          }}>
            {status === 'active' ? '活跃' : '非活跃'}
          </Text>
        </div>
      ),
    },
    {
      title: '性能指标',
      key: 'metrics',
      render: (_: any, record: Agent) => (
        <div>
          <Text style={{ fontSize: '14px', color: '#000000', display: 'block' }}>
            成功率: <strong>{record.metrics.successRate}%</strong>
          </Text>
          <Text style={{ fontSize: '12px', color: '#676767' }}>
            请求数: {record.metrics.requestCount} • 响应时间: {record.metrics.averageResponseTime}ms
          </Text>
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Agent) => (
        <Space size="small">
          <Tooltip title={record.status === 'active' ? '暂停' : '启动'}>
            <Button
              type="text"
              icon={record.status === 'active' ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
              style={{
                color: record.status === 'active' ? '#F5A623' : '#00B341',
                border: 'none',
                borderRadius: '8px'
              }}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              style={{ color: '#676767', border: 'none', borderRadius: '8px' }}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              icon={<DeleteOutlined />}
              style={{ color: '#E53E3E', border: 'none', borderRadius: '8px' }}
            />
          </Tooltip>
          <Tooltip title="更多">
            <Button
              type="text"
              icon={<MoreOutlined />}
              style={{ color: '#676767', border: 'none', borderRadius: '8px' }}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '0' }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: '32px'
      }}>
        <div>
          <Title level={1} style={{
            fontSize: '32px',
            fontWeight: '700',
            color: '#000000',
            margin: '0 0 8px 0',
            letterSpacing: '-1px'
          }}>
            Agent 管理
          </Title>
          <Text style={{ fontSize: '16px', color: '#676767' }}>
            管理和监控您的 AI Agent，查看性能指标和运行状态
          </Text>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          size="large"
          style={{
            backgroundColor: '#000000',
            borderColor: '#000000',
            borderRadius: '12px',
            height: '48px',
            fontSize: '16px',
            fontWeight: '500',
            padding: '0 24px'
          }}
        >
          新建 Agent
        </Button>
      </div>

      <Card
        className="uber-card"
        style={{
          overflow: 'hidden'
        }}
        bodyStyle={{ padding: '0' }}
      >
        <Table
          className="uber-table"
          columns={columns}
          dataSource={mockAgents}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `显示 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            style: { padding: '16px 24px' }
          }}
        />
      </Card>
    </div>
  );
};

export default AgentList;