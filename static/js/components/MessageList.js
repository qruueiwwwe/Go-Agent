/* ============================================
   MessageList.js - 消息列表组件
   ============================================ */

const { defineComponent, h, nextTick, ref, watch } = Vue;
import { MessageItem } from './MessageItem.js';

/**
 * MessageList 组件 - 显示所有消息列表
 * 
 * Props:
 *   - messages: Array 消息数组
 *   - loading: Boolean 是否加载中
 *   - showTimestamp: Boolean 是否显示时间戳
 * 
 * Events:
 *   - message-click: 消息被点击
 *   - regenerate: 重新生成消息
 *   - delete: 删除消息
 */
export const MessageList = defineComponent({
    name: 'MessageList',
    
    components: {
        MessageItem
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
        showTimestamp: {
            type: Boolean,
            default: true
        }
    },
    
    data() {
        return {
            autoScroll: true
        };
    },
    
    watch: {
        messages: {
            handler() {
                if (this.autoScroll) {
                    this.$nextTick(() => this.scrollToBottom());
                }
            },
            deep: true
        },
        
        loading(newVal) {
            if (!newVal && this.autoScroll) {
                this.$nextTick(() => this.scrollToBottom());
            }
        }
    },
    
    mounted() {
        this.scrollToBottom();
        
        const container = this.$refs.messagesContainer;
        if (container) {
            container.addEventListener('scroll', this.handleScroll);
        }
    },
    
    beforeUnmount() {
        const container = this.$refs.messagesContainer;
        if (container) {
            container.removeEventListener('scroll', this.handleScroll);
        }
    },
    
    methods: {
        scrollToBottom() {
            nextTick(() => {
                const container = this.$refs.messagesContainer;
                if (container) {
                    container.scrollTo({
                        top: container.scrollHeight,
                        behavior: 'smooth'
                    });
                }
            });
        },
        
        handleScroll() {
            const container = this.$refs.messagesContainer;
            if (!container) return;
            
            const threshold = 100;
            const isAtBottom = container.scrollHeight - container.scrollTop - container.clientHeight < threshold;
            this.autoScroll = isAtBottom;
        },
        
        handleRegenerate(message) {
            this.$emit('regenerate', message);
        },
        
        handleDelete(message) {
            this.$emit('delete', message);
        },
        
        renderEmptyState() {
            return h('div', { class: 'empty-state animate-fadeIn' }, [
                h('div', { class: 'empty-state-icon' }, [
                    h('span', { class: 'empty-state-emoji' }, '💬')
                ]),
                h('div', { class: 'empty-state-content' }, [
                    h('h3', { class: 'empty-state-title' }, '开始对话'),
                    h('p', { class: 'empty-state-desc' }, '输入消息开始与 AI 助手交流')
                ])
            ]);
        },
        
        renderLoading() {
            if (!this.loading) return null;
            
            return h('div', { class: 'message assistant' }, [
                h('div', { class: 'message-content loading-content' }, [
                    h('div', { class: 'loading-wave' }, [
                        h('span'),
                        h('span'),
                        h('span'),
                        h('span'),
                        h('span')
                    ])
                ])
            ]);
        },
        
        renderMessages() {
            const items = this.messages.map(msg => 
                h(MessageItem, {
                    key: msg.id,
                    message: msg,
                    showTimestamp: this.showTimestamp,
                    onRegenerate: this.handleRegenerate,
                    onDelete: this.handleDelete
                })
            );
            
            items.push(this.renderLoading());
            
            return items;
        }
    },
    
    render() {
        const hasMessages = this.messages.length > 0;
        
        return h('div', {
            ref: 'messagesContainer',
            class: 'messages-container'
        }, hasMessages ? this.renderMessages() : this.renderEmptyState());
    }
});

export default MessageList;