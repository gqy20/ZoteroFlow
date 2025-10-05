package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"zoteroflow2-server/cli"
	"zoteroflow2-server/config"
	"zoteroflow2-server/web"
)

const version = "2.1.0"

func main() {
	// 命令行参数解析
	var (
		webMode = flag.Bool("web", false, "启动Web服务模式")
		port    = flag.String("port", "9876", "Web服务端口 (默认: 9876)")
		help    = flag.Bool("help", false, "显示帮助信息")
		showVer = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *help {
		handler := cli.NewCommandHandler(nil)
		if err := handler.ShowHelp(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *showVer {
		showVersion()
		return
	}

	if *webMode {
		if err := startWebServer(*port); err != nil {
			log.Fatal("启动Web服务失败:", err)
		}
		return
	}

	// CLI模式
	if len(os.Args) > 1 {
		cfg := loadConfigWithCheck()
		if cfg == nil {
			log.Fatal("配置加载失败")
		}

		handler := cli.NewCommandHandler(cfg)
		if err := handler.HandleCommand(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	// 默认行为：显示帮助
	handler := cli.NewCommandHandler(nil)
	if err := handler.ShowHelp(); err != nil {
		log.Fatal(err)
	}
}

// loadConfigWithCheck 加载配置并进行错误检查
func loadConfigWithCheck() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("配置加载失败: %v", err)
		return nil
	}
	return cfg
}

// startWebServer 启动Web服务
func startWebServer(port string) error {
	// 设置路由
	router := web.SetupRouter()

	// 启动信息
	log.Printf("🚀 ZoteroFlow Web服务启动成功!")
	log.Printf("📱 访问地址: http://localhost:%s", port)
	log.Printf("💡 提示: 确保已配置好Zotero数据库和AI API")
	log.Printf("🔧 停止服务: Ctrl+C")

	// 启动服务器
	if err := router.Run(":" + port); err != nil {
		return fmt.Errorf("启动Web服务失败: %w", err)
	}
	return nil
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Printf("ZoteroFlow2 v%s\n", version)
	fmt.Printf("Go语言智能文献管理工具\n")
	fmt.Printf("构建时间: %s\n", "2025-10-05")
}
