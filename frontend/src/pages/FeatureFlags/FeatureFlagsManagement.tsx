import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Modal, Form, Input, Switch, Tag, Space, Tabs, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SettingOutlined } from '@ant-design/icons';

const { TabPane } = Tabs;
const { TextArea } = Input;

interface FeatureFlag {
    id: string;
    name: string;
    description: string;
    enabled: boolean;
    environment: string;
    tenant_id: string;
    rules: Rule[];
    rollout?: RolloutConfig;
    created_by: string;
    created_at: string;
    updated_at: string;
}

interface Rule {
    id: string;
    attribute: string;
    operator: string;
    values: string[];
    percentage?: number;
    enabled: boolean;
    priority: number;
}

interface RolloutConfig {
    strategy: string;
    percentage: number;
    user_groups?: string[];
    start_time?: string;
    end_time?: string;
}

const FeatureFlagsManagement: React.FC = () => {
    const [flags, setFlags] = useState<FeatureFlag[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [editingFlag, setEditingFlag] = useState<FeatureFlag | null>(null);
    const [environment, setEnvironment] = useState('development');
    const [form] = Form.useForm();

    // Mock data for demonstration
    const mockFlags: FeatureFlag[] = [
        {
            id: 'flag_001',
            name: 'new_ui_enabled',
            description: 'Enable new UI features for enhanced user experience',
            enabled: true,
            environment: 'development',
            tenant_id: 'default',
            rules: [
                {
                    id: 'rule_001',
                    attribute: 'user_id',
                    operator: 'in',
                    values: ['admin', 'beta_user'],
                    enabled: true,
                    priority: 1,
                },
            ],
            rollout: {
                strategy: 'percentage',
                percentage: 25,
            },
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T12:00:00Z',
        },
        {
            id: 'flag_002',
            name: 'advanced_analytics',
            description: 'Enable advanced analytics dashboard',
            enabled: false,
            environment: 'development',
            tenant_id: 'default',
            rules: [],
            created_by: 'admin',
            created_at: '2024-01-02T00:00:00Z',
            updated_at: '2024-01-02T00:00:00Z',
        },
    ];

    useEffect(() => {
        loadFlags();
    }, [environment]);

    const loadFlags = async () => {
        setLoading(true);
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setFlags(mockFlags);
        } catch (error) {
            message.error('Failed to load feature flags');
        } finally {
            setLoading(false);
        }
    };

    const handleCreateFlag = () => {
        setEditingFlag(null);
        form.resetFields();
        setModalVisible(true);
    };

    const handleEditFlag = (flag: FeatureFlag) => {
        setEditingFlag(flag);
        form.setFieldsValue({
            name: flag.name,
            description: flag.description,
            enabled: flag.enabled,
            rules: flag.rules,
            rollout: flag.rollout,
        });
        setModalVisible(true);
    };

    const handleDeleteFlag = async (flagId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setFlags(flags.filter(flag => flag.id !== flagId));
            message.success('Feature flag deleted successfully');
        } catch (error) {
            message.error('Failed to delete feature flag');
        }
    };

    const handleToggleFlag = async (flagId: string, enabled: boolean) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 300));
            setFlags(flags.map(flag =>
                flag.id === flagId ? { ...flag, enabled } : flag
            ));
            message.success(`Feature flag ${enabled ? 'enabled' : 'disabled'}`);
        } catch (error) {
            message.error('Failed to update feature flag');
        }
    };

    const handleSubmit = async (values: any) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));

            if (editingFlag) {
                // Update existing flag
                setFlags(flags.map(flag =>
                    flag.id === editingFlag.id
                        ? { ...flag, ...values, updated_at: new Date().toISOString() }
                        : flag
                ));
                message.success('Feature flag updated successfully');
            } else {
                // Create new flag
                const newFlag: FeatureFlag = {
                    id: `flag_${Date.now()}`,
                    ...values,
                    environment,
                    tenant_id: 'default',
                    created_by: 'current_user',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                };
                setFlags([...flags, newFlag]);
                message.success('Feature flag created successfully');
            }

            setModalVisible(false);
            form.resetFields();
        } catch (error) {
            message.error('Failed to save feature flag');
        }
    };

    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
            render: (name: string, record: FeatureFlag) => (
                <div>
                    <div style={{ fontWeight: 500 }}>{name}</div>
                    <div style={{ fontSize: '12px', color: '#666' }}>{record.description}</div>
                </div>
            ),
        },
        {
            title: 'Status',
            dataIndex: 'enabled',
            key: 'enabled',
            render: (enabled: boolean, record: FeatureFlag) => (
                <Switch
                    checked={enabled}
                    onChange={(checked: boolean) => handleToggleFlag(record.id, checked)}
                    checkedChildren="ON"
                    unCheckedChildren="OFF"
                />
            ),
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
            title: 'Rules',
            dataIndex: 'rules',
            key: 'rules',
            render: (rules: Rule[]) => (
                <Tag>{rules.length} rule{rules.length !== 1 ? 's' : ''}</Tag>
            ),
        },
        {
            title: 'Rollout',
            dataIndex: 'rollout',
            key: 'rollout',
            render: (rollout: RolloutConfig) => (
                rollout ? (
                    <Tag color="green">
                        {rollout.strategy} ({rollout.percentage}%)
                    </Tag>
                ) : (
                    <Tag color="default">No rollout</Tag>
                )
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
            render: (_: any, record: FeatureFlag) => (
                <Space>
                    <Button
                        type="link"
                        icon={<EyeOutlined />}
                        onClick={() => handleEditFlag(record)}
                    >
                        View
                    </Button>
                    <Button
                        type="link"
                        icon={<EditOutlined />}
                        onClick={() => handleEditFlag(record)}
                    >
                        Edit
                    </Button>
                    <Popconfirm
                        title="Are you sure you want to delete this feature flag?"
                        onConfirm={() => handleDeleteFlag(record.id)}
                        okText="Yes"
                        cancelText="No"
                    >
                        <Button
                            type="link"
                            danger
                            icon={<DeleteOutlined />}
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
                title="Feature Flags Management"
                extra={
                    <Space>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={handleCreateFlag}
                        >
                            Create Flag
                        </Button>
                        <Button
                            icon={<SettingOutlined />}
                            onClick={() => {/* Open settings */ }}
                        >
                            Settings
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
                    dataSource={flags}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total, range) =>
                            `${range[0]}-${range[1]} of ${total} feature flags`,
                    }}
                />
            </Card>

            <Modal
                title={editingFlag ? 'Edit Feature Flag' : 'Create Feature Flag'}
                open={modalVisible}
                onCancel={() => setModalVisible(false)}
                footer={null}
                width={800}
            >
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleSubmit}
                    initialValues={{
                        enabled: true,
                        rules: [],
                    }}
                >
                    <Form.Item
                        name="name"
                        label="Flag Name"
                        rules={[
                            { required: true, message: 'Please enter flag name' },
                            { pattern: /^[a-z0-9_]+$/, message: 'Only lowercase letters, numbers, and underscores allowed' },
                        ]}
                    >
                        <Input placeholder="e.g., new_feature_enabled" />
                    </Form.Item>

                    <Form.Item
                        name="description"
                        label="Description"
                        rules={[{ required: true, message: 'Please enter description' }]}
                    >
                        <TextArea
                            placeholder="Describe what this feature flag controls"
                            rows={3}
                        />
                    </Form.Item>

                    <Form.Item
                        name="enabled"
                        label="Enabled"
                        valuePropName="checked"
                    >
                        <Switch checkedChildren="ON" unCheckedChildren="OFF" />
                    </Form.Item>

                    <Tabs defaultActiveKey="basic">
                        <TabPane tab="Basic Settings" key="basic">
                            {/* Basic settings already covered above */}
                        </TabPane>

                        <TabPane tab="Targeting Rules" key="rules">
                            <Form.List name="rules">
                                {(fields, { add, remove }) => (
                                    <>
                                        {fields.map(({ key, name, ...restField }) => (
                                            <Card
                                                key={key}
                                                size="small"
                                                title={`Rule ${name + 1}`}
                                                extra={
                                                    <Button
                                                        type="link"
                                                        danger
                                                        onClick={() => remove(name)}
                                                    >
                                                        Remove
                                                    </Button>
                                                }
                                                style={{ marginBottom: 16 }}
                                            >
                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'attribute']}
                                                    label="Attribute"
                                                    rules={[{ required: true, message: 'Missing attribute' }]}
                                                >
                                                    <Input placeholder="e.g., user_id, group, email" />
                                                </Form.Item>

                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'operator']}
                                                    label="Operator"
                                                    rules={[{ required: true, message: 'Missing operator' }]}
                                                >
                                                    <Input placeholder="e.g., equals, in, contains" />
                                                </Form.Item>

                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'values']}
                                                    label="Values"
                                                    rules={[{ required: true, message: 'Missing values' }]}
                                                >
                                                    <Input placeholder="Comma-separated values" />
                                                </Form.Item>

                                                <Form.Item
                                                    {...restField}
                                                    name={[name, 'enabled']}
                                                    valuePropName="checked"
                                                    initialValue={true}
                                                >
                                                    <Switch size="small" /> Enabled
                                                </Form.Item>
                                            </Card>
                                        ))}
                                        <Button
                                            type="dashed"
                                            onClick={() => add()}
                                            block
                                            icon={<PlusOutlined />}
                                        >
                                            Add Targeting Rule
                                        </Button>
                                    </>
                                )}
                            </Form.List>
                        </TabPane>

                        <TabPane tab="Rollout Strategy" key="rollout">
                            <Form.Item
                                name={['rollout', 'strategy']}
                                label="Rollout Strategy"
                            >
                                <Input placeholder="e.g., percentage, user_group, time_based" />
                            </Form.Item>

                            <Form.Item
                                name={['rollout', 'percentage']}
                                label="Percentage"
                            >
                                <Input
                                    type="number"
                                    min={0}
                                    max={100}
                                    placeholder="0-100"
                                    suffix="%"
                                />
                            </Form.Item>

                            <Form.Item
                                name={['rollout', 'user_groups']}
                                label="User Groups"
                            >
                                <Input placeholder="Comma-separated group names" />
                            </Form.Item>
                        </TabPane>
                    </Tabs>

                    <Form.Item style={{ marginTop: 24, marginBottom: 0 }}>
                        <Space>
                            <Button type="primary" htmlType="submit">
                                {editingFlag ? 'Update' : 'Create'} Flag
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

export default FeatureFlagsManagement;