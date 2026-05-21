/* ============================================
   PersonaSelector.js - 角色卡选择组件
   ============================================ */

const { defineComponent, h, ref, onMounted, watch } = Vue;
import API, { errorHandler } from '../api.js';

/**
 * PersonaSelector 组件 - 角色卡选择面板
 */
export const PersonaSelector = defineComponent({
    name: 'PersonaSelector',

    props: {
        visible: {
            type: Boolean,
            default: false
        },
        userRole: {
            type: String,
            default: 'user'
        }
    },

    emits: ['select', 'close'],

    setup(props, { emit }) {
        const personas = ref([]);
        const loading = ref(false);
        const activeTab = ref('public');
        const error = ref('');
        const showCreateForm = ref(false);
        const creating = ref(false);
        const newPersona = ref({
            name: '',
            tagline: '',
            personality: '',
            system_prompt: '',
            is_public: false
        });

        // 加载角色列表
        async function loadPersonas() {
            loading.value = true;
            error.value = '';
            try {
                const result = await API.persona.list(activeTab.value);
                personas.value = result.list || [];
            } catch (err) {
                error.value = errorHandler.handle(err);
                console.error('加载角色列表失败:', err);
            } finally {
                loading.value = false;
            }
        }

        // 切换标签
        function switchTab(tab) {
            activeTab.value = tab;
            showCreateForm.value = false;
            loadPersonas();
        }

        // 选择角色
        function selectPersona(persona) {
            emit('select', persona);
        }

        // 关闭面板
        function closePanel() {
            showCreateForm.value = false;
            resetForm();
            emit('close');
        }

        // 重置表单
        function resetForm() {
            newPersona.value = {
                name: '',
                tagline: '',
                personality: '',
                system_prompt: '',
                is_public: false
            };
        }

        // 显示创建表单
        function openCreateForm() {
            showCreateForm.value = true;
            resetForm();
        }

        // 取消创建
        function cancelCreate() {
            showCreateForm.value = false;
            resetForm();
        }

        // 创建角色
        async function createPersona() {
            if (!newPersona.value.name.trim()) {
                error.value = '请输入角色名称';
                return;
            }
            if (!newPersona.value.system_prompt.trim()) {
                error.value = '请输入系统提示词';
                return;
            }

            creating.value = true;
            error.value = '';

            try {
                await API.persona.create(newPersona.value);
                showCreateForm.value = false;
                resetForm();
                loadPersonas();
            } catch (err) {
                error.value = errorHandler.handle(err);
                console.error('创建角色失败:', err);
            } finally {
                creating.value = false;
            }
        }

        // 监听显示状态
        watch(() => props.visible, (newVal) => {
            if (newVal) {
                loadPersonas();
                showCreateForm.value = false;
            }
        });

        // 初始加载
        onMounted(() => {
            if (props.visible) {
                loadPersonas();
            }
        });

        return {
            personas,
            loading,
            activeTab,
            error,
            showCreateForm,
            creating,
            newPersona,
            loadPersonas,
            switchTab,
            selectPersona,
            closePanel,
            openCreateForm,
            cancelCreate,
            createPersona
        };
    },

    render() {
        if (!this.visible) return null;

        const user = JSON.parse(localStorage.getItem('user') || '{}');
        const isVIP = user.role === 'vip' || user.role === 'admin';

        return h('div', { class: 'persona-selector-overlay' }, [
            h('div', { class: 'persona-selector-panel' }, [
                // 头部
                h('div', { class: 'persona-selector-header' }, [
                    h('h3', this.showCreateForm ? '创建角色' : '选择角色'),
                    h('button', {
                        class: 'close-btn',
                        onClick: this.closePanel
                    }, '✕')
                ]),

                // 非VIP提示
                !isVIP ? h('div', { class: 'persona-vip-notice' }, [
                    h('p', '角色对话功能仅限 VIP 用户使用'),
                    h('p', { class: 'sub' }, '请联系管理员升级账号')
                ]) : [
                    // 创建角色表单
                    this.showCreateForm ? [
                        h('div', { class: 'persona-create-form' }, [
                            // 角色名称
                            h('div', { class: 'form-group' }, [
                                h('label', '角色名称 *'),
                                h('input', {
                                    type: 'text',
                                    placeholder: '例如：面试官 Lisa',
                                    value: this.newPersona.name,
                                    onInput: (e) => this.newPersona.name = e.target.value
                                })
                            ]),
                            // 一句话介绍
                            h('div', { class: 'form-group' }, [
                                h('label', '一句话介绍'),
                                h('input', {
                                    type: 'text',
                                    placeholder: '例如：严厉但专业的技术面试官',
                                    value: this.newPersona.tagline,
                                    onInput: (e) => this.newPersona.tagline = e.target.value
                                })
                            ]),
                            // 性格特质
                            h('div', { class: 'form-group' }, [
                                h('label', '性格特质'),
                                h('input', {
                                    type: 'text',
                                    placeholder: '例如：严肃、追问细节、给建设性反馈',
                                    value: this.newPersona.personality,
                                    onInput: (e) => this.newPersona.personality = e.target.value
                                })
                            ]),
                            // 系统提示词
                            h('div', { class: 'form-group' }, [
                                h('label', '系统提示词 *'),
                                h('textarea', {
                                    placeholder: '描述角色的行为方式、专业知识等...',
                                    value: this.newPersona.system_prompt,
                                    onInput: (e) => this.newPersona.system_prompt = e.target.value,
                                    rows: 5
                                })
                            ]),
                            // 公开设置
                            h('div', { class: 'form-group checkbox' }, [
                                h('label', [
                                    h('input', {
                                        type: 'checkbox',
                                        checked: this.newPersona.is_public,
                                        onChange: (e) => this.newPersona.is_public = e.target.checked
                                    }),
                                    ' 公开此角色（其他用户可见）'
                                ])
                            ]),
                            // 错误提示
                            this.error ? h('div', { class: 'persona-error' }, this.error) : null,
                            // 按钮组
                            h('div', { class: 'form-actions' }, [
                                h('button', {
                                    class: 'btn-cancel',
                                    onClick: this.cancelCreate,
                                    disabled: this.creating
                                }, '取消'),
                                h('button', {
                                    class: 'btn-create',
                                    onClick: this.createPersona,
                                    disabled: this.creating
                                }, this.creating ? '创建中...' : '创建角色')
                            ])
                        ])
                    ] : [
                        // 标签切换
                        h('div', { class: 'persona-tabs' }, [
                            h('button', {
                                class: ['tab-btn', this.activeTab === 'public' && 'active'],
                                onClick: () => this.switchTab('public')
                            }, '公开角色'),
                            h('button', {
                                class: ['tab-btn', this.activeTab === 'mine' && 'active'],
                                onClick: () => this.switchTab('mine')
                            }, '我的角色')
                        ]),

                        // 我的角色 - 创建按钮
                        this.activeTab === 'mine' ? h('div', { class: 'persona-actions' }, [
                            h('button', {
                                class: 'create-btn',
                                onClick: this.openCreateForm
                            }, '+ 创建新角色')
                        ]) : null,

                        // 错误提示
                        this.error ? h('div', { class: 'persona-error' }, this.error) : null,

                        // 加载状态
                        this.loading ? h('div', { class: 'persona-loading' }, [
                            h('div', { class: 'spinner' }),
                            h('span', '加载中...')
                        ]) : null,

                        // 角色列表
                        !this.loading && this.personas.length === 0 ? 
                            h('div', { class: 'persona-empty' }, [
                                h('p', this.activeTab === 'mine' ? '你还没有创建角色，点击上方按钮创建' : '暂无可用角色')
                            ]) :
                            h('div', { class: 'persona-list' }, 
                                this.personas.map(persona => 
                                    h('div', {
                                        class: 'persona-card',
                                        key: persona.id,
                                        onClick: () => this.selectPersona(persona)
                                    }, [
                                        // 头像
                                        h('div', { class: 'persona-avatar' }, 
                                            persona.avatar ? 
                                                h('img', { src: persona.avatar, alt: persona.name }) :
                                                h('div', { class: 'avatar-placeholder' }, persona.name.charAt(0))
                                        ),
                                        // 信息
                                        h('div', { class: 'persona-info' }, [
                                            h('div', { class: 'persona-name' }, persona.name),
                                            persona.tagline ? h('div', { class: 'persona-tagline' }, persona.tagline) : null,
                                            h('div', { class: 'persona-meta' }, [
                                                h('span', { class: 'usage-count' }, `${persona.usage_count || 0} 次对话`)
                                            ])
                                        ])
                                    ])
                                )
                            )
                    ]
                ].flat().filter(Boolean)
            ])
        ]);
    }
});

export default PersonaSelector;
