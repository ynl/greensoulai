package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/crew"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 创建一个简单的研究团队示例，展示Crew协作逻辑的实际使用
func main() {
	// 初始化基础组件
	baseLogger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(baseLogger)

	// 创建Mock LLM (在真实场景中可以替换为OpenAI LLM)
	mockLLM := &MockLLM{
		responses: []string{
			"Based on my research, artificial intelligence trends in 2024 include: 1) Generative AI adoption, 2) Multimodal AI systems, 3) AI safety and alignment",
			"Analysis complete: The research shows strong momentum in GenAI, with significant investments in safety measures and multimodal capabilities",
		},
		currentCall: 0,
	}

	fmt.Println("🚀 Starting Research Team Collaboration Example...")

	// 1. 演示Sequential模式
	fmt.Println("\n=== Sequential Process Example ===")
	if err := demonstrateSequential(eventBus, baseLogger, mockLLM); err != nil {
		log.Fatalf("Sequential demo failed: %v", err)
	}

	// 重置Mock LLM
	mockLLM.currentCall = 0

	// 2. 演示Hierarchical模式
	fmt.Println("\n=== Hierarchical Process Example ===")
	if err := demonstrateHierarchical(eventBus, baseLogger, mockLLM); err != nil {
		log.Fatalf("Hierarchical demo failed: %v", err)
	}

	fmt.Println("\n✅ All demonstrations completed successfully!")
}

func demonstrateSequential(eventBus events.EventBus, baseLogger logger.Logger, mockLLM llm.LLM) error {
	// 创建研究团队agents
	researcherConfig := agent.AgentConfig{
		Role:      "Senior Researcher",
		Goal:      "Conduct thorough research on emerging technologies",
		Backstory: "You are an experienced researcher with expertise in technology trends",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	researcher, err := agent.NewBaseAgent(researcherConfig)
	if err != nil {
		return fmt.Errorf("failed to create researcher: %w", err)
	}

	analystConfig := agent.AgentConfig{
		Role:      "Data Analyst",
		Goal:      "Analyze research findings and extract insights",
		Backstory: "You are a skilled data analyst who can find patterns in research data",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	analyst, err := agent.NewBaseAgent(analystConfig)
	if err != nil {
		return fmt.Errorf("failed to create analyst: %w", err)
	}

	// 创建research crew
	crewConfig := &crew.CrewConfig{
		Name:    "ResearchTeam",
		Process: crew.ProcessSequential,
		Verbose: true,
	}
	researchCrew := crew.NewBaseCrew(crewConfig, eventBus, baseLogger)

	// 初始化agents
	if err := researcher.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize researcher: %w", err)
	}
	if err := analyst.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize analyst: %w", err)
	}

	// 添加agents
	if err := researchCrew.AddAgent(researcher); err != nil {
		return fmt.Errorf("failed to add researcher: %w", err)
	}
	if err := researchCrew.AddAgent(analyst); err != nil {
		return fmt.Errorf("failed to add analyst: %w", err)
	}

	// 创建任务
	researchTask := agent.NewBaseTask(
		"Research current AI trends and technologies for 2024",
		"A comprehensive report on AI trends including key developments, major players, and future predictions",
	)

	analysisTask := agent.NewBaseTask(
		"Analyze the research findings and provide strategic insights",
		"A detailed analysis with actionable insights and recommendations based on the research",
	)

	// 添加任务
	if err := researchCrew.AddTask(researchTask); err != nil {
		return fmt.Errorf("failed to add research task: %w", err)
	}
	if err := researchCrew.AddTask(analysisTask); err != nil {
		return fmt.Errorf("failed to add analysis task: %w", err)
	}

	// 执行crew
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("📊 Starting sequential research process...")
	output, err := researchCrew.Kickoff(ctx, map[string]interface{}{
		"focus_area": "artificial intelligence",
		"year":       "2024",
		"depth":      "comprehensive",
	})

	if err != nil {
		return fmt.Errorf("crew execution failed: %w", err)
	}

	fmt.Printf("✅ Sequential process completed with %d task outputs\n", len(output.TasksOutput))
	for i, taskOutput := range output.TasksOutput {
		fmt.Printf("   Task %d (%s): %s...\n", i+1, taskOutput.Agent,
			truncateString(taskOutput.Raw, 80))
	}

	return nil
}

