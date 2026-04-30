/* ============================================
   Toast.js - Toast 通知组件
   ============================================ */

const { defineComponent, h, ref, reactive } = Vue;

/**
 * Toast 通知类型
 */
const ToastTypes = {
    SUCCESS: 'success',
    ERROR: 'error',
    WARNING: 'warning',
    INFO: 'info'
};

/**
 * Toast 图标映射
 */
const ToastIcons = {
    success: '✓',
    error: '✕',
    warning: '⚠',
    info: 'ℹ'
};

/**
 * Toast 容器组件
 */
export const ToastContainer = defineComponent({
    name: 'ToastContainer',
    
    setup() {
        const toasts = ref([]);
        let toastId = 0;
        
        /**
         * 添加 Toast
         */
        const add = (message, type = 'info', duration = 3000) => {
            const id = ++toastId;
            const toast = {
                id,
                message,
                type,
                visible: true
            };
            
            toasts.value.push(toast);
            
            // 自动移除
            if (duration > 0) {
                setTimeout(() => {
                    remove(id);
                }, duration);
            }
            
            return id;
        };
        
        /**
         * 移除 Toast
         */
        const remove = (id) => {
            const index = toasts.value.findIndex(t => t.id === id);
            if (index !== -1) {
                toasts.value[index].visible = false;
                // 动画结束后移除
                setTimeout(() => {
                    toasts.value = toasts.value.filter(t => t.id !== id);
                }, 300);
            }
        };
        
        // 暴露方法到全局
        if (typeof window !== 'undefined') {
            window.__toast = { add, remove };
        }
        
        return {
            toasts,
            add,
            remove
        };
    },
    
    render() {
        if (this.toasts.length === 0) return null;
        
        return h('div', { class: 'toast-container' }, 
            this.toasts.map(toast => 
                h('div', {
                    key: toast.id,
                    class: ['toast', `toast-${toast.type}`, toast.visible ? 'toast-visible' : 'toast-hidden'],
                    onClick: () => this.remove(toast.id)
                }, [
                    h('span', { class: 'toast-icon' }, ToastIcons[toast.type] || ToastIcons.info),
                    h('span', { class: 'toast-message' }, toast.message)
                ])
            )
        );
    }
});

/**
 * Toast 服务
 */
export const Toast = {
    /**
     * 显示 Toast
     * @param {string} message - 消息内容
     * @param {string} type - 类型：success | error | warning | info
     * @param {number} duration - 持续时间（毫秒），0 表示不自动关闭
     */
    show(message, type = 'info', duration = 3000) {
        if (typeof window !== 'undefined' && window.__toast) {
            return window.__toast.add(message, type, duration);
        }
        console.warn('Toast 容器未初始化');
        return null;
    },
    
    /**
     * 成功提示
     */
    success(message, duration = 3000) {
        return this.show(message, ToastTypes.SUCCESS, duration);
    },
    
    /**
     * 错误提示
     */
    error(message, duration = 4000) {
        return this.show(message, ToastTypes.ERROR, duration);
    },
    
    /**
     * 警告提示
     */
    warning(message, duration = 3500) {
        return this.show(message, ToastTypes.WARNING, duration);
    },
    
    /**
     * 信息提示
     */
    info(message, duration = 3000) {
        return this.show(message, ToastTypes.INFO, duration);
    },
    
    /**
     * 关闭指定 Toast
     */
    close(id) {
        if (typeof window !== 'undefined' && window.__toast) {
            window.__toast.remove(id);
        }
    }
};

export default Toast;