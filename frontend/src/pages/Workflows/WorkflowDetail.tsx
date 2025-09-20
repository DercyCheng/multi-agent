import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Card,
    Row,
    Col,
    Descriptions,
    Tag,
    Button,
    Space,
    Typography,
    Tabs,
    Table,
    Timeline,
    Steps,
    Progress,
    Alert,
} from 'antd';
import {
    ArrowLeftOutlined,
    PlayCircleOutlined,
    PauseCircleOutlined,
    SettingOutlined,
    ReloadOutlined,
    ClockCircleOutlined,
} from '@ant-design/icons';
import type { Workflow, WorkflowExecution } from '../../types';

const { Title } = Typography;
const { TabPane } = Tabs;
const { Step } = Steps;

const WorkflowDetail: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [workflow, setWorkflow] = useState<Workflow | null>(null);
    const [executions, setExecutions] = useState<WorkflowExecution[]>([]);
    const [loading, setLoading] = useState(true);

    // 模拟工作流详情数据
    const mockWorkflow: Workflow = {
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
                configuration: { timeout: 5000, model: 'intent-classifier' },
                dependencies: [],
                timeout: 5000,
                retryPolicy: { maxRetries: 3, backoffStrategy: 'exponential' },
            },
            {
                id: 'step2',
                name: '问题分类',
                type: 'agent',
                agentId: '2',
                configuration: { model: 'classification', confidence: 0.8 },
                dependencies: ['step1'],
                timeout: 3000,
                retryPolicy: { maxRetries: 2, backoffStrategy: 'linear' },
            },
            {
                id: 'step3',
                name: '响应生成',
                type: 'agent',
                agentId: '3',
                configuration: { temperature: 0.7, maxTokens: 500 },
                dependencies: ['step2'],
                timeout: 8000,
                retryPolicy: { maxRetries: 1, backoffStrategy: 'linear' },
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
    };

    const mockExecutions: WorkflowExecution[] = [
        {
            id: 'exec1',
            workflowId: '1',
            status: 'completed',
            startTime: '2024-01-20T10:30:00Z',
            endTime: '2024-01-20T10:30:08Z',
            steps: [
                {
                    stepId: 'step1',
                    status: 'completed',
                    startTime: '2024-01-20T10:30:00Z',
                    endTime: '2024-01-20T10:30:02Z',
                    result: { intent: 'question', confidence: 0.95 },
                    retryCount: 0,
                },
                {
                    stepId: 'step2',
                    status: 'completed',
                    startTime: '2024-01-20T10:30:02Z',
                    endTime: '2024-01-20T10:30:05Z',
                    result: { category: 'technical', subcategory: 'api' },
                    retryCount: 0,
                },
                {
                    stepId: 'step3',
                    status: 'completed',
                    startTime: '2024-01-20T10:30:05Z',
                    endTime: '2024-01-20T10:30:08Z',
                    result: { response: '根据您的问题...' },
                    retryCount: 0,
                },
            ],
            metadata: { userId: 'user123', sessionId: 'session456' },
        },
        {
            id: 'exec2',
            workflowId: '1',
            status: 'running',
            startTime: '2024-01-20T10:35:00Z',
            steps: [
                {
                    stepId: 'step1',
                    status: 'completed',
                    startTime: '2024-01-20T10:35:00Z',
                    endTime: '2024-01-20T10:35:01Z',
                    result: { intent: 'complaint', confidence: 0.88 },
                    retryCount: 0,
                },
                {
                    stepId: 'step2',
                    status: 'running',
                    startTime: '2024-01-20T10:35:01Z',
                    retryCount: 0,
                },
                {
                    stepId: 'step3',
                    status: 'pending',
                    retryCount: 0,
                },
            ],
            metadata: { userId: 'user789', sessionId: 'session012' },
        },
    ];

    useEffect(() => {
        loadWorkflowDetail();
    }, [id]);

    const loadWorkflowDetail = async () => {
        setLoading(true);
        try {
            await new Promise(resolve => setTimeout(resolve, 1000));
            setWorkflow(mockWorkflow);
            setExecutions(mockExecutions);
        } catch (error) {
            console.error('Failed to load workflow detail:', error);
        } finally {
            setLoading(false);
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

    const getStepStatus = (status: string) => {
        const statusMap = {
            pending: 'wait',
            running: 'process',
            completed: 'finish',
            failed: 'error',
            skipped: 'wait',
        };
        return statusMap[status as keyof typeof statusMap] || 'wait';
    };

    const executionColumns = [
        {
            title: '执行ID',
            dataIndex: 'id',
            key: 'id',
            width: 120,
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => {
                const statusConfig = {
                    running: { color: 'processing', text: '运行中' },
                    completed: { color: 'success', text: '已完成' },
                    failed: { color: 'error', text: '失败' },
                    cancelled: { color: 'default', text: '已取消' },
                };
                const config = statusConfig[status as keyof typeof statusConfig];
                return <Tag color={config.color}>{config.text}</Tag>;
            },
        },
        {
            title: '开始时间',
            dataIndex: 'startTime',
            key: 'startTime',
            render: (time: string) => new Date(time).toLocaleString(),
        },
        {
            title: '结束时间',
            dataIndex: 'endTime',
            key: 'endTime',
            render: (time: string) => time ? new Date(time).toLocaleString() : '-',
        },
        {
            title: '执行时长',
            key: 'duration',
            render: (_: any, execution: WorkflowExecution) => {
                if (!execution.endTime) return '-';
                const duration = (new Date(execution.endTime).getTime() - new Date(execution.startTime).getTime()) / 1000;
                return `${duration}s`;
            },
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, execution: WorkflowExecution) => (
                <Space size="small">
                    <Button type="text" size="small">查看详情</Button>
                    {execution.status === 'running' && (
                        <Button type="text" size="small" danger>取消</Button>
                    )}
                </Space>
            ),
        },
    ];

    if (loading) {
        return <div>加载中...</div>;
    }

    if (!workflow) {
        return <div>工作流未找到</div>;
    }

    return (
        <div className="fade-in">
            <div style={{ marginBottom: 24 }}>
                <Button
                    icon={<ArrowLeftOutlined />}
                    onClick={() => navigate('/workflows')}
                    style={{ marginRight: 16 }}
                >
                    返回
                </Button>
                <Title level={2} style={{ display: 'inline', margin: 0 }}>
                    {workflow.name}
                </Title>
            </div>

            <Row gutter={[24, 24]}>
                <Col span={16}>
                    <Card
                        title="基本信息"
                        extra={
                            <Space>
                                <Button icon={<SettingOutlined />}>编辑</Button>
                                {workflow.status === 'active' ? (
                                    <Button icon={<PauseCircleOutlined />}>暂停</Button>
                                ) : (
                                    <Button type="primary" icon={<PlayCircleOutlined />}>执行</Button>
                                )}
                                <Button icon={<ReloadOutlined />} onClick={loadWorkflowDetail}>
                                    刷新
                                </Button>
                            </Space>
                        }
                    >
                        <Descriptions column={2}>
                            <Descriptions.Item label="工作流ID">{workflow.id}</Descriptions.Item>
                            <Descriptions.Item label="版本">v{workflow.metadata.version}</Descriptions.Item>
                            <Descriptions.Item label="状态">
                                {getStatusTag(workflow.status)}
                            </Descriptions.Item>
                            <Descriptions.Item label="步骤数">{workflow.steps.length}</Descriptions.Item>
                            <Descriptions.Item label="创建者">{workflow.metadata.createdBy}</Descriptions.Item>
                            <Descriptions.Item label="创建时间">
                                {new Date(workflow.metadata.createdAt).toLocaleString()}
                            </Descriptions.Item>
                            <Descriptions.Item label="最后更新">
                                {new Date(workflow.metadata.updatedAt).toLocaleString()}
                            </Descriptions.Item>
                            <Descriptions.Item label="平均执行时间">
                                {workflow.metrics.averageExecutionTime}s
                            </Descriptions.Item>
                        </Descriptions>

                        <div style={{ marginTop: 16 }}>
                            <strong>描述：</strong>
                            <p>{workflow.description}</p>
                        </div>
                    </Card>
                </Col>

                <Col span={8}>
                    <Card title="执行统计">
                        <Row gutter={[16, 16]}>
                            <Col span={24}>
                                <div style={{ textAlign: 'center', marginBottom: 16 }}>
                                    <div style={{ fontSize: 24, fontWeight: 'bold', color: '#1890ff' }}>
                                        {workflow.metrics.totalRuns}
                                    </div>
                                    <div style={{ color: '#666' }}>总执行次数</div>
                                </div>
                            </Col>
                            <Col span={12}>
                                <div style={{ textAlign: 'center' }}>
                                    <div style={{ fontSize: 18, fontWeight: 'bold', color: '#52c41a' }}>
                                        {workflow.metrics.successfulRuns}
                                    </div>
                                    <div style={{ fontSize: 12, color: '#666' }}>成功</div>
                                </div>
                            </Col>
                            <Col span={12}>
                                <div style={{ textAlign: 'center' }}>
                                    <div style={{ fontSize: 18, fontWeight: 'bold', color: '#ff4d4f' }}>
                                        {workflow.metrics.failedRuns}
                                    </div>
                                    <div style={{ fontSize: 12, color: '#666' }}>失败</div>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div style={{ marginTop: 16 }}>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                                        <span>成功率</span>
                                        <span>{((workflow.metrics.successfulRuns / workflow.metrics.totalRuns) * 100).toFixed(1)}%</span>
                                    </div>
                                    <Progress
                                        percent={(workflow.metrics.successfulRuns / workflow.metrics.totalRuns) * 100}
                                        strokeColor="#52c41a"
                                    />
                                </div>
                            </Col>
                        </Row>
                    </Card>
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                <Col span={24}>
                    <Card>
                        <Tabs defaultActiveKey="steps">
                            <TabPane tab="工作流步骤" key="steps">
                                <Steps direction="vertical" current={-1}>
                                    {workflow.steps.map((step, index) => (
                                        <Step
                                            key={step.id}
                                            title={step.name}
                                            description={
                                                <div>
                                                    <div>类型: {step.type}</div>
                                                    <div>超时: {step.timeout}ms</div>
                                                    <div>重试策略: {step.retryPolicy.maxRetries}次 ({step.retryPolicy.backoffStrategy})</div>
                                                    <div style={{ marginTop: 8 }}>
                                                        <Tag size="small">Agent ID: {step.agentId}</Tag>
                                                    </div>
                                                </div>
                                            }
                                            icon={<ClockCircleOutlined />}
                                        />
                                    ))}
                                </Steps>
                            </TabPane>

                            <TabPane tab="执行历史" key="executions">
                                <Table
                                    columns={executionColumns}
                                    dataSource={executions}
                                    rowKey="id"
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                        showQuickJumper: true,
                                    }}
                                />
                            </TabPane>

                            <TabPane tab="实时监控" key="monitoring">
                                <Alert
                                    message="实时监控"
                                    description="显示当前运行中的工作流执行状态"
                                    type="info"
                                    showIcon
                                    style={{ marginBottom: 16 }}
                                />

                                {executions.filter(e => e.status === 'running').map(execution => (
                                    <Card key={execution.id} title={`执行 ${execution.id}`} style={{ marginBottom: 16 }}>
                                        <Timeline>
                                            {execution.steps.map(step => {
                                                const workflowStep = workflow.steps.find(s => s.id === step.stepId);
                                                return (
                                                    <Timeline.Item
                                                        key={step.stepId}
                                                        color={
                                                            step.status === 'completed' ? 'green' :
                                                                step.status === 'running' ? 'blue' :
                                                                    step.status === 'failed' ? 'red' : 'gray'
                                                        }
                                                    >
                                                        <div>
                                                            <strong>{workflowStep?.name}</strong>
                                                            <Tag style={{ marginLeft: 8 }}>
                                                                {step.status === 'completed' ? '已完成' :
                                                                    step.status === 'running' ? '运行中' :
                                                                        step.status === 'failed' ? '失败' : '等待中'}
                                                            </Tag>
                                                        </div>
                                                        {step.startTime && (
                                                            <div style={{ fontSize: 12, color: '#666' }}>
                                                                开始时间: {new Date(step.startTime).toLocaleString()}
                                                            </div>
                                                        )}
                                                        {step.endTime && (
                                                            <div style={{ fontSize: 12, color: '#666' }}>
                                                                结束时间: {new Date(step.endTime).toLocaleString()}
                                                            </div>
                                                        )}
                                                        {step.retryCount > 0 && (
                                                            <div style={{ fontSize: 12, color: '#ff4d4f' }}>
                                                                重试次数: {step.retryCount}
                                                            </div>
                                                        )}
                                                    </Timeline.Item>
                                                );
                                            })}
                                        </Timeline>
                                    </Card>
                                ))}
                            </TabPane>
                        </Tabs>
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default WorkflowDetail;