/* ============================================
   InputArea.js - 输入区域组件
   ============================================ */

const { defineComponent, h, ref } = Vue;

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
 *   - 粘贴图片预处理（预留）
 */
export const InputArea = defineComponent({
    name: 'InputArea',
    
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
            inputRef: ref(null)
        };
    },
    
    computed: {
        /**
         * 发送按钮是否禁用
         */
        sendDisabled() {
            return this.disabled || !this.input.trim();
        }
    },
    
    watch: {
        /**
         * 监听输入框值变化，自动调整高度
         */
        input() {
            this.$nextTick(() => {
                this.autoResize();
            });
        },
        
        /**
         * 监听禁用状态变化
         */
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
        /**
         * 自动调整输入框高度
         */
        autoResize() {
            const textarea = this.$refs.inputRef;
            if (!textarea) return;
            
            // 重置高度以获得准确的 scrollHeight
            textarea.style.height = 'auto';
            
            // 设置新高度（最大 120px）
            const newHeight = Math.min(textarea.scrollHeight, 120);
            textarea.style.height = `${newHeight}px`;
        },
        
        /**
         * 处理键盘事件
         */
        handleKeyDown(e) {
            // Enter 发送，Shift+Enter 换行
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.handleSend();
            }
        },
        
        /**
         * 处理输入框输入
         */
        handleInput(e) {
            this.input = e.target.value;
        },
        
        /**
         * 发送消息
         */
        handleSend() {
            if (this.sendDisabled) return;
            
            const message = this.input.trim();
            if (!message) return;
            
            // 发出 send 事件
            this.$emit('send', message);
            
            // 清空输入框
            this.input = '';
            this.$nextTick(() => {
                this.$refs.inputRef?.focus();
                this.autoResize();
            });
        },
        
        /**
         * 处理粘贴事件（预留用于处理图片）
         */
        handlePaste(e) {
            // 未来可以在这里处理粘贴的图片
            // const items = e.clipboardData?.items;
            // for (const item of items) {
            //     if (item.type.startsWith('image/')) {
            //         // 处理图片
            //     }
            // }
        },
        
        /**
         * 焦点处理
         */
        handleFocus() {
            this.$emit('focus');
        },
        
        handleBlur() {
            this.$emit('blur');
        }
    },
    
    render() {
        return h('div', { class: 'input-area' }, [
            h('div', { class: 'input-wrapper' }, [
                h('textarea', {
                    ref: 'inputRef',
                    class: 'message-input',
                    value: this.input,
                    placeholder: this.placeholder,
                    disabled: this.disabled,
                    onInput: this.handleInput,
                    onKeydown: this.handleKeyDown,
                    onPaste: this.handlePaste,
                    onFocus: this.handleFocus,
                    onBlur: this.handleBlur
                })
            ]),
            h('button', {
                class: ['send-btn'],
                disabled: this.sendDisabled,
                onClick: this.handleSend
            }, [
                h('span', '📤'),
                h('span', '发送')
            ])
        ]);
    }
});

export default InputArea;
