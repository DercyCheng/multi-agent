import React, { useEffect, useState } from 'react';
import {
    Row,
    Col,
    Card,
    Statistic,
    Progress,
    Table,
    Tag,
    Typography,
    Space,
    Button,
    Alert,
} from 'antd';
import {
    RobotOutlined,
    ApartmentOutlined,
    ThunderboltOutlined,
    CheckCircleOutlined,
    ExclamationCircleOutlined,
    ClockCircleOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';

const { Title, Text } = Typography;

interface MetricCardProps {
    title: string;
    value: number;
    suffix?: string;
    prefix?: React.ReactNode;
    status?: 'default' | 'success' | 'warning' | 'error';
    trend?: number;
}

const MetricCard: React.FC<MetricCardProps> = ({
    title,
    value,
    suffix = '',
    prefix,
    status = 'default',
    trend
}) => {
    const getStatusColor = (status: string) => {
        switch (status) {
            case 'success': return '#52c41a';
            case 'warning': return '#faad14';
            case 'error': return '#ff4d4f';
            default: return '#1890ff';
        }
    };

    return (
        <Card>
            <Statistic
                title={title}
                value={value}
                suffix={suffix}
                prefix={prefix}
                valueStyle={{ color: getStatusColor(status) }}
            />
            {trend !== undefined && (
                <div style={{ marginTop: 8 }}>
                    <Text type={trend >= 0 ? 'success' : 'danger'} style={{ fontSize: 12 }}>
                        {trend >= 0 ? '↗' : '↘'} {Math.abs(trend)}% 较上周
                    </Text>
                </div>
            )}
        </Card>
    );
};

const Dashboard: React.FC = () => {
    const [loading, setLoading] = useState(true);
    const [systemHealth, setSystemHealth] = useState('healthy');

    useEffect(() => {
        // 模拟数据加载
        const timer = setTimeout(() => {
            setLoading(false);
        }, 1000);
        return () => clearTimeout(timer);
    }, []);

    // 模拟数据
    const recentActivities = [
        {
            key: '1',
            agent: 'GPT-4 Agent',
            action: '处理用户查询',
            status: 'success',
            time: '2 分钟前',
        },
        {
            key: '2',
            agent: 'Search Agent',
            action: '执行网络搜索',
            status: 'running',
            time: '5 分钟前',
        },
        {
            key: '3',
            agent: 'Code Agent',
            action: '代码分析完成',
            status: 'success',
            time: '8 分钟前',
        },
        {
            key: '4',
            agent: 'Data Agent',
            action: '数据处理失败',
            status: 'error',
            time: '12 分钟前',
        },
    ];

    const columns = [
        {
            title: 'Agent',
            dataIndex: 'agent',
            key: 'agent',
        },
        {
            title: '操作',
            dataIndex: 'action',
            key: 'action',
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => {
                const statusConfig = {
                    success: { color: 'success', text: '成功', icon: <CheckCircleOutlined /> },
                    running: { color: 'processing', text: '运行中', icon: <ClockCircleOutlined /> },
                    error: { color: 'error', text: '失败', icon: <ExclamationCircleOutlined /> },
                };
                const config = statusConfig[status as keyof typeof statusConfig];
                return (
                    <Tag color={config.color} icon={config.icon}>
                        {config.text}
                    </Tag>
                );
            },
        },
        {
            title: '时间',
            dataIndex: 'time',
            key: 'time',
        },
    ];

    // 图表配置
    const throughputOption = {
        title: {
            text: '系统吞吐量',
            textStyle: { fontSize: 14, fontWeight: 'normal' },
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
            name: '请求/秒',
        },
        series: [
            {
                data: [120, 200, 150, 80, 70, 110],
                type: 'line',
                smooth: true,
                areaStyle: {},
            },
        ],
    };

    const agentStatusOption = {
        title: {
            text: 'Agent状态分布',
            textStyle: { fontSize: 14, fontWeight: 'normal' },
        },
        tooltip: {
            trigger: 'item',
        },
        series: [
            {
                type: 'pie',
                radius: '70%',
                data: [
                    { value: 12, name: '运行中' },
                    { value: 3, name: '空闲' },
                    { value: 1, name: '错误' },
                    { value: 2, name: '维护中' },
                ],
            },
        ],
    };

    return (
        <div className="fade-in">
            <Title level={2} style={{ marginBottom: 24 }}>
                系统仪表板
            </Title>

            {systemHealth !== 'healthy' && (
                <Alert
                    message="系统健康状态异常"
                    description="检测到部分服务运行异常，请及时处理"
                    type="warning"
                    showIcon
                    closable
                    style={{ marginBottom: 24 }}
                />
            )}

            <Row gutter={[24, 24]}>
                <Col xs={24} sm={12} lg={6}>
                    <MetricCard
                        title="活跃 Agents"
                        value={15}
                        prefix={<RobotOutlined />}
                        status="success"
                        trend={12.5}
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <MetricCard
                        title="运行中工作流"
                        value={8}
                        prefix={<ApartmentOutlined />}
                        status="default"
                        trend={-2.1}
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <MetricCard
                        title="今日处理量"
                        value={1247}
                        prefix={<ThunderboltOutlined />}
                        status="success"
                        trend={18.9}
                    />
                </Col>
                <Col xs={24} sm={12} lg={6}>
                    <MetricCard
                        title="成功率"
                        value={98.5}
                        suffix="%"
                        prefix={<CheckCircleOutlined />}
                        status="success"
                        trend={0.8}
                    />
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                <Col xs={24} lg={12}>
                    <Card>
                        <ReactECharts option={throughputOption} style={{ height: 300 }} />
                    </Card>
                </Col>
                <Col xs={24} lg={12}>
                    <Card>
                        <ReactECharts option={agentStatusOption} style={{ height: 300 }} />
                    </Card>
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
                <Col xs={24} lg={16}>
                    <Card
                        title="最近活动"
                        extra={
                            <Button type="link" size="small">
                                查看全部
                            </Button>
                        }
                    >
                        <Table
                            columns={columns}
                            dataSource={recentActivities}
                            pagination={false}
                            size="small"
                            loading={loading}
                        />
                    </Card>
                </Col>
                <Col xs={24} lg={8}>
                    <Card title="系统状态">
                        <Space direction="vertical" style={{ width: '100%' }}>
                            <div>
                                <Text>CPU 使用率</Text>
                                <Progress percent={45} size="small" status="normal" />
                            </div>
                            <div>
                                <Text>内存使用率</Text>
                                <Progress percent={67} size="small" status="active" />
                            </div>
                            <div>
                                <Text>磁盘使用率</Text>
                                <Progress percent={23} size="small" status="normal" />
                            </div>
                            <div>
                                <Text>网络延迟</Text>
                                <Progress percent={12} size="small" status="success" />
                            </div>
                        </Space>
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default Dashboard;