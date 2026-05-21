/* ============================================
   api.js - API 调用封装和错误处理
   ============================================ */

const API_BASE_URL = '/api';

/**
 * 判断错误是否可重试
 * @param {Error} error - 错误对象
 * @returns {boolean} 是否可重试
 */
function isRetryableError(error) {
    // 网络断开
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
        return true;
    }
    // 超时
    if (error.name === 'AbortError') {
        return true;
    }
    // 网络错误
    if (error.message.includes('network') || error.message.includes('Network')) {
        return true;
    }
    return false;
}

/**
 * 延迟函数
 * @param {number} ms - 毫秒
 */
function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 基础 HTTP 请求（带重试机制）
 * @param {string} endpoint - API 端点
 * @param {Object} options - 请求选项
 * @param {number} retryOptions.maxRetries - 最大重试次数（默认3次）
 * @param {number} retryOptions.retryDelay - 重试延迟毫秒（默认1000ms）
 * @returns {Promise<Object>} API 响应
 */
async function request(endpoint, options = {}, retryOptions = {}) {
    const { maxRetries = 3, retryDelay = 1000 } = retryOptions;
    const url = `${API_BASE_URL}${endpoint}`;
    
    // 获取 token
    const token = localStorage.getItem('token');
    
    const defaultOptions = {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        timeout: 300000 // 5分钟超时
    };
    
    // 添加 Authorization 头
    if (token) {
        defaultOptions.headers['Authorization'] = `Bearer ${token}`;
    }
    
    const config = { ...defaultOptions, ...options };
    
    // 合并 headers
    if (options.headers) {
        config.headers = { ...defaultOptions.headers, ...options.headers };
    }
    
    let lastError = null;
    
    // 重试循环
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
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
                // 401 表示未授权，跳转到登录页
                if (response.status === 401) {
                    localStorage.removeItem('token');
                    localStorage.removeItem('user');
                    window.location.href = '/login.html';
                    throw new Error('登录已过期，请重新登录');
                }
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();

            // 检查业务状态码（errno === 0 表示成功）
            if (data.errno !== 0) {
                throw new Error(data.errmsg || '请求失败');
            }

            return data;
        } catch (error) {
            lastError = error;
            
            // 超时错误
            if (error.name === 'AbortError') {
                error.message = '请求超时，请检查网络连接';
            }
            
            // 判断是否可重试
            if (attempt < maxRetries && isRetryableError(error)) {
                console.log(`[API] 请求失败，第 ${attempt + 1} 次重试...`, error.message);
                await delay(retryDelay * (attempt + 1)); // 递增延迟
                continue;
            }
            
            throw error;
        }
    }
    
    throw lastError;
}

/**
 * 聊天 API（带自动重试）
 */
export const chatAPI = {
    /**
     * 发送聊天消息
     * @param {string} message - 消息内容
     * @param {Array} history - 消息历史
     * @param {Object} retryOptions - 重试选项
     * @returns {Promise<Object>} 响应数据
     */
    async send(message, history = [], retryOptions = { maxRetries: 2, retryDelay: 500 }) {
        const response = await request('/chat', {
            method: 'POST',
            body: JSON.stringify({ message, history })
        }, retryOptions);
        
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
        
        const token = localStorage.getItem('token');
        const headers = {};
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        
        const response = await fetch(`${API_BASE_URL}/upload`, {
            method: 'POST',
            headers,
            body: formData
        });
        
        if (!response.ok) {
            throw new Error(`上传失败: ${response.statusText}`);
        }
        
        const data = await response.json();

        if (data.errno !== 0) {
            throw new Error(data.errmsg || '上传失败');
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
            method: 'POST'
        });
        
        return response.data || response;
    }
};

/**
 * 角色卡 API
 */
export const personaAPI = {
    /**
     * 获取角色卡列表
     * @param {string} type - 'public' 或 'mine'
     * @param {number} page - 页码
     * @param {number} size - 每页数量
     * @returns {Promise<Object>} 角色列表
     */
    async list(type = 'public', page = 1, size = 20) {
        const response = await request(`/persona/list?type=${type}&page=${page}&size=${size}`);
        return response.data || { list: [], total: 0 };
    },

    /**
     * 获取角色卡详情
     * @param {number} id - 角色卡ID
     * @returns {Promise<Object>} 角色详情
     */
    async detail(id) {
        const response = await request(`/persona/detail?id=${id}`);
        return response.data;
    },

    /**
     * 创建角色卡
     * @param {Object} data - 角色卡数据
     * @returns {Promise<Object>} 创建结果
     */
    async create(data) {
        const response = await request('/persona/create', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        return response.data;
    },

    /**
     * 更新角色卡
     * @param {Object} data - 角色卡数据（含id）
     * @returns {Promise<Object>} 更新结果
     */
    async update(data) {
        const response = await request('/persona/update', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        return response.data;
    },

    /**
     * 删除角色卡
     * @param {number} id - 角色卡ID
     * @returns {Promise<Object>} 删除结果
     */
    async delete(id) {
        const response = await request('/persona/delete', {
            method: 'POST',
            body: JSON.stringify({ id })
        });
        return response.data;
    },

    /**
     * 发送角色卡对话消息
     * @param {number} personaId - 角色卡ID
     * @param {string} message - 消息内容
     * @param {string} sessionId - 会话ID（可选）
     * @returns {Promise<Object>} 对话响应
     */
    async chat(personaId, message, sessionId = null) {
        const body = { persona_id: personaId, message };
        if (sessionId) {
            body.session_id = sessionId;
        }
        const response = await request('/persona/chat', {
            method: 'POST',
            body: JSON.stringify(body)
        }, { maxRetries: 1, retryDelay: 1000 });
        return response.data;
    },

    /**
     * 获取角色卡会话列表
     * @param {number} personaId - 角色卡ID（可选）
     * @returns {Promise<Array>} 会话列表
     */
    async sessions(personaId = null) {
        let url = '/persona/sessions';
        if (personaId) {
            url += `?persona_id=${personaId}`;
        }
        const response = await request(url);
        return response.data?.sessions || [];
    },

    /**
     * 获取会话历史
     * @param {string} sessionId - 会话ID
     * @returns {Promise<Object>} 会话历史
     */
    async history(sessionId) {
        const response = await request(`/persona/history?session_id=${sessionId}`);
        return response.data;
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
        
        // 频率限制
        if (error.message.includes('频繁') || error.message.includes('等待')) {
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
    }
};

/**
 * 检查 API 可用性
 * @returns {Promise<boolean>} 是否可用
 */
export async function checkAPIHealth() {
    try {
        const response = await fetch('/api/files');
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
    persona: personaAPI,
    errorHandler,
    checkAPIHealth
};
