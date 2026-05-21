/* ============================================
   ChatWindow.js - 聊天窗口组件
   ============================================ */

const { defineComponent, h } = Vue;
import { MessageList } from './MessageList.js';
import { InputArea } from './InputArea.js';
import { FileManager } from './FileManager.js';
import { ToastContainer } from './Toast.js';

/**
 * 获取当前用户信息
 */
function getCurrentUser() {
    const userStr = localStorage.getItem('user');
    if (userStr) {
        try {
            return JSON.parse(userStr);
        } catch (e) {
            return null;
        }
    }
    return null;
}

/**
 * ChatWindow 组件 - 聊天主窗口
 *
 * Props:
 *   - messages: Array 消息列表
 *   - loading: Boolean 是否加载中
 *   - title: String 窗口标题
 *   - subtitle: String 窗口副标题
 *   - files: Array 文件列表
 *   - uploading: Boolean 是否上传中
 *
 * Events:
 *   - send: 发送消息
 *   - regenerate: 重新生成消息
 *   - file-selected: 文件被选择
 *   - delete-file: 删除文件
 *   - analyze-file: 分析文件
 *   - toggle-sidebar: 切换侧边栏
 */
export const ChatWindow = defineComponent({
    name: 'ChatWindow',

    components: {
        MessageList,
        InputArea,
        FileManager,
        ToastContainer
    },

    props: {
        messages: {
            type: Array,
            default: () => []
        },
        loading: {
            type: Boolean,
            default: false
        },
        title: {
            type: String,
            default: '智能助手'
        },
        subtitle: {
            type: String,
            default: '支持天气查询、数学计算、文件处理'
        },
        files: {
            type: Array,
            default: () => []
        },
        uploading: {
            type: Boolean,
            default: false
        },
        showTimestamp: {
            type: Boolean,
            default: true
        }
    },

    data() {
        return {
            user: getCurrentUser()
        };
    },

    methods: {
        handleSend(message) {
            this.$emit('send', message);
        },

        handleRegenerate(message) {
            this.$emit('regenerate', message);
        },

        handleToggleSidebar() {
            this.$emit('toggle-sidebar');
        },

        handleFileSelected(file) {
            this.$emit('file-selected', file);
        },

        handleDeleteFile(filename) {
            this.$emit('delete-file', filename);
        },

        handleAnalyzeFile(filename) {
            this.$emit('analyze-file', filename);
        },

        handleFileError(message) {
            this.$emit('file-error', message);
        },

        goToAdmin() {
            window.location.href = '/admin.html';
        },

        logout() {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login.html';
        }
    },

    render() {
        const user = this.user;
        const isAdmin = user && user.role === 'admin';

        return h('div', { class: 'main-content' }, [
            // 头部
            h('div', { class: 'app-header' }, [
                h('div', { class: 'header-left' }, [
                    h('button', {
                        class: 'sidebar-toggle-btn',
                        onClick: this.handleToggleSidebar
                    }, '☰'),
                    h('div', { class: 'header-title' }, [
                        h('h1', this.title),
                        h('p', this.subtitle)
                    ])
                ]),
                // 用户信息区
                h('div', { class: 'header-right' }, [
                    // 角色对话按钮
                    h('button', {
                        class: 'persona-header-btn',
                        onClick: () => this.$emit('open-persona'),
                        title: '角色对话'
                    }, [
                        h('span', { class: 'btn-icon' }, '🎭'),
                        h('span', { class: 'btn-text' }, '角色对话')
                    ]),
                    isAdmin ? h('button', {
                        class: 'admin-btn',
                        onClick: this.goToAdmin
                    }, '后台管理') : null,
                    h('div', { class: 'user-info' }, [
                        h('span', { class: 'username' }, user ? user.username : '用户'),
                        user && user.role ? h('span', { class: 'user-role ' + user.role }, 
                            user.role === 'admin' ? '管理员' : 
                            user.role === 'vip' ? 'VIP' : '用户'
                        ) : null
                    ]),
                    h('button', {
                        class: 'logout-btn',
                        onClick: this.logout
                    }, '退出')
                ])
            ]),

            // 文件管理
            h(FileManager, {
                files: this.files,
                uploading: this.uploading,
                onFileSelected: this.handleFileSelected,
                onDeleteFile: this.handleDeleteFile,
                onAnalyzeFile: this.handleAnalyzeFile,
                onError: this.handleFileError
            }),

            // 消息列表
            h(MessageList, {
                messages: this.messages,
                loading: this.loading,
                showTimestamp: this.showTimestamp,
                onRegenerate: this.handleRegenerate
            }),

            // 输入区
            h(InputArea, {
                disabled: this.loading,
                onSend: this.handleSend
            }),

            // Toast 容器
            h(ToastContainer)
        ]);
    }
});

export default ChatWindow;
