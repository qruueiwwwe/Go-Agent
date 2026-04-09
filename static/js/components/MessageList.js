/* ============================================
   MessageList.js - 消息列表组件
   ============================================ */

const { defineComponent, h, nextTick, ref } = Vue;
import { MessageItem } from './MessageItem.js';

/**
 * MessageList 组件 - 显示所有消息列表
 * 
 * Props:
 *   - messages: Array 消息数组
 *   - loading: Boolean 是否加载中
 * 
 * Events:
 *   - message-click: 消息被点击
 * 
 * Features:
 *   - 自动滚动到最新消息
 *   - 空状态显示
 *   - 加载动画
 *   - 预留虚拟滚动空间
 */
export const MessageList = defineComponent({
    name: 'MessageList',
    
    components: {
        MessageItem
    },
    
    props: {
        messages: {
            type: Array,
            default: () => [],
            validator: (arr) => Array.isArray(arr)
        },
        loading: {
            type: Boolean,
            default: false
        }
    },
    
    data() {
        return {
            containerRef: ref(null),
            autoScroll: true
        };
    },
    
    watch: {
        /**
         * 监听消息变化，自动滚动到底部
         */
        messages: {
            handler(newMessages) {
                this.$nextTick(() => {
                    this.scrollToBottom();
                });
            },
            deep: false
        },
        
        /**
         * 监听加载状态变化
         */
        loading(newVal) {
            if (!newVal) {
                this.$nextTick(() => {
                    this.scrollToBottom();
                });
            }
        }
    },
    
    mounted() {
        // 初始滚动到底部
        this.$nextTick(() => {
            this.scrollToBottom();
        });
        
        // 监听容器滚动，判断是否需要自动滚动
        this.$refs.messagesContainer?.addEventListener('scroll', () => {
            this.handleScroll();
        });
    },
    
    beforeUnmount() {
        this.$refs.messagesContainer?.removeEventListener('scroll', () => {
            this.handleScroll();
        });
    },
    
    methods: {
        /**
         * 滚动到底部
         */
        scrollToBottom() {
            if (!this.$refs.messagesContainer) return;
            
            nextTick(() => {
                const container = this.$refs.messagesContainer;
                container.scrollTop = container.scrollHeight;
            });
        },
        
        /**
         * 处理容器滚动
         */
        handleScroll() {
            const container = this.$refs.messagesContainer;
            if (!container) return;
            
            // 判断是否在底部（留 100px 缓冲区）
            const isAtBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 100;
            this.autoScroll = isAtBottom;
        },
        
        /**
         * 处理消息点击事件
         */
        handleMessageClick(message) {
            this.$emit('message-click', message);
        },
        
        /**
         * 渲染空状态
         */
        renderEmptyState() {
            return h('div', { class: 'empty-state' }, [
                h('div', { class: 'empty-state-icon' }, '💬'),
                h('div', { class: 'empty-state-text' }, '还没有消息，开始聊天吧！')
            ]);
        },
        
        /**
         * 渲染加载状态
         */
        renderLoading() {
            if (!this.loading) return null;
            
            return h('div', { class: 'loading active' }, '正在思考中...');
        },
        
        /**
         * 渲染消息列表
         */
        renderMessages() {
            if (this.messages.length === 0 && !this.loading) {
                return this.renderEmptyState();
            }
            
            return [
                ...this.messages.map(msg => 
                    h(MessageItem, {
                        key: msg.id,
                        message: msg,
                        onMessageClick: (data) => {
                            this.$emit('message-click', data);
                        }
                    })
                ),
                this.renderLoading()
            ];
        }
    },
    
    render() {
        return h(
            'div',
            {
                ref: 'messagesContainer',
                class: 'messages-container'
            },
            this.renderMessages()
        );
    }
});

export default MessageList;
