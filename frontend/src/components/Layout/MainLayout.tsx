import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
    Layout,
    Menu,
    Avatar,
    Dropdown,
    Button,
    Typography,
    Space,
    Badge,
} from 'antd';
import {
    DashboardOutlined,
    RobotOutlined,
    ApartmentOutlined,
    SettingOutlined,
    UserOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    BellOutlined,
    FlagOutlined,
    ControlOutlined,
    ScheduleOutlined,
    CloudServerOutlined,
} from '@ant-design/icons';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const MainLayout: React.FC = () => {
    const [collapsed, setCollapsed] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();

    const menuItems = [
        {
            key: '/dashboard',
            icon: <DashboardOutlined />,
            label: '仪表板',
        },
        {
            key: '/agents',
            icon: <RobotOutlined />,
            label: 'Agent管理',
        },
        {
            key: '/workflows',
            icon: <ApartmentOutlined />,
            label: '工作流',
        },
        {
            key: '/feature-flags',
            icon: <FlagOutlined />,
            label: '特性开关',
        },
        {
            key: '/configuration',
            icon: <ControlOutlined />,
            label: '配置中心',
        },
        {
            key: '/cronjobs',
            icon: <ScheduleOutlined />,
            label: '定时任务',
        },
        {
            key: '/service-discovery',
            icon: <CloudServerOutlined />,
            label: '服务发现',
        },
        {
            key: '/settings',
            icon: <SettingOutlined />,
            label: '系统设置',
        },
    ];

    const userMenuItems = [
        {
            key: 'profile',
            icon: <UserOutlined />,
            label: '个人资料',
        },
        {
            key: 'settings',
            icon: <SettingOutlined />,
            label: '设置',
        },
        {
            type: 'divider' as const,
        },
        {
            key: 'logout',
            icon: <LogoutOutlined />,
            label: '退出登录',
            danger: true,
        },
    ];

    const handleMenuClick = ({ key }: { key: string }) => {
        navigate(key);
    };

    const handleUserMenuClick = ({ key }: { key: string }) => {
        if (key === 'logout') {
            // Handle logout logic
            localStorage.removeItem('auth_token');
            navigate('/login');
        } else {
            console.log('User menu clicked:', key);
        }
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider
                trigger={null}
                collapsible
                collapsed={collapsed}
                style={{
                    background: '#001529',
                }}
            >
                <div
                    style={{
                        height: 64,
                        margin: 16,
                        background: 'rgba(255, 255, 255, 0.2)',
                        borderRadius: 6,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'white',
                        fontWeight: 'bold',
                        fontSize: collapsed ? 14 : 16,
                    }}
                >
                    {collapsed ? 'MA' : 'Multi-Agent'}
                </div>
                <Menu
                    theme="dark"
                    mode="inline"
                    selectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={handleMenuClick}
                />
            </Sider>
            <Layout>
                <Header
                    style={{
                        padding: '0 24px',
                        background: '#fff',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        borderBottom: '1px solid #f0f0f0',
                    }}
                >
                    <Space>
                        <Button
                            type="text"
                            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                            onClick={() => setCollapsed(!collapsed)}
                            style={{ fontSize: 16, width: 40, height: 40 }}
                        />
                        <Text strong style={{ fontSize: 18 }}>
                            Multi-Agent 管理平台
                        </Text>
                    </Space>

                    <Space size="middle">
                        <Badge count={3} size="small">
                            <Button
                                type="text"
                                icon={<BellOutlined />}
                                style={{ fontSize: 16, width: 40, height: 40 }}
                            />
                        </Badge>

                        <Dropdown
                            menu={{
                                items: userMenuItems,
                                onClick: handleUserMenuClick,
                            }}
                            placement="bottomRight"
                        >
                            <Space style={{ cursor: 'pointer' }}>
                                <Avatar size="small" icon={<UserOutlined />} />
                                <Text>管理员</Text>
                            </Space>
                        </Dropdown>
                    </Space>
                </Header>

                <Content
                    style={{
                        margin: 24,
                        padding: 24,
                        minHeight: 280,
                        background: '#f5f5f5',
                        borderRadius: 8,
                    }}
                >
                    <Outlet />
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;