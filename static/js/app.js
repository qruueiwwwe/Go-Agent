/* ============================================
   app.js - Vue 应用入口和根组件
   ============================================ */

import { ChatWindow } from './components/ChatWindow.js';
import { FileManager } from './components/FileManager.js';
import { Settings } from './components/Settings.js';
import { ToastContainer } from './components/Toast.js';
import API, { errorHandler } from './api.js';
import { generateId, ThemeManager } from './utils.js';

const { createApp, ref, reactive, computed, nextTick } = Vue;

/**
 * 创建 Vue 应用
 */
const app = createApp({
    name: 'App',
    
    components: {
        ChatWindow,
        FileManager,
        Settings,
        ToastContainer
    },
    
    data() {
        return {
            // 消息列表
            messages: [],
            
            // 文件列表
            files: [],
            
            // 加载状态
            loading: false,
            
            // 文件上传状态
            uploading: false,
            
            // 消息历史（用于 API 上下文）
            history: [],
            
            // 设置面板显示
            showSettings: false,
            
            // 侧边栏状态（默认收起）
            sidebarCollapsed: true,
            
            // 会话列表
            sessions: [],
            
            // 当前会话ID
            currentSessionId: null,
            
            // 当前主题
            currentTheme: 'light'
        };
    },
    
    computed: {
        lastMessage() {
            return this.messages[this.messages.length - 1];
        }
    },
    
    methods: {
        /**
         * 初始化应用
         */
        async initialize() {
            await this.loadFileList();
            this.initTheme();
            this.loadSessions();
            
            // 添加欢迎消息
            this.addMessage({
                type: 'assistant',
                content: '你好！我是智能助手，可以帮你查询天气、进行数学计算或处理文件。\n\n**你可以问我：**\n- 今天北京的天气怎么样？\n- 计算：123 + 456\n- 帮我分析这个文件\n\n有什么可以帮你的吗？'
            });
        },
        
        /**
         * 初始化主题
         */
        initTheme() {
            this.currentTheme = ThemeManager.getResolvedTheme();
            ThemeManager.onChange((theme) => {
                this.currentTheme = theme;
            });
        },
        
        /**
         * 加载会话列表
         */
        loadSessions() {
            const saved = localStorage.getItem('chat-sessions');
            if (saved) {
                try {
                    this.sessions = JSON.parse(saved);
                } catch (e) {
                    this.sessions = [];
                }
            }
            
            // 如果没有会话，创建一个新会话
            if (this.sessions.length === 0) {
                this.createNewSession();
            } else {
                this.currentSessionId = this.sessions[0].id;
            }
        },
        
        /**
         * 保存会话列表
         */
        saveSessions() {
            localStorage.setItem('chat-sessions', JSON.stringify(this.sessions));
        },
        
        /**
         * 创建新会话
         */
        createNewSession() {
            const session = {
                id: generateId(),
                title: '新会话',
                preview: '',
                time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
                messages: []
            };
            
            this.sessions.unshift(session);
            this.currentSessionId = session.id;
            this.messages = [];
            this.history = [];
            this.saveSessions();
        },
        
        /**
         * 选择会话
         */
        selectSession(session) {
            this.currentSessionId = session.id;
            this.messages = session.messages || [];
            // 重建历史
            this.history = this.messages
                .filter(m => m.type === 'user' || m.type === 'assistant')
                .map(m => ({
                    role: m.type === 'user' ? 'user' : 'assistant',
                    content: m.content
                }));
        },
        
        /**
         * 删除会话
         */
        deleteSession(session) {
            const index = this.sessions.findIndex(s => s.id === session.id);
            if (index !== -1) {
                this.sessions.splice(index, 1);
                
                if (this.currentSessionId === session.id) {
                    if (this.sessions.length > 0) {
                        this.selectSession(this.sessions[0]);
                    } else {
                        this.createNewSession();
                    }
                }
                
                this.saveSessions();
            }
        },
        
        /**
         * 更新当前会话
         */
        updateCurrentSession() {
            const session = this.sessions.find(s => s.id === this.currentSessionId);
            if (session) {
                session.messages = [...this.messages];
                
                // 更新标题和预览
                const firstUserMsg = this.messages.find(m => m.type === 'user');
                if (firstUserMsg) {
                    session.title = firstUserMsg.content.slice(0, 20) + (firstUserMsg.content.length > 20 ? '...' : '');
                }
                
                const lastMsg = this.messages[this.messages.length - 1];
                if (lastMsg) {
                    session.preview = lastMsg.content.slice(0, 30) + (lastMsg.content.length > 30 ? '...' : '');
                    session.time = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
                }
                
                this.saveSessions();
            }
        },
        
        /**
         * 切换侧边栏
         */
        toggleSidebar() {
            this.sidebarCollapsed = !this.sidebarCollapsed;
        },
        
        /**
         * 加载文件列表
         */
        async loadFileList() {
            try {
                this.files = await API.files.list();
            } catch (error) {
                console.error('加载文件列表失败:', error);
            }
        },
        
        /**
         * 添加消息到列表
         */
        addMessage(messageData) {
            const message = {
                id: generateId(),
                type: messageData.type || 'assistant',
                content: messageData.content || '',
                timestamp: Date.now()
            };
            
            this.messages.push(message);
            
            // 将消息加入历史
            if (message.type === 'user' || message.type === 'assistant') {
                this.history.push({
                    role: message.type === 'user' ? 'user' : 'assistant',
                    content: message.content
                });
            }
            
            // 更新会话
            this.updateCurrentSession();
            
            return message;
        },
        
        /**
         * 处理发送消息
         */
        async handleSendMessage(userMessage) {
            if (!userMessage || !userMessage.trim()) return;
            
            // 添加用户消息
            this.addMessage({
                type: 'user',
                content: userMessage
            });
            
            // 设置加载状态
            this.loading = true;
            
            try {
                // 调用聊天 API
                const response = await API.chat.send(userMessage, this.history);
                
                // 提取结果内容
                let content = '';
                if (typeof response === 'string') {
                    content = response;
                } else if (response.result) {
                    content = response.result;
                } else if (response.content) {
                    content = response.content;
                } else {
                    content = JSON.stringify(response);
                }
                
                // 添加助手消息
                this.addMessage({
                    type: 'assistant',
                    content: content
                });
            } catch (error) {
                // 处理错误
                const errorMsg = errorHandler.handle(error);
                this.addMessage({
                    type: 'error',
                    content: `错误: ${errorMsg}`
                });
                
                errorHandler.log(error, '发送消息失败');
            } finally {
                this.loading = false;
            }
        },
        
        /**
         * 处理重新生成
         */
        async handleRegenerate(message) {
            // 找到上一条用户消息
            const msgIndex = this.messages.findIndex(m => m.id === message.id);
            if (msgIndex <= 0) return;
            
            const userMsg = this.messages[msgIndex - 1];
            if (userMsg.type !== 'user') return;
            
            // 删除当前的助手回复
            this.messages.splice(msgIndex, 1);
            this.history.pop();
            
            // 重新发送
            await this.handleSendMessage(userMsg.content);
        },
        
        /**
         * 处理文件选择（上传）
         */
        async handleFileSelected(file) {
            this.uploading = true;
            
            try {
                await API.files.upload(file);
                await this.loadFileList();
                
                this.addMessage({
                    type: 'system',
                    content: `✓ 文件 ${file.name} 上传成功`
                });
            } catch (error) {
                const errorMsg = errorHandler.handle(error);
                this.addMessage({
                    type: 'error',
                    content: `文件上传失败: ${errorMsg}`
                });
                
                errorHandler.log(error, '文件上传失败');
            } finally {
                this.uploading = false;
            }
        },
        
        /**
         * 处理删除文件
         */
        async handleDeleteFile(filename) {
            try {
                await API.files.delete(filename);
                await this.loadFileList();
                
                this.addMessage({
                    type: 'system',
                    content: `✓ 文件 ${filename} 已删除`
                });
            } catch (error) {
                const errorMsg = errorHandler.handle(error);
                this.addMessage({
                    type: 'error',
                    content: `删除文件失败: ${errorMsg}`
                });
                
                errorHandler.log(error, '删除文件失败');
            }
        },
        
        /**
         * 处理快速分析文件
         */
        handleAnalyzeFile(filename) {
            const message = JSON.stringify({
                action: 'parse',
                file: `data/${filename}`,
                mode: 'summary'
            });
            
            this.handleSendMessage(message);
        },
        
        /**
         * 切换设置面板
         */
        toggleSettings() {
            this.showSettings = !this.showSettings;
        },
        
        /**
         * 关闭设置面板
         */
        closeSettings() {
            this.showSettings = false;
        },
        
        /**
         * 主题变化处理
         */
        handleThemeChange(theme) {
            this.currentTheme = theme;
        }
    },
    
    mounted() {
        this.initialize();
        
        // 监听网络连接状态
        window.addEventListener('online', () => {
            this.addMessage({
                type: 'system',
                content: '网络已恢复'
            });
        });
        
        window.addEventListener('offline', () => {
            this.addMessage({
                type: 'error',
                content: '网络连接已断开'
            });
        });
        
        // 快捷键
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + , 打开设置
            if ((e.ctrlKey || e.metaKey) && e.key === ',') {
                e.preventDefault();
                this.toggleSettings();
            }
            // Ctrl/Cmd + B 切换侧边栏
            if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
                e.preventDefault();
                this.toggleSidebar();
            }
        });
    },
    
    beforeUnmount() {
        window.removeEventListener('online', null);
        window.removeEventListener('offline', null);
    },
    
    template: `
        <div id="app">
            <div class="app-layout">
                <!-- 侧边栏 -->
                <aside :class="['sidebar', sidebarCollapsed && 'collapsed']">
                    <div class="sidebar-header">
                        <button class="sidebar-toggle" @click="toggleSidebar" :title="sidebarCollapsed ? '展开' : '收起'">
                            <span class="toggle-icon">{{ sidebarCollapsed ? '☰' : '✕' }}</span>
                        </button>
                        <h2 v-if="!sidebarCollapsed" class="sidebar-title">会话历史</h2>
                    </div>

                    <button v-if="!sidebarCollapsed" class="new-session-btn" @click="createNewSession">
                        <span class="btn-icon">+</span>
                        <span class="btn-text">新建会话</span>
                    </button>

                    <div v-if="!sidebarCollapsed" class="session-list">
                        <div
                            v-for="session in sessions"
                            :key="session.id"
                            :class="['session-item', currentSessionId === session.id && 'active']"
                            @click="selectSession(session)"
                        >
                            <div class="session-icon">💬</div>
                            <div class="session-content">
                                <div class="session-title">{{ session.title }}</div>
                                <div v-if="session.preview" class="session-preview">{{ session.preview }}</div>
                                <div v-if="session.time" class="session-time">{{ session.time }}</div>
                            </div>
                        </div>
                    </div>
                </aside>

                <!-- 聊天窗口 -->
                <chat-window
                    :messages="messages"
                    :loading="loading"
                    :files="files"
                    :uploading="uploading"
                    @send="handleSendMessage"
                    @regenerate="handleRegenerate"
                    @toggle-sidebar="toggleSidebar"
                    @file-selected="handleFileSelected"
                    @delete-file="handleDeleteFile"
                    @analyze-file="handleAnalyzeFile"
                />
            </div>

            <!-- 设置面板 -->
            <settings
                :visible="showSettings"
                @close="closeSettings"
                @theme-change="handleThemeChange"
            />

            <!-- 设置按钮 -->
            <button class="settings-fab" @click="toggleSettings" title="设置 (Ctrl+,)">
                ⚙️
            </button>

            <!-- Toast 容器 -->
            <toast-container />
        </div>
    `
});

/**
 * 导出应用
 */
export default app;