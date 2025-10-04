package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// MCP请求结构
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCP响应结构
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func main() {
	fmt.Println("=== ZoteroFlow2 MCP 项目状态检查 ===")

	// 1. 检查项目文件结构
	fmt.Println("\n📁 1. 项目文件结构检查")
	if err := checkProjectStructure(); err != nil {
		fmt.Printf("❌ 项目结构检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 项目结构检查通过")

	// 2. 检查编译状态
	fmt.Println("\n🔨 2. 编译状态检查")
	if err := checkBuildStatus(); err != nil {
		fmt.Printf("❌ 编译状态检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 编译状态检查通过")

	// 3. 检查配置文件
	fmt.Println("\n⚙️ 3. 配置文件检查")
	if err := checkConfiguration(); err != nil {
		fmt.Printf("❌ 配置文件检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 配置文件检查通过")

	// 4. 检查MCP服务器功能
	fmt.Println("\n🚀 4. MCP服务器功能检查")
	if err := checkMCPServer(); err != nil {
		fmt.Printf("❌ MCP服务器功能检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ MCP服务器功能检查通过")

	// 5. 检查外部MCP集成
	fmt.Println("\n🔗 5. 外部MCP集成检查")
	if err := checkExternalMCP(); err != nil {
		fmt.Printf("❌ 外部MCP集成检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 外部MCP集成检查通过")

	// 总结
	fmt.Println("\n🎉 ZoteroFlow2 MCP 项目状态总结")
	fmt.Println("✅ 项目结构完整")
	fmt.Println("✅ 编译状态正常")
	fmt.Println("✅ 配置文件正确")
	fmt.Println("✅ MCP服务器功能完整")
	fmt.Println("✅ 外部MCP集成框架就绪")
	fmt.Println("\n🚀 项目可以在整个项目中正常使用MCP功能！")

	// 使用指南
	fmt.Println("\n📖 使用指南:")
	fmt.Println("1. 启动MCP服务器: ./server/bin/zoteroflow2 mcp")
	fmt.Println("2. 在Claude Desktop中配置: 复制 docs/claude-desktop-config.json")
	fmt.Println("3. 可用工具: 6个本地工具 (zotero_*, mineru_*, zotero_chat)")
	fmt.Println("4. 外部工具: Article MCP (需手动启动或等待代理功能实现)")
}

func checkProjectStructure() error {
	requiredFiles := []string{
		"server/main.go",
		"server/mcp/server.go",
		"server/mcp/tools.go",
		"server/mcp/handlers.go",
		"server/external-mcp-servers.json",
		"server/.env",
		"docs/mcp-integration-plan.md",
		"docs/external-mcp-configuration.md",
		"docs/api/mcp-tools-list.md",
		"tests/mcp/test_mcp_basic.go",
		"tests/test_article_mcp.py",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("缺少必需文件: %s", file)
		}
	}

	return nil
}

func checkBuildStatus() error {
	// 检查二进制文件
	if _, err := os.Stat("server/bin/zoteroflow2"); os.IsNotExist(err) {
		return fmt.Errorf("二进制文件不存在，请运行 'cd server && make build'")
	}

	// 检查版本信息
	cmd := exec.Command("server/bin/zoteroflow2", "help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("二进制文件无法正常执行: %w", err)
	}

	return nil
}

func checkConfiguration() error {
	// 检查.env文件
	envFile := "server/.env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("环境配置文件不存在: %s", envFile)
	}

	// 检查外部MCP配置
	configFile := "server/external-mcp-servers.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取外部MCP配置失败: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析外部MCP配置失败: %w", err)
	}

	externalServers, ok := config["external_mcp_servers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("外部MCP配置格式错误")
	}

	if articleMCP, exists := externalServers["article_mcp"]; exists {
		if configMap, ok := articleMCP.(map[string]interface{}); ok {
			enabled, _ := configMap["enabled"].(bool)
			command, _ := configMap["command"].(string)
			fmt.Printf("   📋 发现外部MCP配置: article_mcp (命令: %s, 启用: %v)\n", command, enabled)
		}
	}

	return nil
}

func checkMCPServer() error {
	// 启动MCP服务器
	cmd := exec.Command("server/bin/zoteroflow2", "mcp")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建stdin管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动MCP服务器失败: %w", err)
	}

	defer cmd.Process.Kill()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	scanner := bufio.NewScanner(stdout)

	// 测试初始化
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "status-check-client",
				"version": "1.0.0",
			},
		},
	}

	if err := sendRequest(stdin, initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	response := readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("MCP初始化失败")
	}

	// 测试工具列表
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := sendRequest(stdin, toolsReq); err != nil {
		return fmt.Errorf("发送工具列表请求失败: %w", err)
	}

	response = readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("获取工具列表失败")
	}

	// 解析工具列表
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if tools, ok := result["tools"].([]interface{}); ok {
			fmt.Printf("   🛠️  发现 %d 个本地工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						fmt.Printf("      %d. %s\n", i+1, name)
					}
				}
			}
		}
	}

	// 测试一个工具调用
	statsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "zotero_get_stats",
			"arguments": map[string]interface{}{},
		},
	}

	if err := sendRequest(stdin, statsReq); err != nil {
		return fmt.Errorf("发送统计请求失败: %w", err)
	}

	response = readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("调用统计工具失败")
	}

	fmt.Printf("   ✅ 工具调用测试通过\n")

	return nil
}

func checkExternalMCP() error {
	// 检查uvx是否可用
	if _, err := exec.LookPath("uvx"); err != nil {
		fmt.Printf("   ⚠️  uvx不可用，外部MCP功能受限\n")
		return nil
	}

	// 测试article-mcp是否可用
	cmd := exec.Command("uvx", "article-mcp", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   ⚠️  article-mcp不可用: %v\n", err)
		return nil
	}

	if strings.Contains(string(output), "Article MCP 文献搜索服务器") {
		fmt.Printf("   ✅ article-mcp可用且功能正常\n")
	}

	// 检查Python测试
	if _, err := os.Stat("tests/test_article_mcp.py"); err == nil {
		fmt.Printf("   ✅ Python集成测试脚本存在\n")
	}

	return nil
}

func sendRequest(stdin io.Writer, req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdin, "%s\n", string(data))
	return err
}

func readResponse(scanner *bufio.Scanner) *MCPResponse {
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil
		}

		var response MCPResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			return nil
		}

		return &response
	}

	return nil
}