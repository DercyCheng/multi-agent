import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Modal, Tag, Space, Tabs, message, Popconfirm, Badge, Progress, Descriptions, Typography } from 'antd';
import { ReloadOutlined, EyeOutlined, DeleteOutlined, CloudServerOutlined, CheckCircleOutlined, ExclamationCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface Service {
    id: string;
    name: string;
    version: string;
    address: string;
    port: number;
    health_check_url: string;
    status: 'healthy' | 'unhealthy' | 'unknown';
    last_seen: string;
    registered_at: string;
    metadata: Record<string, any>;
    tags: string[];
    health_check_interval: number;
    health_check_timeout: number;
    environment: string;
}

interface HealthCheck {
    service_id: string;
    timestamp: string;
    status: 'pass' | 'fail' | 'warn';
    duration: number;
    output: string;
    error?: string;
}

interface LoadBalancer {
    id: string;
    name: string;
    strategy: 'round_robin' | 'least_connections' | 'weighted_round_robin' | 'ip_hash';
    services: string[];
    health_check_enabled: boolean;
    created_at: string;
}

const ServiceDiscovery: React.FC = () => {
    const [services, setServices] = useState<Service[]>([]);
    const [healthChecks, setHealthChecks] = useState<HealthCheck[]>([]);
    const [loadBalancers, setLoadBalancers] = useState<LoadBalancer[]>([]);
    const [loading, setLoading] = useState(false);
    const [serviceModalVisible, setServiceModalVisible] = useState(false);
    const [healthModalVisible, setHealthModalVisible] = useState(false);
    const [selectedService, setSelectedService] = useState<Service | null>(null);
    const [activeTab, setActiveTab] = useState('services');

    // Mock data for demonstration
    const mockServices: Service[] = [
        {
            id: 'svc_001',
            name: 'api-gateway',
            version: '1.2.3',
            address: '10.0.1.10',
            port: 8080,
            health_check_url: '/health',
            status: 'healthy',
            last_seen: '2024-01-01T12:00:00Z',
            registered_at: '2024-01-01T10:00:00Z',
            metadata: {
                region: 'us-west-2',
                datacenter: 'dc1',
                instance_type: 't3.medium'
            },
            tags: ['gateway', 'api', 'frontend'],
            health_check_interval: 30,
            health_check_timeout: 5,
            environment: 'production',
        },
        {
            id: 'svc_002',
            name: 'orchestrator',
            version: '2.1.0',
            address: '10.0.1.11',
            port: 8081,
            health_check_url: '/health',
            status: 'healthy',
            last_seen: '2024-01-01T12:00:30Z',
            registered_at: '2024-01-01T10:05:00Z',
            metadata: {
                region: 'us-west-2',
                datacenter: 'dc1',
                instance_type: 't3.large'
            },
            tags: ['orchestrator', 'backend', 'core'],
            health_check_interval: 15,
            health_check_timeout: 10,
            environment: 'production',
        },
        {
            id: 'svc_003',
            name: 'llm-service',
            version: '1.0.5',
            address: '10.0.1.12',
            port: 8082,
            health_check_url: '/health',
            status: 'unhealthy',
            last_seen: '2024-01-01T11:45:00Z',
            registered_at: '2024-01-01T10:10:00Z',
            metadata: {
                region: 'us-west-2',
                datacenter: 'dc1',
                instance_type: 'g4dn.xlarge',
                gpu_enabled: true
            },
            tags: ['llm', 'ai', 'gpu'],
            health_check_interval: 60,
            health_check_timeout: 30,
            environment: 'production',
        },
        {
            id: 'svc_004',
            name: 'config-service',
            version: '1.1.2',
            address: '10.0.1.13',
            port: 8083,
            health_check_url: '/health',
            status: 'unknown',
            last_seen: '2024-01-01T10:30:00Z',
            registered_at: '2024-01-01T10:15:00Z',
            metadata: {
                region: 'us-west-2',
                datacenter: 'dc1',
                instance_type: 't3.small'
            },
            tags: ['config', 'backend', 'utility'],
            health_check_interval: 45,
            health_check_timeout: 15,
            environment: 'staging',
        },
    ];

    const mockHealthChecks: HealthCheck[] = [
        {
            service_id: 'svc_001',
            timestamp: '2024-01-01T12:00:00Z',
            status: 'pass',
            duration: 120,
            output: 'Service is healthy',
        },
        {
            service_id: 'svc_002',
            timestamp: '2024-01-01T12:00:30Z',
            status: 'pass',
            duration: 85,
            output: 'All systems operational',
        },
        {
            service_id: 'svc_003',
            timestamp: '2024-01-01T11:45:00Z',
            status: 'fail',
            duration: 5000,
            output: 'Connection timeout',
            error: 'Failed to connect to health endpoint',
        },
        {
            service_id: 'svc_004',
            timestamp: '2024-01-01T10:30:00Z',
            status: 'warn',
            duration: 200,
            output: 'High memory usage detected',
        },
    ];

    const mockLoadBalancers: LoadBalancer[] = [
        {
            id: 'lb_001',
            name: 'api-gateway-lb',
            strategy: 'round_robin',
            services: ['svc_001'],
            health_check_enabled: true,
            created_at: '2024-01-01T09:00:00Z',
        },
        {
            id: 'lb_002',
            name: 'backend-services-lb',
            strategy: 'least_connections',
            services: ['svc_002', 'svc_004'],
            health_check_enabled: true,
            created_at: '2024-01-01T09:15:00Z',
        },
    ];

    useEffect(() => {
        loadServices();
        loadHealthChecks();
        loadLoadBalancers();
    }, []);

    const loadServices = async () => {
        setLoading(true);
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setServices(mockServices);
        } catch (error) {
            message.error('Failed to load services');
        } finally {
            setLoading(false);
        }
    };

    const loadHealthChecks = async () => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 300));
            setHealthChecks(mockHealthChecks);
        } catch (error) {
            message.error('Failed to load health checks');
        }
    };

    const loadLoadBalancers = async () => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 300));
            setLoadBalancers(mockLoadBalancers);
        } catch (error) {
            message.error('Failed to load load balancers');
        }
    };

    const handleDeregisterService = async (serviceId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setServices(services.filter(svc => svc.id !== serviceId));
            message.success('Service deregistered successfully');
        } catch (error) {
            message.error('Failed to deregister service');
        }
    };

    const handleViewService = (service: Service) => {
        setSelectedService(service);
        setServiceModalVisible(true);
    };

    const handleViewHealthChecks = (service: Service) => {
        setSelectedService(service);
        setHealthModalVisible(true);
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'healthy':
                return <Badge status="success" text="Healthy" />;
            case 'unhealthy':
                return <Badge status="error" text="Unhealthy" />;
            case 'unknown':
                return <Badge status="default" text="Unknown" />;
            default:
                return <Badge status="default" text={status} />;
        }
    };

    const getHealthIcon = (status: string) => {
        switch (status) {
            case 'pass':
                return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
            case 'fail':
                return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
            case 'warn':
                return <ClockCircleOutlined style={{ color: '#faad14' }} />;
            default:
                return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
        }
    };

    const calculateUptime = (registeredAt: string, lastSeen: string) => {
        const registered = new Date(registeredAt);
        const last = new Date(lastSeen);
        const total = Date.now() - registered.getTime();
        const offline = Date.now() - last.getTime();
        const uptime = ((total - offline) / total) * 100;
        return Math.max(0, Math.min(100, uptime));
    };

    const serviceColumns = [
        {
            title: 'Service',
            key: 'service',
            render: (_: any, record: Service) => (
                <div>
                    <div style={{ fontWeight: 500, display: 'flex', alignItems: 'center' }}>
                        <CloudServerOutlined style={{ marginRight: 8 }} />
                        {record.name}
                        <Tag style={{ marginLeft: 8 }}>v{record.version}</Tag>
                    </div>
                    <div style={{ fontSize: '12px', color: '#666' }}>
                        {record.address}:{record.port}
                    </div>
                </div>
            ),
        },
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => getStatusBadge(status),
        },
        {
            title: 'Environment',
            dataIndex: 'environment',
            key: 'environment',
            render: (env: string) => (
                <Tag color={env === 'production' ? 'red' : env === 'staging' ? 'orange' : 'blue'}>
                    {env.toUpperCase()}
                </Tag>
            ),
        },
        {
            title: 'Uptime',
            key: 'uptime',
            render: (_: any, record: Service) => {
                const uptime = calculateUptime(record.registered_at, record.last_seen);
                return (
                    <div style={{ minWidth: 100 }}>
                        <Progress
                            percent={uptime}
                            size="small"
                            status={uptime > 95 ? 'success' : uptime > 80 ? 'normal' : 'exception'}
                            format={(percent) => `${percent?.toFixed(1)}%`}
                        />
                    </div>
                );
            },
        },
        {
            title: 'Tags',
            dataIndex: 'tags',
            key: 'tags',
            render: (tags: string[]) => (
                <div>
                    {tags.slice(0, 2).map(tag => (
                        <Tag key={tag}>{tag}</Tag>
                    ))}
                    {tags.length > 2 && (
                        <Tag>+{tags.length - 2}</Tag>
                    )}
                </div>
            ),
        },
        {
            title: 'Last Seen',
            dataIndex: 'last_seen',
            key: 'last_seen',
            render: (date: string) => {
                const lastSeen = new Date(date);
                const now = new Date();
                const diffMinutes = Math.floor((now.getTime() - lastSeen.getTime()) / (1000 * 60));

                return (
                    <div>
                        <div>{lastSeen.toLocaleString()}</div>
                        <div style={{ fontSize: '12px', color: '#666' }}>
                            {diffMinutes === 0 ? 'Just now' : `${diffMinutes}m ago`}
                        </div>
                    </div>
                );
            },
        },
        {
            title: 'Actions',
            key: 'actions',
            render: (_: any, record: Service) => (
                <Space>
                    <Button
                        type="link"
                        icon={<EyeOutlined />}
                        onClick={() => handleViewService(record)}
                        size="small"
                    >
                        Details
                    </Button>
                    <Button
                        type="link"
                        icon={<CheckCircleOutlined />}
                        onClick={() => handleViewHealthChecks(record)}
                        size="small"
                    >
                        Health
                    </Button>
                    <Popconfirm
                        title="Are you sure you want to deregister this service?"
                        onConfirm={() => handleDeregisterService(record.id)}
                        okText="Yes"
                        cancelText="No"
                    >
                        <Button
                            type="link"
                            danger
                            icon={<DeleteOutlined />}
                            size="small"
                        >
                            Deregister
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    const loadBalancerColumns = [
        {
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
        },
        {
            title: 'Strategy',
            dataIndex: 'strategy',
            key: 'strategy',
            render: (strategy: string) => (
                <Tag>{strategy.replace('_', ' ').toUpperCase()}</Tag>
            ),
        },
        {
            title: 'Services',
            dataIndex: 'services',
            key: 'services',
            render: (serviceIds: string[]) => (
                <div>
                    {serviceIds.map(id => {
                        const service = services.find(s => s.id === id);
                        return service ? (
                            <Tag key={id} color="blue">{service.name}</Tag>
                        ) : (
                            <Tag key={id} color="red">Unknown</Tag>
                        );
                    })}
                </div>
            ),
        },
        {
            title: 'Health Check',
            dataIndex: 'health_check_enabled',
            key: 'health_check_enabled',
            render: (enabled: boolean) => (
                <Badge status={enabled ? 'success' : 'default'} text={enabled ? 'Enabled' : 'Disabled'} />
            ),
        },
        {
            title: 'Created',
            dataIndex: 'created_at',
            key: 'created_at',
            render: (date: string) => new Date(date).toLocaleString(),
        },
    ];

    return (
        <div>
            <Card
                title="Service Discovery"
                extra={
                    <Button
                        icon={<ReloadOutlined />}
                        onClick={() => {
                            loadServices();
                            loadHealthChecks();
                            loadLoadBalancers();
                        }}
                    >
                        Refresh
                    </Button>
                }
            >
                <Tabs
                    activeKey={activeTab}
                    onChange={setActiveTab}
                    items={[
                        {
                            key: 'services',
                            label: `Services (${services.length})`,
                            children: (
                                <Table
                                    columns={serviceColumns}
                                    dataSource={services}
                                    rowKey="id"
                                    loading={loading}
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                        showQuickJumper: true,
                                        showTotal: (total, range) =>
                                            `${range[0]}-${range[1]} of ${total} services`,
                                    }}
                                />
                            ),
                        },
                        {
                            key: 'loadbalancers',
                            label: `Load Balancers (${loadBalancers.length})`,
                            children: (
                                <Table
                                    columns={loadBalancerColumns}
                                    dataSource={loadBalancers}
                                    rowKey="id"
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                    }}
                                />
                            ),
                        },
                        {
                            key: 'monitoring',
                            label: 'Monitoring',
                            children: (
                                <div>
                                    <p>Service monitoring dashboard will be displayed here.</p>
                                </div>
                            ),
                        },
                    ]}
                />
            </Card>

            {/* Service Details Modal */}
            <Modal
                title="Service Details"
                open={serviceModalVisible}
                onCancel={() => setServiceModalVisible(false)}
                footer={null}
                width={700}
            >
                {selectedService && (
                    <div>
                        <Descriptions column={2} bordered>
                            <Descriptions.Item label="Service Name">{selectedService.name}</Descriptions.Item>
                            <Descriptions.Item label="Version">{selectedService.version}</Descriptions.Item>
                            <Descriptions.Item label="Address">{selectedService.address}:{selectedService.port}</Descriptions.Item>
                            <Descriptions.Item label="Status">{getStatusBadge(selectedService.status)}</Descriptions.Item>
                            <Descriptions.Item label="Environment">
                                <Tag color={selectedService.environment === 'production' ? 'red' : 'orange'}>
                                    {selectedService.environment.toUpperCase()}
                                </Tag>
                            </Descriptions.Item>
                            <Descriptions.Item label="Health Check URL">{selectedService.health_check_url}</Descriptions.Item>
                            <Descriptions.Item label="Check Interval">{selectedService.health_check_interval}s</Descriptions.Item>
                            <Descriptions.Item label="Check Timeout">{selectedService.health_check_timeout}s</Descriptions.Item>
                            <Descriptions.Item label="Registered At">
                                {new Date(selectedService.registered_at).toLocaleString()}
                            </Descriptions.Item>
                            <Descriptions.Item label="Last Seen">
                                {new Date(selectedService.last_seen).toLocaleString()}
                            </Descriptions.Item>
                        </Descriptions>

                        <div style={{ marginTop: 16 }}>
                            <h4>Tags</h4>
                            <div>
                                {selectedService.tags.map(tag => (
                                    <Tag key={tag}>{tag}</Tag>
                                ))}
                            </div>
                        </div>

                        <div style={{ marginTop: 16 }}>
                            <h4>Metadata</h4>
                            <pre style={{
                                background: '#f5f5f5',
                                padding: 12,
                                borderRadius: 4,
                                fontSize: 12
                            }}>
                                {JSON.stringify(selectedService.metadata, null, 2)}
                            </pre>
                        </div>
                    </div>
                )}
            </Modal>

            {/* Health Check History Modal */}
            <Modal
                title="Health Check History"
                open={healthModalVisible}
                onCancel={() => setHealthModalVisible(false)}
                footer={null}
                width={800}
            >
                {selectedService && (
                    <div>
                        <div style={{ marginBottom: 16 }}>
                            <Text strong>Service: </Text>{selectedService.name}
                        </div>

                        <Table
                            columns={[
                                {
                                    title: 'Status',
                                    dataIndex: 'status',
                                    key: 'status',
                                    render: (status: string) => (
                                        <div style={{ display: 'flex', alignItems: 'center' }}>
                                            {getHealthIcon(status)}
                                            <span style={{ marginLeft: 8, textTransform: 'uppercase' }}>{status}</span>
                                        </div>
                                    ),
                                },
                                {
                                    title: 'Timestamp',
                                    dataIndex: 'timestamp',
                                    key: 'timestamp',
                                    render: (date: string) => new Date(date).toLocaleString(),
                                },
                                {
                                    title: 'Duration',
                                    dataIndex: 'duration',
                                    key: 'duration',
                                    render: (duration: number) => `${duration}ms`,
                                },
                                {
                                    title: 'Output',
                                    dataIndex: 'output',
                                    key: 'output',
                                },
                            ]}
                            dataSource={healthChecks.filter(hc => hc.service_id === selectedService.id)}
                            rowKey="timestamp"
                            pagination={false}
                            size="small"
                        />
                    </div>
                )}
            </Modal>
        </div>
    );
};

export default ServiceDiscovery;