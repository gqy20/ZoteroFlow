// ZoteroFlow Web应用主要JavaScript文件

// PDF查看器状态管理
let currentPDF = null;
let currentPage = 1;
let totalPages = 1;
let currentZoom = 1.0;
let pdfDoc = null;

// 应用状态
let isLoading = false;

// 主要功能函数
async function askQuestion() {
    if (isLoading) return;

    const queryInput = document.getElementById('queryInput');
    const askBtn = document.getElementById('askBtn');
    const btnText = document.getElementById('btnText');
    const loading = document.getElementById('loading');
    const resultSection = document.getElementById('resultSection');
    const resultContent = document.getElementById('resultContent');

    const query = queryInput.value.trim();
    if (!query) {
        showMessage('请输入问题', 'error');
        return;
    }

    // 显示加载状态
    isLoading = true;
    askBtn.disabled = true;
    btnText.textContent = '思考中';
    loading.style.display = 'block';

    try {
        const response = await fetch('/api/ask', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ query: query })
        });

        if (!response.ok) {
            throw new Error(`请求失败: ${response.status}`);
        }

        const data = await response.json();

        // 显示结果
        resultContent.innerHTML = formatAnswer(data.answer);
        resultSection.style.display = 'block';

        // 检查是否有PDF文件可以查看
        if (data.pdfUrl) {
            currentPDF = data.pdfUrl;
            document.getElementById('pdfViewBtn').style.display = 'inline-block';
        } else {
            document.getElementById('pdfViewBtn').style.display = 'none';
        }

        resultSection.scrollIntoView({ behavior: 'smooth' });

    } catch (error) {
        console.error('Error:', error);
        resultContent.innerHTML = `<div class="error-message">
            <strong>请求失败:</strong> ${error.message}<br>
            <small>请检查网络连接或稍后重试</small>
        </div>`;
        resultSection.style.display = 'block';
    } finally {
        // 恢复按钮状态
        isLoading = false;
        askBtn.disabled = false;
        btnText.textContent = '提问';
        loading.style.display = 'none';
    }
}

// 格式化答案文本
function formatAnswer(answer) {
    // 简单的Markdown到HTML转换
    return answer
        .replace(/\n\n/g, '</p><p>')
        .replace(/\n/g, '<br>')
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/^- (.*)/gm, '<li>$1</li>')
        .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
        .replace(/^<p>/, '<p>')
        .replace(/<\/p>$/, '</p>')
        .replace(/^/, '<p>')
        .replace(/$/, '</p>');
}

// 复制结果功能
function copyResult() {
    const resultContent = document.getElementById('resultContent');
    const textContent = resultContent.textContent || resultContent.innerText;

    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(textContent).then(() => {
            showMessage('已复制到剪贴板', 'success');
        }).catch(() => {
            fallbackCopyText(textContent);
        });
    } else {
        fallbackCopyText(textContent);
    }
}

// 备用复制方法
function fallbackCopyText(text) {
    try {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        const successful = document.execCommand('copy');
        document.body.removeChild(textArea);

        if (successful) {
            showMessage('已复制到剪贴板', 'success');
        } else {
            showMessage('复制失败，请手动选择文本复制', 'error');
        }
    } catch (err) {
        showMessage('复制失败，请手动选择文本复制', 'error');
    }
}

// PDF查看器功能
function togglePDFViewer() {
    const pdfViewerSection = document.getElementById('pdfViewerSection');
    const isVisible = pdfViewerSection.style.display !== 'none';

    if (isVisible) {
        closePDFViewer();
    } else {
        openPDFViewer();
    }
}

function openPDFViewer() {
    if (!currentPDF) return;

    const pdfViewerSection = document.getElementById('pdfViewerSection');
    const nativeViewer = document.getElementById('nativePDFViewer');
    const pdfjsViewer = document.getElementById('pdfjsViewer');

    pdfViewerSection.style.display = 'block';

    // 首先尝试使用原生PDF查看器
    tryNativePDFViewer(currentPDF);
}

function tryNativePDFViewer(pdfUrl) {
    const nativeViewer = document.getElementById('nativePDFViewer');
    const pdfjsViewer = document.getElementById('pdfjsViewer');

    // 设置iframe加载PDF
    nativeViewer.src = pdfUrl;

    // 检测iframe是否成功加载PDF
    nativeViewer.onload = function() {
        try {
            // 检查iframe内容是否存在PDF查看器
            const iframeDoc = nativeViewer.contentDocument;
            if (iframeDoc && iframeDoc.body.children.length > 0) {
                // 原生查看器可用
                nativeViewer.style.display = 'block';
                pdfjsViewer.style.display = 'none';
                console.log('使用浏览器原生PDF查看器');
            } else {
                throw new Error('原生PDF查看器不可用');
            }
        } catch (e) {
            // 降级到PDF.js
            console.log('原生PDF查看器不可用，使用PDF.js');
            nativeViewer.style.display = 'none';
            initPDFJSViewer(pdfUrl);
        }
    };

    nativeViewer.onerror = function() {
        // 加载失败，降级到PDF.js
        console.log('原生PDF查看器加载失败，使用PDF.js');
        nativeViewer.style.display = 'none';
        initPDFJSViewer(pdfUrl);
    };
}

