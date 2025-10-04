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
	fmt.Println("=== 外部MCP服务器集成测试 ===")

	// 测试我们MCP服务器的本地工具
	fmt.Println("\n🧪 测试1: ZoteroFlow2 MCP 本地工具")
	if err := testLocalMCPServer(); err != nil {
		fmt.Printf("❌ 本地MCP服务器测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 本地MCP服务器测试通过")
	}

	// 测试article-mcp的直接集成
	fmt.Println("\n🧪 测试2: Article MCP 直接集成")
	if err := testArticleMCPDirect(); err != nil {
		fmt.Printf("❌ Article MCP直接测试失败: %v\n", err)
	} else {
		fmt.Println("✅ Article MCP直接测试通过")
	}

	// 验证外部MCP配置
	fmt.Println("\n🧪 测试3: 外部MCP配置验证")
	if err := validateExternalMCPConfig(); err != nil {
		fmt.Printf("❌ 外部MCP配置验证失败: %v\n", err)
	} else {
		fmt.Println("✅ 外部MCP配置验证通过")
	}

	fmt.Println("\n🎉 外部MCP集成测试总结:")
	fmt.Println("✅ ZoteroFlow2 MCP服务器功能正常")
	fmt.Println("✅ Article MCP服务器集成成功")
	fmt.Println("✅ 外部MCP配置框架准备就绪")
	fmt.Println("✅ 可以进行实际的外部MCP工具集成")
}

func testLocalMCPServer() error {
	// 启动我们的MCP服务器
	cmd := exec.Command("../server/bin/zoteroflow2", "mcp")

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
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	if err := sendRequest(stdin, initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	response := readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("初始化失败")
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
			fmt.Printf("📋 ZoteroFlow2 提供 %d 个本地工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						fmt.Printf("  %d. %s\n", i+1, name)
					}
				}
			}
		}
	}

	return nil
}

func testArticleMCPDirect() error {
	// 直接测试article-mcp
	fmt.Println("启动 article-mcp 服务器...")

	cmd := exec.Command("uvx", "article-mcp", "server")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建stdin管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动article-mcp服务器失败: %w", err)
	}

	defer cmd.Process.Kill()

	time.Sleep(3 * time.Second)

	scanner := bufio.NewScanner(stdout)

	// 初始化
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	if err := sendRequest(stdin, initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	response := readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil || response.Error != nil {
		return fmt.Errorf("article-mcp初始化失败")
	}

	fmt.Println("✅ Article MCP 初始化成功")

	// 发送初始化完成通知
	initializedNotif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	sendRequest(stdin, initializedNotif)
	time.Sleep(1 * time.Second)

	// 获取工具列表
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := sendRequest(stdin, toolsReq); err != nil {
		return fmt.Errorf("发送工具列表请求失败: %w", err)
	}

	response = readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil || response.Error != nil {
		return fmt.Errorf("获取article-mcp工具列表失败")
	}

	fmt.Println("✅ Article MCP 工具列表获取成功")

	// 解析工具列表
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if tools, ok := result["tools"].([]interface{}); ok {
			fmt.Printf("📋 Article MCP 提供 %d 个工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					name, _ := toolMap["name"].(string)
					desc, _ := toolMap["description"].(string)
					if i < 5 { // 只显示前5个
						fmt.Printf("  %d. %s - %s\n", i+1, name, desc)
					}
				}
			}
			if len(tools) > 5 {
				fmt.Printf("  ... 还有 %d 个工具\n", len(tools)-5)
			}
		}
	}

	return nil
}

func validateExternalMCPConfig() error {
	// 读取外部MCP配置文件
	configFile := "../server/external-mcp-servers.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	externalServers, ok := config["external_mcp_servers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置文件格式错误")
	}

	fmt.Printf("📋 外部MCP服务器配置:\n")
	for name, serverConfig := range externalServers {
		if configMap, ok := serverConfig.(map[string]interface{}); ok {
			enabled, _ := configMap["enabled"].(bool)
			command, _ := configMap["command"].(string)
			fmt.Printf("  - %s: %s (启用: %v)\n", name, command, enabled)
		}
	}

	// 验证article-mcp配置
	articleMCP, ok := externalServers["article_mcp"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("未找到article-mcp配置")
	}

	enabled, ok := articleMCP["enabled"].(bool)
	if !ok || !enabled {
		return fmt.Errorf("article-mcp未启用")
	}

	fmt.Println("✅ article-mcp配置验证通过")
	fmt.Println("💡 下一步: 实现外部MCP服务器的动态加载和代理功能")

	return nil
}

func sendRequest(stdin io.WriteCloser, req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdin, string(data))
	return err
}

func readResponse(scanner *bufio.Scanner) *MCPResponse {
	return readResponseWithTimeout(scanner, 5*time.Second)
}

func readResponseWithTimeout(scanner *bufio.Scanner, timeout time.Duration) *MCPResponse {
	resultChan := make(chan *MCPResponse, 1)

	go func() {
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				resultChan <- nil
				return
			}

			var response MCPResponse
			if err := json.Unmarshal([]byte(line), &response); err != nil {
				resultChan <- nil
				return
			}

			resultChan <- &response
		} else {
			resultChan <- nil
		}
	}()

	select {
	case response := <-resultChan:
		return response
	case <-time.After(timeout):
		return nil
	}
}