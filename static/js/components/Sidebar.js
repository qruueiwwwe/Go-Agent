/* ============================================
   Sidebar.js - 侧边栏组件
   ============================================ */

const { defineComponent, h, ref, computed } = Vue;

/**
 * Sidebar 组件 - 侧边栏导航
 * 
 * Props:
 *   - collapsed: Boolean 是否收起
 *   - sessions: Array 会话列表
 *   - currentSessionId: String 当前会话ID
 * 
 * Events:
 *   - toggle: 切换侧边栏
 *   - new-session: 新建会话
 *   - select-session: 选择会话
 *   - delete-session: 删除会话
 */
export const Sidebar = defineComponent({
    name: 'Sidebar',
    
    props: {
        collapsed: {
            type: Boolean,
            default: false
        },
        sessions: {
            type: Array,
            default: () => []
        },
        currentSessionId: {
            type: String,
            default: null
        }
    },
    
    data() {
        return {
            searchQuery: '',
            hoveredSession: null
        };
    },
    
    computed: {
        filteredSessions() {
            if (!this.searchQuery) return this.sessions;
            const query = this.searchQuery.toLowerCase();
            return this.sessions.filter(s => 
                s.title?.toLowerCase().includes(query) ||
                s.preview?.toLowerCase().includes(query)
            );
        }
    },
    
    methods: {
        handleToggle() {
            this.$emit('toggle');
        },
        
        handleNewSession() {
            this.$emit('new-session');
        },
        
        handleSelectSession(session) {
            this.$emit('select-session', session);
        },
        
        handleDeleteSession(session, e) {
            e.stopPropagation();
            if (confirm(`确定删除会话 "${session.title || '未命名会话'}" 吗？`)) {
                this.$emit('delete-session', session);
            }
        },
        
        renderHeader() {
            return h('div', { class: 'sidebar-header' }, [
                h('button', {
                    class: 'sidebar-toggle',
                    onClick: this.handleToggle,
                    title: this.collapsed ? '展开' : '收起'
                }, [
                    h('span', { class: 'toggle-icon' }, this.collapsed ? '☰' : '✕')
                ]),
                !this.collapsed && h('h2', { class: 'sidebar-title' }, '会话历史')
            ]);
        },
        
        renderNewButton() {
            if (this.collapsed) return null;
            
            return h('button', {
                class: 'new-session-btn',
                onClick: this.handleNewSession
            }, [
                h('span', { class: 'btn-icon' }, '+'),
                h('span', { class: 'btn-text' }, '新建会话')
            ]);
        },
        
        renderSearch() {
            if (this.collapsed) return null;
            
            return h('div', { class: 'sidebar-search' }, [
                h('input', {
                    type: 'text',
                    class: 'search-input',
                    placeholder: '搜索会话...',
                    value: this.searchQuery,
                    onInput: (e) => { this.searchQuery = e.target.value; }
                })
            ]);
        },
        
        renderSessionItem(session) {
            const isActive = session.id === this.currentSessionId;
            const isHovered = session.id === this.hoveredSession;
            
            return h('div', {
                key: session.id,
                class: ['session-item', isActive && 'active'],
                onClick: () => this.handleSelectSession(session),
                onMouseenter: () => { this.hoveredSession = session.id; },
                onMouseleave: () => { this.hoveredSession = null; }
            }, [
                h('div', { class: 'session-icon' }, '💬'),
                !this.collapsed && h('div', { class: 'session-content' }, [
                    h('div', { class: 'session-title' }, session.title || '新会话'),
                    session.preview && h('div', { class: 'session-preview' }, session.preview),
                    session.time && h('div', { class: 'session-time' }, session.time)
                ]),
                !this.collapsed && (isHovered || isActive) && h('button', {
                    class: 'session-delete',
                    onClick: (e) => this.handleDeleteSession(session, e),
                    title: '删除会话'
                }, '🗑')
            ]);
        },
        
        renderSessionList() {
            if (this.filteredSessions.length === 0) {
                return h('div', { class: 'empty-sessions' }, 
                    this.collapsed ? null : '暂无会话'
                );
            }
            
            return h('div', { class: 'session-list' }, 
                this.filteredSessions.map(s => this.renderSessionItem(s))
            );
        }
    },
    
    render() {
        return h('aside', {
            class: ['sidebar', this.collapsed && 'collapsed']
        }, [
            this.renderHeader(),
            this.renderNewButton(),
            this.renderSearch(),
            this.renderSessionList()
        ]);
    }
});

export default Sidebar;