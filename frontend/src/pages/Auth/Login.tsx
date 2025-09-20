import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Layout,
    Card,
    Form,
    Input,
    Button,
    Typography,
    Space,
    Divider,
    message,
    Checkbox,
} from 'antd';
import {
    UserOutlined,
    LockOutlined,
    SafetyCertificateOutlined,
} from '@ant-design/icons';

const { Content } = Layout;
const { Title, Text } = Typography;

const Login: React.FC = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            // 模拟登录API调用
            await new Promise(resolve => setTimeout(resolve, 1000));

            // 模拟登录成功
            if (values.username === 'admin' && values.password === 'admin123') {
                localStorage.setItem('auth_token', 'mock-jwt-token');
                message.success('登录成功');
                navigate('/dashboard');
            } else {
                message.error('用户名或密码错误');
            }
        } catch (error) {
            message.error('登录失败，请重试');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Layout style={{ minHeight: '100vh', background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}>
            <Content style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '50px' }}>
                <Card
                    style={{
                        width: 400,
                        boxShadow: '0 10px 30px rgba(0, 0, 0, 0.2)',
                        borderRadius: 12,
                        overflow: 'hidden'
                    }}
                    bodyStyle={{ padding: '40px 40px 30px' }}
                >
                    <div style={{ textAlign: 'center', marginBottom: 32 }}>
                        <div
                            style={{
                                width: 64,
                                height: 64,
                                background: 'linear-gradient(135deg, #1890ff, #722ed1)',
                                borderRadius: 12,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                margin: '0 auto 16px',
                            }}
                        >
                            <SafetyCertificateOutlined style={{ fontSize: 32, color: 'white' }} />
                        </div>
                        <Title level={2} style={{ margin: '0 0 8px 0', color: '#1f1f1f' }}>
                            Multi-Agent 平台
                        </Title>
                        <Text type="secondary">
                            智能多代理工作流管理系统
                        </Text>
                    </div>

                    <Form
                        name="login"
                        onFinish={onFinish}
                        autoComplete="off"
                        size="large"
                    >
                        <Form.Item
                            name="username"
                            rules={[
                                { required: true, message: '请输入用户名' },
                                { min: 3, message: '用户名至少3个字符' }
                            ]}
                        >
                            <Input
                                prefix={<UserOutlined style={{ color: '#999' }} />}
                                placeholder="用户名"
                                style={{ borderRadius: 8 }}
                            />
                        </Form.Item>

                        <Form.Item
                            name="password"
                            rules={[
                                { required: true, message: '请输入密码' },
                                { min: 6, message: '密码至少6个字符' }
                            ]}
                        >
                            <Input.Password
                                prefix={<LockOutlined style={{ color: '#999' }} />}
                                placeholder="密码"
                                style={{ borderRadius: 8 }}
                            />
                        </Form.Item>

                        <Form.Item>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Form.Item name="remember" valuePropName="checked" noStyle>
                                    <Checkbox>记住我</Checkbox>
                                </Form.Item>
                                <Button type="link" style={{ padding: 0 }}>
                                    忘记密码？
                                </Button>
                            </div>
                        </Form.Item>

                        <Form.Item>
                            <Button
                                type="primary"
                                htmlType="submit"
                                loading={loading}
                                block
                                style={{
                                    height: 48,
                                    borderRadius: 8,
                                    background: 'linear-gradient(135deg, #1890ff, #722ed1)',
                                    border: 'none',
                                    fontSize: 16,
                                    fontWeight: 500,
                                }}
                            >
                                登录
                            </Button>
                        </Form.Item>
                    </Form>

                    <Divider style={{ margin: '24px 0' }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                            演示账号
                        </Text>
                    </Divider>

                    <div style={{ textAlign: 'center' }}>
                        <Space direction="vertical" size="small">
                            <Text type="secondary" style={{ fontSize: 12 }}>
                                用户名: admin | 密码: admin123
                            </Text>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                                用户名: operator | 密码: operator123
                            </Text>
                        </Space>
                    </div>

                    <div style={{ textAlign: 'center', marginTop: 24 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                            © 2024 Multi-Agent Platform. All rights reserved.
                        </Text>
                    </div>
                </Card>
            </Content>
        </Layout>
    );
};

export default Login;