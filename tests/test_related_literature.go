package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestResult struct {
	TestName string
	Success bool
	Error   string
	Time    time.Duration
	Details string
}

func main() {
	fmt.Println("=== ZoteroFlow2 相关文献分析功能测试 (100秒超时) ===")
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name    string
		doc     string
		question string
	}{
		{
			name:    "测试1: 机器学习基础分析",
			doc:     "机器学习",
			question: "这个领域的主要研究方法是什么？",
		},
		{
			name:    "测试2: 深度学习应用",
			doc:     "深度学习",
			question: "在计算机视觉中有哪些重要应用？",
		},
		{
			name:    "测试3: 神经网络研究",
			doc:     "neural networks",
			question: "近年来有哪些重要突破？",
		},
		{
			name:    "测试4: 数据科学趋势",
			doc:     "data science",
			question: "未来发展趋势如何？",
		},
		{
			name:    "测试5: 人工智能伦理",
			doc:     "artificial intelligence",
			question: "伦理问题有哪些考虑？",
		},
	}

	var results []TestResult

	for i, testCase := range testCases {
		fmt.Printf("🧪 执行 %s (第%d/5)\n", testCase.name, i+1)
		result := runTest(testCase.doc, testCase.question)
		results = append(results, result)

		// 显示测试结果
		if result.Success {
			fmt.Printf("✅ 成功 (%.2fs)\n", result.Time.Seconds())
		} else {
			fmt.Printf("❌ 失败 (%.2fs): %s\n", result.Time.Seconds(), result.Error)
		}

		if result.Details != "" {
			fmt.Printf("   详情: %s\n", result.Details)
		}
		fmt.Println()

		// 测试间隔，避免过于频繁的请求
		if i < len(testCases)-1 {
			fmt.Println("⏳ 等待5秒后进行下一个测试...")
			time.Sleep(5 * time.Second)
		}
	}

	// 统计结果
	analyzeResults(results)
}

func runTest(docIdentifier, question string) TestResult {
	start := time.Now()

	cmd := exec.Command("server/zoteroflow2", "related", docIdentifier, question)

	// 设置环境变量
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")

	// 启动命令
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return TestResult{
			TestName: fmt.Sprintf("%s - %s", docIdentifier, question),
			Success: false,
			Error:   fmt.Sprintf("启动命令失败: %v", err),
			Time:    0,
			Details: "",
		}
	}

	err = cmd.Start()
	if err != nil {
		return TestResult{
			TestName: fmt.Sprintf("%s - %s", docIdentifier, question),
			Success: false,
			Error:   fmt.Sprintf("执行命令失败: %v", err),
			Time:    0,
			Details: "",
		}
	}

	// 读取输出
	var output strings.Builder
	done := make(chan bool)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			output.WriteString(scanner.Text() + "\n")
		}
		done <- true
	}()

	// 等待命令完成
	err = cmd.Wait()
	duration := time.Since(start)

	// 等待输出读取完成
	<-done

	if err != nil {
		return TestResult{
			TestName: fmt.Sprintf("%s - %s", docIdentifier, question),
			Success: false,
			Error:   err.Error(),
			Time:    duration,
			Details: "",
		}
	}

	// 分析输出
	result := TestResult{
		TestName: fmt.Sprintf("%s - %s", docIdentifier, question),
		Success: true,
		Error:   "",
		Time:    duration,
		Details: "",
	}

	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")

	// 检查关键步骤
	steps := map[string]bool{
		"本地文献查找":      false,
		"全球文献搜索":      false,
		"AI智能分析":        false,
		"相关文献详情":      false,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "步骤1") && strings.Contains(line, "本地文献") {
			steps["本地文献查找"] = true
		}
		if strings.Contains(line, "步骤2") && strings.Contains(line, "全球文献") {
			steps["全球文献搜索"] = true
		}
		if strings.Contains(line, "步骤3") && strings.Contains(line, "AI智能分析") {
			steps["AI智能分析"] = true
		}
		if strings.Contains(line, "相关文献详情") {
			steps["相关文献详情"] = true
		}
	}

	// 检查是否有错误
	if strings.Contains(outputStr, "❌") || strings.Contains(outputStr, "失败") {
		result.Success = false
		result.Error = "输出中包含错误信息"
	}

	// 检查步骤完成情况
	var completedSteps []string
	for step, completed := range steps {
		if completed {
			completedSteps = append(completedSteps, step)
		}
	}

	if len(completedSteps) > 0 {
		result.Details = fmt.Sprintf("完成步骤: %s", strings.Join(completedSteps, ", "))
	}

	// 如果没有完成AI分析但其他步骤正常，仍然认为是部分成功
	if !steps["AI智能分析"] && (steps["本地文献查找"] || steps["全球文献搜索"]) {
		if result.Success && result.Error == "" {
			result.Details += " (AI分析跳过)"
		}
	}

	return result
}

