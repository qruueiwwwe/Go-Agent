/* ============================================
   QuickActions.js - 快捷操作组件
   ============================================ */

const { defineComponent, h } = Vue;

/**
 * 预设快捷命令
 */
const QUICK_COMMANDS = [
    { id: 'weather', label: '天气查询', icon: '🌤️', template: '今天{city}的天气怎么样？', placeholder: '请输入城市名' },
    { id: 'calc', label: '数学计算', icon: '🔢', template: '计算：{expression}', placeholder: '请输入算式' },
    { id: 'translate', label: '翻译', icon: '🌐', template: '翻译成中文：{text}', placeholder: '请输入要翻译的内容' },
    { id: 'summarize', label: '摘要', icon: '📝', template: '请总结以下内容：\n{content}', placeholder: '请输入内容' }
];

/**
 * QuickActions 组件 - 快捷操作按钮
 * 
 * Props:
 *   - disabled: Boolean 是否禁用
 * 
 * Events:
 *   - action: 执行快捷命令 (payload: { id, template })
 */
export const QuickActions = defineComponent({
    name: 'QuickActions',
    
    props: {
        disabled: {
            type: Boolean,
            default: false
        }
    },
    
    data() {
        return {
            expanded: false
        };
    },
    
    methods: {
        handleToggle() {
            this.expanded = !this.expanded;
        },
        
        handleAction(action) {
            this.$emit('action', action);
            this.expanded = false;
        },
        
        renderToggleButton() {
            return h('button', {
                class: ['quick-toggle', this.expanded && 'active'],
                disabled: this.disabled,
                onClick: this.handleToggle,
                title: '快捷命令'
            }, '⚡');
        },
        
        renderActionList() {
            if (!this.expanded) return null;
            
            return h('div', { class: 'quick-menu animate-slideInUp' }, 
                QUICK_COMMANDS.map(action => 
                    h('button', {
                        key: action.id,
                        class: 'quick-action-btn',
                        onClick: () => this.handleAction(action)
                    }, [
                        h('span', { class: 'action-icon' }, action.icon),
                        h('span', { class: 'action-label' }, action.label)
                    ])
                )
            );
        }
    },
    
    render() {
        return h('div', { class: 'quick-actions' }, [
            this.renderToggleButton(),
            this.renderActionList()
        ]);
    }
});

export { QUICK_COMMANDS };
export default QuickActions;