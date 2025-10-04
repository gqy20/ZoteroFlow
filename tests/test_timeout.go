package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== ZoteroFlow2 100秒超时测试 ===")
	fmt.Println("将进行3次测试，验证100秒超时设置是否有效")
	fmt.Println()

	testCases := []struct {
		name string
		doc  string
		q    string
	}{
		{"测试1: 机器学习", "机器学习", "主要研究方法"},
		{"测试2: 深度学习", "深度学习", "计算机视觉应用"},
		{"测试3: 神经网络", "neural networks", "最新突破"},
	}

	successCount := 0
	var totalTime time.Duration

	for i, test := range testCases {
		fmt.Printf("🧪 执行 %s (第%d/3)\n", test.name, i+1)

		start := time.Now()
		cmd := exec.Command("../server/zoteroflow2", "related", test.doc, test.q)

		// 设置超时为120秒，给系统留出余量
		timer := time.AfterFunc(120*time.Second, func() {
			cmd.Process.Kill()
		})

		output, err := cmd.CombinedOutput()
		timer.Stop()

		duration := time.Since(start)

		if err != nil {
			if strings.Contains(err.Error(), "signal: killed") {
				fmt.Printf("❌ 超时 (>120s): %v\n", err)
			} else {
				fmt.Printf("❌ 失败 (%.2fs): %v\n", duration.Seconds(), err)
			}
		} else {
			fmt.Printf("✅ 成功 (%.2fs)\n", duration.Seconds())
			successCount++

			// 检查是否包含AI分析结果
			outputStr := string(output)
			if strings.Contains(outputStr, "AI分析完成") {
				fmt.Printf("   ✅ 包含AI分析结果\n")
			} else {
				fmt.Printf("   ⚠️  缺少AI分析结果\n")
			}
		}

		totalTime += duration
		fmt.Println()

		// 测试间隔
		if i < len(testCases)-1 {
			fmt.Println("⏳ 等待3秒...")
			time.Sleep(3 * time.Second)
		}
	}

	// 统计结果
	fmt.Printf("=== 测试总结 ===\n")
	fmt.Printf("成功测试: %d/%d\n", successCount, len(testCases))
	fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(len(testCases))*100)
	fmt.Printf("平均耗时: %.2f秒\n", float64(totalTime.Nanoseconds())/float64(len(testCases))/1e9)

	if successCount == len(testCases) {
		fmt.Printf("✅ 所有测试通过！100秒超时设置工作正常\n")
	} else {
		fmt.Printf("⚠️  部分测试失败，可能需要进一步优化\n")
	}
}