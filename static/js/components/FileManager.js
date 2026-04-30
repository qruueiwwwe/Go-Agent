/* ============================================
   FileManager.js - 文件管理组件（抽拉模式）
   ============================================ */

const { defineComponent, h, ref } = Vue;
import { formatFileSize, validateFileType } from '../utils.js';

/**
 * FileManager 组件 - 文件管理面板（抽拉模式）
 *
 * Props:
 *   - files: Array 文件列表
 *   - uploading: Boolean 是否上传中
 *
 * Events:
 *   - file-selected: 文件被选择上传
 *   - delete-file: 请求删除文件
 *   - analyze-file: 快速分析文件
 */
export const FileManager = defineComponent({
    name: 'FileManager',

    props: {
        files: {
            type: Array,
            default: () => [],
            validator: (arr) => Array.isArray(arr)
        },
        uploading: {
            type: Boolean,
            default: false
        }
    },

    data() {
        return {
            expanded: false,
            fileInputRef: ref(null)
        };
    },

    computed: {
        hasFiles() {
            return this.files && this.files.length > 0;
        },

        fileCount() {
            return this.files?.length || 0;
        }
    },

    methods: {
        toggle() {
            this.expanded = !this.expanded;
        },

        handleUploadClick() {
            this.$refs.fileInputRef?.click();
        },

        handleFileChange(e) {
            const files = e.target.files;
            if (!files || files.length === 0) return;

            for (let file of files) {
                if (!validateFileType(file.name)) {
                    this.$emit('error', `不支持的文件类型: ${file.name}`);
                    continue;
                }
                this.$emit('file-selected', file);
            }

            e.target.value = '';
        },

        handleDeleteFile(filename) {
            if (!confirm(`确定要删除文件 ${filename} 吗？`)) return;
            this.$emit('delete-file', filename);
        },

        handleQuickAnalyze(filename) {
            this.$emit('analyze-file', filename);
        },

        renderToggleBar() {
            return h('div', {
                class: ['file-toggle-bar', this.expanded && 'active'],
                onClick: this.toggle
            }, [
                h('span', { class: 'file-toggle-icon' }, '📎'),
                h('span', { class: 'file-toggle-text' }, '文件'),
                this.fileCount > 0 && h('span', { class: 'file-count-badge' }, this.fileCount),
                h('span', { class: 'file-toggle-arrow' }, this.expanded ? '▲' : '▼')
            ]);
        },

        renderFileItem(file) {
            const size = formatFileSize(file.size || 0);

            return h('div', { class: 'file-item', key: file.name }, [
                h('span', {
                    class: 'file-name',
                    onClick: () => this.handleQuickAnalyze(file.name),
                    title: `点击快速分析 ${file.name}`
                }, file.name),
                h('span', { class: 'file-size' }, size),
                h('button', {
                    class: 'file-item-delete',
                    onClick: () => this.handleDeleteFile(file.name),
                    title: `删除 ${file.name}`
                }, '🗑️')
            ]);
        },

        renderContent() {
            return h('div', { class: 'file-panel-content' }, [
                // 上传按钮
                h('div', { class: 'file-upload-area' }, [
                    h('div', { class: 'file-input-wrapper' }, [
                        h('button', {
                            class: 'file-upload-btn',
                            disabled: this.uploading,
                            onClick: this.handleUploadClick
                        }, '📤 上传文件'),

                        h('input', {
                            ref: 'fileInputRef',
                            type: 'file',
                            accept: '.txt,.md,.json,.go,.py,.js,.pdf',
                            multiple: false,
                            onChange: this.handleFileChange
                        })
                    ])
                ]),

                // 文件列表
                !this.hasFiles
                    ? h('div', { class: 'file-empty' }, '暂无文件')
                    : h('div', { class: 'file-list' },
                        this.files.map(file => this.renderFileItem(file))
                    ),

                // 上传进度
                this.uploading && h('div', { class: 'file-uploading' }, '上传中...')
            ]);
        }
    },

    render() {
        return h('div', { class: ['file-manager-drawer', this.expanded && 'expanded'] }, [
            // 抽拉按钮
            this.renderToggleBar(),

            // 展开内容
            h('div', { class: 'file-panel' }, this.renderContent())
        ]);
    }
});

export default FileManager;
