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
    Statistic,
    Typography,
    Tabs,
    // Table,
    Progress,
    Alert,
    List,
} from 'antd';
import {
    ArrowLeftOutlined,
    PlayCircleOutlined,
    PauseCircleOutlined,
    SettingOutlined,
    ReloadOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { Agent } from '../../types';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const AgentDetail: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [agent, setAgent] = useState<Agent | null>(null);
    const [loading, setLoading] = useState(true);

    // 模拟Agent详情数据
    const mockAgent: Agent = {
        id: '1',
        name: 'GPT-4 对话Agent',
        type: 'LLM',
        status: 'active',
        capabilities: ['自然语言理解', '对话生成', '问答', '文本摘要'],
        configuration: {
            model: 'gpt-4',
            temperature: 0.7,
            maxTokens: 4096,
            timeout: 30000
        },
        metrics: {
            requestCount: 1247,
            successRate: 98.5,
            averageResponseTime: 1.2,
            lastActive: '2024-01-20T10:30:00Z',
        },
        createdAt: '2024-01-15T09:00:00Z',
        updatedAt: '2024-01-20T10:30:00Z',
    };

    useEffect(() => {
        loadAgentDetail();
    }, [id]);

    const loadAgentDetail = async () => {
        setLoading(true);
        try {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 1000));
            setAgent(mockAgent);
        } catch (error) {
            console.error('Failed to load agent detail:', error);
        } finally {
            setLoading(false);
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

    // 性能图表配置
    const performanceOption = {
        title: {
            text: '响应时间趋势',
            textStyle: { fontSize: 14 },
        },
        tooltip: {
            trigger: 'axis',
        },
        xAxis: {
            type: 'category',
            data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
        },
        yAxis: {
            type: 'value',
            name: '响应时间(秒)',
        },
        series: [
            {
                data: [1.2, 1.5, 1.1, 1.8, 1.3, 1.2],
                type: 'line',
                smooth: true,
            },
        ],
    };

    // 请求量图表配置
    const requestOption = {
        title: {
            text: '请求量统计',
            textStyle: { fontSize: 14 },
        },
        tooltip: {
            trigger: 'axis',
        },
        xAxis: {
            type: 'category',
            data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
        },
        yAxis: {
            type: 'value',
            name: '请求数',
        },
        series: [
            {
                data: [150, 200, 180, 220, 190, 160, 140],
                type: 'bar',
                itemStyle: {
                    color: '#1890ff',
                },
            },
        ],
    };

    // 模拟日志数据
    const logs = [
        {
            timestamp: '2024-01-20 10:30:15',
            level: 'INFO',
            message: '处理用户查询: "什么是人工智能?"',
        },
        {
            timestamp: '2024-01-20 10:29:45',
            level: 'INFO',
            message: '成功响应请求，耗时 1.2s',
        },
        {
            timestamp: '2024-01-20 10:28:30',
            level: 'WARN',
            message: '响应时间超过阈值 (1.5s)',
        },
        {
            timestamp: '2024-01-20 10:27:20',
            level: 'INFO',
            message: '开始处理新的对话会话',
        },
    ];

    if (loading) {
        return <div>加载中...</div>;
    }

    if (!agent) {
        return <div>Agent未找到</div>;
    }

    return (
        <div className="fade-in">
            <div style={{ marginBottom: 24 }}>
                <Button
                    icon={<ArrowLeftOutlined />}
                    onClick={() => navigate('/agents')}
                    style={{ marginRight: 16 }}
                >
                    返回
                </Button>
                <Title level={2} style={{ display: 'inline', margin: 0 }}>
                    {agent.name}
                </Title>
            </div>

            <Row gutter={[24, 24]}>
                <Col span={24}>
                    <Card
                        title="基本信息"
                        extra={
                            <Space>
                                <Button icon={<SettingOutlined />}>配置</Button>
                                {agent.status === 'active' ? (
                                    <Button icon={<PauseCircleOutlined />}>停止</Button>
                                ) : (
                                    <Button type="primary" icon={<PlayCircleOutlined />}>启动</Button>
                                )}
                                <Button icon={<ReloadOutlined />} onClick={loadAgentDetail}>
                                    刷新
                                </Button>
                            </Space>
                        }
                    >
                        <Row gutter={[24, 24]}>
                            <Col span={12}>
                                <Descriptions column={1}>
                                    <Descriptions.Item label="Agent ID">{agent.id}</Descriptions.Item>
                                    <Descriptions.Item label="名称">{agent.name}</Descriptions.Item>
                                    <Descriptions.Item label="类型">
                                        <Tag>{agent.type}</Tag>
                                    </Descriptions.Item>
                                    <Descriptions.Item label="状态">
                                        {getStatusTag(agent.status)}
                                    </Descriptions.Item>
                                    <Descriptions.Item label="创建时间">
                                        {new Date(agent.createdAt).toLocaleString()}
                                    </Descriptions.Item>
                                    <Descriptions.Item label="最后更新">
                                        {new Date(agent.updatedAt).toLocaleString()}
                                    </Descriptions.Item>
                                </Descriptions>
                            </Col>
                            <Col span={12}>
                                <Space direction="vertical" style={{ width: '100%' }}>
                                    <Card size="small">
                                        <Statistic
                                            title="总请求数"
                                            value={agent.metrics.requestCount}
                                            valueStyle={{ color: '#3f8600' }}
                                        />
                                    </Card>
                                    <Card size="small">
                                        <Statistic
                                            title="成功率"
                                            value={agent.metrics.successRate}
                                            suffix="%"
                                            valueStyle={{ color: '#1890ff' }}
                                        />
                                    </Card>
                                    <Card size="small">
                                        <Statistic
                                            title="平均响应时间"
                                            value={agent.metrics.averageResponseTime}
                                            suffix="s"
                                            precision={1}
                                            valueStyle={{ color: '#722ed1' }}
                                        />
                                    </Card>
                                </Space>
                            </Col>
                        </Row>
                    </Card>
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                <Col span={24}>
                    <Card title="能力描述">
                        <Space wrap>
                            {agent.capabilities.map(capability => (
                                <Tag key={capability} color="blue" style={{ margin: '4px 0' }}>
                                    {capability}
                                </Tag>
                            ))}
                        </Space>
                    </Card>
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                <Col span={24}>
                    <Card>
                        <Tabs defaultActiveKey="performance">
                            <TabPane tab="性能监控" key="performance">
                                <Row gutter={[24, 24]}>
                                    <Col span={12}>
                                        <ReactECharts option={performanceOption} style={{ height: 300 }} />
                                    </Col>
                                    <Col span={12}>
                                        <ReactECharts option={requestOption} style={{ height: 300 }} />
                                    </Col>
                                </Row>

                                <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                                    <Col span={8}>
                                        <Card>
                                            <Text>CPU 使用率</Text>
                                            <Progress percent={45} status="normal" />
                                        </Card>
                                    </Col>
                                    <Col span={8}>
                                        <Card>
                                            <Text>内存使用率</Text>
                                            <Progress percent={67} status="active" />
                                        </Card>
                                    </Col>
                                    <Col span={8}>
                                        <Card>
                                            <Text>错误率</Text>
                                            <Progress percent={2} status="success" />
                                        </Card>
                                    </Col>
                                </Row>
                            </TabPane>

                            <TabPane tab="配置信息" key="configuration">
                                <Alert
                                    message="配置信息"
                                    description="以下是当前Agent的详细配置参数"
                                    type="info"
                                    showIcon
                                    style={{ marginBottom: 16 }}
                                />
                                <pre style={{
                                    background: '#f5f5f5',
                                    padding: 16,
                                    borderRadius: 4,
                                    overflow: 'auto'
                                }}>
                                    {JSON.stringify(agent.configuration, null, 2)}
                                </pre>
                            </TabPane>

                            <TabPane tab="运行日志" key="logs">
                                <List
                                    dataSource={logs}
                                    renderItem={(item: any) => (
                                        <List.Item>
                                            <List.Item.Meta
                                                title={
                                                    <Space>
                                                        <Text type="secondary">{item.timestamp}</Text>
                                                        <Tag color={
                                                            item.level === 'ERROR' ? 'red' :
                                                                item.level === 'WARN' ? 'orange' : 'blue'
                                                        }>
                                                            {item.level}
                                                        </Tag>
                                                    </Space>
                                                }
                                                description={item.message}
                                            />
                                        </List.Item>
                                    )}
                                />
                            </TabPane>
                        </Tabs>
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default AgentDetail;