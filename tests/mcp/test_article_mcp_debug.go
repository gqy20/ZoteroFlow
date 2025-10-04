package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	fmt.Println("=== Article MCP 详细调试测试 ===")

	// 测试article-mcp
	if err := testArticleMCPDetailed(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
	} else {
		fmt.Println("✅ article-mcp 测试成功")
	}
}

func testArticleMCPDetailed() error {
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

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %w", err)
	}

	// 启动服务器
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动article-mcp服务器失败: %w", err)
	}

	defer cmd.Process.Kill()

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 启动stderr读取器
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[STDERR] %s\n", scanner.Text())
		}
	}()

	// 创建stdout扫描器
	scanner := bufio.NewScanner(stdout)

	// 测试初始化
	fmt.Println("🧪 发送初始化请求...")
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	if err := sendRequest(stdin, initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	// 读取初始化响应
	response := readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil {
		return fmt.Errorf("未收到初始化响应")
	}

	if response.Error != nil {
		return fmt.Errorf("初始化失败: %s", response.Error.Message)
	}

	fmt.Println("✅ 初始化成功")

	// 解析初始化响应查看服务器能力
	var initResult map[string]interface{}
	if err := json.Unmarshal(response.Result, &initResult); err == nil {
		if capabilities, ok := initResult["capabilities"].(map[string]interface{}); ok {
			fmt.Printf("🔧 服务器能力: %v\n", capabilities)
		}
		if serverInfo, ok := initResult["serverInfo"].(map[string]interface{}); ok {
			fmt.Printf("ℹ️  服务器信息: %v\n", serverInfo)
		}
	}

	// 发送初始化完成通知（某些MCP服务器需要这个）
	fmt.Println("🧪 发送初始化完成通知...")
	initializedNotif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	if err := sendRequest(stdin, initializedNotif); err != nil {
		return fmt.Errorf("发送初始化完成通知失败: %w", err)
	}

	// 等待一下确保服务器处理完成
	time.Sleep(1 * time.Second)

	// 测试获取工具列表
	fmt.Println("🧪 发送工具列表请求...")
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	if err := sendRequest(stdin, toolsReq); err != nil {
		return fmt.Errorf("发送工具列表请求失败: %w", err)
	}

	response = readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil {
		return fmt.Errorf("未收到工具列表响应")
	}

	if response.Error != nil {
		fmt.Printf("❌ 工具列表获取失败: %s\n", response.Error.Message)
		if response.Error.Data != "" {
			fmt.Printf("   详细信息: %s\n", response.Error.Data)
		}
		return fmt.Errorf("工具列表错误: %s", response.Error.Message)
	}

	fmt.Println("✅ 工具列表获取成功")

	// 解析工具列表
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if tools, ok := result["tools"].([]interface{}); ok {
			fmt.Printf("📋 Article MCP 提供 %d 个工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					name, _ := toolMap["name"].(string)
					desc, _ := toolMap["description"].(string)
					fmt.Printf("  %d. %s - %s\n", i+1, name, desc)
				}
			}

			// 如果有工具，测试第一个工具
			if len(tools) > 0 {
				return testFirstTool(stdin, scanner, tools[0])
			}
		}
	}

	return nil
}

func testFirstTool(stdin io.WriteCloser, scanner *bufio.Scanner, tool interface{}) error {
	toolMap, ok := tool.(map[string]interface{})
	if !ok {
		return fmt.Errorf("工具格式错误")
	}

	toolName, ok := toolMap["name"].(string)
	if !ok {
		return fmt.Errorf("工具名称缺失")
	}

	fmt.Printf("\n🧪 测试工具: %s\n", toolName)

	// 构造工具调用参数
	var args map[string]interface{}

	// 根据工具类型构造不同的参数
	switch toolName {
	case "search_europe_pmc":
		args = map[string]interface{}{
			"keyword":     "machine learning",
			"max_results": 3,
		}
	case "search_arxiv_papers":
		args = map[string]interface{}{
			"keyword":     "artificial intelligence",
			"max_results": 3,
		}
	default:
		fmt.Printf("⚠️  跳过未知工具: %s\n", toolName)
		return nil
	}

	toolReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	if err := sendRequest(stdin, toolReq); err != nil {
		return fmt.Errorf("发送工具调用请求失败: %w", err)
	}

	response := readResponseWithTimeout(scanner, 30*time.Second)
	if response == nil {
		return fmt.Errorf("未收到工具调用响应")
	}

	if response.Error != nil {
		fmt.Printf("❌ 工具调用失败: %s\n", response.Error.Message)
		if response.Error.Data != "" {
			fmt.Printf("   详细信息: %s\n", response.Error.Data)
		}
		return fmt.Errorf("工具调用错误: %s", response.Error.Message)
	}

	fmt.Printf("✅ 工具 %s 调用成功\n", toolName)

	// 解析工具调用结果
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err == nil {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					// 只显示前300个字符
					if len(text) > 300 {
						text = text[:300] + "..."
					}
					fmt.Printf("📄 调用结果预览: %s\n", text)
				}
			}
		}
	}

	return nil
}

func sendRequest(stdin io.WriteCloser, req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// 打印请求内容
	fmt.Printf("[SEND] %s\n", string(data))

	_, err = fmt.Fprintln(stdin, string(data))
	return err
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

			// 打印响应内容
			fmt.Printf("[RECV] %s\n", line)

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