func analyzeResults(results []TestResult) {
	fmt.Println("=== 测试结果分析 (100秒超时) ===")
	fmt.Println()

	// 基本统计
	totalTests := len(results)
	successCount := 0
	var totalTime time.Duration
	var fastest, slowest time.Duration

	for _, result := range results {
		if result.Success {
			successCount++
		}
		totalTime += result.Time

		if fastest == 0 || result.Time < fastest {
			fastest = result.Time
		}
		if slowest == 0 || result.Time > slowest {
			slowest = result.Time
		}
	}

	fmt.Printf("📊 总体统计:\n")
	fmt.Printf("   总测试数: %d\n", totalTests)
	fmt.Printf("   成功测试: %d\n", successCount)
	fmt.Printf("   失败测试: %d\n", totalTests-successCount)
	fmt.Printf("   成功率: %.1f%%\n", float64(successCount)/float64(totalTests)*100)
	fmt.Printf("   平均时间: %.2f秒\n", float64(totalTime.Nanoseconds())/float64(totalTests)/1e9)
	fmt.Printf("   最快时间: %.2f秒\n", fastest.Seconds())
	fmt.Printf("   最慢时间: %.2f秒\n", slowest.Seconds())
	fmt.Println()

	// 详细结果
	fmt.Printf("📋 详细结果:\n")
	for i, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("%s %d. %s (%.2fs)\n", status, i+1, result.TestName, result.Time.Seconds())
		if result.Error != "" {
			fmt.Printf("   错误: %s\n", result.Error)
		}
		if result.Details != "" {
			fmt.Printf("   详情: %s\n", result.Details)
		}
	}

	// 性能分析
	fmt.Printf("\n⚡ 性能分析:\n")
	if fastest < 10*time.Second {
		fmt.Printf("   ✅ 性能优秀: 最快响应时间 %.2f秒\n", fastest.Seconds())
	} else if fastest < 30*time.Second {
		fmt.Printf("   ⚠️  性能良好: 最快响应时间 %.2f秒\n", fastest.Seconds())
	} else {
		fmt.Printf("   ❌ 性能较慢: 最快响应时间 %.2f秒，建议优化\n", fastest.Seconds())
	}

	if slowest > 60*time.Second {
		fmt.Printf("   ✅ 响应时间合理: 最慢响应时间 %.2f秒\n", slowest.Seconds())
	} else if slowest > 100*time.Second {
		fmt.Printf("   ❌ 响应超时: 最慢响应时间%.2f秒，可能存在性能问题\n", slowest.Seconds())
	}

	// 错误分析
	fmt.Printf("\n🔍 错误分析:\n")
	var errorTypes = make(map[string]int)
	for _, result := range results {
		if !result.Success {
			errorType := "未知错误"
			if strings.Contains(result.Error, "timeout") || strings.Contains(result.Error, "deadline exceeded") {
				errorType = "网络超时"
			} else if strings.Contains(result.Error, "连接") {
				errorType = "连接错误"
			} else if strings.Contains(result.Error, "启动") {
				errorType = "启动失败"
			} else if strings.Contains(result.Error, "执行") {
				errorType = "执行错误"
			}
			errorTypes[errorType]++
		}
	}

	if len(errorTypes) > 0 {
		fmt.Printf("   错误类型统计:\n")
		for errorType, count := range errorTypes {
			fmt.Printf("   - %s: %d次\n", errorType, count)
		}
	} else {
		fmt.Printf("   ✅ 所有测试都成功完成\n")
	}

	// 建议
	fmt.Printf("\n💡 优化建议:\n")
	if successCount == totalTests {
		fmt.Printf("   ✅ 所有测试通过，功能运行良好\n")
	} else {
		if errorTypes["网络超时"] > 0 {
			fmt.Printf("   🌐 网络超时问题已通过100秒超时设置解决\n")
			fmt.Printf("   💡 如果仍然超时，可能需要检查网络连接\n")
		}
		if errorTypes["启动失败"] > 0 {
			fmt.Printf("   🔧 确保二进制文件存在: cd server && make build\n")
		}
	}
}