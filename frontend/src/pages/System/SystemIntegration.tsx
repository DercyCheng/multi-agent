import React, { useState, useEffect } from 'react';
import {
    Card,
    Row,
    Col,
    Statistic,
    Table,
    Tag,
    Alert,
    Timeline,
    Tabs,
    Space,
    Button,
    Badge,
    Progress,
    Modal,
    Form,
    Select,
    message,
    Typography
} from 'antd';
import {
    DashboardOutlined,
    WarningOutlined,
    CheckCircleOutlined,
    ExclamationCircleOutlined,
    SyncOutlined,
    SettingOutlined,
    ApiOutlined
} from '@ant-design/icons';
import { systemIntegrationApi } from '../../services/platform';

const { Text } = Typography;
const { Option } = Select;

interface SystemStatus {
    feature_flags_count: number;
    configurations_count: number;
    active_cronjobs: number;
    healthy_services: number;
    total_services: number;
    system_health: 'healthy' | 'warning' | 'critical';
}

interface DependencyInfo {
    feature_flag_dependencies: Array<{
        flag_name: string;
        dependent_configs: string[];
        dependent_services: string[];
        dependent_jobs: string[];
    }>;
    service_dependencies: Array<{
        service_name: string;
        depends_on: string[];
        dependents: string[];
    }>;
}

interface ChangeHistoryItem {
    timestamp: string;
    type: 'feature_flag' | 'configuration' | 'cronjob' | 'service';
    action: 'create' | 'update' | 'delete';
    entity_id: string;
    entity_name: string;
    changes: Record<string, any>;
    user: string;
    impact: {
        affected_services: string[];
        estimated_users: number;
        risk_level: 'low' | 'medium' | 'high';
    };
}

