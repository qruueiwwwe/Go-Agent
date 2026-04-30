/* ============================================
   InputArea.js - 输入区域组件
   ============================================ */

const { defineComponent, h, ref } = Vue;
import { QuickActions, QUICK_COMMANDS } from './QuickActions.js';

/**
 * InputArea 组件 - 消息输入框
 * 
 * Props:
 *   - disabled: Boolean 是否禁用
 *   - placeholder: String 占位符文本
 * 
 * Events:
 *   - send: 发送消息 (payload: 消息内容)
 * 
 * Features:
 *   - 回车发送消息（Shift+Enter 换行）
 *   - 文本自动扩展高度
 *   - 禁用状态处理
 *   - 快捷命令支持
 */
export const InputArea = defineComponent({
    name: 'InputArea',
    
    components: {
        QuickActions
    },
    
    props: {
        disabled: {
            type: Boolean,
            default: false
        },
        placeholder: {
            type: String,
            default: '请输入你的问题...'
        }
    },
    
    data() {
        return {
            input: '',
            isComposing: false
        };
    },
    
    computed: {
        sendDisabled() {
            return this.disabled || !this.input.trim();
        }
    },
    
    watch: {
        input() {
            this.$nextTick(() => this.autoResize());
        },
        
        disabled(newVal) {
            if (!newVal) {
                this.$nextTick(() => {
                    this.$refs.inputRef?.focus();
                });
            }
        }
    },
    
    mounted() {
        this.$nextTick(() => {
            this.$refs.inputRef?.focus();
        });
    },
    
    methods: {
        autoResize() {
            const textarea = this.$refs.inputRef;
            if (!textarea) return;
            
            textarea.style.height = 'auto';
            const newHeight = Math.min(textarea.scrollHeight, 120);
            textarea.style.height = `${newHeight}px`;
        },
        
        handleKeyDown(e) {
            // IME 输入过程中（拼音/日文等），回车不发送
            if (this.isComposing) return;
            
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.handleSend();
            }
        },
        
        handleCompositionStart() {
            this.isComposing = true;
        },
        
        handleCompositionEnd(e) {
            this.isComposing = false;
            this.input = e.target.value;
        },
        
        handleInput(e) {
            // compositionend 已经更新过值，避免重复
            if (!this.isComposing) {
                this.input = e.target.value;
            }
        },
        
        handleSend() {
            if (this.sendDisabled) return;
            
            const message = this.input.trim();
            if (!message) return;
            
            this.$emit('send', message);
            
            this.input = '';
            this.$nextTick(() => {
                this.$refs.inputRef?.focus();
                this.autoResize();
            });
        },
        
        handleQuickAction(action) {
            // 将快捷命令模板插入输入框
            if (action.template) {
                // 用占位符提示用户填写
                const placeholderText = action.placeholder || '请输入内容';
                this.input = action.template.replace(/\{[^}]+\}/, `[${placeholderText}]`);
            }
            this.$refs.inputRef?.focus();
        }
    },
    
    render() {
        return h('div', { class: 'input-area' }, [
            h('div', { class: 'input-wrapper' }, [
                h(QuickActions, {
                    disabled: this.disabled,
                    onAction: this.handleQuickAction
                }),
                h('textarea', {
                    ref: 'inputRef',
                    class: 'message-input',
                    value: this.input,
                    placeholder: this.placeholder,
                    disabled: this.disabled,
                    onInput: this.handleInput,
                    onKeydown: this.handleKeyDown,
                    onCompositionstart: this.handleCompositionStart,
                    onCompositionend: this.handleCompositionEnd
                })
            ]),
            h('button', {
                class: ['send-btn'],
                disabled: this.sendDisabled,
                onClick: this.handleSend
            }, [
                h('span', '发送')
            ])
        ]);
    }
});

export default InputArea;