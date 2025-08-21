package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/storage"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 演示Mem0Storage的使用方法
func main() {
	fmt.Println("🧠 Mem0 存储示例 - 基于crewAI实现逻辑")

	// 创建logger
	consoleLogger := logger.NewConsoleLogger()

	// 配置Mem0存储（云端模式）
	config := map[string]interface{}{
		// 如果设置了API Key，则使用云端模式
		// "api_key": os.Getenv("MEM0_API_KEY"),
		// "user_id": "demo-user-123",
		// "org_id": "my-org",
		// "project_id": "my-project",

		// 本地模式配置（当前Go版本仅支持云端模式）
		"user_id":  "demo-user-123",
		"agent_id": "demo-agent-456",
		"run_id":   "run-789",
		"infer":    true,
		"includes": "important,facts",
		"excludes": "temporary,debug",
	}

	// 创建crew（模拟）
	crew := map[string]interface{}{
		"name":   "demo-crew",
		"agents": []string{"researcher", "writer"},
	}

	// 创建不同类型的Mem0存储实例
	storageTypes := []string{"short_term", "long_term", "entities", "external"}

	for _, storageType := range storageTypes {
		fmt.Printf("\n📝 测试 %s 存储类型\n", storageType)

		// 创建Mem0存储实例
		mem0Storage := storage.NewMem0Storage(storageType, crew, config, consoleLogger)

		// 创建测试记忆项
		testItem := memory.MemoryItem{
			ID:    fmt.Sprintf("test-%s-%d", storageType, time.Now().Unix()),
			Value: fmt.Sprintf("这是一个%s类型的测试记忆：用户偏好使用高质量产品", storageType),
			Metadata: map[string]interface{}{
				"category":   "user_preference",
				"confidence": 0.95,
				"source":     "user_interaction",
				"timestamp":  time.Now().Format(time.RFC3339),
			},
		}

		// 演示配置验证
		fmt.Printf("✅ 配置信息:\n")
		configInfo := mem0Storage.GetConfig()
		for k, v := range configInfo {
			fmt.Printf("  %s: %v\n", k, v)
		}

		fmt.Printf("🔧 配置状态: %v\n", mem0Storage.IsConfigured())

		// 如果是本地模式，显示相应信息
		ctx := context.Background()
		if !mem0Storage.IsConfigured() {
			fmt.Printf("💡 提示：设置 MEM0_API_KEY 环境变量以启用云端模式\n")

			// 测试连接会失败，但我们可以演示其他功能
			fmt.Printf("⚠️  当前为本地模式（Go版本暂不支持）\n")
		} else {
			// 测试连接
			fmt.Printf("🔍 测试连接...\n")
			if err := mem0Storage.TestConnection(ctx); err != nil {
				fmt.Printf("❌ 连接测试失败: %v\n", err)
			} else {
				fmt.Printf("✅ 连接测试成功\n")
			}
		}

		// 演示保存功能（即使在本地模式也会显示逻辑）
		fmt.Printf("💾 尝试保存记忆...\n")
		if err := mem0Storage.Save(ctx, testItem); err != nil {
			fmt.Printf("❌ 保存失败（预期）: %v\n", err)
		} else {
			fmt.Printf("✅ 保存成功\n")
		}

		// 演示搜索功能
		fmt.Printf("🔍 尝试搜索记忆...\n")
		results, err := mem0Storage.Search(ctx, "用户偏好", 5, 0.7)
		if err != nil {
			fmt.Printf("❌ 搜索失败（预期）: %v\n", err)
		} else {
			fmt.Printf("✅ 搜索成功，找到 %d 条记忆\n", len(results))
		}

		// 关闭存储
		if err := mem0Storage.Close(); err != nil {
			log.Printf("关闭存储失败: %v", err)
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	fmt.Println("\n🎉 Mem0存储示例完成！")
	fmt.Println("\n📖 使用说明:")
	fmt.Println("1. 设置 MEM0_API_KEY 环境变量以启用云端模式")
	fmt.Println("2. 配置 user_id, org_id, project_id 等参数")
	fmt.Println("3. 根据crewAI的逻辑，支持多种存储类型和过滤器")
	fmt.Println("4. 完全兼容crewAI的Mem0存储接口和行为")
}
