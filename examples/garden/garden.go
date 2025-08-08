package garden

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/crew"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// RunGarden 构建一个包含5种花朵（Agent）的花园（Crew），
// 以顺序流程模拟“自我介绍-相互交流-资源协商-最终布局”的业务链路。
func RunGarden(ctx context.Context) (*crew.CrewOutput, error) {
	// 默认静默日志，仅显示“💬 对话”。设置 GARDEN_VERBOSE=1 可启用详细日志。
	var baseLogger logger.Logger
	if os.Getenv("GARDEN_VERBOSE") == "1" {
		baseLogger = logger.NewConsoleLogger()
	} else {
		baseLogger = &silentLogger{}
	}
	bus := events.NewEventBus(baseLogger)

	// 配置对话延迟（毫秒），默认0；设置 GARDEN_DELAY_MS 可开启，例如 300
	var speakDelay time.Duration
	if v := os.Getenv("GARDEN_DELAY_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			speakDelay = time.Duration(ms) * time.Millisecond
		}
	}
	// 延迟抖动（百分比 0-100），默认 30（±30% 抖动）
	jitterPct := 30
	if v := os.Getenv("GARDEN_DELAY_JITTER_PCT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			if p < 0 {
				p = 0
			}
			if p > 100 {
				p = 100
			}
			jitterPct = p
		}
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建 LLM（仅支持 OpenRouter），未配置将报错
	var gardenLLM llm.LLM
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		baseURL := "https://openrouter.ai/api/v1"
		model := os.Getenv("GARDEN_MODEL")
		if model == "" {
			model = "moonshotai/kimi-k2:free"
		}
		gardenLLM = llm.NewOpenAILLM(
			model,
			llm.WithAPIKey(apiKey),
			llm.WithBaseURL(baseURL),
			llm.WithTimeout(30*time.Second),
			llm.WithMaxRetries(3),
			llm.WithCustomHeader("HTTP-Referer", "https://github.com/ynl/greensoulai"),
			llm.WithCustomHeader("X-Title", "GreenSoulAI Garden"),
		)
		// 可选：把事件总线传给 LLM
		gardenLLM.SetEventBus(bus)
	} else {
		return nil, fmt.Errorf("no API key found, set OPENROUTER_API_KEY")
	}

	// 创建5个花朵Agent
	rose, err := agent.NewBaseAgent(agent.AgentConfig{
		Role:      "Rose",
		Goal:      "表达玫瑰的生长需求与协作建议",
		Backstory: "你是玫瑰，偏香气，喜微酸土，怕积水并需适度日照。",
		LLM:       gardenLLM,
		EventBus:  bus,
		Logger:    baseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("create rose: %w", err)
	}

	sunflower, err := agent.NewBaseAgent(agent.AgentConfig{
		Role:      "Sunflower",
		Goal:      "表达向日葵的日照需求并提出整体布局思路",
		Backstory: "你是向日葵，需充足日照，耐旱，常作为高杆背景与授粉引导。",
		LLM:       gardenLLM,
		EventBus:  bus,
		Logger:    baseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("create sunflower: %w", err)
	}

	lavender, err := agent.NewBaseAgent(agent.AgentConfig{
		Role:      "Lavender",
		Goal:      "表达薰衣草的排水与日照偏好，并提出共植益处",
		Backstory: "你是薰衣草，喜排水良好、日照充足，且有一定驱虫帮助。",
		LLM:       gardenLLM,
		EventBus:  bus,
		Logger:    baseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("create lavender: %w", err)
	}

	lily, err := agent.NewBaseAgent(agent.AgentConfig{
		Role:      "Lily",
		Goal:      "表达百合的半阴偏好与养分诉求",
		Backstory: "你是百合，偏好半阴或散射光环境，土壤养分需求中等。",
		LLM:       gardenLLM,
		EventBus:  bus,
		Logger:    baseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("create lily: %w", err)
	}

	tulip, err := agent.NewBaseAgent(agent.AgentConfig{
		Role:      "Tulip",
		Goal:      "表达郁金香的季节性与球根管理要点",
		Backstory: "你是郁金香，球根类，适合季节性种植与花期错峰。",
		LLM:       gardenLLM,
		EventBus:  bus,
		Logger:    baseLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("create tulip: %w", err)
	}

	// 初始化Agents
	for _, ag := range []agent.Agent{rose, sunflower, lavender, lily, tulip} {
		if err := ag.Initialize(); err != nil {
			return nil, fmt.Errorf("initialize agent %s: %w", ag.GetRole(), err)
		}
	}

	// 创建Crew（顺序流程）
	c := crew.NewBaseCrew(&crew.CrewConfig{
		Name:    "GardenCrew",
		Process: crew.ProcessSequential,
		Verbose: true,
	}, bus, baseLogger)

	// 为对话增加实时感：以非阻塞方式收集“发言”，由独立协程按节奏输出
	utterances := make(chan string, 64)
	donePrint := make(chan struct{})
	go func() {
		defer close(donePrint)
		for msg := range utterances {
			fmt.Print(msg)
			if speakDelay > 0 {
				// 计算抖动：在 [-jitterPct, +jitterPct]% 范围内随机
				jitterFactor := 1.0
				if jitterPct > 0 {
					jitter := (rng.Float64()*2 - 1) * float64(jitterPct) / 100.0
					jitterFactor += jitter
					if jitterFactor < 0.0 {
						jitterFactor = 0.0
					}
				}
				time.Sleep(time.Duration(float64(speakDelay) * jitterFactor))
			}
		}
	}()

	_ = c.AddTaskCallback(func(ctx context.Context, t agent.Task, out *agent.TaskOutput) error {
		if out == nil {
			return nil
		}
		msg := fmt.Sprintf("💬 %s：%s\n", out.Agent, out.Raw)
		select {
		case utterances <- msg:
		default:
			// 丢弃过载的发言，保证不阻塞主流程
		}
		return nil
	})

	// 添加Agents
	for _, ag := range []agent.Agent{rose, sunflower, lavender, lily, tulip} {
		if err := c.AddAgent(ag); err != nil {
			return nil, fmt.Errorf("add agent %s: %w", ag.GetRole(), err)
		}
	}

	// 创建并分配任务（多轮顺序群聊）：
	// Round 1: 5个自我介绍
	tRose := agent.NewTaskWithOptions(
		"作为玫瑰，请介绍你的生长需求（土壤、光照、浇灌）与与邻居的相处建议。",
		"简要说明要点，并给1-2条与其他花共植建议。",
		agent.WithAssignedAgent(rose),
	)

	tSunflower := agent.NewTaskWithOptions(
		"作为向日葵，请介绍你的日照与土壤需求，并说明你作为高杆背景/授粉引导的作用。",
		"简要说明要点，并给1-2条与其他花共植建议。",
		agent.WithAssignedAgent(sunflower),
	)

	tLavender := agent.NewTaskWithOptions(
		"作为薰衣草，请介绍排水、日照偏好与可能的驱虫互助价值。",
		"简要说明要点，并给1-2条与其他花共植建议。",
		agent.WithAssignedAgent(lavender),
	)

	tLily := agent.NewTaskWithOptions(
		"作为百合，请介绍你的半阴偏好、土壤养分诉求与与邻居的相处建议。",
		"简要说明要点，并给1-2条与其他花共植建议。",
		agent.WithAssignedAgent(lily),
	)

	tTulip := agent.NewTaskWithOptions(
		"作为郁金香，请介绍你的季节性与球根管理要点，以及花期错峰建议。",
		"简要说明要点，并给1-2条与其他花共植建议。",
		agent.WithAssignedAgent(tulip),
	)

	// Round 2: 5个回应（基于 aggregated_context）
	r2Rose := agent.NewTaskWithOptions(
		"第二轮（玫瑰）：请基于 aggregated_context 回应他花建议，指出与你的友好或冲突点，并给出1条妥协方案。",
		"给出简要回应与妥协点。",
		agent.WithAssignedAgent(rose),
	)
	r2Sunflower := agent.NewTaskWithOptions(
		"第二轮（向日葵）：请基于 aggregated_context 回应他花建议，描述对光照分配和授粉动线的考量。",
		"给出简要回应与动线建议。",
		agent.WithAssignedAgent(sunflower),
	)
	r2Lavender := agent.NewTaskWithOptions(
		"第二轮（薰衣草）：请基于 aggregated_context 回应他花建议，提出与驱虫、排水相关的共植辅助。",
		"给出简要回应与共植辅助建议。",
		agent.WithAssignedAgent(lavender),
	)
	r2Lily := agent.NewTaskWithOptions(
		"第二轮（百合）：请基于 aggregated_context 回应他花建议，强调半阴分区与直射避让的细化。",
		"给出简要回应与细化条目。",
		agent.WithAssignedAgent(lily),
	)
	r2Tulip := agent.NewTaskWithOptions(
		"第二轮（郁金香）：请基于 aggregated_context 回应他花建议，给出季节轮作与错峰的补充。",
		"给出简要回应与季节轮作建议。",
		agent.WithAssignedAgent(tulip),
	)

	// Round 3: 5个总结（基于两轮讨论给出最终偏好/约束）
	r3Rose := agent.NewTaskWithOptions(
		"第三轮（玫瑰）：请给出你的最终偏好与约束清单（条目化，基于前两轮讨论）。",
		"输出若干约束与偏好条目。",
		agent.WithAssignedAgent(rose),
	)
	r3Sunflower := agent.NewTaskWithOptions(
		"第三轮（向日葵）：请给出你的最终偏好与约束清单（条目化，基于前两轮讨论）。",
		"输出若干约束与偏好条目。",
		agent.WithAssignedAgent(sunflower),
	)
	r3Lavender := agent.NewTaskWithOptions(
		"第三轮（薰衣草）：请给出你的最终偏好与约束清单（条目化，基于前两轮讨论）。",
		"输出若干约束与偏好条目。",
		agent.WithAssignedAgent(lavender),
	)
	r3Lily := agent.NewTaskWithOptions(
		"第三轮（百合）：请给出你的最终偏好与约束清单（条目化，基于前两轮讨论）。",
		"输出若干约束与偏好条目。",
		agent.WithAssignedAgent(lily),
	)
	r3Tulip := agent.NewTaskWithOptions(
		"第三轮（郁金香）：请给出你的最终偏好与约束清单（条目化，基于前两轮讨论）。",
		"输出若干约束与偏好条目。",
		agent.WithAssignedAgent(tulip),
	)

	// 综合协商与最终布局
	tNegotiate := agent.NewTaskWithOptions(
		"请根据三轮交流，综合提出资源分配与初步布局建议（分区、光照/排水/邻里）。",
		"输出建议要点清单。",
		agent.WithAssignedAgent(sunflower),
	)

	tFinal := agent.NewTaskWithOptions(
		"整合上述建议，输出最终花园布局，建议包含 JSON 字段：zones, neighbors, seasonal_plan, care_notes。",
		"输出 JSON 规划或清晰的结构化说明。",
		agent.WithAssignedAgent(rose),
	)

	// 组装回合任务，支持可选乱序
	r1Tasks := []agent.Task{tRose, tSunflower, tLavender, tLily, tTulip}
	r2Tasks := []agent.Task{r2Rose, r2Sunflower, r2Lavender, r2Lily, r2Tulip}
	r3Tasks := []agent.Task{r3Rose, r3Sunflower, r3Lavender, r3Lily, r3Tulip}

	if os.Getenv("GARDEN_SHUFFLE_ROUNDS") == "1" {
		shuffle := func(ts []agent.Task) {
			rng.Shuffle(len(ts), func(i, j int) { ts[i], ts[j] = ts[j], ts[i] })
		}
		shuffle(r1Tasks)
		shuffle(r2Tasks)
		shuffle(r3Tasks)
	}

	// 添加全部任务：5（R1）+5（R2）+5（R3）+2（协商与最终）= 17
	allTasks := make([]agent.Task, 0, 17)
	allTasks = append(allTasks, r1Tasks...)
	allTasks = append(allTasks, r2Tasks...)
	allTasks = append(allTasks, r3Tasks...)
	allTasks = append(allTasks, tNegotiate, tFinal)
	for _, t := range allTasks {
		if err := c.AddTask(t); err != nil {
			return nil, fmt.Errorf("add task: %w", err)
		}
	}

	// 统一设置上下文输入
	inputs := map[string]interface{}{
		"garden_size": "10x10m",
		"location":    "温带",
		"drainage":    "中-良好",
		"constraints": "避免强风直吹；局部有半阴区域",
		"preferences": "四季有景、花香为主、兼顾授粉与生态",
	}

	// 带超时执行，确保测试稳定
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
	}

	output, err := c.Kickoff(ctx, inputs)
	if err != nil {
		close(utterances)
		<-donePrint
		return nil, err
	}
	// 关闭并等待打印完成
	close(utterances)
	<-donePrint
	return output, nil
}

// silentLogger 把内部实现日志静默，避免暴露实现细节，提升“对话感”。
type silentLogger struct{}

func (s *silentLogger) Debug(msg string, fields ...logger.Field) {}
func (s *silentLogger) Info(msg string, fields ...logger.Field)  {}
func (s *silentLogger) Warn(msg string, fields ...logger.Field)  {}
func (s *silentLogger) Error(msg string, fields ...logger.Field) {}
func (s *silentLogger) Fatal(msg string, fields ...logger.Field) {}
