import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { ConfigProvider, App as AntdApp } from 'antd';
import { store } from './store';
import MainLayout from './components/Layout/MainLayout';
import Dashboard from './pages/Dashboard';
import AgentList from './pages/Agents/AgentList';
import AgentDetail from './pages/Agents/AgentDetail';
import WorkflowList from './pages/Workflows/WorkflowList';
import WorkflowDetail from './pages/Workflows/WorkflowDetail';
import FeatureFlagsManagement from './pages/FeatureFlags/FeatureFlagsManagement';
import ConfigurationCenter from './pages/Configuration/ConfigurationCenter';
import CronJobsManagement from './pages/CronJobs/CronJobsManagement';
import ServiceDiscovery from './pages/ServiceDiscovery/ServiceDiscovery';
import Login from './pages/Auth/Login';
import './App.css';

// Ant Design 主题配置
const theme = {
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 6,
    wireframe: false,
  },
  components: {
    Layout: {
      siderBg: '#001529',
      triggerBg: '#002140',
    },
  },
};

function App() {
  return (
    <Provider store={store}>
      <ConfigProvider theme={theme}>
        <AntdApp>
          <Router>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/" element={<MainLayout />}>
                <Route index element={<Navigate to="/dashboard" replace />} />
                <Route path="dashboard" element={<Dashboard />} />
                <Route path="agents" element={<AgentList />} />
                <Route path="agents/:id" element={<AgentDetail />} />
                <Route path="workflows" element={<WorkflowList />} />
                <Route path="workflows/:id" element={<WorkflowDetail />} />
                <Route path="feature-flags" element={<FeatureFlagsManagement />} />
                <Route path="configuration" element={<ConfigurationCenter />} />
                <Route path="cronjobs" element={<CronJobsManagement />} />
                <Route path="service-discovery" element={<ServiceDiscovery />} />
              </Route>
            </Routes>
          </Router>
        </AntdApp>
      </ConfigProvider>
    </Provider>
  );
}

export default App;