function initPDFJSViewer(pdfUrl) {
    const pdfjsViewer = document.getElementById('pdfjsViewer');
    const pdfjsScript = document.getElementById('pdfjsScript');

    // 动态加载PDF.js
    if (typeof pdfjsLib === 'undefined') {
        pdfjsScript.style.display = 'block';
        pdfjsScript.onload = function() {
            loadPDFWithJS(pdfUrl);
        };
        pdfjsScript.onerror = function() {
            showError('PDF.js加载失败，无法显示PDF文件');
        };
    } else {
        loadPDFWithJS(pdfUrl);
    }

    pdfjsViewer.style.display = 'block';
}

function loadPDFWithJS(pdfUrl) {
    if (typeof pdfjsLib === 'undefined') {
        console.error('PDF.js未加载');
        showError('PDF.js加载失败，请刷新页面重试');
        return;
    }

    // 设置PDF.js worker路径
    pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';

    // 显示加载状态
    showPDFLoading(true);

    // 加载PDF文档
    pdfjsLib.getDocument(pdfUrl).promise.then(function(pdf) {
        pdfDoc = pdf;
        totalPages = pdf.numPages;
        currentPage = 1;
        currentZoom = 1.0;

        updatePageInfo();
        renderPage(currentPage);
        showPDFLoading(false);
    }).catch(function(error) {
        console.error('加载PDF失败:', error);
        showPDFLoading(false);
        showError('PDF加载失败: ' + error.message);
    });
}

function renderPage(pageNumber) {
    if (!pdfDoc) return;

    pdfDoc.getPage(pageNumber).then(function(page) {
        const canvas = document.getElementById('pdfCanvas');
        const context = canvas.getContext('2d');

        // 计算渲染尺寸
        const viewport = page.getViewport({ scale: currentZoom });
        canvas.height = viewport.height;
        canvas.width = viewport.width;

        // 渲染PDF页面
        const renderContext = {
            canvasContext: context,
            viewport: viewport
        };

        page.render(renderContext);
    });

    updatePageInfo();
}

function updatePageInfo() {
    document.querySelector('.page-info').textContent = `${currentPage} / ${totalPages}`;
    document.querySelector('.zoom-level').textContent = Math.round(currentZoom * 100) + '%';
}

function prevPage() {
    if (currentPage <= 1) return;
    currentPage--;
    renderPage(currentPage);
}

function nextPage() {
    if (currentPage >= totalPages) return;
    currentPage++;
    renderPage(currentPage);
}

function zoomIn() {
    if (currentZoom >= 3.0) return;
    currentZoom += 0.25;
    renderPage(currentPage);
}

function zoomOut() {
    if (currentZoom <= 0.5) return;
    currentZoom -= 0.25;
    renderPage(currentPage);
}

function closePDFViewer() {
    document.getElementById('pdfViewerSection').style.display = 'none';
    document.getElementById('nativePDFViewer').src = '';

    // 重置PDF.js状态
    pdfDoc = null;
    currentPage = 1;
    currentZoom = 1.0;

    const canvas = document.getElementById('pdfCanvas');
    const context = canvas.getContext('2d');
    context.clearRect(0, 0, canvas.width, canvas.height);
}

// 辅助函数
function showPDFLoading(show) {
    const canvas = document.getElementById('pdfCanvas');
    if (show) {
        canvas.style.opacity = '0.5';
        canvas.innerHTML = '加载PDF中...';
    } else {
        canvas.style.opacity = '1';
    }
}

function showError(message) {
    showMessage(message, 'error');
}

function showMessage(message, type = 'info') {
    const messageDiv = document.createElement('div');
    messageDiv.className = `${type}-message`;
    messageDiv.textContent = message;
    messageDiv.style.position = 'fixed';
    messageDiv.style.top = '20px';
    messageDiv.style.right = '20px';
    messageDiv.style.zIndex = '1000';
    messageDiv.style.maxWidth = '300px';

    document.body.appendChild(messageDiv);

    setTimeout(() => {
        if (messageDiv.parentNode) {
            messageDiv.parentNode.removeChild(messageDiv);
        }
    }, 3000);
}

// 显示帮助信息
function showHelp() {
    const helpText = `
📚 ZoteroFlow 使用帮助

🔍 功能介绍：
• 文献搜索：搜索相关学术论文
• PDF查看：在线浏览PDF文献
• 智能分析：AI分析文献内容
• 学术问答：专业学术问题解答

💡 使用示例：
• "搜索机器学习论文"
• "查看Attention Is All You Need"
• "分析这篇论文的贡献"
• "什么是Transformer模型？"

🔧 快捷键：
• Enter：发送问题
• Shift+Enter：换行
• Ctrl+C：复制回答

📄 PDF查看：
• 支持缩放和翻页
• 自动适配屏幕大小
• 支持文本选择

有问题请查看系统状态或联系开发者。
    `;

    alert(helpText);
}

// 事件监听器
document.addEventListener('DOMContentLoaded', function() {
    // 支持按Enter发送，但不支持Shift+Enter换行
    const queryInput = document.getElementById('queryInput');
    queryInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            askQuestion();
        }
    });

    // 自动调整文本框高度
    queryInput.addEventListener('input', function() {
        this.style.height = 'auto';
        this.style.height = Math.min(this.scrollHeight, 200) + 'px';
    });

    // 页面加载完成提示
    console.log('ZoteroFlow Web应用已加载完成');
    console.log('支持的PDF查看模式：浏览器原生 + PDF.js降级');
});

// 错误处理
window.addEventListener('error', function(e) {
    console.error('页面错误:', e.error);
});

window.addEventListener('unhandledrejection', function(e) {
    console.error('未处理的Promise拒绝:', e.reason);
});