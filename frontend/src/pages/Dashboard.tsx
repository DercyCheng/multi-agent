import React from 'react';
import { Card, Row, Col, Statistic, Typography, Progress, Button, Space, Divider } from 'antd';
import { UserOutlined, RobotOutlined, ApiOutlined, ClockCircleOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  // Mock data for metrics
  const metrics = [
    {
      title: '活跃 Agent',
      value: 12,
      icon: <RobotOutlined />,
      change: '+8.2%',
      isPositive: true,
      color: '#00B341'
    },
    {
      title: '运行中工作流',
      value: 8,
      icon: <ApiOutlined />,
      change: '+12.5%',
      isPositive: true,
      color: '#000000'
    },
    {
      title: '在线用户',
      value: 24,
      icon: <UserOutlined />,
      change: '-2.1%',
      isPositive: false,
      color: '#F5A623'
    },
    {
      title: '系统可用性',
      value: '99.9%',
      icon: <ClockCircleOutlined />,
      change: '+0.1%',
      isPositive: true,
      color: '#00B341'
    }
  ];

  return (
    <div style={{ padding: '0' }}>
      <div style={{ marginBottom: '32px' }}>
        <Title level={1} style={{
          fontSize: '32px',
          fontWeight: '700',
          color: '#000000',
          margin: '0 0 8px 0',
          letterSpacing: '-1px'
        }}>
          仪表板
        </Title>
        <Text style={{ fontSize: '16px', color: '#676767' }}>
          欢迎回来，以下是您系统的实时概览
        </Text>
      </div>

      <Row gutter={[24, 24]}>
        {metrics.map((metric, index) => (
          <Col xs={24} sm={12} xl={6} key={index}>
            <Card
              style={{
                borderRadius: '16px',
                border: '1px solid #E5E5E5',
                boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)',
                background: '#FFFFFF'
              }}
              bodyStyle={{ padding: '24px' }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
                <div style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '12px',
                  backgroundColor: `${metric.color}10`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '20px',
                  color: metric.color
                }}>
                  {metric.icon}
                </div>
                <Space size={4} style={{ color: metric.isPositive ? '#00B341' : '#E53E3E' }}>
                  {metric.isPositive ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
                  <Text style={{ fontSize: '12px', fontWeight: '500', color: metric.isPositive ? '#00B341' : '#E53E3E' }}>
                    {metric.change}
                  </Text>
                </Space>
              </div>
              <div>
                <Text style={{ fontSize: '14px', color: '#676767', display: 'block', marginBottom: '4px' }}>
                  {metric.title}
                </Text>
                <Text style={{ fontSize: '28px', fontWeight: '700', color: '#000000', lineHeight: 1 }}>
                  {metric.value}
                </Text>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[24, 24]} style={{ marginTop: '32px' }}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text style={{ fontSize: '18px', fontWeight: '600', color: '#000000' }}>系统性能</Text>
                <Button type="text" style={{ fontSize: '14px', color: '#676767' }}>查看详情</Button>
              </div>
            }
            style={{
              borderRadius: '16px',
              border: '1px solid #E5E5E5',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)'
            }}
            bodyStyle={{ padding: '24px' }}
          >
            <Row gutter={[24, 24]}>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Progress
                    type="circle"
                    percent={85}
                    strokeColor="#00B341"
                    trailColor="#F6F6F6"
                    strokeWidth={8}
                    size={120}
                    format={(percent) => <span style={{ fontSize: '18px', fontWeight: '600' }}>{percent}%</span>}
                  />
                  <Text style={{ display: 'block', marginTop: '12px', fontSize: '14px', color: '#676767' }}>
                    CPU 使用率
                  </Text>
                </div>
              </Col>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Progress
                    type="circle"
                    percent={67}
                    strokeColor="#F5A623"
                    trailColor="#F6F6F6"
                    strokeWidth={8}
                    size={120}
                    format={(percent) => <span style={{ fontSize: '18px', fontWeight: '600' }}>{percent}%</span>}
                  />
                  <Text style={{ display: 'block', marginTop: '12px', fontSize: '14px', color: '#676767' }}>
                    内存使用率
                  </Text>
                </div>
              </Col>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Progress
                    type="circle"
                    percent={43}
                    strokeColor="#000000"
                    trailColor="#F6F6F6"
                    strokeWidth={8}
                    size={120}
                    format={(percent) => <span style={{ fontSize: '18px', fontWeight: '600' }}>{percent}%</span>}
                  />
                  <Text style={{ display: 'block', marginTop: '12px', fontSize: '14px', color: '#676767' }}>
                    磁盘使用率
                  </Text>
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card
            title={<Text style={{ fontSize: '18px', fontWeight: '600', color: '#000000' }}>快速操作</Text>}
            style={{
              borderRadius: '16px',
              border: '1px solid #E5E5E5',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.04)'
            }}
            bodyStyle={{ padding: '24px' }}
          >
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <Button
                type="primary"
                size="large"
                block
                style={{
                  height: '48px',
                  borderRadius: '12px',
                  backgroundColor: '#000000',
                  borderColor: '#000000',
                  fontSize: '16px',
                  fontWeight: '500'
                }}
              >
                创建新 Agent
              </Button>
              <Button
                size="large"
                block
                style={{
                  height: '48px',
                  borderRadius: '12px',
                  fontSize: '16px',
                  fontWeight: '500',
                  border: '1px solid #E5E5E5',
                  color: '#000000'
                }}
              >
                查看工作流
              </Button>
              <Button
                size="large"
                block
                style={{
                  height: '48px',
                  borderRadius: '12px',
                  fontSize: '16px',
                  fontWeight: '500',
                  border: '1px solid #E5E5E5',
                  color: '#000000'
                }}
              >
                系统设置
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;