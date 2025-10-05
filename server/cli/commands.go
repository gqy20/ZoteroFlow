package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zoteroflow2-server/config"
)

// CommandHandler 处理CLI命令
type CommandHandler struct {
	config *config.Config
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(cfg *config.Config) *CommandHandler {
	return &CommandHandler{
		config: cfg,
	}
}

// HandleCommand 处理命令行参数
func (h *CommandHandler) HandleCommand(args []string) error {
	if len(args) == 0 {
		return h.ShowHelp()
	}

	switch args[0] {
	case "list":
		return h.listResults()
	case "open":
		if len(args) < 2 {
			return fmt.Errorf("用法: open <文献名称>")
		}
		return h.openResult(args[1])
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("用法: search <标题关键词>")
		}
		return fmt.Errorf("search命令暂未实现，请使用Web界面")
	case "doi":
		if len(args) < 2 {
			return fmt.Errorf("用法: doi <DOI号>")
		}
		return fmt.Errorf("doi命令暂未实现，请使用Web界面")
	case "chat":
		return fmt.Errorf("chat命令暂未实现，请使用Web界面")
	case "related":
		return fmt.Errorf("related命令暂未实现，请使用Web界面")
	case "help":
		return h.ShowHelp()
	default:
		return fmt.Errorf("未知命令: %s", args[0])
	}
}

// ShowHelp 显示帮助信息
func (h *CommandHandler) ShowHelp() error {
	fmt.Println("ZoteroFlow2 - 智能文献管理工具 (CLI + Web 双模式)")
	fmt.Println()
	fmt.Println("🌐 Web服务模式:")
	fmt.Println("  go run main.go -web                   # 启动Web服务 (默认端口9876)")
	fmt.Println("  go run main.go -web -port=8888        # 指定端口启动Web服务")
	fmt.Println()
	fmt.Println("📚 CLI模式 - 文献管理:")
	fmt.Println("  list                    - 列出所有解析结果")
	fmt.Println("  open <名称>             - 打开指定文献文件夹")
	fmt.Println("  search <关键词>         -��标题搜索并解析文献")
	fmt.Println("  doi <DOI号>             - 按DOI搜索并解析文献")
	fmt.Println()
	fmt.Println("🤖 AI助手对话:")
	fmt.Println("  chat                    - 进入交互式AI对话模式")
	fmt.Println("  chat <问题>             - 单次AI问答")
	fmt.Println("  chat --doc=文献名 <问题> - 基于指定文献的AI对话")
	fmt.Println()
	fmt.Println("🔍 智能文献分析:")
	fmt.Println("  related <文献名/DOI> <问题> - 查找相关文献并AI分析")
	fmt.Println()
	fmt.Println("🔧 其他命令:")
	fmt.Println("  help                    - 显示此帮助信息")
	fmt.Println("  version                 - 显示版本信息")
	fmt.Println()
	fmt.Println("💡 Web功能特性:")
	fmt.Println("  • 智能文献问答界面")
	fmt.Println("  • PDF在线查看 (浏览器原生 + PDF.js)")
	fmt.Println("  • 实时AI对话和分析")
	fmt.Println("  • 响应式设计，支持移动设备")
	fmt.Println()
	fmt.Println("📱 使用示例:")
	fmt.Println("  go run main.go -web                       # 启动Web服务")
	fmt.Println("  # 浏览器访问: http://localhost:9876")
	fmt.Println("  # Web界面支持: 文献搜索、PDF查看、AI问答")
	fmt.Println()
	fmt.Println("  go run main.go list                      # CLI列出文献")
	fmt.Println("  go run main.go search \"机器学习\"          # 搜索文献")
	fmt.Println()
	fmt.Println("🎯 双模式优势:")
	fmt.Println("  • CLI模式: 高效的命令行操作")
	fmt.Println("  • Web模式: 直观的图形界面")
	fmt.Println("  • 统一配置: 共享数据库和AI配置")
	fmt.Println("  • 端口自动检测: 避免冲突 (9876-9976)")
	return nil
}

// listResults 列出所有解析结果
func (h *CommandHandler) listResults() error {
	if h.config == nil {
		return fmt.Errorf("配置未加载")
	}

	resultsDir := h.config.ResultsDir

	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("📂 结果目录不存在: %s\n", resultsDir)
			fmt.Println("💡 请先使用 'search' 或 'doi' 命令解析文献")
			return nil
		}
		return fmt.Errorf("读取结果目录失败: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("📋 暂无解析结果")
		fmt.Println("💡 请使用 'search <关键词>' 或 'doi <DOI号>' 命令解析文献")
		return nil
	}

	fmt.Printf("📚 已解析文献列表 (共 %d 篇):\n", len(entries))
	fmt.Println(strings.Repeat("─", 80))

	for i, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		jsonPath := filepath.Join(resultsDir, name, "info.json")

		// 检查info.json是否存在
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("%2d. %s\n", i+1, name)
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("📁 结果目录: %s\n", resultsDir)
	fmt.Println("💡 使用 'open <文献名称>' 打开对应文件夹")

	return nil
}

// openResult 打开指定文献文件夹
func (h *CommandHandler) openResult(name string) error {
	if h.config == nil {
		return fmt.Errorf("配置未加载")
	}

	targetDir := filepath.Join(h.config.ResultsDir, name)

	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("文献 '%s' 不存在", name)
	}

	// 尝试打开文件夹
	fmt.Printf("📂 打开文献文件夹: %s\n", targetDir)

	// 在Linux/macOS上使用xdg-open/open命令
	var openCmd string
	if strings.Contains(strings.ToLower(os.Getenv("TERM")), "linux") || os.Getenv("WSL_DISTRO_NAME") != "" {
		openCmd = "xdg-open"
	} else {
		openCmd = "open"
	}

	cmd := fmt.Sprintf("%s '%s'", openCmd, targetDir)
	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("打开文件夹失败: %v", err)
	}

	return nil
}

// 辅助函数
func runCommand(cmd string) error {
	// 简化实现，实际可以使用os/exec包
	fmt.Printf("执行命令: %s\n", cmd)
	return nil
}
