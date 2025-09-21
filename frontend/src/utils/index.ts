import type { Agent, Workflow } from '../types';

// 格式化日期时间
export const formatDateTime = (dateString: string): string => {
    return new Date(dateString).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    });
};

// 格式化相对时间
export const formatRelativeTime = (dateString: string): string => {
    const now = new Date();
    const date = new Date(dateString);
    const diffMs = now.getTime() - date.getTime();

    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMinutes < 1) {
        return '刚刚';
    } else if (diffMinutes < 60) {
        return `${diffMinutes}分钟前`;
    } else if (diffHours < 24) {
        return `${diffHours}小时前`;
    } else if (diffDays < 30) {
        return `${diffDays}天前`;
    } else {
        return formatDateTime(dateString);
    }
};

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
};

// 格式化数字
export const formatNumber = (num: number): string => {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
};

// 格式化百分比
export const formatPercentage = (value: number, total: number): string => {
    if (total === 0) return '0%';
    return `${((value / total) * 100).toFixed(1)}%`;
};

// 格式化时长
export const formatDuration = (seconds: number): string => {
    if (seconds < 60) {
        return `${seconds.toFixed(1)}秒`;
    } else if (seconds < 3600) {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;
        return `${minutes}分${remainingSeconds.toFixed(0)}秒`;
    } else {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        return `${hours}小时${minutes}分`;
    }
};

// 生成随机ID
export const generateId = (): string => {
    return Math.random().toString(36).substr(2, 9);
};

// 深拷贝对象
export const deepClone = <T>(obj: T): T => {
    return JSON.parse(JSON.stringify(obj));
};

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
    func: T,
    wait: number
): ((...args: Parameters<T>) => void) => {
    let timeout: number;
    return (...args: Parameters<T>) => {
        clearTimeout(timeout);
        timeout = window.setTimeout(() => func(...args), wait);
    };
};

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
    func: T,
    limit: number
): ((...args: Parameters<T>) => void) => {
    let inThrottle: boolean;
    return (...args: Parameters<T>) => {
        if (!inThrottle) {
            func(...args);
            inThrottle = true;
            window.setTimeout(() => (inThrottle = false), limit);
        }
    };
};

// 下载文件
export const downloadFile = (content: string, filename: string, contentType = 'text/plain') => {
    const blob = new Blob([content], { type: contentType });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
};

// 复制到剪贴板
export const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        // 降级方案
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        try {
            document.execCommand('copy');
            return true;
        } catch (err) {
            return false;
        } finally {
            document.body.removeChild(textArea);
        }
    }
};

// 获取Agent状态颜色
export const getAgentStatusColor = (status: Agent['status']): string => {
    const colorMap = {
        active: '#52c41a',
        inactive: '#999',
        error: '#ff4d4f',
        pending: '#1890ff',
    };
    return colorMap[status] || '#999';
};

// 获取工作流状态颜色
export const getWorkflowStatusColor = (status: Workflow['status']): string => {
    const colorMap: Record<string, string> = {
        draft: '#999',
        active: '#52c41a',
        paused: '#faad14',
        completed: '#1890ff',
        failed: '#ff4d4f',
        running: '#52c41a',
        stopped: '#999',
        error: '#ff4d4f',
    };
    return colorMap[status] || '#999';
};

// 验证JSON字符串
export const isValidJSON = (str: string): boolean => {
    try {
        JSON.parse(str);
        return true;
    } catch (e) {
        return false;
    }
};

// 获取错误信息
export const getErrorMessage = (error: unknown): string => {
    if (error instanceof Error) {
        return error.message;
    }
    if (typeof error === 'string') {
        return error;
    }
    return '未知错误';
};

// 本地存储操作
export const storage = {
    get: <T>(key: string, defaultValue: T): T => {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : defaultValue;
        } catch {
            return defaultValue;
        }
    },

    set: <T>(key: string, value: T): void => {
        try {
            localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error('Failed to save to localStorage:', error);
        }
    },

    remove: (key: string): void => {
        try {
            localStorage.removeItem(key);
        } catch (error) {
            console.error('Failed to remove from localStorage:', error);
        }
    },

    clear: (): void => {
        try {
            localStorage.clear();
        } catch (error) {
            console.error('Failed to clear localStorage:', error);
        }
    },
};

// URL参数操作
export const urlParams = {
    get: (key: string): string | null => {
        const params = new URLSearchParams(window.location.search);
        return params.get(key);
    },

    set: (key: string, value: string): void => {
        const url = new URL(window.location.href);
        url.searchParams.set(key, value);
        window.history.replaceState({}, '', url.toString());
    },

    remove: (key: string): void => {
        const url = new URL(window.location.href);
        url.searchParams.delete(key);
        window.history.replaceState({}, '', url.toString());
    },
};