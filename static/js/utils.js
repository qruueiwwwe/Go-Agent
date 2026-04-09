/* ============================================
   utils.js - 工具函数库
   ============================================ */

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @returns {string} 格式化后的文件大小
 */
export function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

/**
 * 格式化时间
 * @param {Date|number} date - 日期对象或时间戳
 * @param {string} format - 格式字符串 (默认: 'HH:mm')
 * @returns {string} 格式化后的时间
 */
export function formatTime(date, format = 'HH:mm') {
    const d = typeof date === 'number' ? new Date(date) : date;
    
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const year = d.getFullYear();
    
    const replacements = {
        'HH': hours,
        'mm': minutes,
        'ss': seconds,
        'DD': day,
        'MM': month,
        'YYYY': year
    };
    
    let result = format;
    Object.entries(replacements).forEach(([key, value]) => {
        result = result.replace(key, value);
    });
    
    return result;
}

/**
 * 验证文件类型
 * @param {string} filename - 文件名
 * @returns {boolean} 是否为允许的文件类型
 */
export function validateFileType(filename) {
    const allowedExtensions = ['.txt', '.md', '.json', '.go', '.py', '.js', '.pdf'];
    const ext = filename.substring(filename.lastIndexOf('.')).toLowerCase();
    return allowedExtensions.includes(ext);
}

/**
 * 生成唯一 ID
 * @returns {string} 唯一 ID
 */
export function generateId() {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * 防抖函数
 * @param {Function} func - 要防抖的函数
 * @param {number} delay - 延迟时间（毫秒）
 * @returns {Function} 防抖后的函数
 */
export function debounce(func, delay = 300) {
    let timeoutId;
    
    return function debounced(...args) {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => func(...args), delay);
    };
}

/**
 * 节流函数
 * @param {Function} func - 要节流的函数
 * @param {number} delay - 间隔时间（毫秒）
 * @returns {Function} 节流后的函数
 */
export function throttle(func, delay = 300) {
    let lastCall = 0;
    
    return function throttled(...args) {
        const now = Date.now();
        if (now - lastCall >= delay) {
            lastCall = now;
            func(...args);
        }
    };
}

/**
 * 深拷贝
 * @param {any} obj - 要拷贝的对象
 * @returns {any} 拷贝后的对象
 */
export function deepClone(obj) {
    if (obj === null || typeof obj !== 'object') {
        return obj;
    }
    
    if (obj instanceof Date) {
        return new Date(obj.getTime());
    }
    
    if (Array.isArray(obj)) {
        return obj.map(item => deepClone(item));
    }
    
    if (obj instanceof Object) {
        const cloned = {};
        for (const key in obj) {
            if (obj.hasOwnProperty(key)) {
                cloned[key] = deepClone(obj[key]);
            }
        }
        return cloned;
    }
}

/**
 * 延迟执行
 * @param {number} ms - 延迟毫秒数
 * @returns {Promise} 延迟 Promise
 */
export function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 检查是否为空
 * @param {any} value - 要检查的值
 * @returns {boolean} 是否为空
 */
export function isEmpty(value) {
    if (value === null || value === undefined) return true;
    if (typeof value === 'string') return value.trim() === '';
    if (Array.isArray(value)) return value.length === 0;
    if (typeof value === 'object') return Object.keys(value).length === 0;
    return false;
}

/**
 * 获取查询参数
 * @param {string} name - 参数名
 * @returns {string|null} 参数值
 */
export function getQueryParam(name) {
    const url = new URL(window.location);
    return url.searchParams.get(name);
}

/**
 * 复制到剪贴板
 * @param {string} text - 要复制的文本
 * @returns {Promise<boolean>} 是否复制成功
 */
export async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        console.error('复制失败:', err);
        return false;
    }
}

/**
 * 高亮搜索关键词
 * @param {string} text - 原文本
 * @param {string} keyword - 关键词
 * @returns {string} 高亮后的 HTML
 */
export function highlightKeyword(text, keyword) {
    if (!keyword) return text;
    const regex = new RegExp(`(${keyword})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
}

/**
 * 验证邮箱格式
 * @param {string} email - 邮箱
 * @returns {boolean} 是否有效
 */
export function isValidEmail(email) {
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return regex.test(email);
}

/**
 * 获取浏览器信息
 * @returns {Object} 浏览器信息
 */
export function getBrowserInfo() {
    const ua = navigator.userAgent;
    
    let browserName = 'Unknown';
    let version = 'Unknown';
    
    if (ua.indexOf('Chrome') > -1) {
        browserName = 'Chrome';
        version = ua.match(/Chrome\/(\d+)/)?.[1] || 'Unknown';
    } else if (ua.indexOf('Safari') > -1) {
        browserName = 'Safari';
        version = ua.match(/Version\/(\d+)/)?.[1] || 'Unknown';
    } else if (ua.indexOf('Firefox') > -1) {
        browserName = 'Firefox';
        version = ua.match(/Firefox\/(\d+)/)?.[1] || 'Unknown';
    } else if (ua.indexOf('Edge') > -1) {
        browserName = 'Edge';
        version = ua.match(/Edg\/(\d+)/)?.[1] || 'Unknown';
    }
    
    return {
        name: browserName,
        version: version,
        userAgent: ua
    };
}

/**
 * 检查是否支持某个 API
 * @param {string} api - API 名称
 * @returns {boolean} 是否支持
 */
export function isSupported(api) {
    const apis = {
        fetch: typeof fetch !== 'undefined',
        clipboard: navigator.clipboard !== undefined,
        localStorage: typeof localStorage !== 'undefined',
        sessionStorage: typeof sessionStorage !== 'undefined',
        indexedDB: typeof indexedDB !== 'undefined',
        webWorker: typeof Worker !== 'undefined',
        serviceWorker: 'serviceWorker' in navigator
    };
    
    return apis[api] !== undefined ? apis[api] : false;
}
