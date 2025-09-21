import React from 'react';
import { Table, Card, Button, Tag, Space, Typography } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { Agent } from '../../types';

const { Title } = Typography;

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
    }
  ];

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="blue">{type}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '活跃' : '非活跃'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space size="middle">
          <Button type="link" icon={<EditOutlined />}>编辑</Button>
          <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '16px' }}>
        <Title level={2}>Agent管理</Title>
        <Button type="primary" icon={<PlusOutlined />}>
          新建Agent
        </Button>
      </div>
      <Card>
        <Table
          columns={columns}
          dataSource={mockAgents}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  );
};

export default AgentList;