package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/crew"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// AI研究助手 - 完整的端到端示例
// 展示 Agent + Tool + LLM 的完整工作流
func main() {
	fmt.Println("🚀 GreenSoulAI 完整端到端示例：AI研究助手")
	fmt.Println("===============================================")
	fmt.Println()

	// 检查OpenAI API密钥
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 错误：未设置 OPENAI_API_KEY 环境变量")
		fmt.Println()
		fmt.Println("请先设置您的OpenAI API密钥：")
		fmt.Println("export OPENAI_API_KEY='your-api-key-here'")
		fmt.Println()
		fmt.Println("或者在程序中直接设置（不推荐用于生产环境）：")
		fmt.Println("apiKey := \"your-api-key-here\"")
		return
	}

	// 1. 初始化基础组件
	fmt.Println("🔧 初始化系统组件...")
	baseLogger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(baseLogger)

	// 设置事件监听器，展示完整的事件系统
	setupEventListeners(eventBus)

	// 2. 创建真实的OpenAI LLM
	fmt.Println("🤖 创建OpenAI LLM实例...")
	config := &llm.Config{
		Provider:    "openai",
		Model:       "gpt-4o-mini", // 使用成本较低的模型
		APIKey:      apiKey,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 1500; return &t }(),
	}

	llmInstance, err := llm.CreateLLM(config)
	if err != nil {
		log.Fatalf("❌ 创建LLM失败: %v", err)
	}
	defer llmInstance.Close()

	fmt.Printf("✅ 成功创建 %s 实例\n", llmInstance.GetModel())
	fmt.Printf("🎯 支持函数调用: %v\n", llmInstance.SupportsFunctionCalling())

	// 3. 演示不同的使用场景
	fmt.Println("\n" + strings.Repeat("=", 50))

	// 场景1: 单个Agent使用工具
	if err := demonstrateSingleAgentWithTools(llmInstance, eventBus, baseLogger); err != nil {
		log.Printf("❌ 单Agent演示失败: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// 场景2: Crew协作研究
	if err := demonstrateCrewResearch(llmInstance, eventBus, baseLogger); err != nil {
		log.Printf("❌ Crew协作演示失败: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// 场景3: 复杂工作流
	if err := demonstrateComplexWorkflow(llmInstance, eventBus, baseLogger); err != nil {
		log.Printf("❌ 复杂工作流演示失败: %v", err)
	}

	fmt.Println("\n🎉 所有演示完成！")
	fmt.Println("✨ 本示例展示了：")
	fmt.Println("   - Agent与LLM的完整集成")
	fmt.Println("   - 工具的智能使用")
	fmt.Println("   - Crew团队协作")
	fmt.Println("   - 事件系统监控")
	fmt.Println("   - 错误处理和恢复")
	fmt.Println("   - 真实的OpenAI API调用")
}

// 场景1: 单个Agent使用工具进行研究
func demonstrateSingleAgentWithTools(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) error {
	fmt.Println("📊 场景1: 单个Agent使用工具进行技术研究")

	// 创建研究员Agent
	researcherConfig := agent.AgentConfig{
		Role:      "高级技术研究员",
		Goal:      "对新兴技术进行全面研究并提供详细洞察",
		Backstory: "你是一位经验丰富的技术研究专家，在AI、软件开发和新兴技术趋势方面有深度专业知识。你总是提供详细、有据的洞察，并用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}

	researcher, err := agent.NewBaseAgent(researcherConfig)
	if err != nil {
		return fmt.Errorf("创建研究员失败: %w", err)
	}

	// 为Agent添加研究工具
	if err := addResearchTools(researcher); err != nil {
		return fmt.Errorf("添加工具失败: %w", err)
	}

	// 初始化Agent
	if err := researcher.Initialize(); err != nil {
		return fmt.Errorf("初始化研究员失败: %w", err)
	}

	// 创建研究任务
	researchTask := agent.NewBaseTask(
		"研究2024年大语言模型（LLMs）的现状和未来趋势。重点关注：1）最新的模型架构 2）性能改进 3）实际应用 4）挑战和限制",
		"一份全面的研究报告，涵盖LLMs的现状，包括最新发展、性能指标、应用领域和未来趋势。报告应该详细且结构清晰。",
	)

	// 执行任务
	fmt.Println("🔍 开始执行研究任务...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()
	output, err := researcher.Execute(ctx, researchTask)
	if err != nil {
		return fmt.Errorf("任务执行失败: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ 任务完成! 耗时: %v\n", duration)
	fmt.Printf("📄 生成内容长度: %d 字符\n", len(output.Raw))
	fmt.Printf("🔢 使用Token: %d\n", output.TokensUsed)

	// 显示研究结果摘要
	fmt.Println("\n📋 研究结果摘要:")
	fmt.Println(strings.Repeat("-", 40))
	lines := strings.Split(output.Raw, "\n")
	for i, line := range lines {
		if i >= 10 { // 只显示前10行
			fmt.Println("... (更多内容已省略)")
			break
		}
		if strings.TrimSpace(line) != "" {
			fmt.Printf("   %s\n", line)
		}
	}

	// 展示工具使用统计
	tools := researcher.GetTools()
	fmt.Printf("\n🔧 工具使用统计:\n")
	for _, tool := range tools {
		fmt.Printf("   - %s: %d次使用\n", tool.GetName(), tool.GetUsageCount())
	}

	return nil
}

// 场景2: Crew协作研究
func demonstrateCrewResearch(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) error {
	fmt.Println("👥 场景2: Crew团队协作进行技术调研")

	// 创建多个专业Agent
	agents, err := createResearchTeamAgents(llmInstance, eventBus, baseLogger)
	if err != nil {
		return fmt.Errorf("创建研究团队失败: %w", err)
	}

	// 创建研究Crew
	crewConfig := &crew.CrewConfig{
		Name:    "TechResearchCrew",
		Process: crew.ProcessSequential,
		Verbose: true,
	}
	researchCrew := crew.NewBaseCrew(crewConfig, eventBus, baseLogger)

	// 添加所有Agent到Crew
	for _, agent := range agents {
		if err := agent.Initialize(); err != nil {
			return fmt.Errorf("初始化Agent失败: %w", err)
		}
		if err := researchCrew.AddAgent(agent); err != nil {
			return fmt.Errorf("添加Agent到Crew失败: %w", err)
		}
	}

	// 创建协作任务序列
	tasks := createCollaborativeTasks()
	for _, task := range tasks {
		if err := researchCrew.AddTask(task); err != nil {
			return fmt.Errorf("添加任务失败: %w", err)
		}
	}

	// 执行Crew任务
	fmt.Println("🚀 开始团队协作研究...")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	startTime := time.Now()
	output, err := researchCrew.Kickoff(ctx, map[string]interface{}{
		"research_topic":  "Artificial Intelligence in Software Development",
		"focus_areas":     []string{"code generation", "testing automation", "architecture design"},
		"depth":           "comprehensive",
		"target_audience": "senior developers and architects",
	})

	if err != nil {
		return fmt.Errorf("crew执行失败: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ 团队协作完成! 总耗时: %v\n", duration)
	fmt.Printf("📊 完成任务数: %d\n", len(output.TasksOutput))

	// 显示每个任务的结果摘要
	fmt.Println("\n📋 团队协作结果:")
	for i, taskOutput := range output.TasksOutput {
		fmt.Printf("\n%d. %s 的输出:\n", i+1, taskOutput.Agent)
		fmt.Println(strings.Repeat("-", 30))
		lines := strings.Split(taskOutput.Raw, "\n")
		for j, line := range lines {
			if j >= 5 { // 每个任务显示前5行
				fmt.Println("   ... (更多内容已省略)")
				break
			}
			if strings.TrimSpace(line) != "" {
				fmt.Printf("   %s\n", line)
			}
		}
	}

	// 显示Crew统计信息
	metrics := researchCrew.GetUsageMetrics()
	fmt.Printf("\n📈 团队执行统计:\n")
	fmt.Printf("   - 总执行时间: %v\n", duration)
	fmt.Printf("   - Agent数量: %d\n", len(researchCrew.GetAgents()))
	fmt.Printf("   - 任务数量: %d\n", len(researchCrew.GetTasks()))
	if metrics != nil {
		fmt.Printf("   - 总Token使用: %d\n", metrics.TotalTokens)
		fmt.Printf("   - 成功任务数: %d\n", metrics.SuccessfulTasks)
	}

	return nil
}

// 场景3: 复杂工作流演示
func demonstrateComplexWorkflow(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) error {
	fmt.Println("🔄 场景3: 复杂工作流 - AI产品需求分析")

	// 这里将展示一个更复杂的场景：
	// 1. 市场研究Agent收集信息
	// 2. 产品经理Agent分析需求
	// 3. 技术架构师Agent设计方案
	// 4. 项目经理Agent制定计划

	// 创建专业团队
	marketResearcher, err := createMarketResearcher(llmInstance, eventBus, baseLogger)
	if err != nil {
		return err
	}

	productManager, err := createProductManager(llmInstance, eventBus, baseLogger)
	if err != nil {
		return err
	}

	// 这个场景展示了更复杂的工作流，包括条件任务、依赖关系等
	fmt.Println("📊 市场研究阶段...")

	marketTask := agent.NewBaseTask(
		"分析AI驱动开发工具的当前市场。重点关注：1）市场规模和增长 2）主要竞争对手 3）用户需求和痛点 4）市场机会",
		"一份全面的市场分析报告，包括市场规模、竞争格局、用户需求和所识别的AI开发工具机会。",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	marketOutput, err := marketResearcher.Execute(ctx, marketTask)
	if err != nil {
		return fmt.Errorf("市场研究失败: %w", err)
	}

	fmt.Printf("✅ 市场研究完成 (Token: %d)\n", marketOutput.TokensUsed)

	fmt.Println("🎯 产品需求分析阶段...")

	productTask := agent.NewBaseTask(
		fmt.Sprintf("基于市场研究结果，为AI开发助手定义产品需求。市场研究结果：%s",
			truncateString(marketOutput.Raw, 500)),
		"详细的产品需求文档，包括功能、用户故事、成功指标和技术需求。",
	)

	productOutput, err := productManager.Execute(ctx, productTask)
	if err != nil {
		return fmt.Errorf("产品分析失败: %w", err)
	}

	fmt.Printf("✅ 产品需求分析完成 (Token: %d)\n", productOutput.TokensUsed)

	// 展示最终结果
	fmt.Println("\n🎉 复杂工作流完成!")
	fmt.Println("\n📋 工作流结果摘要:")
	fmt.Println("\n1. 市场研究结果:")
	fmt.Printf("   %s\n", truncateString(marketOutput.Raw, 200))

	fmt.Println("\n2. 产品需求分析:")
	fmt.Printf("   %s\n", truncateString(productOutput.Raw, 200))

	totalTokens := marketOutput.TokensUsed + productOutput.TokensUsed
	fmt.Printf("\n📊 总Token使用量: %d\n", totalTokens)

	return nil
}

// 辅助函数：为Agent添加研究工具
func addResearchTools(agent agent.Agent) error {
	// 添加网络搜索工具（模拟）
	searchTool := createWebSearchTool()
	if err := agent.AddTool(searchTool); err != nil {
		return err
	}

	// 添加数据分析工具
	analysisTool := createDataAnalysisTool()
	if err := agent.AddTool(analysisTool); err != nil {
		return err
	}

	// 添加文档生成工具
	docTool := createDocumentTool()
	if err := agent.AddTool(docTool); err != nil {
		return err
	}

	return nil
}

// 创建网络搜索工具（模拟实现）
func createWebSearchTool() agent.Tool {
	return agent.NewBaseTool(
		"web_search",
		"Search the web for current information about technology trends, companies, and developments",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			query, ok := args["query"].(string)
			if !ok {
				return nil, fmt.Errorf("query parameter is required")
			}

			// 模拟搜索结果（在实际应用中，这里会调用真实的搜索API）
			results := map[string]interface{}{
				"query": query,
				"results": []map[string]interface{}{
					{
						"title":   "Latest LLM Developments - 2024 Trends",
						"url":     "https://example.com/llm-trends-2024",
						"summary": "Recent advances in Large Language Models include improved efficiency, multimodal capabilities, and better reasoning abilities.",
					},
					{
						"title":   "OpenAI GPT-4 Turbo Performance Analysis",
						"url":     "https://example.com/gpt4-turbo-analysis",
						"summary": "GPT-4 Turbo shows significant improvements in coding tasks and mathematical reasoning compared to previous versions.",
					},
					{
						"title":   "Google Gemini vs ChatGPT Comparison",
						"url":     "https://example.com/gemini-vs-chatgpt",
						"summary": "Comparative analysis of Google Gemini and ChatGPT across various benchmarks including coding, reasoning, and creative tasks.",
					},
				},
				"search_time": time.Now().Format(time.RFC3339),
			}

			return results, nil
		},
	)
}

// 创建数据分析工具
func createDataAnalysisTool() agent.Tool {
	return agent.NewBaseTool(
		"data_analysis",
		"Analyze numerical data and generate insights, statistics, and trends",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			dataType, ok := args["data_type"].(string)
			if !ok {
				return nil, fmt.Errorf("data_type parameter is required")
			}

			// 模拟数据分析
			analysis := map[string]interface{}{
				"data_type":   dataType,
				"analyzed_at": time.Now().Format(time.RFC3339),
				"key_insights": []string{
					"Market growth rate: 35% YoY",
					"Primary use cases: Code generation (45%), Testing (30%), Documentation (25%)",
					"User satisfaction score: 4.2/5.0",
				},
				"trends": map[string]interface{}{
					"adoption_rate":     "increasing",
					"market_maturity":   "early growth",
					"competition_level": "moderate",
				},
			}

			return analysis, nil
		},
	)
}

// 创建文档生成工具
func createDocumentTool() agent.Tool {
	return agent.NewBaseTool(
		"document_generator",
		"Generate structured documents, reports, and summaries from research data",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content parameter is required")
			}

			docType := "report"
			if dt, exists := args["type"].(string); exists {
				docType = dt
			}

			// 生成结构化文档
			document := map[string]interface{}{
				"type":       docType,
				"title":      "Generated Research Document",
				"created_at": time.Now().Format(time.RFC3339),
				"content":    content,
				"word_count": len(strings.Fields(content)),
				"sections": []string{
					"Executive Summary",
					"Market Analysis",
					"Technical Overview",
					"Recommendations",
				},
				"metadata": map[string]interface{}{
					"format":  "structured_report",
					"version": "1.0",
				},
			}

			return document, nil
		},
	)
}

// 创建研究团队的Agent们
func createResearchTeamAgents(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) ([]agent.Agent, error) {
	var agents []agent.Agent

	// 1. 数据收集专家
	dataCollectorConfig := agent.AgentConfig{
		Role:      "数据收集专家",
		Goal:      "从各种源收集全面的数据和信息",
		Backstory: "你是一位擅长从多个源查找和收集相关信息的专家。你可以使用各种研究工具和数据库，总是用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	dataCollector, err := agent.NewBaseAgent(dataCollectorConfig)
	if err != nil {
		return nil, err
	}
	addResearchTools(dataCollector) // 添加研究工具
	agents = append(agents, dataCollector)

	// 2. 趋势分析师
	trendAnalystConfig := agent.AgentConfig{
		Role:      "趋势分析专家",
		Goal:      "分析技术和市场数据中的趋势和模式",
		Backstory: "你是一位技能高超的分析师，能够识别技术市场中的趋势、模式和未来方向。你擅长解释数据和做出预测，总是用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	trendAnalyst, err := agent.NewBaseAgent(trendAnalystConfig)
	if err != nil {
		return nil, err
	}
	agents = append(agents, trendAnalyst)

	// 3. 技术评估专家
	techEvaluatorConfig := agent.AgentConfig{
		Role:      "技术评估专家",
		Goal:      "评估技术的技术方面、能力和局限性",
		Backstory: "你是一位技术专家，能够评估各种技术的技术优劣、实施挑战和实用应用。你总是用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	techEvaluator, err := agent.NewBaseAgent(techEvaluatorConfig)
	if err != nil {
		return nil, err
	}
	agents = append(agents, techEvaluator)

	return agents, nil
}

// 创建协作任务序列
func createCollaborativeTasks() []agent.Task {
	var tasks []agent.Task

	// 任务1: 数据收集
	task1 := agent.NewBaseTask(
		"收集关于AI在软件开发中应用的全面数据，包括当前工具、市场采用情况、用户反馈和技术能力",
		"一份详细的数据收集报告，包含来自多个源关于AI开发工具的信息，包括统计数据、用户评价和技本规格",
	)
	tasks = append(tasks, task1)

	// 任务2: 趋势分析
	task2 := agent.NewBaseTask(
		"分析收集的数据，识别关键趋势、增长模式和AI驱动的软件开发的未来方向",
		"一份全面的趋势分析报告，突出关键模式、增长轨迹和对AI开发工具未来发展的预测",
	)
	tasks = append(tasks, task2)

	// 任务3: 技术评估
	task3 := agent.NewBaseTask(
		"评估当前AI开发工具的技术方面，包括优势、局限性和潜在改进空间",
		"一份技术评估报告，评估当前AI开发工具的能力、局限性和改进建议",
	)
	tasks = append(tasks, task3)

	return tasks
}

// 创建市场研究员
func createMarketResearcher(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) (agent.Agent, error) {
	config := agent.AgentConfig{
		Role:      "高级市场研究分析师",
		Goal:      "进行全面的市场研究和竞争分析",
		Backstory: "你是一位经验丰富的市场研究分析师，擅长技术市场、用户行为分析和竞争情报。你提供数据驱动的洞察，总是用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}

	researcher, err := agent.NewBaseAgent(config)
	if err != nil {
		return nil, err
	}

	// 添加市场研究工具
	if err := addResearchTools(researcher); err != nil {
		return nil, err
	}

	if err := researcher.Initialize(); err != nil {
		return nil, err
	}

	return researcher, nil
}

// 创建产品经理
func createProductManager(llmInstance llm.LLM, eventBus events.EventBus, baseLogger logger.Logger) (agent.Agent, error) {
	config := agent.AgentConfig{
		Role:      "高级产品经理",
		Goal:      "基于市场研究定义产品需求和策略",
		Backstory: "你是一位经验丰富的产品经理，在AI/ML产品、用户体验设计和产品策略方面有专业知识。你擅长将市场需求转化为可执行的产品需求，总是用中文回答。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}

	pm, err := agent.NewBaseAgent(config)
	if err != nil {
		return nil, err
	}

	if err := pm.Initialize(); err != nil {
		return nil, err
	}

	return pm, nil
}

// 设置事件监听器
func setupEventListeners(eventBus events.EventBus) {
	// 监听Agent执行事件
	eventBus.Subscribe("agent_execution_started", func(ctx context.Context, event events.Event) error {
		fmt.Printf("🤖 Agent开始执行任务: %v\n", event.GetPayload())
		return nil
	})

	eventBus.Subscribe("agent_execution_completed", func(ctx context.Context, event events.Event) error {
		payload := event.GetPayload()
		if success, ok := payload["success"].(bool); ok && success {
			fmt.Printf("✅ Agent任务完成: %v\n", payload["agent"])
		} else {
			fmt.Printf("❌ Agent任务失败: %v\n", payload["agent"])
		}
		return nil
	})

	// 监听LLM调用事件
	eventBus.Subscribe("llm_call_started", func(ctx context.Context, event events.Event) error {
		fmt.Printf("🧠 LLM调用开始: %v\n", event.GetPayload()["model"])
		return nil
	})

	eventBus.Subscribe("llm_call_completed", func(ctx context.Context, event events.Event) error {
		payload := event.GetPayload()
		fmt.Printf("🧠 LLM调用完成: %vms\n", payload["duration_ms"])
		return nil
	})

	// 监听工具使用事件
	eventBus.Subscribe("tool_usage_started", func(ctx context.Context, event events.Event) error {
		fmt.Printf("🔧 工具调用: %v\n", event.GetPayload()["tool_name"])
		return nil
	})
}

// 辅助函数：截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
