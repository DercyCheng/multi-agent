import { io, Socket } from 'socket.io-client';
import { store } from '../store';
import { setMetrics, setConnectionStatus } from '../store/slices/systemSlice';
import type { SystemMetrics } from '../types';

class WebSocketService {
    private socket: Socket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectInterval = 1000;

    connect() {
        if (this.socket?.connected) {
            return;
        }

        const token = localStorage.getItem('auth_token');

        this.socket = io('/ws', {
            auth: {
                token,
            },
            transports: ['websocket'],
        });

        this.setupEventListeners();
    }

    disconnect() {
        if (this.socket) {
            this.socket.disconnect();
            this.socket = null;
        }
        (store.dispatch as any)({ type: 'system/setConnectionStatus', payload: false });
    }

    private setupEventListeners() {
        if (!this.socket) return;

        this.socket.on('connect', () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            (store.dispatch as any)({ type: 'system/setConnectionStatus', payload: true });
        });

        this.socket.on('disconnect', () => {
            console.log('WebSocket disconnected');
            (store.dispatch as any)({ type: 'system/setConnectionStatus', payload: false });
            this.handleReconnect();
        });

        this.socket.on('connect_error', (error) => {
            console.error('WebSocket connection error:', error);
            (store.dispatch as any)({ type: 'system/setConnectionStatus', payload: false });
            this.handleReconnect();
        });

        // 系统指标更新
        this.socket.on('system_metrics', (metrics: SystemMetrics) => {
            (store.dispatch as any)({ type: 'system/setMetrics', payload: metrics });
        });

        // Agent状态更新
        this.socket.on('agent_status_update', (data: { agentId: string; status: string }) => {
            console.log('Agent status update:', data);
            // 这里可以更新agent状态
        });

        // 工作流状态更新
        this.socket.on('workflow_status_update', (data: { workflowId: string; status: string }) => {
            console.log('Workflow status update:', data);
            // 这里可以更新工作流状态
        });

        // 工作流执行更新
        this.socket.on('workflow_execution_update', (data: any) => {
            console.log('Workflow execution update:', data);
            // 这里可以更新工作流执行状态
        });

        // 日志事件
        this.socket.on('log_event', (data: any) => {
            console.log('Log event:', data);
            // 这里可以处理实时日志
        });

        // 错误事件
        this.socket.on('error_event', (data: any) => {
            console.error('Error event:', data);
            // 这里可以处理错误通知
        });
    }

    private handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

            setTimeout(() => {
                this.connect();
            }, this.reconnectInterval * this.reconnectAttempts);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    // 发送消息
    emit(event: string, data?: any) {
        if (this.socket?.connected) {
            this.socket.emit(event, data);
        } else {
            console.warn('WebSocket not connected, cannot emit event:', event);
        }
    }

    // 订阅特定Agent的事件
    subscribeToAgent(agentId: string) {
        this.emit('subscribe_agent', { agentId });
    }

    // 取消订阅Agent事件
    unsubscribeFromAgent(agentId: string) {
        this.emit('unsubscribe_agent', { agentId });
    }

    // 订阅特定工作流的事件
    subscribeToWorkflow(workflowId: string) {
        this.emit('subscribe_workflow', { workflowId });
    }

    // 取消订阅工作流事件
    unsubscribeFromWorkflow(workflowId: string) {
        this.emit('unsubscribe_workflow', { workflowId });
    }

    // 获取连接状态
    isConnected(): boolean {
        return this.socket?.connected || false;
    }
}

export const websocketService = new WebSocketService();