const SystemIntegration: React.FC = () => {
    const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null);
    const [dependencies, setDependencies] = useState<DependencyInfo | null>(null);
    const [changeHistory, setChangeHistory] = useState<ChangeHistoryItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [syncModalVisible, setSyncModalVisible] = useState(false);
    const [activeTab, setActiveTab] = useState('overview');
    const [form] = Form.useForm();

    useEffect(() => {
        loadSystemStatus();
        loadDependencies();
        loadChangeHistory();

        // 设置实时事件监听
        const unsubscribe = systemIntegrationApi.subscribeToEvents((event) => {
            handleRealTimeEvent(event);
        });

        return () => {
            if (unsubscribe) unsubscribe();
        };
    }, []);

    const loadSystemStatus = async () => {
        try {
            const response = await systemIntegrationApi.getSystemStatus();
            setSystemStatus(response.data);
        } catch (error) {
            message.error('获取系统状态失败');
        }
    };

    const loadDependencies = async () => {
        try {
            const response = await systemIntegrationApi.analyzeDependencies();
            setDependencies(response.data);
        } catch (error) {
            message.error('分析依赖关系失败');
        }
    };

    const loadChangeHistory = async () => {
        try {
            const response = await systemIntegrationApi.getChangeHistory({ hours: 24 });
            setChangeHistory(response.data);
        } catch (error) {
            message.error('获取变更历史失败');
        }
    };

    const handleRealTimeEvent = (event: any) => {
        // 处理实时事件，更新状态
        console.log('Real-time event:', event);

        // 更新变更历史
        if (event.type === 'change') {
            setChangeHistory(prev => [event.data, ...prev.slice(0, 49)]);
        }

        // 刷新系统状态
        if (event.type === 'status_change') {
            loadSystemStatus();
        }
    };

    const handleEnvironmentSync = async (values: any) => {
        setLoading(true);
        try {
            const response = await systemIntegrationApi.syncEnvironments(
                values.source_environment,
                values.target_environment,
                values.sync_types
            );

            message.success(`同步完成: ${response.data.synced_items} 个项目已同步`);

            if (response.data.conflicts.length > 0) {
                Modal.warning({
                    title: '同步冲突',
                    content: (
                        <div>
                            <p>以下项目存在冲突:</p>
                            <ul>
                                {response.data.conflicts.map((conflict, index) => (
                                    <li key={index}>
                                        {conflict.type}: {conflict.key} - {conflict.reason}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    ),
                });
            }

            setSyncModalVisible(false);
            form.resetFields();
            loadSystemStatus();
        } catch (error) {
            message.error('环境同步失败');
        } finally {
            setLoading(false);
        }
    };

    const getHealthColor = (health: string) => {
        switch (health) {
            case 'healthy': return '#52c41a';
            case 'warning': return '#faad14';
            case 'critical': return '#ff4d4f';
            default: return '#d9d9d9';
        }
    };

    const getRiskColor = (risk: string) => {
        switch (risk) {
            case 'low': return 'green';
            case 'medium': return 'orange';
            case 'high': return 'red';
            default: return 'default';
        }
    };

    const dependencyColumns = [
        {
            title: '资源名称',
            dataIndex: 'flag_name',
            key: 'name',
            render: (name: string, record: any) => (
                <div>
                    <Text strong>{name || record.service_name}</Text>
                    <br />
                    <Tag color="blue">
                        {record.flag_name ? '特性开关' : '服务'}
                    </Tag>
                </div>
            ),
        },
        {
            title: '依赖项',
            key: 'dependencies',
            render: (_: any, record: any) => (
                <div>
                    {record.dependent_configs?.length > 0 && (
                        <div style={{ marginBottom: 4 }}>
                            <Text type="secondary">配置: </Text>
                            {record.dependent_configs.slice(0, 3).map((config: string) => (
                                <Tag key={config}>{config}</Tag>
                            ))}
                            {record.dependent_configs.length > 3 && (
                                <Tag>+{record.dependent_configs.length - 3}</Tag>
                            )}
                        </div>
                    )}
                    {record.dependent_services?.length > 0 && (
                        <div style={{ marginBottom: 4 }}>
                            <Text type="secondary">服务: </Text>
                            {record.dependent_services.slice(0, 3).map((service: string) => (
                                <Tag key={service} color="blue">{service}</Tag>
                            ))}
                            {record.dependent_services.length > 3 && (
                                <Tag>+{record.dependent_services.length - 3}</Tag>
                            )}
                        </div>
                    )}
                    {record.depends_on?.length > 0 && (
                        <div>
                            <Text type="secondary">依赖于: </Text>
                            {record.depends_on.slice(0, 3).map((dep: string) => (
                                <Tag key={dep} color="orange">{dep}</Tag>
                            ))}
                            {record.depends_on.length > 3 && (
                                <Tag>+{record.depends_on.length - 3}</Tag>
                            )}
                        </div>
                    )}
                </div>
            ),
        },
        {
            title: '影响范围',
            key: 'impact',
            render: (_: any, record: any) => {
                const totalDeps = (record.dependent_configs?.length || 0) +
                    (record.dependent_services?.length || 0) +
                    (record.dependent_jobs?.length || 0) +
                    (record.dependents?.length || 0);

                return (
                    <div>
                        <Badge count={totalDeps} style={{ backgroundColor: '#52c41a' }} />
                        <Text style={{ marginLeft: 8 }}>个组件</Text>
                    </div>
                );
            },
        },
    ];

    const changeHistoryColumns = [
        {
            title: '时间',
            dataIndex: 'timestamp',
            key: 'timestamp',
            render: (timestamp: string) => new Date(timestamp).toLocaleString(),
            width: 150,
        },
        {
            title: '变更',
            key: 'change',
            render: (_: any, record: ChangeHistoryItem) => (
                <div>
                    <div>
                        <Tag color={record.action === 'create' ? 'green' : record.action === 'delete' ? 'red' : 'blue'}>
                            {record.action.toUpperCase()}
                        </Tag>
                        <Tag>{record.type.replace('_', ' ').toUpperCase()}</Tag>
                    </div>
                    <Text strong>{record.entity_name}</Text>
                    <br />
                    <Text type="secondary">by {record.user}</Text>
                </div>
            ),
        },
        {
            title: '影响分析',
            key: 'impact',
            render: (_: any, record: ChangeHistoryItem) => (
                <div>
                    <div style={{ marginBottom: 4 }}>
                        <Tag color={getRiskColor(record.impact.risk_level)}>
                            {record.impact.risk_level.toUpperCase()} RISK
                        </Tag>
                    </div>
                    <div>
                        <Text type="secondary">影响服务: </Text>
                        <Badge count={record.impact.affected_services.length} style={{ backgroundColor: '#1890ff' }} />
                    </div>
                    <div>
                        <Text type="secondary">预估用户: </Text>
                        <Text>{record.impact.estimated_users.toLocaleString()}</Text>
                    </div>
                </div>
            ),
        },
    ];

    return (
        <div>
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col span={24}>
                    <Card>
                        <Row gutter={16}>
                            <Col span={6}>
                                <Statistic
                                    title="系统健康状态"
                                    value={systemStatus?.system_health || 'unknown'}
                                    valueStyle={{ color: getHealthColor(systemStatus?.system_health || 'unknown') }}
                                    prefix={
                                        systemStatus?.system_health === 'healthy' ?
                                            <CheckCircleOutlined /> :
                                            systemStatus?.system_health === 'warning' ?
                                                <ExclamationCircleOutlined /> :
                                                <WarningOutlined />
                                    }
                                />
                            </Col>
                            <Col span={4}>
                                <Statistic
                                    title="特性开关"
                                    value={systemStatus?.feature_flags_count || 0}
                                    prefix={<DashboardOutlined />}
                                />
                            </Col>
                            <Col span={4}>
                                <Statistic
                                    title="配置项"
                                    value={systemStatus?.configurations_count || 0}
                                    prefix={<SettingOutlined />}
                                />
                            </Col>
                            <Col span={4}>
                                <Statistic
                                    title="活跃任务"
                                    value={systemStatus?.active_cronjobs || 0}
                                    prefix={<SyncOutlined />}
                                />
                            </Col>
                            <Col span={6}>
                                <Statistic
                                    title="服务健康率"
                                    value={systemStatus ?
                                        Math.round((systemStatus.healthy_services / systemStatus.total_services) * 100) : 0
                                    }
                                    suffix="%"
                                    prefix={<ApiOutlined />}
                                />
                                <Progress
                                    percent={systemStatus ?
                                        Math.round((systemStatus.healthy_services / systemStatus.total_services) * 100) : 0
                                    }
                                    size="small"
                                    showInfo={false}
                                />
                            </Col>
                        </Row>
                    </Card>
                </Col>
            </Row>

            <Card
                title="系统集成管理"
                extra={
                    <Space>
                        <Button
                            type="primary"
                            icon={<SyncOutlined />}
                            onClick={() => setSyncModalVisible(true)}
                        >
                            环境同步
                        </Button>
                        <Button
                            icon={<SyncOutlined />}
                            onClick={() => {
                                loadSystemStatus();
                                loadDependencies();
                                loadChangeHistory();
                            }}
                        >
                            刷新
                        </Button>
                    </Space>
                }
            >
                <Tabs
                    activeKey={activeTab}
                    onChange={setActiveTab}
                    items={[
                        {
                            key: 'overview',
                            label: '系统概览',
                            children: (
                                <Row gutter={[16, 16]}>
                                    <Col span={24}>
                                        {systemStatus?.system_health === 'warning' && (
                                            <Alert
                                                message="系统警告"
                                                description="检测到部分服务状态异常，建议立即检查相关服务健康状况。"
                                                type="warning"
                                                showIcon
                                                style={{ marginBottom: 16 }}
                                            />
                                        )}
                                        {systemStatus?.system_health === 'critical' && (
                                            <Alert
                                                message="系统严重警告"
                                                description="系统存在严重问题，可能影响业务正常运行，请立即处理。"
                                                type="error"
                                                showIcon
                                                style={{ marginBottom: 16 }}
                                            />
                                        )}
                                    </Col>
                                </Row>
                            ),
                        },
                        {
                            key: 'dependencies',
                            label: `依赖关系 (${(dependencies?.feature_flag_dependencies?.length || 0) + (dependencies?.service_dependencies?.length || 0)})`,
                            children: (
                                <Table
                                    columns={dependencyColumns}
                                    dataSource={[
                                        ...(dependencies?.feature_flag_dependencies || []),
                                        ...(dependencies?.service_dependencies || [])
                                    ]}
                                    rowKey={(record: any) => record.flag_name || record.service_name}
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                    }}
                                />
                            ),
                        },
                        {
                            key: 'changes',
                            label: `变更历史 (${changeHistory.length})`,
                            children: (
                                <Table
                                    columns={changeHistoryColumns}
                                    dataSource={changeHistory}
                                    rowKey="entity_id"
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                    }}
                                />
                            ),
                        },
                        {
                            key: 'timeline',
                            label: '实时事件',
                            children: (
                                <Timeline>
                                    {changeHistory.slice(0, 20).map((item, index) => (
                                        <Timeline.Item
                                            key={index}
                                            color={item.impact.risk_level === 'high' ? 'red' :
                                                item.impact.risk_level === 'medium' ? 'orange' : 'green'}
                                            dot={item.action === 'create' ? <CheckCircleOutlined /> :
                                                item.action === 'delete' ? <ExclamationCircleOutlined /> :
                                                    <SyncOutlined />}
                                        >
                                            <div>
                                                <Text strong>{item.entity_name}</Text>
                                                <Tag style={{ marginLeft: 8 }}>{item.type.replace('_', ' ')}</Tag>
                                                <Tag color={item.action === 'create' ? 'green' :
                                                    item.action === 'delete' ? 'red' : 'blue'}>
                                                    {item.action}
                                                </Tag>
                                                <br />
                                                <Text type="secondary">
                                                    {new Date(item.timestamp).toLocaleString()} by {item.user}
                                                </Text>
                                                <br />
                                                <Text type="secondary">
                                                    影响 {item.impact.affected_services.length} 个服务，
                                                    预估 {item.impact.estimated_users.toLocaleString()} 用户
                                                </Text>
                                            </div>
                                        </Timeline.Item>
                                    ))}
                                </Timeline>
                            ),
                        },
                    ]}
                />
            </Card>

            {/* 环境同步模态框 */}
            <Modal
                title="环境同步"
                open={syncModalVisible}
                onCancel={() => setSyncModalVisible(false)}
                footer={null}
                width={600}
            >
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleEnvironmentSync}
                >
                    <Form.Item
                        name="source_environment"
                        label="源环境"
                        rules={[{ required: true, message: '请选择源环境' }]}
                    >
                        <Select placeholder="选择源环境">
                            <Option value="production">生产环境</Option>
                            <Option value="staging">预发布环境</Option>
                            <Option value="development">开发环境</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="target_environment"
                        label="目标环境"
                        rules={[{ required: true, message: '请选择目标环境' }]}
                    >
                        <Select placeholder="选择目标环境">
                            <Option value="production">生产环境</Option>
                            <Option value="staging">预发布环境</Option>
                            <Option value="development">开发环境</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="sync_types"
                        label="同步类型"
                        rules={[{ required: true, message: '请选择同步类型' }]}
                    >
                        <Select mode="multiple" placeholder="选择要同步的内容">
                            <Option value="feature_flags">特性开关</Option>
                            <Option value="configurations">配置项</Option>
                            <Option value="cronjobs">定时任务</Option>
                            <Option value="services">服务注册</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item>
                        <Space>
                            <Button type="primary" htmlType="submit" loading={loading}>
                                开始同步
                            </Button>
                            <Button onClick={() => setSyncModalVisible(false)}>
                                取消
                            </Button>
                        </Space>
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default SystemIntegration;