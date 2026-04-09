/* ============================================
   app.js - Vue 应用入口和根组件
   ============================================ */

import { ChatWindow } from './components/ChatWindow.js';
import { FileManager } from './components/FileManager.js';
import API, { errorHandler } from './api.js';
import { generateId } from './utils.js';

const { createApp, ref, reactive, computed, nextTick } = Vue;

/**
 * 创建 Vue 应用
 */
const app = createApp({
    name: 'App',
    
    components: {
        ChatWindow,
        FileManager
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
            history: []
        };
    },
    
    computed: {
        /**
         * 最后一条消息
         */
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
            
            // 添加欢迎消息
            this.addMessage({
                type: 'assistant',
                content: '你好！我是智能助手，可以帮你查询天气、进行数学计算或处理文件。有什么可以帮你的吗？'
            });
        },
        
        /**
         * 加载文件列表
         */
        async loadFileList() {
            try {
                this.files = await API.files.list();
            } catch (error) {
                console.error('加载文件列表失败:', error);
                this.addMessage({
                    type: 'error',
                    content: `加载文件失败: ${error.message}`
                });
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
                
                // 记录错误
                errorHandler.log(error, '发送消息失败');
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
                const result = await API.files.upload(file);
                
                // 重新加载文件列表
                await this.loadFileList();
                
                // 添加成功消息
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
                
                // 重新加载文件列表
                await this.loadFileList();
                
                // 添加成功消息
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
         * 处理消息点击（预留用于复制等）
         */
        handleMessageClick(data) {
            console.log('消息点击:', data);
            // 可以在这里实现消息复制、编辑等功能
        }
    },
    
    mounted() {
        // 初始化应用
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
    },
    
    beforeUnmount() {
        window.removeEventListener('online', null);
        window.removeEventListener('offline', null);
    },
    
    template: `
        <div id="app">
            <div class="app-container">
                <!-- 文件管理 -->
                <file-manager
                    :files="files"
                    :uploading="uploading"
                    @file-selected="handleFileSelected"
                    @delete-file="handleDeleteFile"
                    @analyze-file="handleAnalyzeFile"
                    @error="(msg) => addMessage({ type: 'error', content: msg })"
                />
                
                <!-- 聊天窗口 -->
                <chat-window
                    :messages="messages"
                    :loading="loading"
                    @send="handleSendMessage"
                    @message-click="handleMessageClick"
                />
            </div>
        </div>
    `
});

/**
 * 导出应用
 */
export default app;
