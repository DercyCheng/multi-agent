import React, { useEffect, useState } from 'react';
import {
    Table,
    Button,
    Space,
    Tag,
    Card,
    Input,
    Select,
    Modal,
    Form,
    message,
    Tooltip,
    Typography,
    Row,
    Col,
    Statistic,
} from 'antd';
import {
    PlusOutlined,
    SearchOutlined,
    PlayCircleOutlined,
    PauseCircleOutlined,
    DeleteOutlined,
    SettingOutlined,
    EyeOutlined,
    ReloadOutlined,
} from '@ant-design/icons';
import type { Agent } from '../../types';

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const AgentList: React.FC = () => {
    const [agents, setAgents] = useState<Agent[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchText, setSearchText] = useState('');
    const [statusFilter, setStatusFilter] = useState<string>('all');
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingAgent, setEditingAgent] = useState<Agent | null>(null);
    const [form] = Form.useForm();

    // 模拟数据
    const mockAgents: Agent[] = [
        {
            id: '1',
            name: 'GPT-4 对话Agent',
            type: 'LLM',
            status: 'active',
            capabilities: ['自然语言理解', '对话生成', '问答'],
            configuration: { model: 'gpt-4', temperature: 0.7 },
            metrics: {
                requestCount: 1247,
                successRate: 98.5,
                averageResponseTime: 1.2,
                lastActive: '2024-01-20T10:30:00Z',
            },
            createdAt: '2024-01-15T09:00:00Z',
            updatedAt: '2024-01-20T10:30:00Z',
        },
        {
            id: '2',
            name: '网络搜索Agent',
            type: 'Tool',
            status: 'active',
            capabilities: ['网络搜索', '信息检索', '内容摘要'],
            configuration: { searchEngine: 'google', maxResults: 10 },
            metrics: {
                requestCount: 856,
                successRate: 95.2,
                averageResponseTime: 2.8,
                lastActive: '2024-01-20T10:25:00Z',
            },
            createdAt: '2024-01-16T14:00:00Z',
            updatedAt: '2024-01-20T10:25:00Z',
        },
        {
            id: '3',
            name: '代码分析Agent',
            type: 'Code',
            status: 'inactive',
            capabilities: ['代码分析', '语法检查', '性能优化建议'],
            configuration: { languages: ['python', 'javascript', 'go'] },
            metrics: {
                requestCount: 342,
                successRate: 89.1,
                averageResponseTime: 5.6,
                lastActive: '2024-01-19T16:45:00Z',
            },
            createdAt: '2024-01-17T11:00:00Z',
            updatedAt: '2024-01-19T16:45:00Z',
        },
        {
            id: '4',
            name: '数据处理Agent',
            type: 'Data',
            status: 'error',
            capabilities: ['数据清洗', '格式转换', '统计分析'],
            configuration: { maxFileSize: '100MB', formats: ['csv', 'json', 'xml'] },
            metrics: {
                requestCount: 123,
                successRate: 78.5,
                averageResponseTime: 8.2,
                lastActive: '2024-01-20T08:15:00Z',
            },
            createdAt: '2024-01-18T13:00:00Z',
            updatedAt: '2024-01-20T08:15:00Z',
        },
    ];

    useEffect(() => {
        loadAgents();
    }, []);

    const loadAgents = async () => {
        setLoading(true);
        try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000));
            setAgents(mockAgents);
        } catch (error) {
            message.error('加载Agent列表失败');
        } finally {
            setLoading(false);
        }
    };

    const handleStartAgent = async (agent: Agent) => {
        try {
            message.success(`Agent ${agent.name} 启动成功`);
            // 更新状态
            setAgents(prev =>
                prev.map(a =>
                    a.id === agent.id ? { ...a, status: 'active' as const } : a
                )
            );
        } catch (error) {
            message.error('启动Agent失败');
        }
    };

    const handleStopAgent = async (agent: Agent) => {
        try {
            message.success(`Agent ${agent.name} 已停止`);
            setAgents(prev =>
                prev.map(a =>
                    a.id === agent.id ? { ...a, status: 'inactive' as const } : a
                )
            );
        } catch (error) {
            message.error('停止Agent失败');
        }
    };

    const handleDeleteAgent = (agent: Agent) => {
        Modal.confirm({
            title: '确认删除',
            content: `确定要删除Agent "${agent.name}" 吗？此操作不可恢复。`,
            okText: '删除',
            okType: 'danger',
            cancelText: '取消',
            onOk: async () => {
                try {
                    message.success(`Agent ${agent.name} 删除成功`);
                    setAgents(prev => prev.filter(a => a.id !== agent.id));
                } catch (error) {
                    message.error('删除Agent失败');
                }
            },
        });
    };

    const handleCreateAgent = () => {
        setEditingAgent(null);
        form.resetFields();
        setIsModalVisible(true);
    };

    const handleEditAgent = (agent: Agent) => {
        setEditingAgent(agent);
        form.setFieldsValue({
            name: agent.name,
            type: agent.type,
            capabilities: agent.capabilities.join(', '),
            configuration: JSON.stringify(agent.configuration, null, 2),
        });
        setIsModalVisible(true);
    };

    const handleModalOk = async () => {
        try {
            const values = await form.validateFields();
            const agentData = {
                ...values,
                capabilities: values.capabilities.split(',').map((s: string) => s.trim()),
                configuration: JSON.parse(values.configuration),
            };

            if (editingAgent) {
                message.success('Agent更新成功');
                setAgents(prev =>
                    prev.map(a =>
                        a.id === editingAgent.id
                            ? { ...a, ...agentData, updatedAt: new Date().toISOString() }
                            : a
                    )
                );
            } else {
                const newAgent: Agent = {
                    id: Date.now().toString(),
                    ...agentData,
                    status: 'inactive' as const,
                    metrics: {
                        requestCount: 0,
                        successRate: 0,
                        averageResponseTime: 0,
                        lastActive: new Date().toISOString(),
                    },
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                };
                message.success('Agent创建成功');
                setAgents(prev => [...prev, newAgent]);
            }

            setIsModalVisible(false);
        } catch (error) {
            message.error('保存失败，请检查输入内容');
        }
    };

    const getStatusTag = (status: string) => {
        const statusConfig = {
            active: { color: 'success', text: '运行中' },
            inactive: { color: 'default', text: '已停止' },
            error: { color: 'error', text: '错误' },
            pending: { color: 'processing', text: '启动中' },
        };
        const config = statusConfig[status as keyof typeof statusConfig];
        return <Tag color={config.color}>{config.text}</Tag>;
    };

    const columns = [
        {
            title: 'Agent名称',
            dataIndex: 'name',
            key: 'name',
            filterable: true,
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            render: (type: string) => <Tag>{type}</Tag>,
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => getStatusTag(status),
        },
        {
            title: '能力',
            dataIndex: 'capabilities',
            key: 'capabilities',
            render: (capabilities: string[]) => (
                <div>
                    {capabilities.slice(0, 2).map(cap => (
                        <Tag key={cap}>{cap}</Tag>
                    ))}
                    {capabilities.length > 2 && (
                        <Tooltip title={capabilities.slice(2).join(', ')}>
                            <Tag>+{capabilities.length - 2}</Tag>
                        </Tooltip>
                    )}
                </div>
            ),
        },
        {
            title: '请求数',
            dataIndex: ['metrics', 'requestCount'],
            key: 'requestCount',
            render: (count: number) => count.toLocaleString(),
        },
        {
            title: '成功率',
            dataIndex: ['metrics', 'successRate'],
            key: 'successRate',
            render: (rate: number) => `${rate}%`,
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, agent: Agent) => (
                <Space size={8}>
                    <Tooltip title="查看详情">
                        <Button type="text" icon={<EyeOutlined />} />
                    </Tooltip>
                    <Tooltip title="配置">
                        <Button
                            type="text"
                            icon={<SettingOutlined />}
                            size="small"
                            onClick={() => handleEditAgent(agent)}
                        />
                    </Tooltip>
                    {agent.status === 'active' ? (
                        <Tooltip title="停止">
                            <Button
                                type="text"
                                icon={<PauseCircleOutlined />}
                                size="small"
                                onClick={() => handleStopAgent(agent)}
                            />
                        </Tooltip>
                    ) : (
                        <Tooltip title="启动">
                            <Button
                                type="text"
                                icon={<PlayCircleOutlined />}
                                size="small"
                                onClick={() => handleStartAgent(agent)}
                            />
                        </Tooltip>
                    )}
                    <Tooltip title="删除">
                        <Button
                            type="text"
                            icon={<DeleteOutlined />}
                            size="small"
                            danger
                            onClick={() => handleDeleteAgent(agent)}
                        />
                    </Tooltip>
                </Space>
            ),
        },
    ];

    const filteredAgents = agents.filter(agent => {
        const matchesSearch = agent.name.toLowerCase().includes(searchText.toLowerCase()) ||
            agent.type.toLowerCase().includes(searchText.toLowerCase());
        const matchesStatus = statusFilter === 'all' || agent.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const stats = {
        total: agents.length,
        active: agents.filter(a => a.status === 'active').length,
        inactive: agents.filter(a => a.status === 'inactive').length,
        error: agents.filter(a => a.status === 'error').length,
    };

    return (
        <div className="fade-in">
            <Title level={2}>Agent 管理</Title>

            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col span={6}>
                    <Card>
                        <Statistic title="总计" value={stats.total} />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic title="运行中" value={stats.active} valueStyle={{ color: '#3f8600' }} />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic title="已停止" value={stats.inactive} valueStyle={{ color: '#999' }} />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic title="错误" value={stats.error} valueStyle={{ color: '#cf1322' }} />
                    </Card>
                </Col>
            </Row>

            <Card>
                <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space>
                        <Input
                            placeholder="搜索Agent名称或类型"
                            prefix={<SearchOutlined />}
                            value={searchText}
                            onChange={(e) => setSearchText(e.target.value)}
                            style={{ width: 300 }}
                        />
                        <Select
                            value={statusFilter}
                            onChange={setStatusFilter}
                            style={{ width: 120 }}
                        >
                            <Option value="all">全部状态</Option>
                            <Option value="active">运行中</Option>
                            <Option value="inactive">已停止</Option>
                            <Option value="error">错误</Option>
                        </Select>
                    </Space>

                    <Space>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadAgents}
                            loading={loading}
                        >
                            刷新
                        </Button>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={handleCreateAgent}
                        >
                            创建Agent
                        </Button>
                    </Space>
                </div>

                <Table
                    columns={columns}
                    dataSource={filteredAgents}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        total: filteredAgents.length,
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
                    }}
                />
            </Card>

            <Modal
                title={editingAgent ? '编辑Agent' : '创建Agent'}
                open={isModalVisible}
                onOk={handleModalOk}
                onCancel={() => setIsModalVisible(false)}
                width={600}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        name="name"
                        label="Agent名称"
                        rules={[{ required: true, message: '请输入Agent名称' }]}
                    >
                        <Input placeholder="请输入Agent名称" />
                    </Form.Item>

                    <Form.Item
                        name="type"
                        label="Agent类型"
                        rules={[{ required: true, message: '请选择Agent类型' }]}
                    >
                        <Select placeholder="请选择Agent类型">
                            <Option value="LLM">LLM</Option>
                            <Option value="Tool">工具</Option>
                            <Option value="Code">代码</Option>
                            <Option value="Data">数据</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="capabilities"
                        label="能力描述"
                        rules={[{ required: true, message: '请输入能力描述' }]}
                    >
                        <Input placeholder="用逗号分隔，例如：自然语言理解, 对话生成" />
                    </Form.Item>

                    <Form.Item
                        name="configuration"
                        label="配置信息"
                        rules={[
                            { required: true, message: '请输入配置信息' },
                            {
                                validator: (_, value) => {
                                    try {
                                        JSON.parse(value);
                                        return Promise.resolve();
                                    } catch {
                                        return Promise.reject(new Error('请输入有效的JSON格式'));
                                    }
                                },
                            },
                        ]}
                    >
                        <TextArea
                            rows={6}
                            placeholder='{"model": "gpt-4", "temperature": 0.7}'
                        />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default AgentList;