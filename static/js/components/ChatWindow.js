/* ============================================
   ChatWindow.js - 聊天窗口组件
   ============================================ */

const { defineComponent, h } = Vue;
import { MessageList } from './MessageList.js';
import { InputArea } from './InputArea.js';

/**
 * ChatWindow 组件 - 聊天主窗口
 * 
 * Props:
 *   - messages: Array 消息列表
 *   - loading: Boolean 是否加载中
 *   - title: String 窗口标题
 *   - subtitle: String 窗口副标题
 * 
 * Events:
 *   - send: 发送消息 (payload: 消息内容)
 *   - message-click: 消息被点击
 * 
 * 聚合 MessageList 和 InputArea 组件
 * 管理聊天窗口的整体交互
 */
export const ChatWindow = defineComponent({
    name: 'ChatWindow',
    
    components: {
        MessageList,
        InputArea
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
        }
    },
    
    methods: {
        /**
         * 处理消息发送
         */
        handleSend(message) {
            this.$emit('send', message);
        },
        
        /**
         * 处理消息点击
         */
        handleMessageClick(data) {
            this.$emit('message-click', data);
        }
    },
    
    render() {
        return h('div', { class: 'app-container' }, [
            // 头部
            h('div', { class: 'app-header' }, [
                h('h1', this.title),
                h('p', this.subtitle)
            ]),
            
            // 消息列表
            h(MessageList, {
                messages: this.messages,
                loading: this.loading,
                onMessageClick: this.handleMessageClick
            }),
            
            // 输入区
            h(InputArea, {
                disabled: this.loading,
                onSend: this.handleSend
            })
        ]);
    }
});

export default ChatWindow;
