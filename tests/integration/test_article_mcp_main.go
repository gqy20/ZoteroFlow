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
	fmt.Println("=== 主流程测试：Article MCP 工具 ===")

	// 测试1: 验证主MCP服务器
	fmt.Println("\n🧪 测试1: 主MCP服务器功能验证")
	if err := testMainServer(); err != nil {
		fmt.Printf("❌ 主服务器测试失败: %v\n", err)
		return
	}
	fmt.Println("✅ 主MCP服务器功能正常")

	// 测试2: 验证Article MCP服务器
	fmt.Println("\n🧪 测试2: Article MCP服务器功能验证")
	if err := testArticleMCPServer(); err != nil {
		fmt.Printf("❌ Article MCP测试失败: %v\n", err)
		return
	}
	fmt.Println("✅ Article MCP服务器功能正常")

	// 测试3: 模拟实际工作流
	fmt.Println("\n🧪 测试3: 模拟学术研究工作流")
	testWorkflow()

	fmt.Println("\n🎉 主流程测试完成!")
	fmt.Println("✅ 两个MCP服务器都可以独立正常工作")
	fmt.Println("✅ 准备用于实际的学术研究工作流")
}

func testMainServer() error {
	cmd := exec.Command("./server/bin/zoteroflow2", "mcp")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	defer cmd.Process.Kill()
	time.Sleep(2 * time.Second)

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
				"name":    "workflow-test",
				"version": "1.0.0",
			},
		},
	}

	sendRequest(stdin, initReq)
	response := readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("主服务器初始化失败")
	}

	// 获取工具列表
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	sendRequest(stdin, toolsReq)
	response = readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("获取主服务器工具列表失败")
	}

	var result map[string]interface{}
	json.Unmarshal(response.Result, &result)
	if tools, ok := result["tools"].([]interface{}); ok {
		fmt.Printf("   🛠️  主服务器工具: %d个\n", len(tools))
		for i, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if name, ok := toolMap["name"].(string); ok {
					if i < 3 {
						fmt.Printf("      - %s\n", name)
					}
				}
			}
		}
		if len(tools) > 3 {
			fmt.Printf("      ... 还有%d个工具\n", len(tools)-3)
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

	sendRequest(stdin, statsReq)
	response = readResponse(scanner)
	if response == nil || response.Error != nil {
		return fmt.Errorf("主服务器工具调用失败")
	}

	fmt.Printf("   ✅ 主服务器工具调用正常\n")
	return nil
}

func testArticleMCPServer() error {
	cmd := exec.Command("uvx", "article-mcp", "server")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
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
				"name":    "article-test",
				"version": "1.0.0",
			},
		},
	}

	sendRequest(stdin, initReq)
	response := readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil || response.Error != nil {
		return fmt.Errorf("Article MCP初始化失败")
	}

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

	sendRequest(stdin, toolsReq)
	response = readResponseWithTimeout(scanner, 10*time.Second)
	if response == nil || response.Error != nil {
		return fmt.Errorf("获取Article MCP工具列表失败")
	}

	var result map[string]interface{}
	json.Unmarshal(response.Result, &result)
	if tools, ok := result["tools"].([]interface{}); ok {
		fmt.Printf("   📚 Article MCP工具: %d个\n", len(tools))
		for i, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if name, ok := toolMap["name"].(string); ok {
					if i < 3 {
						fmt.Printf("      - %s\n", name)
					}
				}
			}
		}
		if len(tools) > 3 {
			fmt.Printf("      ... 还有%d个工具\n", len(tools)-3)
		}
	}

	// 测试Europe PMC搜索
	searchReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "search_europe_pmc",
			"arguments": map[string]interface{}{
				"keyword":     "machine learning",
				"max_results": 2,
			},
		},
	}

	sendRequest(stdin, searchReq)
	response = readResponseWithTimeout(scanner, 30*time.Second)
	if response == nil || response.Error != nil {
		return fmt.Errorf("Article MCP工具调用失败")
	}

	// 解析搜索结果
	json.Unmarshal(response.Result, &result)
	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if textContent, ok := content[0].(map[string]interface{}); ok {
			if text, ok := textContent["text"].(string); ok {
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				fmt.Printf("   📄 搜索结果预览: %s\n", text)
			}
		}
	}

	fmt.Printf("   ✅ Article MCP工具调用正常\n")
	return nil
}

func testWorkflow() {
	fmt.Println("   📋 模拟学术研究工作流程:")
	fmt.Println()
	fmt.Println("   💭 用户需求: '我想研究机器学习在医学影像中的应用'")
	fmt.Println("   🔄 推荐工作流程:")
	fmt.Println("      1️⃣  使用 zotero_search 搜索本地相关文献")
	fmt.Println("      2️⃣  使用 search_europe_pmc 搜索最新研究")
	fmt.Println("      3️⃣  使用 zotero_chat 基于文献进行AI分析")
	fmt.Println("      4️⃣  使用 mineru_parse 解析重要PDF文件")
	fmt.Println("      5️⃣  使用 get_article_details 获取详细信息")
	fmt.Println()
	fmt.Println("   🎯 当前状态:")
	fmt.Println("      ✅ 主MCP服务器: 提供本地文献管理和AI分析")
	fmt.Println("      ✅ Article MCP: 提供全球学术文献搜索")
	fmt.Println("      ⚠️  集成方式: 需要在MCP客户端中配置两个服务器")
	fmt.Println()
	fmt.Println("   🚀 使用建议:")
	fmt.Println("      1. 在Claude Desktop中同时配置两个MCP服务器")
	fmt.Println("      2. 让AI助手根据需求选择合适的工具")
	fmt.Println("      3. 实现本地+全球的完整学术研究工作流")
}

func sendRequest(stdin io.Writer, req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdin, "%s\n", string(data))
	return nil
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