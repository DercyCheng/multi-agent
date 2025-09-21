import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Modal, Form, Input, Select, Tag, Space, Tabs, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, ReloadOutlined } from '@ant-design/icons';

const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;

interface Configuration {
    id: string;
    key: string;
    value: any;
    environment: string;
    tenant_id: string;
    version: number;
    description: string;
    created_by: string;
    created_at: string;
    updated_at: string;
    metadata: Record<string, any>;
}

const ConfigurationCenter: React.FC = () => {
    const [configs, setConfigs] = useState<Configuration[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [editingConfig, setEditingConfig] = useState<Configuration | null>(null);
    const [environment, setEnvironment] = useState('development');
    const [form] = Form.useForm();

    // Mock data for demonstration
    const mockConfigs: Configuration[] = [
        {
            id: 'config_001',
            key: 'app.title',
            value: 'Multi-Agent Platform',
            environment: 'development',
            tenant_id: 'default',
            version: 1,
            description: 'Application title displayed in header',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            metadata: { type: 'string' },
        },
        {
            id: 'config_002',
            key: 'database.max_connections',
            value: 100,
            environment: 'development',
            tenant_id: 'default',
            version: 2,
            description: 'Maximum database connections',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-02T00:00:00Z',
            metadata: { type: 'number', min: 1, max: 1000 },
        },
        {
            id: 'config_003',
            key: 'features.enabled',
            value: {
                analytics: true,
                notifications: false,
                beta_features: true,
            },
            environment: 'development',
            tenant_id: 'default',
            version: 1,
            description: 'Feature enablement configuration',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            metadata: { type: 'object' },
        },
    ];

    useEffect(() => {
        loadConfigs();
    }, [environment]);

    const loadConfigs = async () => {
        setLoading(true);
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setConfigs(mockConfigs);
        } catch (error) {
            message.error('Failed to load configurations');
        } finally {
            setLoading(false);
        }
    };

    const handleCreateConfig = () => {
        setEditingConfig(null);
        form.resetFields();
        setModalVisible(true);
    };

    const handleEditConfig = (config: Configuration) => {
        setEditingConfig(config);
        form.setFieldsValue({
            key: config.key,
            value: typeof config.value === 'object' ? JSON.stringify(config.value, null, 2) : config.value,
            description: config.description,
            type: config.metadata.type || 'string',
        });
        setModalVisible(true);
    };

    const handleDeleteConfig = async (configId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setConfigs(configs.filter(config => config.id !== configId));
            message.success('Configuration deleted successfully');
        } catch (error) {
            message.error('Failed to delete configuration');
        }
    };

    const handleSubmit = async (values: any) => {
        try {
            // Parse value based on type
            let parsedValue = values.value;
            switch (values.type) {
                case 'number':
                    parsedValue = Number(values.value);
                    break;
                case 'boolean':
                    parsedValue = values.value === 'true' || values.value === true;
                    break;
                case 'object':
                    try {
                        parsedValue = JSON.parse(values.value);
                    } catch (e) {
                        message.error('Invalid JSON format');
                        return;
                    }
                    break;
                case 'array':
                    try {
                        parsedValue = JSON.parse(values.value);
                        if (!Array.isArray(parsedValue)) {
                            throw new Error('Not an array');
                        }
                    } catch (e) {
                        message.error('Invalid array format');
                        return;
                    }
                    break;
            }

            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));

            if (editingConfig) {
                // Update existing config
                setConfigs(configs.map(config =>
                    config.id === editingConfig.id
                        ? {
                            ...config,
                            value: parsedValue,
                            description: values.description,
                            version: config.version + 1,
                            updated_at: new Date().toISOString(),
                            metadata: { ...config.metadata, type: values.type }
                        }
                        : config
                ));
                message.success('Configuration updated successfully');
            } else {
                // Create new config
                const newConfig: Configuration = {
                    id: `config_${Date.now()}`,
                    key: values.key,
                    value: parsedValue,
                    description: values.description,
                    environment,
                    tenant_id: 'default',
                    version: 1,
                    created_by: 'current_user',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    metadata: { type: values.type },
                };
                setConfigs([...configs, newConfig]);
                message.success('Configuration created successfully');
            }

            setModalVisible(false);
            form.resetFields();
        } catch (error) {
            message.error('Failed to save configuration');
        }
    };

    const handleReloadConfig = async (configId: string) => {
        try {
            // Mock API call to reload config
            await new Promise(resolve => setTimeout(resolve, 300));
            message.success('Configuration reloaded successfully');
        } catch (error) {
            message.error('Failed to reload configuration');
        }
    };

    const columns = [
        {
            title: 'Key',
            dataIndex: 'key',
            key: 'key',
            render: (key: string, record: Configuration) => (
                <div>
                    <div style={{ fontWeight: 500 }}>{key}</div>
                    <div style={{ fontSize: '12px', color: '#666' }}>{record.description}</div>
                </div>
            ),
        },
        {
            title: 'Value',
            dataIndex: 'value',
            key: 'value',
            render: (value: any, record: Configuration) => {
                const displayValue = typeof value === 'object'
                    ? JSON.stringify(value, null, 2)
                    : String(value);

                return (
                    <div style={{ maxWidth: 200 }}>
                        <div style={{
                            fontFamily: 'monospace',
                            background: '#f5f5f5',
                            padding: '4px 8px',
                            borderRadius: '4px',
                            fontSize: '12px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap'
                        }}>
                            {displayValue}
                        </div>
                        <Tag size="small" style={{ marginTop: 4 }}>
                            {record.metadata.type}
                        </Tag>
                    </div>
                );
            },
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
            title: 'Version',
            dataIndex: 'version',
            key: 'version',
            render: (version: number) => (
                <Tag color="purple">v{version}</Tag>
            ),
        },
        {
            title: 'Last Updated',
            dataIndex: 'updated_at',
            key: 'updated_at',
            render: (date: string) => new Date(date).toLocaleString(),
        },
        {
            title: 'Actions',
            key: 'actions',
            render: (_: any, record: Configuration) => (
                <Space>
                    <Button
                        type="link"
                        icon={<ReloadOutlined />}
                        onClick={() => handleReloadConfig(record.id)}
                        size="small"
                    >
                        Reload
                    </Button>
                    <Button
                        type="link"
                        icon={<EyeOutlined />}
                        onClick={() => handleEditConfig(record)}
                        size="small"
                    >
                        View
                    </Button>
                    <Button
                        type="link"
                        icon={<EditOutlined />}
                        onClick={() => handleEditConfig(record)}
                        size="small"
                    >
                        Edit
                    </Button>
                    <Popconfirm
                        title="Are you sure you want to delete this configuration?"
                        onConfirm={() => handleDeleteConfig(record.id)}
                        okText="Yes"
                        cancelText="No"
                    >
                        <Button
                            type="link"
                            danger
                            icon={<DeleteOutlined />}
                            size="small"
                        >
                            Delete
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Card
                title="Configuration Center"
                extra={
                    <Space>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={handleCreateConfig}
                        >
                            Add Configuration
                        </Button>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadConfigs}
                        >
                            Refresh
                        </Button>
                    </Space>
                }
            >
                <Tabs
                    activeKey={environment}
                    onChange={setEnvironment}
                    items={[
                        { key: 'development', label: 'Development' },
                        { key: 'staging', label: 'Staging' },
                        { key: 'production', label: 'Production' },
                    ]}
                />

                <Table
                    columns={columns}
                    dataSource={configs}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total, range) =>
                            `${range[0]}-${range[1]} of ${total} configurations`,
                    }}
                />
            </Card>

            <Modal
                title={editingConfig ? 'Edit Configuration' : 'Create Configuration'}
                open={modalVisible}
                onCancel={() => setModalVisible(false)}
                footer={null}
                width={600}
            >
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleSubmit}
                    initialValues={{
                        type: 'string',
                    }}
                >
                    <Form.Item
                        name="key"
                        label="Configuration Key"
                        rules={[
                            { required: true, message: 'Please enter configuration key' },
                            { pattern: /^[a-z0-9._]+$/, message: 'Only lowercase letters, numbers, dots, and underscores allowed' },
                        ]}
                    >
                        <Input
                            placeholder="e.g., app.title, database.url, features.enabled"
                            disabled={!!editingConfig}
                        />
                    </Form.Item>

                    <Form.Item
                        name="type"
                        label="Value Type"
                        rules={[{ required: true, message: 'Please select value type' }]}
                    >
                        <Select placeholder="Select value type">
                            <Option value="string">String</Option>
                            <Option value="number">Number</Option>
                            <Option value="boolean">Boolean</Option>
                            <Option value="object">Object (JSON)</Option>
                            <Option value="array">Array (JSON)</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item
                        name="value"
                        label="Value"
                        rules={[{ required: true, message: 'Please enter configuration value' }]}
                    >
                        <TextArea
                            placeholder="Enter configuration value"
                            rows={4}
                            style={{ fontFamily: 'monospace' }}
                        />
                    </Form.Item>

                    <Form.Item
                        name="description"
                        label="Description"
                        rules={[{ required: true, message: 'Please enter description' }]}
                    >
                        <TextArea
                            placeholder="Describe what this configuration controls"
                            rows={2}
                        />
                    </Form.Item>

                    <Form.Item style={{ marginBottom: 0 }}>
                        <Space>
                            <Button type="primary" htmlType="submit">
                                {editingConfig ? 'Update' : 'Create'} Configuration
                            </Button>
                            <Button onClick={() => setModalVisible(false)}>
                                Cancel
                            </Button>
                        </Space>
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default ConfigurationCenter;