import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Modal, Form, Input, Switch, Tag, Space, Tabs, message, Popconfirm, Badge, Progress } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, HistoryOutlined, ReloadOutlined } from '@ant-design/icons';

const { TextArea } = Input;

interface CronJob {
    id: string;
    name: string;
    description: string;
    schedule: string;
    command: string;
    enabled: boolean;
    timeout: number;
    retries: number;
    tenant_id: string;
    created_by: string;
    created_at: string;
    updated_at: string;
    last_run?: string;
    next_run?: string;
    success_rate: number;
    total_runs: number;
}

interface Execution {
    id: string;
    job_id: string;
    status: 'scheduled' | 'running' | 'completed' | 'failed' | 'timeout';
    started_at: string;
    finished_at?: string;
    duration: number;
    exit_code?: number;
    output: string;
    error?: string;
    attempt: number;
    trigger_type: 'scheduled' | 'manual';
}

const CronJobsManagement: React.FC = () => {
    const [jobs, setJobs] = useState<CronJob[]>([]);
    const [executions, setExecutions] = useState<Execution[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalVisible, setModalVisible] = useState(false);
    const [executionModalVisible, setExecutionModalVisible] = useState(false);
    const [editingJob, setEditingJob] = useState<CronJob | null>(null);
    const [activeTab, setActiveTab] = useState('jobs');
    const [form] = Form.useForm();

    // Mock data for demonstration
    const mockJobs: CronJob[] = [
        {
            id: 'job_001',
            name: 'cleanup_logs',
            description: 'Clean up application logs older than 30 days',
            schedule: '0 2 * * *',
            command: '/scripts/cleanup-logs.sh',
            enabled: true,
            timeout: 300,
            retries: 2,
            tenant_id: 'default',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            last_run: '2024-01-01T02:00:00Z',
            next_run: '2024-01-02T02:00:00Z',
            success_rate: 95.5,
            total_runs: 128,
        },
        {
            id: 'job_002',
            name: 'backup_database',
            description: 'Create daily database backup',
            schedule: '0 1 * * *',
            command: '/scripts/backup-db.sh',
            enabled: true,
            timeout: 1800,
            retries: 3,
            tenant_id: 'default',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            last_run: '2024-01-01T01:00:00Z',
            next_run: '2024-01-02T01:00:00Z',
            success_rate: 98.2,
            total_runs: 55,
        },
        {
            id: 'job_003',
            name: 'health_check',
            description: 'Monitor system health and send alerts',
            schedule: '*/5 * * * *',
            command: '/scripts/health-check.sh',
            enabled: false,
            timeout: 60,
            retries: 1,
            tenant_id: 'default',
            created_by: 'admin',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            success_rate: 87.3,
            total_runs: 2880,
        },
    ];

    const mockExecutions: Execution[] = [
        {
            id: 'exec_001',
            job_id: 'job_001',
            status: 'completed',
            started_at: '2024-01-01T02:00:00Z',
            finished_at: '2024-01-01T02:05:00Z',
            duration: 300,
            exit_code: 0,
            output: 'Successfully cleaned 1,234 log files (2.3GB freed)',
            attempt: 1,
            trigger_type: 'scheduled',
        },
        {
            id: 'exec_002',
            job_id: 'job_002',
            status: 'completed',
            started_at: '2024-01-01T01:00:00Z',
            finished_at: '2024-01-01T01:15:00Z',
            duration: 900,
            exit_code: 0,
            output: 'Database backup completed successfully (backup size: 4.7GB)',
            attempt: 1,
            trigger_type: 'scheduled',
        },
        {
            id: 'exec_003',
            job_id: 'job_001',
            status: 'failed',
            started_at: '2023-12-31T02:00:00Z',
            finished_at: '2023-12-31T02:01:00Z',
            duration: 60,
            exit_code: 1,
            output: 'Error: Permission denied when accessing /var/log/app/',
            error: 'Script failed with exit code 1',
            attempt: 2,
            trigger_type: 'scheduled',
        },
    ];

    useEffect(() => {
        loadJobs();
    }, []);

    const loadJobs = async () => {
        setLoading(true);
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setJobs(mockJobs);
        } catch (error) {
            message.error('Failed to load cron jobs');
        } finally {
            setLoading(false);
        }
    };

    const loadExecutions = async (jobId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 300));
            const jobExecutions = mockExecutions.filter(exec => exec.job_id === jobId);
            setExecutions(jobExecutions);
        } catch (error) {
            message.error('Failed to load executions');
        }
    };

    const handleCreateJob = () => {
        setEditingJob(null);
        form.resetFields();
        setModalVisible(true);
    };

    const handleEditJob = (job: CronJob) => {
        setEditingJob(job);
        form.setFieldsValue({
            name: job.name,
            description: job.description,
            schedule: job.schedule,
            command: job.command,
            enabled: job.enabled,
            timeout: job.timeout,
            retries: job.retries,
        });
        setModalVisible(true);
    };

    const handleDeleteJob = async (jobId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));
            setJobs(jobs.filter(job => job.id !== jobId));
            message.success('Cron job deleted successfully');
        } catch (error) {
            message.error('Failed to delete cron job');
        }
    };

    const handleToggleJob = async (jobId: string, enabled: boolean) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 300));
            setJobs(jobs.map(job =>
                job.id === jobId ? { ...job, enabled } : job
            ));
            message.success(`Cron job ${enabled ? 'enabled' : 'disabled'}`);
        } catch (error) {
            message.error('Failed to update cron job');
        }
    };

    const handleTriggerJob = async (jobId: string) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));

            // Create new execution record
            const newExecution: Execution = {
                id: `exec_${Date.now()}`,
                job_id: jobId,
                status: 'running',
                started_at: new Date().toISOString(),
                duration: 0,
                output: '',
                attempt: 1,
                trigger_type: 'manual',
            };

            setExecutions([newExecution, ...executions]);
            message.success('Cron job triggered successfully');

            // Simulate completion after 3 seconds
            setTimeout(() => {
                setExecutions(prev => prev.map(exec =>
                    exec.id === newExecution.id
                        ? {
                            ...exec,
                            status: 'completed',
                            finished_at: new Date().toISOString(),
                            duration: 3,
                            exit_code: 0,
                            output: 'Manual execution completed successfully'
                        }
                        : exec
                ));
            }, 3000);
        } catch (error) {
            message.error('Failed to trigger cron job');
        }
    };

    const handleViewExecutions = (jobId: string) => {
        loadExecutions(jobId);
        setExecutionModalVisible(true);
    };

    const handleSubmit = async (values: any) => {
        try {
            // Mock API call
            await new Promise(resolve => setTimeout(resolve, 500));

            if (editingJob) {
                // Update existing job
                setJobs(jobs.map(job =>
                    job.id === editingJob.id
                        ? { ...job, ...values, updated_at: new Date().toISOString() }
                        : job
                ));
                message.success('Cron job updated successfully');
            } else {
                // Create new job
                const newJob: CronJob = {
                    id: `job_${Date.now()}`,
                    ...values,
                    tenant_id: 'default',
                    created_by: 'current_user',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    success_rate: 0,
                    total_runs: 0,
                };
                setJobs([...jobs, newJob]);
                message.success('Cron job created successfully');
            }

            setModalVisible(false);
            form.resetFields();
        } catch (error) {
            message.error('Failed to save cron job');
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'completed': return 'success';
            case 'running': return 'processing';
            case 'failed': return 'error';
            case 'timeout': return 'warning';
            default: return 'default';
        }
    };

    const jobColumns = [
        {
            title: 'Name',
            dataIndex: 'name',
            key: 'name',
            render: (name: string, record: CronJob) => (
                <div>
                    <div style={{ fontWeight: 500, display: 'flex', alignItems: 'center' }}>
                        {name}
                        {record.enabled ? (
                            <Badge status="success" style={{ marginLeft: 8 }} />
                        ) : (
                            <Badge status="default" style={{ marginLeft: 8 }} />
                        )}
                    </div>
                    <div style={{ fontSize: '12px', color: '#666' }}>{record.description}</div>
                </div>
            ),
        },
        {
            title: 'Schedule',
            dataIndex: 'schedule',
            key: 'schedule',
            render: (schedule: string) => (
                <Tag style={{ fontFamily: 'monospace' }}>{schedule}</Tag>
            ),
        },
        {
            title: 'Success Rate',
            dataIndex: 'success_rate',
            key: 'success_rate',
            render: (rate: number, record: CronJob) => (
                <div>
                    <Progress
                        percent={rate}
                        size="small"
                        status={rate > 90 ? 'success' : rate > 70 ? 'normal' : 'exception'}
                    />
                    <div style={{ fontSize: '12px', color: '#666' }}>
                        {record.total_runs} total runs
                    </div>
                </div>
            ),
        },
        {
            title: 'Last Run',
            dataIndex: 'last_run',
            key: 'last_run',
            render: (date: string) => (
                date ? new Date(date).toLocaleString() : 'Never'
            ),
        },
        {
            title: 'Next Run',
            dataIndex: 'next_run',
            key: 'next_run',
            render: (date: string) => (
                date ? new Date(date).toLocaleString() : 'N/A'
            ),
        },
        {
            title: 'Status',
            dataIndex: 'enabled',
            key: 'enabled',
            render: (enabled: boolean, record: CronJob) => (
                <Switch
                    checked={enabled}
                    onChange={(checked) => handleToggleJob(record.id, checked)}
                    checkedChildren="ON"
                    unCheckedChildren="OFF"
                />
            ),
        },
        {
            title: 'Actions',
            key: 'actions',
            render: (_: any, record: CronJob) => (
                <Space>
                    <Button
                        type="link"
                        icon={<PlayCircleOutlined />}
                        onClick={() => handleTriggerJob(record.id)}
                        size="small"
                    >
                        Trigger
                    </Button>
                    <Button
                        type="link"
                        icon={<HistoryOutlined />}
                        onClick={() => handleViewExecutions(record.id)}
                        size="small"
                    >
                        History
                    </Button>
                    <Button
                        type="link"
                        icon={<EditOutlined />}
                        onClick={() => handleEditJob(record)}
                        size="small"
                    >
                        Edit
                    </Button>
                    <Popconfirm
                        title="Are you sure you want to delete this cron job?"
                        onConfirm={() => handleDeleteJob(record.id)}
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

    const executionColumns = [
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => (
                <Badge status={getStatusColor(status)} text={status.toUpperCase()} />
            ),
        },
        {
            title: 'Started',
            dataIndex: 'started_at',
            key: 'started_at',
            render: (date: string) => new Date(date).toLocaleString(),
        },
        {
            title: 'Duration',
            dataIndex: 'duration',
            key: 'duration',
            render: (duration: number) => `${duration}s`,
        },
        {
            title: 'Exit Code',
            dataIndex: 'exit_code',
            key: 'exit_code',
            render: (code: number) => (
                code !== undefined ? (
                    <Tag color={code === 0 ? 'green' : 'red'}>{code}</Tag>
                ) : '-'
            ),
        },
        {
            title: 'Trigger',
            dataIndex: 'trigger_type',
            key: 'trigger_type',
            render: (type: string) => (
                <Tag color={type === 'manual' ? 'blue' : 'default'}>
                    {type.toUpperCase()}
                </Tag>
            ),
        },
        {
            title: 'Output',
            dataIndex: 'output',
            key: 'output',
            render: (output: string) => (
                <div style={{
                    maxWidth: 300,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    fontFamily: 'monospace',
                    fontSize: '12px'
                }}>
                    {output}
                </div>
            ),
        },
    ];

    return (
        <div>
            <Card
                title="CronJob Management"
                extra={
                    <Space>
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={handleCreateJob}
                        >
                            Create Job
                        </Button>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadJobs}
                        >
                            Refresh
                        </Button>
                    </Space>
                }
            >
                <Tabs
                    activeKey={activeTab}
                    onChange={setActiveTab}
                    items={[
                        {
                            key: 'jobs',
                            label: `Jobs (${jobs.length})`,
                            children: (
                                <Table
                                    columns={jobColumns}
                                    dataSource={jobs}
                                    rowKey="id"
                                    loading={loading}
                                    pagination={{
                                        pageSize: 10,
                                        showSizeChanger: true,
                                        showQuickJumper: true,
                                        showTotal: (total, range) =>
                                            `${range[0]}-${range[1]} of ${total} cron jobs`,
                                    }}
                                />
                            ),
                        },
                        {
                            key: 'monitoring',
                            label: 'Monitoring',
                            children: (
                                <div>
                                    <p>System monitoring and alerts for cron jobs will be displayed here.</p>
                                </div>
                            ),
                        },
                    ]}
                />
            </Card>

            {/* Create/Edit Job Modal */}
            <Modal
                title={editingJob ? 'Edit Cron Job' : 'Create Cron Job'}
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
                        enabled: true,
                        timeout: 300,
                        retries: 2,
                    }}
                >
                    <Form.Item
                        name="name"
                        label="Job Name"
                        rules={[
                            { required: true, message: 'Please enter job name' },
                            { pattern: /^[a-z0-9_]+$/, message: 'Only lowercase letters, numbers, and underscores allowed' },
                        ]}
                    >
                        <Input placeholder="e.g., cleanup_logs, backup_database" />
                    </Form.Item>

                    <Form.Item
                        name="description"
                        label="Description"
                        rules={[{ required: true, message: 'Please enter description' }]}
                    >
                        <TextArea
                            placeholder="Describe what this job does"
                            rows={2}
                        />
                    </Form.Item>

                    <Form.Item
                        name="schedule"
                        label="Cron Schedule"
                        rules={[{ required: true, message: 'Please enter cron schedule' }]}
                    >
                        <Input
                            placeholder="e.g., 0 2 * * * (daily at 2 AM)"
                            style={{ fontFamily: 'monospace' }}
                        />
                    </Form.Item>

                    <Form.Item
                        name="command"
                        label="Command"
                        rules={[{ required: true, message: 'Please enter command' }]}
                    >
                        <Input
                            placeholder="e.g., /scripts/cleanup.sh, python /app/backup.py"
                            style={{ fontFamily: 'monospace' }}
                        />
                    </Form.Item>

                    <Space style={{ width: '100%' }} size="large">
                        <Form.Item
                            name="timeout"
                            label="Timeout (seconds)"
                            rules={[{ required: true, message: 'Please enter timeout' }]}
                        >
                            <Input type="number" min={1} max={3600} />
                        </Form.Item>

                        <Form.Item
                            name="retries"
                            label="Retry Attempts"
                            rules={[{ required: true, message: 'Please enter retry count' }]}
                        >
                            <Input type="number" min={0} max={5} />
                        </Form.Item>
                    </Space>

                    <Form.Item
                        name="enabled"
                        label="Enabled"
                        valuePropName="checked"
                    >
                        <Switch checkedChildren="ON" unCheckedChildren="OFF" />
                    </Form.Item>

                    <Form.Item style={{ marginBottom: 0 }}>
                        <Space>
                            <Button type="primary" htmlType="submit">
                                {editingJob ? 'Update' : 'Create'} Job
                            </Button>
                            <Button onClick={() => setModalVisible(false)}>
                                Cancel
                            </Button>
                        </Space>
                    </Form.Item>
                </Form>
            </Modal>

            {/* Execution History Modal */}
            <Modal
                title="Execution History"
                open={executionModalVisible}
                onCancel={() => setExecutionModalVisible(false)}
                footer={null}
                width={1000}
            >
                <Table
                    columns={executionColumns}
                    dataSource={executions}
                    rowKey="id"
                    pagination={{
                        pageSize: 5,
                        showSizeChanger: false,
                    }}
                    expandable={{
                        expandedRowRender: (record: Execution) => (
                            <div style={{ margin: 0 }}>
                                <div style={{ marginBottom: 8 }}>
                                    <strong>Output:</strong>
                                </div>
                                <pre style={{
                                    background: '#f5f5f5',
                                    padding: 12,
                                    borderRadius: 4,
                                    fontSize: 12,
                                    maxHeight: 200,
                                    overflow: 'auto'
                                }}>
                                    {record.output}
                                    {record.error && (
                                        <>
                                            {'\n\nError:\n'}
                                            {record.error}
                                        </>
                                    )}
                                </pre>
                            </div>
                        ),
                        rowExpandable: (record: Execution) => !!record.output,
                    }}
                />
            </Modal>
        </div>
    );
};

export default CronJobsManagement;