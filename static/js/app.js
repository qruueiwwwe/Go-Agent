/* ============================================
   app.js - Vue 应用入口和根组件
   ============================================ */

import { ChatWindow } from './components/ChatWindow.js';
import { FileManager } from './components/FileManager.js';
import { Settings } from './components/Settings.js';
import { ToastContainer, Toast } from './components/Toast.js';
import { PersonaSelector } from './components/PersonaSelector.js';
import API, { errorHandler, AuthExpiredError, isTokenExpiredLocally } from './api.js';
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
        ToastContainer,
        PersonaSelector
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

            // 服务端 session_id（后端持久化的会话 UUID）
            backendSessionId: null,

            // 对话模式：normal / thinking / auto
            chatMode: localStorage.getItem('chat-mode') || 'normal',
            
            // 当前主题
            currentTheme: 'light',

            // 设置项
            showTimestamp: localStorage.getItem('setting-timestamp') !== 'false',
            soundEnabled: localStorage.getItem('setting-sound') === 'true',

            // 角色卡模式
            personaMode: false,
            showPersonaSelector: false,
            currentPersona: null,
            personaSessionId: null,
            personaMessages: [],
            personaInputText: ''
        };
    },
    
    computed: {
        lastMessage() {
            return this.messages[this.messages.length - 1];
        },
        userRole() {
            const user = JSON.parse(localStorage.getItem('user') || '{}');
            return user.role || 'user';
        },
        canUseThinking() {
            return this.userRole === 'vip' || this.userRole === 'admin';
        },
        canUseAuto() {
            return this.userRole === 'admin';
        }
    },
    
    methods: {
        /**
         * 初始化应用
         */
        async initialize() {
            await this.loadFileList();
            this.initTheme();
            this.loadSettings();
            this.loadSessions();

            // 恢复登录前未发送的消息草稿
            let pending = '';
            try { pending = sessionStorage.getItem('pendingChatInput') || ''; } catch (_) {}
            if (pending) {
                sessionStorage.removeItem('pendingChatInput');
            }

            // 仅在空会话时添加欢迎消息，避免污染恢复会话
            if (this.messages.length === 0) {
                this.addMessage({
                    type: 'assistant',
                    content: '你好！我是智能助手，可以帮你查询天气、进行数学计算或处理文件。\n\n**你可以问我：**\n- 今天北京的天气怎么样？\n- 计算：123 + 456\n- 帮我分析这个文件\n\n有什么可以帮你的吗？'
                });
            }

            if (pending) {
                // 稍作延迟让 UI 就绪
                setTimeout(() => this.handleSendMessage(pending), 200);
            }
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
         * 读取设置项
         */
        loadSettings() {
            this.showTimestamp = localStorage.getItem('setting-timestamp') !== 'false';
            this.soundEnabled = localStorage.getItem('setting-sound') === 'true';
        },

        /**
         * 从当前 messages 重建 history
         */
        rebuildHistoryFromMessages() {
            this.history = this.messages
                .filter(m => m.type === 'user' || m.type === 'assistant')
                .map(m => ({
                    role: m.type === 'user' ? 'user' : 'assistant',
                    content: m.content
                }));
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

            // 如果没有会话，创建一个新会话；否则恢复第一个会话的 messages/history
            if (this.sessions.length === 0) {
                this.createNewSession();
            } else {
                this.selectSession(this.sessions[0]);
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
            this.backendSessionId = null; // 新会话，后端 session_id 待生成
            this.messages = [];
            this.history = [];
            this.saveSessions();
        },
        
        /**
         * 选择会话
         */
        selectSession(session) {
            this.currentSessionId = session.id;
            this.backendSessionId = session.backendSessionId || null;
            this.messages = session.messages || [];
            this.rebuildHistoryFromMessages();
        },

        /**
         * 切换对话模式
         */
        setChatMode(mode) {
            if (mode === 'thinking' && !this.canUseThinking) {
                Toast.error('深度思考仅限 VIP / 管理员');
                return;
            }
            if (mode === 'auto' && !this.canUseAuto) {
                Toast.error('Auto 模式仅限管理员');
                return;
            }
            this.chatMode = mode;
            localStorage.setItem('chat-mode', mode);
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

            if (message.type === 'assistant') {
                this.playNotificationSound();
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

            // 提前拦截：本地即可判定 token 过期，直接跳登录并保留草稿
            if (isTokenExpiredLocally()) {
                try { sessionStorage.setItem('pendingChatInput', userMessage); } catch (_) {}
                Toast.warning('登录已过期，正在跳转登录页');
                setTimeout(() => { location.href = '/login.html'; }, 300);
                return;
            }

            // 权限自动降级
            let mode = this.chatMode;
            if (mode === 'thinking' && !this.canUseThinking) mode = 'normal';
            if (mode === 'auto' && !this.canUseAuto) mode = 'normal';

            // 添加用户消息
            this.addMessage({
                type: 'user',
                content: userMessage
            });

            this.loading = true;

            if (mode === 'normal') {
                await this.sendNormal(userMessage);
            } else {
                await this.sendStream(userMessage, mode);
            }
            this.loading = false;
        },

        /**
         * 普通模式发送（非流式）
         */
        async sendNormal(userMessage) {
            try {
                const response = await API.chat.send(userMessage, this.backendSessionId);
                if (response.session_id) {
                    this.backendSessionId = response.session_id;
                }
                this.addMessage({
                    type: 'assistant',
                    content: response.result || ''
                });
                if (response.title) {
                    this.updateCurrentSessionTitle(response.title);
                }
            } catch (error) {
                // 认证失效已由 API 层自动跳转，不再往聊天里塞错误
                if (error instanceof AuthExpiredError) return;
                const errorMsg = errorHandler.handle(error);
                this.addMessage({ type: 'error', content: `错误: ${errorMsg}` });
                errorHandler.log(error, '发送消息失败');
            }
        },

        /**
         * 流式发送（thinking / auto）
         */
        async sendStream(userMessage, mode) {
            const draft = {
                id: generateId(),
                type: 'assistant',
                content: '',
                thought: '',
                streaming: true,
                timestamp: Date.now()
            };
            this.messages.push(draft);
            // 从数组取回响应式代理，直接改本地对象不会触发 Vue 3 Proxy 更新
            const assistantMsg = this.messages[this.messages.length - 1];

            const setStreaming = (v) => { assistantMsg.streaming = v; this.updateCurrentSession(); };

            try {
                await API.chat.stream(userMessage, {
                    sessionId: this.backendSessionId,
                    mode,
                    onSession: (sid, title) => {
                        this.backendSessionId = sid;
                        if (title) this.updateCurrentSessionTitle(title);
                    },
                    onThought: (t) => {
                        assistantMsg.thought = (assistantMsg.thought || '') + t;
                    },
                    onAnswer: (t) => {
                        assistantMsg.content = (assistantMsg.content || '') + t;
                    },
                    onAnswerReset: () => {
                        assistantMsg.content = '';
                    },
                    onToolCall: (data) => {
                        assistantMsg.toolCalls = assistantMsg.toolCalls || [];
                        assistantMsg.toolCalls.push({
                            tool: data.tool || '',
                            input: data.tool_input || data.content || '',
                            result: ''
                        });
                    },
                    onToolResult: (data) => {
                        if (!assistantMsg.toolCalls || assistantMsg.toolCalls.length === 0) {
                            assistantMsg.toolCalls = [{ tool: data.tool || '', input: '', result: data.content || '' }];
                        } else {
                            const last = assistantMsg.toolCalls[assistantMsg.toolCalls.length - 1];
                            last.result = data.content || '';
                        }
                    },
                    onTitle: (title) => {
                        if (title) this.updateCurrentSessionTitle(title);
                    },
                    onDone: () => {
                        setStreaming(false);
                        // 完成后加入 history
                        this.history.push({ role: 'user', content: userMessage });
                        this.history.push({ role: 'assistant', content: assistantMsg.content });
                        this.playNotificationSound();
                    },
                    onError: (msg) => {
                        assistantMsg.content = `错误: ${msg}`;
                        assistantMsg.type = 'error';
                        setStreaming(false);
                    }
                });
            } catch (err) {
                assistantMsg.content = `错误: ${err.message || '未知错误'}`;
                assistantMsg.type = 'error';
                setStreaming(false);
            }
        },

        /** 更新当前会话标题 */
        updateCurrentSessionTitle(title) {
            const session = this.sessions.find(s => s.id === this.currentSessionId);
            if (session) {
                session.title = title;
                this.saveSessions();
            }
        },
        
        /**
         * 处理重新生成
         */
        async handleRegenerate(message) {
            const msgIndex = this.messages.findIndex(m => m.id === message.id);
            if (msgIndex <= 0) return;

            const userMsg = this.messages[msgIndex - 1];
            if (userMsg.type !== 'user') return;

            // 删除当前 assistant 回复并重建 history，避免重复 user 消息
            this.messages.splice(msgIndex, 1);
            this.rebuildHistoryFromMessages();
            this.updateCurrentSession();

            this.loading = true;
            try {
                const response = await API.chat.send(userMsg.content, this.history);

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

                this.addMessage({
                    type: 'assistant',
                    content
                });
            } catch (error) {
                const errorMsg = errorHandler.handle(error);
                this.addMessage({
                    type: 'error',
                    content: `重新生成失败: ${errorMsg}`
                });
                errorHandler.log(error, '重新生成失败');
            } finally {
                this.loading = false;
            }
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
            this.loadSettings();
        },
        
        /**
         * 主题变化处理
         */
        handleThemeChange(theme) {
            this.currentTheme = theme;
        },

        handleFileError(message) {
            Toast.error(message || '文件操作失败');
        },

        playNotificationSound() {
            if (!this.soundEnabled) return;
            try {
                const AudioContextClass = window.AudioContext || window.webkitAudioContext;
                if (!AudioContextClass) return;

                const ctx = new AudioContextClass();
                const oscillator = ctx.createOscillator();
                const gainNode = ctx.createGain();

                oscillator.type = 'sine';
                oscillator.frequency.setValueAtTime(880, ctx.currentTime);

                gainNode.gain.setValueAtTime(0.0001, ctx.currentTime);
                gainNode.gain.exponentialRampToValueAtTime(0.08, ctx.currentTime + 0.01);
                gainNode.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.12);

                oscillator.connect(gainNode);
                gainNode.connect(ctx.destination);

                oscillator.start();
                oscillator.stop(ctx.currentTime + 0.12);

                oscillator.onended = () => {
                    if (typeof ctx.close === 'function') {
                        ctx.close();
                    }
                };
            } catch (e) {
                // 音频不可用时静默降级
            }
        },

        // ========== 角色卡模式方法 ==========

        /**
         * 打开角色选择器
         */
        openPersonaSelector() {
            const user = JSON.parse(localStorage.getItem('user') || '{}');
            if (user.role !== 'vip' && user.role !== 'admin') {
                Toast.error('角色对话功能仅限 VIP 用户使用');
                return;
            }
            this.showPersonaSelector = true;
        },

        /**
         * 关闭角色选择器
         */
        closePersonaSelector() {
            this.showPersonaSelector = false;
        },

        /**
         * 选择角色并进入对话模式
         */
        async selectPersona(persona) {
            this.currentPersona = persona;
            this.personaMode = true;
            this.personaSessionId = null;
            this.personaMessages = [];
            this.showPersonaSelector = false;

            // 添加欢迎消息
            this.addPersonaMessage({
                type: 'assistant',
                content: `你好！我是${persona.name}。${persona.tagline || ''}\n\n有什么我可以帮你的吗？`
            });
        },

        /**
         * 退出角色模式
         */
        exitPersonaMode() {
            // 如果正在加载，提示用户
            if (this.loading) {
                Toast.warning('正在等待响应，请稍候...');
                return;
            }
            this.personaMode = false;
            this.currentPersona = null;
            this.personaSessionId = null;
            this.personaMessages = [];
        },

        /**
         * 添加角色对话消息
         */
        addPersonaMessage(messageData) {
            const message = {
                id: generateId(),
                type: messageData.type || 'assistant',
                content: messageData.content || '',
                timestamp: Date.now()
            };
            this.personaMessages.push(message);

            if (message.type === 'assistant') {
                this.playNotificationSound();
            }

            return message;
        },

        /**
         * 处理角色对话发送（默认流式 SSE）
         */
        async handlePersonaSendMessage(userMessage) {
            if (!userMessage || !userMessage.trim()) return;
            if (!this.currentPersona) return;

            // 添加用户消息
            this.addPersonaMessage({
                type: 'user',
                content: userMessage
            });

            // 预插入 assistant 占位气泡，用于流式追加。
            // Vue 3 的响应式基于 Proxy：必须通过数组下标访问代理对象后再修改属性，
            // 直接改 addPersonaMessage 返回的裸对象不会触发视图更新。
            this.addPersonaMessage({
                type: 'assistant',
                content: ''
            });
            const assistantIdx = this.personaMessages.length - 1;

            this.loading = true;

            try {
                await API.persona.chatStream(
                    this.currentPersona.id,
                    userMessage,
                    this.personaSessionId,
                    {
                        session: (d) => {
                            if (d && d.session_id) {
                                this.personaSessionId = d.session_id;
                            }
                        },
                        answer: (d) => {
                            if (d && d.content) {
                                this.personaMessages[assistantIdx].content += d.content;
                            }
                        },
                        done: (_d) => {
                            // 服务端已发送完整 full_response，无需覆盖
                        },
                        error: (d) => {
                            const m = this.personaMessages[assistantIdx];
                            m.type = 'error';
                            m.content = `错误: ${(d && d.content) || '未知错误'}`;
                        },
                    }
                );
            } catch (error) {
                const errorMsg = errorHandler.handle(error);
                const m = this.personaMessages[assistantIdx];
                m.type = 'error';
                m.content = `错误: ${errorMsg}`;
                errorHandler.log(error, '角色对话失败');
            } finally {
                this.loading = false;
            }
        },

        /**
         * 发送角色消息（从输入框）
         */
        sendPersonaMessage() {
            if (this.personaInputText?.trim()) {
                this.handlePersonaSendMessage(this.personaInputText);
                this.personaInputText = '';
            }
        },

        /**
         * 处理角色输入框 Enter 键
         */
        handlePersonaEnter(event) {
            if (!event.shiftKey) {
                this.sendPersonaMessage();
            }
        },

        /**
         * 渲染消息内容
         */
        renderContent(content, type) {
            // 简单的换行处理
            return content?.replace(/\n/g, '<br>') || '';
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
                <!-- 侧边栏（非角色模式时显示） -->
                <aside v-if="!personaMode" :class="['sidebar', sidebarCollapsed && 'collapsed']">
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

                <!-- 角色模式 -->
                <template v-if="personaMode">
                    <div class="main-content persona-mode">
                        <!-- 角色模式头部 -->
                        <div class="app-header persona-header">
                            <div class="header-left">
                                <button class="back-btn" @click="exitPersonaMode" :disabled="loading" :title="loading ? '正在等待响应...' : '返回'">← 返回</button>
                                <div class="persona-info-header">
                                    <div class="persona-avatar-small">
                                        {{ currentPersona?.name?.charAt(0) || '?' }}
                                    </div>
                                    <div class="header-title">
                                        <h1>{{ currentPersona?.name || '角色对话' }}</h1>
                                        <p>{{ currentPersona?.tagline || '' }}</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 消息列表 -->
                        <div class="messages-container">
                            <div class="message-list">
                                <div v-for="msg in personaMessages" :key="msg.id" :class="['message', msg.type]">
                                    <div class="message-content">
                                        <div class="message-text" v-html="renderContent(msg.content, msg.type)"></div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 输入区 -->
                        <div class="input-area">
                            <textarea
                                v-model="personaInputText"
                                placeholder="输入消息..."
                                @keydown.enter.prevent="handlePersonaEnter"
                                :disabled="loading"
                            ></textarea>
                            <button class="send-btn" @click="sendPersonaMessage" :disabled="loading || !personaInputText?.trim()">
                                {{ loading ? '发送中...' : '发送' }}
                            </button>
                        </div>
                    </div>
                </template>

                <!-- 普通聊天窗口 -->
                <template v-else>
                    <chat-window
                        :messages="messages"
                        :loading="loading"
                        :files="files"
                        :uploading="uploading"
                        :show-timestamp="showTimestamp"
                        :chat-mode="chatMode"
                        :can-use-thinking="canUseThinking"
                        :can-use-auto="canUseAuto"
                        @send="handleSendMessage"
                        @regenerate="handleRegenerate"
                        @toggle-sidebar="toggleSidebar"
                        @file-selected="handleFileSelected"
                        @delete-file="handleDeleteFile"
                        @analyze-file="handleAnalyzeFile"
                        @file-error="handleFileError"
                        @open-persona="openPersonaSelector"
                        @change-mode="setChatMode"
                    />
                </template>
            </div>

            <!-- 角色选择器 -->
            <persona-selector
                :visible="showPersonaSelector"
                :user-role="userRole"
                @select="selectPersona"
                @close="closePersonaSelector"
            />

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