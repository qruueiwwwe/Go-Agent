/* ============================================
   FileManager.js - 文件管理组件
   ============================================ */

const { defineComponent, h, ref } = Vue;
import { formatFileSize, validateFileType } from '../utils.js';

/**
 * FileManager 组件 - 文件管理面板
 * 
 * Props:
 *   - files: Array 文件列表
 *   - uploading: Boolean 是否上传中
 * 
 * Events:
 *   - file-selected: 文件被选择上传
 *   - delete-file: 请求删除文件
 *   - analyze-file: 快速分析文件
 * 
 * Features:
 *   - 拖放上传（预留）
 *   - 文件列表显示
 *   - 快速分析按钮
 *   - 删除确认
 *   - 上传进度显示
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
            fileInputRef: ref(null),
            uploadProgress: 0
        };
    },
    
    computed: {
        /**
         * 是否有文件
         */
        hasFiles() {
            return this.files && this.files.length > 0;
        },
        
        /**
         * 文件总大小
         */
        totalSize() {
            return (this.files || []).reduce((sum, file) => sum + (file.size || 0), 0);
        }
    },
    
    methods: {
        /**
         * 点击上传按钮
         */
        handleUploadClick() {
            if (!this.fileInputRef) return;
            this.$refs.fileInputRef?.click();
        },
        
        /**
         * 处理文件选择
         */
        handleFileChange(e) {
            const files = e.target.files;
            if (!files || files.length === 0) return;
            
            // 验证文件
            for (let file of files) {
                if (!validateFileType(file.name)) {
                    this.$emit('error', `不支持的文件类型: ${file.name}`);
                    continue;
                }
                
                // 发送文件选择事件
                this.$emit('file-selected', file);
            }
            
            // 重置文件输入
            e.target.value = '';
        },
        
        /**
         * 处理拖放上传（预留）
         */
        handleDragOver(e) {
            e.preventDefault();
            e.stopPropagation();
            // 可以在这里添加拖放视觉效果
        },
        
        handleDragLeave(e) {
            e.preventDefault();
            e.stopPropagation();
        },
        
        handleDrop(e) {
            e.preventDefault();
            e.stopPropagation();
            
            const files = e.dataTransfer?.files;
            if (!files) return;
            
            for (let file of files) {
                if (validateFileType(file.name)) {
                    this.$emit('file-selected', file);
                }
            }
        },
        
        /**
         * 删除文件
         */
        handleDeleteFile(filename) {
            if (!confirm(`确定要删除文件 ${filename} 吗？`)) {
                return;
            }
            
            this.$emit('delete-file', filename);
        },
        
        /**
         * 快速分析文件
         */
        handleQuickAnalyze(filename) {
            this.$emit('analyze-file', filename);
        },
        
        /**
         * 渲染单个文件项
         */
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
        
        /**
         * 渲染文件列表或空状态
         */
        renderFileList() {
            if (!this.hasFiles) {
                return h('div', {
                    style: {
                        color: '#999',
                        fontSize: '12px'
                    }
                }, '暂无文件');
            }
            
            return h('div', { class: 'file-list' },
                this.files.map(file => this.renderFileItem(file))
            );
        },
        
        /**
         * 渲染上传进度
         */
        renderUploadProgress() {
            if (!this.uploading) return null;
            
            return h('div', {
                style: {
                    fontSize: '12px',
                    color: '#667eea',
                    marginTop: '8px'
                }
            }, `上传中... ${this.uploadProgress}%`);
        }
    },
    
    render() {
        return h('div', { class: 'file-manager' }, [
            h('h3', '📁 文件管理'),
            
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
                        onChange: this.handleFileChange,
                        onDragover: this.handleDragOver,
                        onDragleave: this.handleDragLeave,
                        onDrop: this.handleDrop
                    })
                ])
            ]),
            
            this.renderFileList(),
            this.renderUploadProgress()
        ]);
    }
});

export default FileManager;