func demonstrateHierarchical(eventBus events.EventBus, baseLogger logger.Logger, mockLLM llm.LLM) error {
	// 创建工作团队agents
	developerConfig := agent.AgentConfig{
		Role:      "Senior Developer",
		Goal:      "Implement software features based on requirements",
		Backstory: "You are an experienced software developer skilled in multiple programming languages",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	developer, err := agent.NewBaseAgent(developerConfig)
	if err != nil {
		return fmt.Errorf("failed to create developer: %w", err)
	}

	testerConfig := agent.AgentConfig{
		Role:      "QA Engineer",
		Goal:      "Test software features and ensure quality",
		Backstory: "You are a meticulous QA engineer focused on finding bugs and ensuring quality",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    baseLogger,
	}
	tester, err := agent.NewBaseAgent(testerConfig)
	if err != nil {
		return fmt.Errorf("failed to create tester: %w", err)
	}

	// 创建development crew
	devCrewConfig := &crew.CrewConfig{
		Name:       "DevelopmentTeam",
		Process:    crew.ProcessHierarchical,
		Verbose:    true,
		ManagerLLM: mockLLM, // 设置Manager LLM
	}
	devCrew := crew.NewBaseCrew(devCrewConfig, eventBus, baseLogger)

	// 初始化agents
	if err := developer.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize developer: %w", err)
	}
	if err := tester.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize tester: %w", err)
	}

	// 添加agents
	if err := devCrew.AddAgent(developer); err != nil {
		return fmt.Errorf("failed to add developer: %w", err)
	}
	if err := devCrew.AddAgent(tester); err != nil {
		return fmt.Errorf("failed to add tester: %w", err)
	}

	// 创建任务
	implementTask := agent.NewBaseTask(
		"Implement a user authentication system",
		"A fully implemented authentication system with proper security measures",
	)

	testTask := agent.NewBaseTask(
		"Test the authentication system thoroughly",
		"A comprehensive test report with all test cases and results",
	)

	// 添加任务
	if err := devCrew.AddTask(implementTask); err != nil {
		return fmt.Errorf("failed to add implement task: %w", err)
	}
	if err := devCrew.AddTask(testTask); err != nil {
		return fmt.Errorf("failed to add test task: %w", err)
	}

	// 执行crew
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("👥 Starting hierarchical development process...")
	output, err := devCrew.Kickoff(ctx, map[string]interface{}{
		"project":  "user_authentication",
		"priority": "high",
		"deadline": "next_sprint",
	})

	if err != nil {
		return fmt.Errorf("crew execution failed: %w", err)
	}

	fmt.Printf("✅ Hierarchical process completed with %d task outputs\n", len(output.TasksOutput))
	fmt.Printf("   Manager: %s\n", output.TasksOutput[0].Agent) // 应该是Crew Manager
	for i, taskOutput := range output.TasksOutput {
		fmt.Printf("   Task %d: %s...\n", i+1, truncateString(taskOutput.Raw, 80))
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MockLLM 简单的Mock实现用于演示
type MockLLM struct {
	responses   []string
	currentCall int
}

func (m *MockLLM) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
	if m.currentCall >= len(m.responses) {
		m.currentCall = 0 // 重置以循环使用
	}

	response := m.responses[m.currentCall]
	m.currentCall++

	return &llm.Response{
		Content:      response,
		Model:        "mock-gpt-3.5-turbo",
		FinishReason: "stop",
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: len(response) / 4, // 粗略估算
			TotalTokens:      10 + len(response)/4,
			Cost:             0.001,
		},
	}, nil
}

func (m *MockLLM) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	responseChan := make(chan llm.StreamResponse, 1)

	go func() {
		defer close(responseChan)

		response, err := m.Call(ctx, messages, options)
		if err != nil {
			responseChan <- llm.StreamResponse{Error: err}
			return
		}

		responseChan <- llm.StreamResponse{
			Delta:        response.Content,
			Usage:        &response.Usage,
			FinishReason: response.FinishReason,
		}
	}()

	return responseChan, nil
}

func (m *MockLLM) SetEventBus(eventBus events.EventBus) {
	// Mock implementation - no-op
}

func (m *MockLLM) GetContextWindowSize() int {
	return 4096
}

func (m *MockLLM) GetModel() string {
	return "mock-gpt-3.5-turbo"
}

func (m *MockLLM) SupportsFunctionCalling() bool {
	return true
}

func (m *MockLLM) Close() error {
	return nil
}
