#!/bin/bash

echo "=== MCP调用各阶段分析测试 ==="
echo "开始时间: $(date)"
echo

# 编译测试
echo "📦 编译测试程序..."
go build -tags=test -o bin/mcp_stages_test tests/mcp/mcp_stages_test.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo

# 运行测试
echo "🚀 运行MCP阶段分析测试..."
echo "注意: 这个测试会分析MCP调用的各个阶段，请耐心等待"
echo

# 设置环境变量（如果需要）
export AI_API_KEY="test-key-for-mcp-testing"

# 运行测试并记录输出
./bin/mcp_stages_test 2>&1 | tee mcp_test_output.log

echo
echo "=== 测试完成 ==="
echo "结束时间: $(date)"
echo "详细日志已保存到: mcp_test_output.log"

# 分析结果
echo
echo "📊 快速结果分析:"
echo "- 查看日志中的错误信息"
echo "- 检查各个阶段的耗时"
echo "- 确认哪个阶段出现了问题"