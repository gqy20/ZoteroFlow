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
	fmt.Println("=== Article MCP 集成测试 ===")

	// 测试1: 直接测试article-mcp
	fmt.Println("\n🧪 测试1: 直接测试 article-mcp 服务器")
	if err := testDirectArticleMCP(); err != nil {
		fmt.Printf("❌ 直接测试失败: %v\n", err)
	} else {
		fmt.Println("✅ article-mcp 直接测试通过")
	}

	// 测试2: 测试外部MCP配置是否被识别
	fmt.Println("\n🧪 测试2: 检查外部MCP配置")
	if err := testExternalMCPConfig(); err != nil {
		fmt.Printf("❌ 外部MCP配置测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 外部MCP配置正常")
	}

	fmt.Println("\n🎉 Article MCP 集成测试完成！")
}

func testDirectArticleMCP() error {
	fmt.Println("启动 article-mcp 服务器...")

	// 启动article-mcp服务器
	cmd := exec.Command("uvx", "article-mcp", "server")

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建stdin管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %w", err)
	}

	// 启动服务器
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动article-mcp服务器失败: %w", err)
	}

	defer cmd.Process.Kill()

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 创建扫描器
	scanner := bufio.NewScanner(stdout)

	// 测试初始化
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	if err := sendRequest(stdin, initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	// 读取响应
	response := readResponse(scanner)
	if response == nil {
		return fmt.Errorf("未收到初始化响应")
	}

	if response.Error != nil {
		return fmt.Errorf("初始化失败: %s", response.Error.Message)
	}

	fmt.Println("✅ article-mcp 初始化成功")

	// 测试获取工具列表
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := sendRequest(stdin, toolsReq); err != nil {
		return fmt.Errorf("发送工具列表请求失败: %w", err)
	}

	response = readResponse(scanner)
	if response == nil {
		return fmt.Errorf("未收到工具列表响应")
	}

	if response.Error != nil {
		return fmt.Errorf("获取工具列表失败: %s", response.Error.Message)
	}

	fmt.Println("✅ article-mcp 工具列表获取成功")

	// 解析工具列表
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if tools, ok := result["tools"].([]interface{}); ok {
			fmt.Printf("📋 Article MCP 提供 %d 个工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						fmt.Printf("  %d. %s\n", i+1, name)
					}
				}
			}
		}
	}

	// 测试一个具体的工具调用
	fmt.Println("🧪 测试 Europe PMC 搜索...")
	searchReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "search_europe_pmc",
			"arguments": map[string]interface{}{
				"keyword":     "machine learning",
				"max_results": 3,
			},
		},
	}

	if err := sendRequest(stdin, searchReq); err != nil {
		return fmt.Errorf("发送搜索请求失败: %w", err)
	}

	response = readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil {
		return fmt.Errorf("未收到搜索响应")
	}

	if response.Error != nil {
		return fmt.Errorf("搜索失败: %s", response.Error.Message)
	}

	fmt.Println("✅ Europe PMC 搜索成功")

	// 解析搜索结果
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					// 只显示前200个字符
					if len(text) > 200 {
						text = text[:200] + "..."
					}
					fmt.Printf("📄 搜索结果预览: %s\n", text)
				}
			}
		}
	}

	return nil
}

func testExternalMCPConfig() error {
	// 读取外部MCP配置文件
	configFile := "server/external-mcp-servers.json"
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
		return fmt.Errorf("配置文件格式错误: 缺少 external_mcp_servers")
	}

	articleMCP, ok := externalServers["article_mcp"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置文件中未找到 article_mcp 配置")
	}

	enabled, ok := articleMCP["enabled"].(bool)
	if !ok || !enabled {
		return fmt.Errorf("article_mcp 未启用")
	}

	command, ok := articleMCP["command"].(string)
	if !ok || command != "uvx" {
		return fmt.Errorf("article_mcp 命令配置错误")
	}

	args, ok := articleMCP["args"].([]interface{})
	if !ok || len(args) == 0 {
		return fmt.Errorf("article_mcp 参数配置错误")
	}

	fmt.Printf("✅ 外部MCP配置验证通过:\n")
	fmt.Printf("  - 命令: %s\n", command)
	fmt.Printf("  - 参数: %v\n", args)
	fmt.Printf("  - 超时: %v 秒\n", articleMCP["timeout"])
	fmt.Printf("  - 自动启动: %v\n", articleMCP["auto_start"])

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
	// 创建一个channel用于接收结果
	resultChan := make(chan *MCPResponse, 1)

	// 启动goroutine读取响应
	go func() {
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				resultChan <- nil
				return
			}

			var response MCPResponse
			if err := json.Unmarshal([]byte(line), &response); err != nil {
				fmt.Printf("⚠️  解析响应失败: %v\n", err)
				fmt.Printf("   原始响应: %s\n", line)
				resultChan <- nil
				return
			}

			resultChan <- &response
		} else {
			resultChan <- nil
		}
	}()

	// 等待结果或超时
	select {
	case response := <-resultChan:
		return response
	case <-time.After(timeout):
		fmt.Println("⚠️  读取响应超时")
		return nil
	}
}