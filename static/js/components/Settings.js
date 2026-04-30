/* ============================================
   Settings.js - 设置面板组件
   ============================================ */

const { defineComponent, h, ref, computed } = Vue;
import { ThemeManager } from '../utils.js';
import Toast from './Toast.js';

/**
 * Settings 组件 - 设置面板
 * 
 * Props:
 *   - visible: Boolean 是否显示
 * 
 * Events:
 *   - close: 关闭设置面板
 *   - theme-change: 主题变化
 */
export const Settings = defineComponent({
    name: 'Settings',
    
    props: {
        visible: {
            type: Boolean,
            default: false
        }
    },
    
    data() {
        return {
            theme: ThemeManager.current,
            typewriterEffect: true,
            showTimestamp: true,
            soundEnabled: false
        };
    },
    
    computed: {
        themeOptions() {
            return [
                { value: 'light', label: '浅色' },
                { value: 'dark', label: '深色' },
                { value: 'system', label: '跟随系统' }
            ];
        }
    },
    
    methods: {
        handleClose() {
            this.$emit('close');
        },
        
        handleOverlayClick(e) {
            if (e.target === e.currentTarget) {
                this.handleClose();
            }
        },
        
        handleThemeChange(theme) {
            this.theme = theme;
            ThemeManager.set(theme);
            this.$emit('theme-change', theme);
            Toast.success(`已切换到${theme === 'system' ? '跟随系统' : (theme === 'dark' ? '深色' : '浅色')}主题`);
        },
        
        handleTypewriterChange(e) {
            this.typewriterEffect = e.target.checked;
            localStorage.setItem('setting-typewriter', this.typewriterEffect);
        },
        
        handleTimestampChange(e) {
            this.showTimestamp = e.target.checked;
            localStorage.setItem('setting-timestamp', this.showTimestamp);
        },
        
        handleSoundChange(e) {
            this.soundEnabled = e.target.checked;
            localStorage.setItem('setting-sound', this.soundEnabled);
        },
        
        loadSettings() {
            this.typewriterEffect = localStorage.getItem('setting-typewriter') !== 'false';
            this.showTimestamp = localStorage.getItem('setting-timestamp') !== 'false';
            this.soundEnabled = localStorage.getItem('setting-sound') === 'true';
        },
        
        renderOverlay() {
            if (!this.visible) return null;
            
            return h('div', {
                class: 'settings-overlay',
                onClick: this.handleOverlayClick
            }, [
                h('div', { class: 'settings-modal animate-scaleIn' }, [
                    // 头部
                    h('div', { class: 'settings-header' }, [
                        h('h2', { class: 'settings-title' }, '设置'),
                        h('button', {
                            class: 'settings-close',
                            onClick: this.handleClose
                        }, '✕')
                    ]),
                    
                    // 内容
                    h('div', { class: 'settings-content' }, [
                        // 主题设置
                        this.renderSection('外观', [
                            this.renderThemeSelector()
                        ]),
                        
                        // 功能设置
                        this.renderSection('功能', [
                            this.renderToggle('打字机效果', this.typewriterEffect, this.handleTypewriterChange, '消息以打字机方式逐字显示'),
                            this.renderToggle('显示时间戳', this.showTimestamp, this.handleTimestampChange, '在消息下方显示发送时间'),
                            this.renderToggle('提示音', this.soundEnabled, this.handleSoundChange, '收到消息时播放提示音')
                        ]),
                        
                        // 关于
                        this.renderSection('关于', [
                            h('div', { class: 'about-info' }, [
                                h('p', { class: 'about-version' }, '版本: 2.0.0'),
                                h('p', { class: 'about-desc' }, '智能助手 - 基于 Ollama 的本地 AI 助手')
                            ])
                        ])
                    ])
                ])
            ]);
        },
        
        renderSection(title, children) {
            return h('div', { class: 'settings-section' }, [
                h('h3', { class: 'section-title' }, title),
                h('div', { class: 'section-content' }, children)
            ]);
        },
        
        renderThemeSelector() {
            return h('div', { class: 'theme-selector' }, [
                h('div', { class: 'theme-options' }, 
                    this.themeOptions.map(option => 
                        h('button', {
                            key: option.value,
                            class: ['theme-option', this.theme === option.value && 'active'],
                            onClick: () => this.handleThemeChange(option.value)
                        }, [
                            h('span', { class: 'theme-icon' }, 
                                option.value === 'light' ? '☀️' : 
                                option.value === 'dark' ? '🌙' : '💻'
                            ),
                            h('span', { class: 'theme-label' }, option.label)
                        ])
                    )
                )
            ]);
        },
        
        renderToggle(label, value, onChange, description) {
            return h('div', { class: 'setting-toggle' }, [
                h('div', { class: 'toggle-info' }, [
                    h('span', { class: 'toggle-label' }, label),
                    description && h('span', { class: 'toggle-desc' }, description)
                ]),
                h('label', { class: 'toggle-switch' }, [
                    h('input', {
                        type: 'checkbox',
                        checked: value,
                        onChange: onChange
                    }),
                    h('span', { class: 'toggle-slider' })
                ])
            ]);
        }
    },
    
    mounted() {
        this.loadSettings();
    },
    
    render() {
        return this.renderOverlay();
    }
});

export default Settings;