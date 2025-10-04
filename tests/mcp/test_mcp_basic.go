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
	fmt.Println("=== ZoteroFlow2 MCP 服务器基础测试 ===")

	// 启动MCP服务器
	cmd := exec.Command("../server/bin/zoteroflow2", "mcp")

	// 创建管道用于stdin/stdout通信
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("❌ 创建stdin管道失败: %v\n", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ 创建stdout管道失败: %v\n", err)
		return
	}

	// 启动服务器
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ 启动MCP服务器失败: %v\n", err)
		return
	}

	defer cmd.Process.Kill()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	fmt.Println("✅ MCP服务器已启动")

	// 创建扫描器读取响应
	scanner := bufio.NewScanner(stdout)

	// 测试1: 初始化
	fmt.Println("\n🧪 测试1: 初始化请求")
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
		fmt.Printf("❌ 发送初始化请求失败: %v\n", err)
		return
	}

	// 读取响应
	if response := readResponse(scanner); response != nil {
		if response.Error != nil {
			fmt.Printf("❌ 初始化失败: %s\n", response.Error.Message)
		} else {
			fmt.Println("✅ 初始化成功")
		}
	} else {
		fmt.Println("❌ 未收到初始化响应")
	}

	// 测试2: 获取工具列表
	fmt.Println("\n🧪 测试2: 获取工具列表")
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := sendRequest(stdin, toolsReq); err != nil {
		fmt.Printf("❌ 发送工具列表请求失败: %v\n", err)
		return
	}

	if response := readResponse(scanner); response != nil {
		if response.Error != nil {
			fmt.Printf("❌ 获取工具列表失败: %s\n", response.Error.Message)
		} else {
			fmt.Println("✅ 工具列表获取成功")

			// 解析工具列表
			var result map[string]interface{}
			if err := json.Unmarshal(response.Result, &result); err == nil {
				if tools, ok := result["tools"].([]interface{}); ok {
					fmt.Printf("📋 发现 %d 个工具:\n", len(tools))
					for i, tool := range tools {
						if toolMap, ok := tool.(map[string]interface{}); ok {
							if name, ok := toolMap["name"].(string); ok {
								fmt.Printf("  %d. %s\n", i+1, name)
							}
						}
					}
				}
			}
		}
	} else {
		fmt.Println("❌ 未收到工具列表响应")
	}

	// 测试3: 调用统计工具
	fmt.Println("\n🧪 测试3: 调用zotero_get_stats工具")
	statsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "zotero_get_stats",
			"arguments": map[string]interface{}{},
		},
	}

	if err := sendRequest(stdin, statsReq); err != nil {
		fmt.Printf("❌ 发送统计请求失败: %v\n", err)
		return
	}

	if response := readResponse(scanner); response != nil {
		if response.Error != nil {
			fmt.Printf("❌ 获取统计信息失败: %s\n", response.Error.Message)
			if response.Error.Data != "" {
				fmt.Printf("   详细信息: %s\n", response.Error.Data)
			}
		} else {
			fmt.Println("✅ 统计信息获取成功")

			// 解析统计结果
			var result map[string]interface{}
			if err := json.Unmarshal(response.Result, &result); err == nil {
				if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
					if textContent, ok := content[0].(map[string]interface{}); ok {
						if text, ok := textContent["text"].(string); ok {
							fmt.Printf("📊 统计结果: %s\n", text)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("❌ 未收到统计响应")
	}

	fmt.Println("\n🎉 MCP服务器基础测试完成！")
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
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil
		}

		var response MCPResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			fmt.Printf("⚠️  解析响应失败: %v\n", err)
			fmt.Printf("   原始响应: %s\n", line)
			return nil
		}

		return &response
	}

	return nil
}