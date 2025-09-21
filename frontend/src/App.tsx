// React import not required with the new JSX transform
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
import './styles/uber-theme.css';

// Uber-style 主题配置
const theme = {
  token: {
    colorPrimary: '#000000', // Uber's black
    colorPrimaryHover: '#333333',
    colorInfo: '#000000',
    colorSuccess: '#00B341', // Uber green
    colorWarning: '#F5A623',
    colorError: '#E53E3E',
    borderRadius: 8,
    wireframe: false,
    fontSize: 14,
    fontFamily: 'UberMove, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    boxShadow: '0 2px 8px rgba(0, 0, 0, 0.12)',
    boxShadowSecondary: '0 4px 12px rgba(0, 0, 0, 0.08)',
  },
  components: {
    Layout: {
      siderBg: '#FFFFFF',
      headerBg: '#FFFFFF',
      bodyBg: '#F6F6F6',
      triggerBg: '#F6F6F6',
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: '#F6F6F6',
      itemHoverBg: '#F9F9F9',
      itemSelectedColor: '#000000',
      itemColor: '#676767',
      iconSize: 20,
    },
    Card: {
      boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
      borderRadius: 12,
    },
    Button: {
      borderRadius: 8,
      fontWeight: 500,
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
