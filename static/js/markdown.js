/* ============================================
   markdown.js - Markdown 渲染和代码高亮模块
   ============================================ */

/**
 * Markdown 渲染器
 * 使用 marked.js 解析 Markdown
 * 使用 highlight.js 实现代码高亮
 */

// 检查依赖是否加载
let marked = null;
let hljs = null;
let DOMPurify = null;

/**
 * 初始化 Markdown 渲染器
 */
function initMarkdown() {
    // 获取全局对象
    if (typeof window !== 'undefined') {
        marked = window.marked;
        hljs = window.hljs;
        DOMPurify = window.DOMPurify;
    }
    
    if (!marked) {
        console.warn('marked.js 未加载，Markdown 渲染将不可用');
        return false;
    }
    
    // 配置 marked
    marked.setOptions({
        breaks: true,        // 支持 GitHub 风格的换行
        gfm: true,           // 启用 GitHub Flavored Markdown
        headerIds: false,    // 禁用 header ids（安全性考虑）
        mangle: false,       // 禁用邮箱混淆
        highlight: function(code, lang) {
            // 代码高亮
            if (hljs) {
                if (lang && hljs.getLanguage(lang)) {
                    try {
                        return hljs.highlight(code, { language: lang }).value;
                    } catch (e) {
                        console.error('代码高亮失败:', e);
                    }
                }
                // 自动检测语言
                try {
                    return hljs.highlightAuto(code).value;
                } catch (e) {
                    console.error('代码自动检测失败:', e);
                }
            }
            return code;
        }
    });
    
    return true;
}

/**
 * 渲染 Markdown 文本
 * @param {string} text - Markdown 文本
 * @returns {string} - 渲染后的 HTML
 */
export function renderMarkdown(text) {
    if (!text) return '';
    
    // 确保 marked 已初始化
    if (!marked) {
        initMarkdown();
    }
    
    if (!marked) {
        // 降级处理：直接返回转义后的文本
        return escapeHtml(text).replace(/\n/g, '<br>');
    }
    
    try {
        let html = marked.parse(text);
        
        // XSS 防护（如果 DOMPurify 可用）
        if (DOMPurify) {
            html = DOMPurify.sanitize(html, {
                ALLOWED_TAGS: [
                    'p', 'br', 'strong', 'em', 'u', 's', 'del', 'ins',
                    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
                    'ul', 'ol', 'li', 'dl', 'dt', 'dd',
                    'blockquote', 'pre', 'code', 'kbd', 'samp',
                    'a', 'img', 'table', 'thead', 'tbody', 'tr', 'th', 'td',
                    'hr', 'div', 'span', 'sup', 'sub', 'mark'
                ],
                ALLOWED_ATTR: [
                    'href', 'src', 'alt', 'title', 'class', 'id',
                    'target', 'rel', 'width', 'height'
                ]
            });
        }
        
        return html;
    } catch (e) {
        console.error('Markdown 渲染失败:', e);
        return escapeHtml(text).replace(/\n/g, '<br>');
    }
}

/**
 * HTML 转义
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * 为代码块添加复制按钮
 * @param {HTMLElement} container - 包含代码块的容器
 */
export function addCopyButtons(container) {
    if (!container) return;
    
    const codeBlocks = container.querySelectorAll('pre code');
    
    codeBlocks.forEach((codeBlock) => {
        const pre = codeBlock.parentElement;
        
        // 避免重复添加
        if (pre.querySelector('.code-copy-btn')) return;
        
        // 创建包装器
        const wrapper = document.createElement('div');
        wrapper.className = 'code-block-wrapper';
        pre.parentNode.insertBefore(wrapper, pre);
        wrapper.appendChild(pre);
        
        // 添加语言标签
        const langClass = codeBlock.className.match(/language-(\w+)/);
        if (langClass) {
            const langLabel = document.createElement('span');
            langLabel.className = 'code-language-label';
            langLabel.textContent = langClass[1].toUpperCase();
            wrapper.appendChild(langLabel);
        }
        
        // 创建复制按钮
        const copyBtn = document.createElement('button');
        copyBtn.className = 'code-copy-btn';
        copyBtn.textContent = '复制';
        copyBtn.type = 'button';
        
        copyBtn.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(codeBlock.textContent);
                copyBtn.textContent = '已复制!';
                copyBtn.classList.add('copied');
                
                setTimeout(() => {
                    copyBtn.textContent = '复制';
                    copyBtn.classList.remove('copied');
                }, 2000);
            } catch (e) {
                console.error('复制失败:', e);
                copyBtn.textContent = '复制失败';
                setTimeout(() => {
                    copyBtn.textContent = '复制';
                }, 2000);
            }
        });
        
        wrapper.appendChild(copyBtn);
    });
}

/**
 * 处理消息中的链接（新窗口打开）
 * @param {HTMLElement} container - 消息容器
 */
export function processLinks(container) {
    if (!container) return;
    
    const links = container.querySelectorAll('a');
    links.forEach(link => {
        link.target = '_blank';
        link.rel = 'noopener noreferrer';
    });
}

/**
 * 完整的消息渲染
 * @param {string} content - 消息内容
 * @returns {string} - 渲染后的 HTML
 */
export function renderMessage(content) {
    const html = renderMarkdown(content);
    return `<div class="message-content-inner">${html}</div>`;
}

// 自动初始化
if (typeof window !== 'undefined') {
    // 等待 DOM 和依赖加载
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initMarkdown);
    } else {
        initMarkdown();
    }
}

export default {
    renderMarkdown,
    addCopyButtons,
    processLinks,
    renderMessage,
    initMarkdown
};