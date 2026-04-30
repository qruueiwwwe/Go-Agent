/* ============================================
   MessageItem.js - 单条消息组件
   ============================================ */

const { defineComponent, h, ref, onMounted, nextTick } = Vue;
import { renderMarkdown, addCopyButtons, processLinks } from '../markdown.js';
import { formatTime, copyToClipboard } from '../utils.js';
import Toast from './Toast.js';

/**
 * MessageItem 组件 - 显示单条消息
 * 
 * Props:
 *   - message: {id, type, content, timestamp} 消息数据
 *   - showTimestamp: Boolean 是否显示时间戳
 * 
 * Events:
 *   - copy: 复制消息
 *   - regenerate: 重新生成
 *   - delete: 删除消息
 */
export const MessageItem = defineComponent({
    name: 'MessageItem',
    
    props: {
        message: {
            type: Object,
            required: true,
            validator: (msg) => msg.id && msg.type && (msg.content || msg.content === '')
        },
        showTimestamp: {
            type: Boolean,
            default: true
        }
    },
    
    data() {
        return {
            showActions: false,
            copied: false
        };
    },
    
    computed: {
        messageClass() {
            return ['message', this.message.type];
        },
        
        formattedTime() {
            if (!this.message.timestamp) return '';
            return formatTime(new Date(this.message.timestamp), 'HH:mm');
        },
        
        /**
         * 是否显示操作按钮（仅助手消息）
         */
        canShowActions() {
            return this.message.type === 'assistant' || this.message.type === 'user';
        }
    },
    
    methods: {
        /**
         * 复制消息内容
         */
        async handleCopy() {
            try {
                await copyToClipboard(this.message.content);
                this.copied = true;
                Toast.success('已复制到剪贴板');
                setTimeout(() => {
                    this.copied = false;
                }, 2000);
            } catch (e) {
                Toast.error('复制失败');
            }
        },
        
        /**
         * 重新生成回复
         */
        handleRegenerate() {
            this.$emit('regenerate', this.message);
        },
        
        /**
         * 删除消息
         */
        handleDelete() {
            this.$emit('delete', this.message);
        },
        
        /**
         * 渲染 Markdown 内容
         */
        renderMarkdownContent() {
            const { content, type } = this.message;
            
            // 用户消息：不渲染 Markdown，保留换行
            if (type === 'user') {
                return h('div', { class: 'message-text' }, 
                    content.split('\n').map((line, i, arr) => 
                        i < arr.length - 1 
                            ? [line, h('br')]
                            : line
                    ).flat()
                );
            }
            
            // 错误消息：简单显示
            if (type === 'error') {
                return h('div', { class: 'message-text' }, content);
            }
            
            // 系统消息：简单显示
            if (type === 'system') {
                return h('div', { class: 'message-text' }, content);
            }
            
            // 助手消息：渲染 Markdown
            const html = renderMarkdown(content);
            return h('div', {
                class: 'message-text markdown-body',
                innerHTML: html,
                ref: 'contentRef'
            });
        },
        
        /**
         * 渲染时间戳
         */
        renderTimestamp() {
            if (!this.showTimestamp || !this.formattedTime) return null;
            
            return h('div', { class: 'message-timestamp' }, this.formattedTime);
        },
        
        /**
         * 渲染操作按钮
         */
        renderActions() {
            if (!this.canShowActions) return null;
            
            const actions = [];
            
            // 复制按钮
            actions.push(h('button', {
                class: ['message-action-btn', this.copied && 'copied'],
                onClick: this.handleCopy,
                title: '复制'
            }, this.copied ? '已复制' : '复制'));
            
            // 重新生成按钮（仅助手消息）
            if (this.message.type === 'assistant') {
                actions.push(h('button', {
                    class: 'message-action-btn',
                    onClick: this.handleRegenerate,
                    title: '重新生成'
                }, '重新生成'));
            }
            
            return h('div', { class: 'message-actions' }, actions);
        }
    },
    
    mounted() {
        // 为代码块添加复制按钮
        nextTick(() => {
            if (this.$refs.contentRef && this.message.type === 'assistant') {
                addCopyButtons(this.$refs.contentRef);
                processLinks(this.$refs.contentRef);
            }
        });
    },
    
    render() {
        const children = [
            this.renderMarkdownContent(),
            this.renderTimestamp()
        ];
        
        // 添加操作按钮（悬浮显示）
        if (this.canShowActions) {
            children.push(this.renderActions());
        }
        
        return h('div', {
            class: this.messageClass,
            key: this.message.id,
            onMouseenter: () => { this.showActions = true; },
            onMouseleave: () => { this.showActions = false; }
        }, [
            h('div', { class: 'message-content' }, children)
        ]);
    }
});

export default MessageItem;