package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
			return fmt.Errorf("MCP服务器 %s 已经在运行", name)
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

	// 初始化MCP连接
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, fmt.Errorf("初始化MCP连接失败: %w", err)
	}

	return client, nil
}

// initialize 初始化MCP连接
func (c *MCPClient) initialize() error {
	// 发送初始化请求
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
				"name":    "zoteroflow2",
				"version": "1.0.0",
			},
		},
	}

	if err := c.sendRequest(initReq); err != nil {
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
			notifReq := MCPRequest{
				JSONRPC: "2.0",
				Method:  "notifications/initialized",
			}
			c.sendRequest(notifReq)
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

	// 首先获取工具列表以确保连接正常
	toolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	if err := c.sendRequest(toolsReq); err != nil {
		return nil, fmt.Errorf("发送工具列表请求失败: %w", err)
	}

	toolsResponse := c.readResponseWithTimeout(5 * time.Second)
	if toolsResponse == nil {
		return nil, fmt.Errorf("未收到工具列表响应")
	}

	if toolsResponse.Error != nil {
		return nil, fmt.Errorf("获取工具列表失败: %s", toolsResponse.Error.Message)
	}

	// 构建工具调用请求
	callReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	// 发送请求
	if err := c.sendRequest(callReq); err != nil {
		return nil, fmt.Errorf("发送工具调用请求失败: %w", err)
	}

	// 读取响应
	response := c.readResponseWithTimeout(c.timeout)
	if response == nil {
		return nil, fmt.Errorf("读取工具调用响应超时")
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP错误: %s", response.Error.Message)
	}

	return response, nil
}

// sendRequest 发送请求
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

// readResponseWithTimeout 读取响应（带超时）
func (c *MCPClient) readResponseWithTimeout(timeout time.Duration) *MCPResponse {
	done := make(chan *MCPResponse)

	go func() {
		for c.scanner.Scan() {
			line := c.scanner.Text()
			if strings.HasPrefix(line, "{") {
				var response MCPResponse
				if err := json.Unmarshal([]byte(line), &response); err == nil {
					done <- &response
					return
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
