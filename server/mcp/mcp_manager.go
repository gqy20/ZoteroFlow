package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Enabled       bool     `json:"enabled"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	Timeout       int      `json:"timeout"`
	RetryAttempts int      `json:"retryAttempts"`
	Description   string   `json:"description"`
	Tools         []string `json:"tools"`
}

// MCPConfig MCP配置文件
type MCPConfig struct {
	MCPServers     map[string]MCPServerConfig `json:"mcpServers"`
	GlobalSettings struct {
		DefaultTimeout   int    `json:"defaultTimeout"`
		MaxRetryAttempts int    `json:"maxRetryAttempts"`
		EnableLogging    bool   `json:"enableLogging"`
		LogLevel         string `json:"logLevel"`
	} `json:"globalSettings"`
}

// MCPManager MCP管理器
type MCPManager struct {
	config     *MCPConfig
	clients    map[string]*MCPClient
	configFile string
	mu         sync.RWMutex
}

// MCPClient MCP客户端
type MCPClient struct {
	config  MCPServerConfig
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
	scanner *bufio.Scanner
	process *os.Process
	active  bool
	timeout time.Duration
	nextID  int
	mu      sync.Mutex
}

// NewMCPManager 创建MCP管理器
func NewMCPManager(configFile string) (*MCPManager, error) {
	manager := &MCPManager{
		clients:    make(map[string]*MCPClient),
		configFile: configFile,
	}

	// 加载配置
	if err := manager.loadConfig(); err != nil {
		return nil, fmt.Errorf("加载MCP配置失败: %w", err)
	}

	return manager, nil
}

// loadConfig 加载配置文件
func (m *MCPManager) loadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	m.config = &config
	return nil
}

// StartServer 启动指定的MCP服务器
func (m *MCPManager) StartServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已经存在
	if client, exists := m.clients[name]; exists {
		if client.active {
			log.Printf("MCP服务器 %s 已经在运行，复用现有连接", name)
			return nil
		}
		client.Close()
		delete(m.clients, name)
	}

	// 检查配置
	serverConfig, exists := m.config.MCPServers[name]
	if !exists {
		return fmt.Errorf("未找到MCP服务器配置: %s", name)
	}

	if !serverConfig.Enabled {
		return fmt.Errorf("MCP服务器 %s 已禁用", name)
	}

	// 创建客户端
	client, err := m.createClient(name, serverConfig)
	if err != nil {
		return fmt.Errorf("创建MCP客户端失败: %w", err)
	}

	m.clients[name] = client
	return nil
}

// createClient 创建MCP客户端
func (m *MCPManager) createClient(name string, config MCPServerConfig) (*MCPClient, error) {
	client := &MCPClient{
		config:  config,
		timeout: time.Duration(config.Timeout) * time.Second,
		nextID:  1,
	}

	// 构建命令
	args := append(config.Args, []string{}...)
	cmd := exec.Command(config.Command, args...)

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stdin管道失败: %w", err)
	}
	client.stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("创建stdout管道失败: %w", err)
	}
	client.stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("创建stderr管道失败: %w", err)
	}
	client.stderr = stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("启动MCP服务器失败: %w", err)
	}
	client.cmd = cmd
	client.process = cmd.Process

	// 等待服务器启动
	time.Sleep(2 * time.Second)
	client.active = true
	client.scanner = bufio.NewScanner(client.stdout)

	// 设置更大的缓冲区以处理大型JSON响应
	buf := make([]byte, 0, 64*1024)
	client.scanner.Buffer(buf, 1024*1024)

	// 初始化MCP连接
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, fmt.Errorf("初始化MCP连接失败: %w", err)
	}

	return client, nil
}

// initialize 初始化MCP连接
func (c *MCPClient) initialize() error {
	// 发送初始化请求（使用map确保格式一致）
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "zoteroflow2",
				"version": "1.0.0",
			},
		},
	}

	if err := c.sendRequestMap(initReq); err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}

	// 等待并读取初始化响应
	for i := 0; i < 5; i++ {
		response := c.readResponseWithTimeout(5 * time.Second)
		if response == nil {
			if i == 0 {
				return fmt.Errorf("未收到初始化响应")
			}
			continue
		}

		if response.Error != nil {
			return fmt.Errorf("初始化失败: %s", response.Error.Message)
		}

		// 如果是初始化响应，发送initialized通知
		if response.ID == 1 {
			notifReq := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "notifications/initialized",
			}
			c.sendRequestMap(notifReq)
			return nil
		}
	}

	return fmt.Errorf("初始化超时")
}

// CallTool 调用MCP工具
func (m *MCPManager) CallTool(serverName, toolName string, arguments map[string]interface{}) (*MCPResponse, error) {
	m.mu.RLock()
	client, exists := m.clients[serverName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP服务器 %s 未启动", serverName)
	}

	return client.CallTool(toolName, arguments)
}

// StopServer 停止指定的MCP服务器
func (m *MCPManager) StopServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[name]
	if !exists {
		return fmt.Errorf("MCP服务器 %s 未找到", name)
	}

	client.Close()
	delete(m.clients, name)
	return nil
}

// ListActiveServers 列出活跃的服务器
func (m *MCPManager) ListActiveServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []string
	for name, client := range m.clients {
		if client.active {
			servers = append(servers, name)
		}
	}
	return servers
}

// GetServerConfig 获取服务器配置
func (m *MCPManager) GetServerConfig(name string) (*MCPServerConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.config.MCPServers[name]
	if !exists {
		return nil, fmt.Errorf("未找到MCP服务器配置: %s", name)
	}

	return &config, nil
}

// Close 关闭所有连接
func (m *MCPManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.clients {
		client.Close()
		delete(m.clients, name)
	}

	return nil
}

// MCPRequest MCP请求
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse MCP响应
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError MCP错误
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// CallTool 调用工具
func (c *MCPClient) CallTool(toolName string, arguments map[string]interface{}) (*MCPResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return nil, fmt.Errorf("MCP客户端未激活")
	}

	// 记录详细的调用信息
	log.Printf("🔧 [MCP调用] 开始调用工具: %s", toolName)
	log.Printf("📥 [MCP输入] 工具参数: %s", formatJSON(arguments))

	// 构建工具调用请求（确保与测试脚本格式一致）
	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.nextID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	requestJSON, _ := json.MarshalIndent(callReq, "", "  ")
	log.Printf("📤 [MCP请求] 完整请求数据 (ID: %d):\n%s", c.nextID, string(requestJSON))
	c.nextID++

	// 发送请求
	if err := c.sendRequestMap(callReq); err != nil {
		log.Printf("❌ [MCP错误] 发送工具调用请求失败: %v", err)
		return nil, fmt.Errorf("发送工具调用请求失败: %w", err)
	}

	log.Printf("⏳ [MCP等待] 等待工具响应，超时时间: %v", c.timeout)

	// 读取响应
	response := c.readResponseWithTimeout(c.timeout)
	if response == nil {
		log.Printf("❌ [MCP错误] 读取工具调用响应超时")
		return nil, fmt.Errorf("读取工具调用响应超时")
	}

	// 记录响应信息
	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("📥 [MCP响应] 完整响应数据:\n%s", string(responseJSON))

	if response.Error != nil {
		errorJSON, _ := json.MarshalIndent(response.Error, "", "  ")
		log.Printf("❌ [MCP错误] 服务器返回错误:\n%s", string(errorJSON))
		return nil, fmt.Errorf("MCP错误: %s", response.Error.Message)
	}

	// 记录成功信息
	if response.Result != nil {
		resultSize := len(response.Result)
		log.Printf("✅ [MCP成功] 工具调用成功，响应大小: %d 字节", resultSize)

		// 尝试解析并记录关键结果信息
		if resultMap, ok := parseJSONToMap(response.Result); ok {
			if content, exists := resultMap["content"]; exists {
				if contentArray, ok := content.([]interface{}); ok && len(contentArray) > 0 {
					if firstItem, ok := contentArray[0].(map[string]interface{}); ok {
						if text, ok := firstItem["text"].(string); ok {
							// 限制显示长度，避免日志过长
							displayText := text
							if len(displayText) > 500 {
								displayText = displayText[:500] + "..."
							}
							log.Printf("📄 [MCP结果] 内容摘要: %s", displayText)
						}
					}
				}
			}
		}
	}

	return response, nil
}

// sendRequest 发送请求（保留向后兼容）
func (c *MCPClient) sendRequest(req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	_, err = c.stdin.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("发送数据失败: %w", err)
	}

	return nil
}

// sendRequestMap 发送map格式请求（确保JSON格式正确）
func (c *MCPClient) sendRequestMap(req map[string]interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	log.Printf("📤 发送JSON: %s", string(data))

	_, err = c.stdin.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("发送数据失败: %w", err)
	}

	return nil
}

// readResponseWithTimeout 读取响应（带超时）
func (c *MCPClient) readResponseWithTimeout(timeout time.Duration) *MCPResponse {
	done := make(chan *MCPResponse)

	go func() {
		for c.scanner.Scan() {
			line := strings.TrimSpace(c.scanner.Text())
			// 尝试解析任何看起来像JSON的行
			if strings.HasPrefix(line, "{") || strings.Contains(line, `"jsonrpc"`) {
				var response MCPResponse
				if err := json.Unmarshal([]byte(line), &response); err == nil {
					done <- &response
					return
				} else {
					// 记录解析错误以便调试
					log.Printf("JSON解析错误: %v, 原始行: %s", err, line)
				}
			}
		}
		done <- nil
	}()

	select {
	case response := <-done:
		return response
	case <-time.After(timeout):
		return nil
	}
}

// Close 关闭客户端
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.active = false

	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}

	return nil
}

// IsActive 检查客户端是否活跃
func (c *MCPClient) IsActive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}

// formatJSON 格式化JSON为可读字符串
func formatJSON(data interface{}) string {
	if data == nil {
		return "null"
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", data)
	}
	return string(jsonBytes)
}

// parseJSONToMap 解析JSON到map
func parseJSONToMap(jsonData json.RawMessage) (map[string]interface{}, bool) {
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		log.Printf("⚠️ [MCP解析] 解析JSON到map失败: %v", err)
		return nil, false
	}
	return result, true
}
