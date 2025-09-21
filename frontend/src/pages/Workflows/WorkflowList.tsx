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
    Typography,
    Row,
    Col,
    Statistic,
    Progress,
    Tooltip,
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
    CopyOutlined,
} from '@ant-design/icons';
import type { Workflow } from '../../types';

const { Title } = Typography;
const { Option } = Select;
// const { TextArea } = Input;

const WorkflowList: React.FC = () => {
    const [workflows, setWorkflows] = useState<Workflow[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchText, setSearchText] = useState('');
    const [statusFilter, setStatusFilter] = useState<string>('all');
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [_editingWorkflow, _setEditingWorkflow] = useState<Workflow | null>(null);
    const [_form] = Form.useForm();

    // 模拟数据
    const mockWorkflows: Workflow[] = [
        {
            id: '1',
            name: '客户服务工作流',
            description: '处理客户咨询和问题解决的完整流程',
            status: 'active',
            steps: [
                {
                    id: 'step1',
                    name: '意图识别',
                    type: 'agent',
                    agentId: '1',
                    configuration: { timeout: 5000 },
                    dependencies: [],
                    timeout: 5000,
                    retryPolicy: { maxRetries: 3, backoffStrategy: 'exponential' },
                },
                {
                    id: 'step2',
                    name: '问题分类',
                    type: 'agent',
                    agentId: '2',
                    configuration: { model: 'classification' },
                    dependencies: ['step1'],
                    timeout: 3000,
                    retryPolicy: { maxRetries: 2, backoffStrategy: 'linear' },
                },
            ],
            metadata: {
                createdBy: 'admin',
                createdAt: '2024-01-15T09:00:00Z',
                updatedAt: '2024-01-20T10:30:00Z',
                version: '1.2.0',
            },
            metrics: {
                totalRuns: 156,
                successfulRuns: 148,
                failedRuns: 8,
                averageExecutionTime: 8.5,
            },
        },
        {
            id: '2',
            name: '数据处理管道',
            description: '自动化数据清洗、转换和分析流程',
            status: 'draft',
            steps: [],
            metadata: {
                createdBy: 'operator',
                createdAt: '2024-01-18T14:00:00Z',
                updatedAt: '2024-01-19T16:45:00Z',
                version: '0.1.0',
            },
            metrics: {
                totalRuns: 0,
                successfulRuns: 0,
                failedRuns: 0,
                averageExecutionTime: 0,
            },
        },
        {
            id: '3',
            name: '内容生成流程',
            description: '基于用户需求生成高质量内容',
            status: 'paused',
            steps: [],
            metadata: {
                createdBy: 'admin',
                createdAt: '2024-01-16T11:00:00Z',
                updatedAt: '2024-01-20T08:15:00Z',
                version: '2.0.1',
            },
            metrics: {
                totalRuns: 89,
                successfulRuns: 82,
                failedRuns: 7,
                averageExecutionTime: 12.3,
            },
        },
    ];

    useEffect(() => {
        loadWorkflows();
    }, []);

    const loadWorkflows = async () => {
        setLoading(true);
        try {
            await new Promise(resolve => setTimeout(resolve, 1000));
            setWorkflows(mockWorkflows);
        } catch (error) {
            message.error('加载工作流列表失败');
        } finally {
            setLoading(false);
        }
    };

    const handleExecuteWorkflow = async (workflow: Workflow) => {
        try {
            message.success(`工作流 ${workflow.name} 开始执行`);
            setWorkflows(prev =>
                prev.map(w =>
                    w.id === workflow.id
                        ? {
                            ...w,
                            status: 'active' as const,
                            metrics: {
                                ...w.metrics,
                                totalRuns: w.metrics.totalRuns + 1
                            }
                        }
                        : w
                )
            );
        } catch (error) {
            message.error('执行工作流失败');
        }
    };

    const handlePauseWorkflow = async (workflow: Workflow) => {
        try {
            message.success(`工作流 ${workflow.name} 已暂停`);
            setWorkflows(prev =>
                prev.map(w =>
                    w.id === workflow.id ? { ...w, status: 'paused' as const } : w
                )
            );
        } catch (error) {
            message.error('暂停工作流失败');
        }
    };

    const handleDeleteWorkflow = (workflow: Workflow) => {
        Modal.confirm({
            title: '确认删除',
            content: `确定要删除工作流 "${workflow.name}" 吗？此操作不可恢复。`,
            okText: '删除',
            okType: 'danger',
            cancelText: '取消',
            onOk: async () => {
                try {
                    message.success(`工作流 ${workflow.name} 删除成功`);
                    setWorkflows(prev => prev.filter(w => w.id !== workflow.id));
                } catch (error) {
                    message.error('删除工作流失败');
                }
            },
        });
    };

    const handleCloneWorkflow = async (workflow: Workflow) => {
        try {
            const clonedWorkflow: Workflow = {
                ...workflow,
                id: Date.now().toString(),
                name: `${workflow.name} (副本)`,
                status: 'draft',
                metadata: {
                    ...workflow.metadata,
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                    version: '1.0.0',
                },
                metrics: {
                    totalRuns: 0,
                    successfulRuns: 0,
                    failedRuns: 0,
                    averageExecutionTime: 0,
                },
            };
            message.success(`工作流 ${workflow.name} 克隆成功`);
            setWorkflows(prev => [clonedWorkflow, ...prev]);
        } catch (error) {
            message.error('克隆工作流失败');
        }
    };

    const getStatusTag = (status: string) => {
        const statusConfig = {
            active: { color: 'success', text: '运行中' },
            draft: { color: 'default', text: '草稿' },
            paused: { color: 'warning', text: '已暂停' },
            completed: { color: 'blue', text: '已完成' },
            failed: { color: 'error', text: '失败' },
        };
        const config = statusConfig[status as keyof typeof statusConfig];
        return <Tag color={config.color}>{config.text}</Tag>;
    };

    const columns = [
        {
            title: '工作流名称',
            dataIndex: 'name',
            key: 'name',
            render: (name: string, workflow: Workflow) => (
                <div>
                    <div style={{ fontWeight: 500 }}>{name}</div>
                    <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                        v{workflow.metadata.version}
                    </div>
                </div>
            ),
        },
        {
            title: '描述',
            dataIndex: 'description',
            key: 'description',
            ellipsis: {
                showTitle: false,
            },
            render: (description: string) => (
                <Tooltip placement="topLeft" title={description}>
                    {description}
                </Tooltip>
            ),
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => getStatusTag(status),
        },
        {
            title: '步骤数',
            dataIndex: 'steps',
            key: 'stepCount',
            render: (steps: any[]) => steps.length,
        },
        {
            title: '执行统计',
            key: 'metrics',
            render: (_: any, workflow: Workflow) => {
                const successRate = workflow.metrics.totalRuns > 0
                    ? (workflow.metrics.successfulRuns / workflow.metrics.totalRuns * 100).toFixed(1)
                    : 0;

                return (
                    <div>
                        <div>总运行: {workflow.metrics.totalRuns}</div>
                        <div style={{ fontSize: 12, color: '#666' }}>
                            成功率: {successRate}%
                        </div>
                        {workflow.metrics.totalRuns > 0 && (
                            <Progress
                                percent={Number(successRate)}
                                size="small"
                                showInfo={false}
                                style={{ marginTop: 4 }}
                            />
                        )}
                    </div>
                );
            },
        },
        {
            title: '平均执行时间',
            dataIndex: ['metrics', 'averageExecutionTime'],
            key: 'executionTime',
            render: (time: number) => time > 0 ? `${time}s` : '-',
        },
        {
            title: '创建时间',
            dataIndex: ['metadata', 'createdAt'],
            key: 'createdAt',
            render: (date: string) => new Date(date).toLocaleDateString(),
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, workflow: Workflow) => (
                <Space size="small">
                    <Tooltip title="查看详情">
                        <Button type="text" icon={<EyeOutlined />} size="small" />
                    </Tooltip>
                    <Tooltip title="编辑">
                        <Button type="text" icon={<SettingOutlined />} size="small" />
                    </Tooltip>
                    <Tooltip title="克隆">
                        <Button
                            type="text"
                            icon={<CopyOutlined />}
                            size="small"
                            onClick={() => handleCloneWorkflow(workflow)}
                        />
                    </Tooltip>
                    {workflow.status === 'active' ? (
                        <Tooltip title="暂停">
                            <Button
                                type="text"
                                icon={<PauseCircleOutlined />}
                                size="small"
                                onClick={() => handlePauseWorkflow(workflow)}
                            />
                        </Tooltip>
                    ) : (
                        <Tooltip title="执行">
                            <Button
                                type="text"
                                icon={<PlayCircleOutlined />}
                                size="small"
                                onClick={() => handleExecuteWorkflow(workflow)}
                            />
                        </Tooltip>
                    )}
                    <Tooltip title="删除">
                        <Button
                            type="text"
                            icon={<DeleteOutlined />}
                            size="small"
                            danger
                            onClick={() => handleDeleteWorkflow(workflow)}
                        />
                    </Tooltip>
                </Space>
            ),
        },
    ];

    const filteredWorkflows = workflows.filter(workflow => {
        const matchesSearch = workflow.name.toLowerCase().includes(searchText.toLowerCase()) ||
            workflow.description.toLowerCase().includes(searchText.toLowerCase());
        const matchesStatus = statusFilter === 'all' || workflow.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const stats = {
        total: workflows.length,
        active: workflows.filter(w => w.status === 'active').length,
        draft: workflows.filter(w => w.status === 'draft').length,
        paused: workflows.filter(w => w.status === 'paused').length,
    };

    return (
        <div className="fade-in">
            <Title level={2}>工作流管理</Title>

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
                        <Statistic title="草稿" value={stats.draft} valueStyle={{ color: '#999' }} />
                    </Card>
                </Col>
                <Col span={6}>
                    <Card>
                        <Statistic title="已暂停" value={stats.paused} valueStyle={{ color: '#faad14' }} />
                    </Card>
                </Col>
            </Row>

            <Card>
                <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Space>
                        <Input
                            placeholder="搜索工作流名称或描述"
                            prefix={<SearchOutlined />}
                            value={searchText}
                            onChange={(e: any) => setSearchText(e.target.value)}
                            style={{ width: 300 }}
                        />
                        <Select
                            value={statusFilter}
                            onChange={setStatusFilter}
                            style={{ width: 120 }}
                        >
                            <Option value="all">全部状态</Option>
                            <Option value="active">运行中</Option>
                            <Option value="draft">草稿</Option>
                            <Option value="paused">已暂停</Option>
                            <Option value="completed">已完成</Option>
                            <Option value="failed">失败</Option>
                        </Select>
                    </Space>

                    <Space>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadWorkflows}
                            loading={loading}
                        >
                            刷新
                        </Button>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={() => setIsModalVisible(true)}
                        >
                            创建工作流
                        </Button>
                    </Space>
                </div>

                <Table
                    columns={columns}
                    dataSource={filteredWorkflows}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        total: filteredWorkflows.length,
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total: any, range: any) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
                    }}
                />
            </Card>

            <Modal
                title="创建工作流"
                open={isModalVisible}
                onCancel={() => setIsModalVisible(false)}
                footer={null}
                width={800}
            >
                <div style={{ textAlign: 'center', padding: '40px 0' }}>
                    <Title level={4}>工作流创建向导</Title>
                    <p style={{ color: '#666', marginBottom: 32 }}>
                        选择一种方式来创建您的工作流
                    </p>

                    <Row gutter={[24, 24]}>
                        <Col span={8}>
                            <Card
                                hoverable
                                onClick={() => message.info('功能开发中')}
                                style={{ textAlign: 'center' }}
                            >
                                <PlusOutlined style={{ fontSize: 32, color: '#1890ff', marginBottom: 16 }} />
                                <div style={{ fontWeight: 500 }}>从模板创建</div>
                                <div style={{ fontSize: 12, color: '#666', marginTop: 8 }}>
                                    使用预定义的工作流模板
                                </div>
                            </Card>
                        </Col>
                        <Col span={8}>
                            <Card
                                hoverable
                                onClick={() => message.info('功能开发中')}
                                style={{ textAlign: 'center' }}
                            >
                                <SettingOutlined style={{ fontSize: 32, color: '#52c41a', marginBottom: 16 }} />
                                <div style={{ fontWeight: 500 }}>可视化编辑</div>
                                <div style={{ fontSize: 12, color: '#666', marginTop: 8 }}>
                                    通过拖拽方式创建工作流
                                </div>
                            </Card>
                        </Col>
                        <Col span={8}>
                            <Card
                                hoverable
                                onClick={() => message.info('功能开发中')}
                                style={{ textAlign: 'center' }}
                            >
                                <CopyOutlined style={{ fontSize: 32, color: '#722ed1', marginBottom: 16 }} />
                                <div style={{ fontWeight: 500 }}>导入配置</div>
                                <div style={{ fontSize: 12, color: '#666', marginTop: 8 }}>
                                    从JSON配置文件导入
                                </div>
                            </Card>
                        </Col>
                    </Row>
                </div>
            </Modal>
        </div>
    );
};

export default WorkflowList;