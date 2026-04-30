/* ============================================
   api.js - API 调用封装和错误处理
   ============================================ */

const API_BASE_URL = '/api';

/**
 * 基础 HTTP 请求
 * @param {string} endpoint - API 端点
 * @param {Object} options - 请求选项
 * @returns {Promise<Object>} API 响应
 */
async function request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const defaultOptions = {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        timeout: 300000
    };
    
    const config = { ...defaultOptions, ...options };
    
    try {
        // 添加超时控制
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), config.timeout);
        
        const response = await fetch(url, {
            ...config,
            signal: controller.signal
        });
        
        clearTimeout(timeoutId);
        
        // 检查 HTTP 状态
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // 检查业务状态码
        if (data.code && data.code !== 1000) {
            throw new Error(data.message || '请求失败');
        }
        
        return data;
    } catch (error) {
        if (error.name === 'AbortError') {
            throw new Error('请求超时，请检查网络连接');
        }
        throw error;
    }
}

/**
 * 聊天 API
 */
export const chatAPI = {
    /**
     * 发送聊天消息
     * @param {string} message - 消息内容
     * @param {Array} history - 消息历史
     * @returns {Promise<Object>} 响应数据
     */
    async send(message, history = []) {
        const response = await request('/chat', {
            method: 'POST',
            body: JSON.stringify({ message, history })
        });
        
        return response.data || response;
    }
};

/**
 * 文件 API
 */
export const filesAPI = {
    /**
     * 上传文件
     * @param {File} file - 文件对象
     * @returns {Promise<Object>} 上传结果
     */
    async upload(file) {
        // 验证文件
        const maxSize = 10 * 1024 * 1024; // 10MB
        if (file.size > maxSize) {
            throw new Error(`文件大小超过 ${maxSize / 1024 / 1024}MB 限制`);
        }
        
        const allowedExtensions = ['.txt', '.md', '.json', '.go', '.py', '.js', '.pdf'];
        const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
        if (!allowedExtensions.includes(ext)) {
            throw new Error(`不支持的文件类型: ${ext}`);
        }
        
        // 上传文件
        const formData = new FormData();
        formData.append('file', file);
        
        const response = await fetch(`${API_BASE_URL}/upload`, {
            method: 'POST',
            body: formData,
            timeout: 60000
        });
        
        if (!response.ok) {
            throw new Error(`上传失败: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        if (data.code !== 1000) {
            throw new Error(data.message || '上传失败');
        }
        
        return data.data || data;
    },
    
    /**
     * 获取文件列表
     * @returns {Promise<Array>} 文件列表
     */
    async list() {
        const response = await request('/files');
        return response.data?.files || [];
    },
    
    /**
     * 删除文件
     * @param {string} filename - 文件名
     * @returns {Promise<Object>} 删除结果
     */
    async delete(filename) {
        const response = await request(`/file/delete?filename=${encodeURIComponent(filename)}`, {
            method: 'DELETE'
        });
        
        return response.data || response;
    }
};

/**
 * 错误处理工具
 */
export const errorHandler = {
    /**
     * 处理 API 错误
     * @param {Error} error - 错误对象
     * @returns {string} 用户友好的错误信息
     */
    handle(error) {
        if (!error) return '未知错误';
        
        // 网络错误
        if (!navigator.onLine) {
            return '网络连接失败，请检查网络';
        }
        
        // 超时错误
        if (error.message.includes('超时')) {
            return '请求超时，请重试';
        }
        
        // HTTP 错误
        if (error.message.includes('HTTP')) {
            return '服务器错误，请稍后重试';
        }
        
        // 文件错误
        if (error.message.includes('文件')) {
            return error.message;
        }
        
        // 其他错误
        return error.message || '发生错误，请重试';
    },
    
    /**
     * 记录错误
     * @param {Error} error - 错误对象
     * @param {string} context - 错误上下文
     */
    log(error, context = '') {
        const timestamp = new Date().toISOString();
        const message = `[${timestamp}] ${context}: ${error.message}`;
        console.error(message, error);
        
        // 可以在这里添加远程错误上报
        // this.reportToServer(message, error);
    }
};

/**
 * 重试机制
 */
export function withRetry(fn, maxAttempts = 3, delay = 1000) {
    return async function retriedFn(...args) {
        for (let attempt = 1; attempt <= maxAttempts; attempt++) {
            try {
                return await fn(...args);
            } catch (error) {
                if (attempt === maxAttempts) {
                    throw error;
                }
                
                // 等待后重试
                await new Promise(resolve => setTimeout(resolve, delay * attempt));
            }
        }
    };
}

/**
 * 检查 API 可用性
 * @returns {Promise<boolean>} 是否可用
 */
export async function checkAPIHealth() {
    try {
        const response = await fetch('/api/files', { timeout: 5000 });
        return response.ok;
    } catch (error) {
        return false;
    }
}

/**
 * 导出 API 对象（用于全局访问）
 */
export default {
    chat: chatAPI,
    files: filesAPI,
    errorHandler,
    withRetry,
    checkAPIHealth
};
