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
                width={280}
                style={{
                    background: '#FFFFFF',
                    borderRight: '1px solid #E5E5E5',
                    boxShadow: '2px 0 8px rgba(0, 0, 0, 0.04)',
                }}
            >
                <div style={{
                    height: '80px',
                    display: 'flex',
                    alignItems: 'center',
                    padding: '0 24px',
                    borderBottom: '1px solid #F0F0F0'
                }}>
                    {!collapsed ? (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <div style={{
                                width: '32px',
                                height: '32px',
                                background: 'linear-gradient(135deg, #000000 0%, #333333 100%)',
                                borderRadius: '8px',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                color: 'white',
                                fontWeight: 'bold',
                                fontSize: '14px'
                            }}>
                                MA
                            </div>
                            <Text style={{
                                fontSize: '18px',
                                fontWeight: '600',
                                color: '#000000',
                                letterSpacing: '-0.5px'
                            }}>
                                Multi-Agent
                            </Text>
                        </div>
                    ) : (
                        <div style={{
                            width: '32px',
                            height: '32px',
                            background: 'linear-gradient(135deg, #000000 0%, #333333 100%)',
                            borderRadius: '8px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: 'white',
                            fontWeight: 'bold',
                            fontSize: '14px',
                            margin: '0 auto'
                        }}>
                            MA
                        </div>
                    )}
                </div>
                <Menu
                    className="uber-menu"
                    mode="inline"
                    selectedKeys={[location.pathname]}
                    items={menuItems}
                    onClick={handleMenuClick}
                    style={{
                        border: 'none',
                        backgroundColor: 'transparent',
                        paddingTop: '16px'
                    }}
                />
            </Sider>
            <Layout>
                <Header style={{
                    padding: '0 32px',
                    background: '#FFFFFF',
                    borderBottom: '1px solid #E5E5E5',
                    height: '80px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    position: 'sticky',
                    top: 0,
                    zIndex: 1000,
                    boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)'
                }}>
                    <Button
                        type="text"
                        icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                        onClick={() => setCollapsed(!collapsed)}
                        style={{
                            fontSize: '18px',
                            width: '48px',
                            height: '48px',
                            color: '#676767',
                            border: 'none',
                            borderRadius: '8px'
                        }}
                    />
                    <Space size="large">
                        <Badge count={3} size="small">
                            <Button
                                type="text"
                                icon={<BellOutlined />}
                                style={{
                                    fontSize: '18px',
                                    width: '48px',
                                    height: '48px',
                                    color: '#676767',
                                    border: 'none',
                                    borderRadius: '8px'
                                }}
                            />
                        </Badge>
                        <Dropdown
                            menu={{
                                items: userMenuItems,
                                onClick: handleUserMenuClick,
                            }}
                            placement="bottomRight"
                        >
                            <Space style={{ cursor: 'pointer', padding: '8px 16px', borderRadius: '12px', transition: 'all 0.2s' }}>
                                <Avatar
                                    size="large"
                                    icon={<UserOutlined />}
                                    style={{
                                        backgroundColor: '#F6F6F6',
                                        color: '#676767',
                                        border: '2px solid #E5E5E5'
                                    }}
                                />
                                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
                                    <Text style={{ fontSize: '14px', fontWeight: '500', color: '#000000', lineHeight: 1.2 }}>
                                        管理员
                                    </Text>
                                    <Text style={{ fontSize: '12px', color: '#676767', lineHeight: 1.2 }}>
                                        admin@multiagent.com
                                    </Text>
                                </div>
                            </Space>
                        </Dropdown>
                    </Space>
                </Header>
                <Content style={{
                    margin: '32px',
                    padding: '0',
                    backgroundColor: 'transparent',
                    overflow: 'auto'
                }}>
                    <Outlet />
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;