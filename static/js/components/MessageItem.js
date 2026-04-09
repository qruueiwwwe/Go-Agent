/* ============================================
   MessageItem.js - 单条消息组件
   ============================================ */

const { defineComponent, h } = Vue;

/**
 * MessageItem 组件 - 显示单条消息
 * 
 * Props:
 *   - message: {id, type, content} 消息数据
 *   - type: string 消息类型 (user/assistant/error/system)
 *   - content: string 消息内容
 * 
 * Events: 无
 * 
 * 消息类型说明:
 *   - user: 用户消息（蓝色背景）
 *   - assistant: 助手消息（灰色背景）
 *   - error: 错误消息（红色背景）
 *   - system: 系统消息（蓝色背景，居中）
 */
export const MessageItem = defineComponent({
    name: 'MessageItem',
    
    props: {
        message: {
            type: Object,
            required: true,
            validator: (msg) => msg.id && msg.type && (msg.content || msg.content === '')
        }
    },
    
    computed: {
        /**
         * 消息的 CSS 类名
         */
        messageClass() {
            return ['message', this.message.type];
        },
        
        /**
         * 消息内容类
         */
        contentClass() {
            return ['message-content'];
        }
    },
    
    methods: {
        /**
         * 处理消息点击（预留用于复制等功能）
         */
        handleClick(e) {
            // 可以在这里实现消息复制、编辑等功能
            this.$emit('message-click', {
                message: this.message,
                event: e
            });
        },
        
        /**
         * 渲染消息内容（支持换行）
         */
        renderContent() {
            const { content, type } = this.message;
            
            // 如果是代码块或包含特殊格式，可以在这里处理
            // 例如：Markdown 渲染、代码高亮等（后续扩展）
            
            if (type === 'assistant' && content.includes('\n')) {
                // 保留换行符格式
                return content.split('\n').map((line, index) => [
                    h('span', line),
                    index < content.split('\n').length - 1 ? h('br') : null
                ]).flat().filter(Boolean);
            }
            
            return content;
        }
    },
    
    render() {
        return h(
            'div',
            {
                class: this.messageClass,
                key: this.message.id
            },
            h(
                'div',
                {
                    class: this.contentClass,
                    onClick: this.handleClick
                },
                this.renderContent()
            )
        );
    }
});

export default MessageItem;
