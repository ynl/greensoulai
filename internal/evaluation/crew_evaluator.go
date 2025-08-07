package evaluation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// CrewEvaluatorImpl CrewEvaluator接口的实现，对应Python版本的CrewEvaluator类
// 负责评估crew的整体性能和任务执行质量
type CrewEvaluatorImpl struct {
	crew              Crew              // 被评估的crew
	llm               llm.LLM           // 评估用的LLM
	config            *EvaluationConfig // 评估配置
	eventBus          events.EventBus   // 事件总线
	logger            logger.Logger     // 日志器
	tasksScores       map[int][]float64 // 任务评分，按迭代分组
	runExecutionTimes map[int][]float64 // 执行时间，按迭代分组
	iteration         int               // 当前评估迭代次数
	mu                sync.RWMutex      // 并发安全锁
}

// NewCrewEvaluator 创建新的CrewEvaluator实例，对应Python版本的__init__()
func NewCrewEvaluator(
	crew Crew,
	evalLLM llm.LLM,
	config *EvaluationConfig,
	eventBus events.EventBus,
	logger logger.Logger,
) *CrewEvaluatorImpl {
	// 如果配置为空，使用默认配置
	if config == nil {
		config = DefaultEvaluationConfig()
	}

	evaluator := &CrewEvaluatorImpl{
		crew:              crew,
		llm:               evalLLM,
		config:            config,
		eventBus:          eventBus,
		logger:            logger,
		tasksScores:       make(map[int][]float64),
		runExecutionTimes: make(map[int][]float64),
		iteration:         0,
	}

	return evaluator
}

// SetupForEvaluating 设置crew进行评估，对应Python版本的_setup_for_evaluating()
func (ce *CrewEvaluatorImpl) SetupForEvaluating(ctx context.Context, crew Crew) error {
	startTime := time.Now()

	if crew == nil {
		return NewEvaluationConfigError("crew", "", "crew cannot be nil")
	}

	ce.mu.Lock()
	ce.crew = crew
	ce.mu.Unlock()

	// 为每个任务设置评估回调
	tasks := crew.GetTasks()
	if len(tasks) == 0 {
		return NewEvaluationConfigError("crew.tasks", "", "crew has no tasks to evaluate")
	}

	// 发射setup开始事件
	if ce.eventBus != nil {
		ce.eventBus.Emit(ctx, ce, NewEvaluationStartedEvent(
			ce, "crew_setup", crew.GetName(), crew.GetName(), fmt.Sprintf("iteration_%d", ce.iteration), ce.config,
		))
	}

	ce.logger.Info("Setting up crew for evaluation",
		logger.Field{Key: "crew_name", Value: crew.GetName()},
		logger.Field{Key: "tasks_count", Value: len(tasks)},
		logger.Field{Key: "iteration", Value: ce.iteration},
	)

	// 为crew设置任务完成回调
	err := crew.SetTaskCallback(func(taskOutput *TaskOutput) {
		// 异步处理评估，避免阻塞主流程
		go func() {
			evalResult, evalErr := ce.Evaluate(ctx, taskOutput)
			if evalErr != nil {
				ce.logger.Error("Task evaluation failed",
					logger.Field{Key: "task_id", Value: taskOutput.TaskID},
					logger.Field{Key: "error", Value: evalErr.Error()},
				)
			} else {
				ce.logger.Debug("Task evaluation completed",
					logger.Field{Key: "task_id", Value: taskOutput.TaskID},
					logger.Field{Key: "quality", Value: evalResult.Quality},
				)
			}
		}()
	})

	if err != nil {
		ce.logger.Error("Failed to set task callback", logger.Field{Key: "error", Value: err.Error()})
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "crew_setup", crew.GetName(), crew.GetName(), fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "callback_setup", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return NewEvaluationExecutionError("callback_setup", crew.GetName(), "crew", ce.iteration, 0, err)
	}

	executionTime := float64(time.Since(startTime).Milliseconds())
	ce.logger.Info("Crew evaluation setup completed",
		logger.Field{Key: "crew_name", Value: crew.GetName()},
		logger.Field{Key: "setup_time_ms", Value: executionTime},
	)

	if ce.eventBus != nil {
		ce.eventBus.Emit(ctx, ce, NewEvaluationCompletedEvent(
			ce, "crew_setup", crew.GetName(), crew.GetName(), fmt.Sprintf("iteration_%d", ce.iteration),
			10.0, "A", executionTime, true,
		))
	}

	return nil
}

// Evaluate 评估任务输出，对应Python版本的evaluate()方法
func (ce *CrewEvaluatorImpl) Evaluate(ctx context.Context, taskOutput *TaskOutput) (*TaskEvaluationPydanticOutput, error) {
	startTime := time.Now()

	if taskOutput == nil {
		return nil, NewTaskOutputError("", "validation", "task output cannot be nil", ErrTaskOutputEmpty)
	}

	ce.logger.Debug("Starting task evaluation",
		logger.Field{Key: "task_id", Value: taskOutput.TaskID},
		logger.Field{Key: "agent", Value: taskOutput.Agent},
		logger.Field{Key: "description", Value: taskOutput.Description},
	)

	// 发射任务评估开始事件
	if ce.eventBus != nil {
		ce.eventBus.Emit(ctx, ce, NewTaskEvaluationStartedEvent(
			ce, taskOutput.TaskID, taskOutput.Description, taskOutput.Agent, fmt.Sprintf("iteration_%d", ce.iteration),
		))
	}

	// 1. 查找对应的任务
	currentTask, err := ce.findTaskByOutput(taskOutput)
	if err != nil {
		ce.logger.Error("Failed to find task for output",
			logger.Field{Key: "task_id", Value: taskOutput.TaskID},
			logger.Field{Key: "error", Value: err.Error()},
		)
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "task_lookup", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return nil, NewTaskOutputError(taskOutput.TaskID, "task_lookup", "failed to find corresponding task", err)
	}

	// 2. 创建评估代理
	evaluatorAgent, err := ce.createEvaluatorAgent(ctx)
	if err != nil {
		ce.logger.Error("Failed to create evaluator agent",
			logger.Field{Key: "error", Value: err.Error()},
		)
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "agent_creation", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return nil, NewEvaluatorCreationError("evaluator_agent", "task_evaluation", "failed to create evaluator agent", err)
	}

	// 3. 创建评估任务
	evaluationTask, err := ce.createEvaluationTask(ctx, evaluatorAgent, currentTask, taskOutput.Raw)
	if err != nil {
		ce.logger.Error("Failed to create evaluation task",
			logger.Field{Key: "error", Value: err.Error()},
		)
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "task_creation", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return nil, NewEvaluatorCreationError("evaluation_task", "task_evaluation", "failed to create evaluation task", err)
	}

	// 4. 执行评估
	evaluationResult, err := evaluationTask.ExecuteSync(ctx)
	if err != nil {
		ce.logger.Error("Failed to execute evaluation task",
			logger.Field{Key: "error", Value: err.Error()},
		)
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "task_execution", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return nil, NewEvaluationExecutionError("task_execution", taskOutput.TaskID, "task", ce.iteration, 0, err)
	}

	// 5. 解析评估结果
	pydanticOutput, err := ce.parseEvaluationResult(evaluationResult)
	if err != nil {
		ce.logger.Error("Failed to parse evaluation result",
			logger.Field{Key: "error", Value: err.Error()},
		)
		if ce.eventBus != nil {
			ce.eventBus.Emit(ctx, ce, NewEvaluationFailedEvent(
				ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
				err.Error(), "result_parsing", float64(time.Since(startTime).Milliseconds()),
			))
		}
		return nil, NewTaskOutputError(taskOutput.TaskID, "result_parsing", "failed to parse evaluation result", err)
	}

	// 6. 记录评估结果
	ce.recordEvaluationResult(pydanticOutput.Quality, currentTask.GetExecutionDuration())

	executionTime := float64(time.Since(startTime).Milliseconds())

	// 发射评估完成事件
	if ce.eventBus != nil {
		// 发射CrewTestResultEvent（对应Python版本）
		ce.eventBus.Emit(ctx, ce, NewCrewTestResultEvent(
			ce, pydanticOutput.Quality, currentTask.GetExecutionDuration().Seconds()*1000, // 转换为毫秒
			ce.llm.GetModel(), ce.crew.GetName(), ce.iteration, taskOutput.TaskID, taskOutput.Agent,
		))

		// 发射任务评估完成事件
		ce.eventBus.Emit(ctx, ce, NewTaskEvaluatedEvent(
			ce, taskOutput.TaskID, taskOutput.Description, taskOutput.Agent, pydanticOutput.Quality,
			pydanticOutput, executionTime, fmt.Sprintf("iteration_%d", ce.iteration),
		))

		// 发射通用评估完成事件
		grade := ce.getGradeFromScore(pydanticOutput.Quality)
		ce.eventBus.Emit(ctx, ce, NewEvaluationCompletedEvent(
			ce, "task_evaluation", taskOutput.TaskID, taskOutput.Description, fmt.Sprintf("iteration_%d", ce.iteration),
			pydanticOutput.Quality, grade, executionTime, true,
		))
	}

	ce.logger.Info("Task evaluation completed successfully",
		logger.Field{Key: "task_id", Value: taskOutput.TaskID},
		logger.Field{Key: "quality", Value: pydanticOutput.Quality},
		logger.Field{Key: "execution_time_ms", Value: executionTime},
	)

	return pydanticOutput, nil
}

// SetIteration 设置评估迭代次数
func (ce *CrewEvaluatorImpl) SetIteration(iteration int) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.iteration = iteration
}

// GetIteration 获取当前评估迭代次数
func (ce *CrewEvaluatorImpl) GetIteration() int {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.iteration
}

// PrintCrewEvaluationResult 打印crew评估结果
func (ce *CrewEvaluatorImpl) PrintCrewEvaluationResult(ctx context.Context) error {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if len(ce.tasksScores) == 0 {
		ce.logger.Info("No evaluation results to display")
		return nil
	}

	// 生成评估报告
	result := ce.generateEvaluationResult()
	result.CalculateStats()

	// 输出评估结果表格
	ce.logger.Info("=== Crew Evaluation Results ===")
	ce.logger.Info("Crew Performance Summary",
		logger.Field{Key: "crew_name", Value: result.CrewName},
		logger.Field{Key: "total_tasks", Value: result.TotalTasks},
		logger.Field{Key: "passed_tasks", Value: result.PassedTasks},
		logger.Field{Key: "success_rate", Value: fmt.Sprintf("%.1f%%", result.SuccessRate)},
		logger.Field{Key: "average_score", Value: fmt.Sprintf("%.2f", result.AverageScore)},
		logger.Field{Key: "average_time", Value: fmt.Sprintf("%.2f ms", result.AverageTime)},
		logger.Field{Key: "performance_grade", Value: result.GetPerformanceGrade()},
	)

	// 按迭代显示详细结果
	for iteration, scores := range ce.tasksScores {
		executionTimes, hasExecTimes := ce.runExecutionTimes[iteration]
		ce.logger.Info("Iteration Results",
			logger.Field{Key: "iteration", Value: iteration},
			logger.Field{Key: "tasks_count", Value: len(scores)},
			logger.Field{Key: "scores", Value: scores},
		)
		if hasExecTimes && len(executionTimes) == len(scores) {
			ce.logger.Info("Execution Times",
				logger.Field{Key: "iteration", Value: iteration},
				logger.Field{Key: "times_ms", Value: executionTimes},
			)
		}
	}

	return nil
}

// GetEvaluationResult 获取评估结果
func (ce *CrewEvaluatorImpl) GetEvaluationResult() *CrewEvaluationResult {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	result := ce.generateEvaluationResult()
	result.CalculateStats()
	return result
}

// GetTasksScores 获取任务评分
func (ce *CrewEvaluatorImpl) GetTasksScores() map[int][]float64 {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	// 深拷贝避免外部修改
	result := make(map[int][]float64)
	for k, v := range ce.tasksScores {
		result[k] = append([]float64{}, v...)
	}
	return result
}

// GetExecutionTimes 获取执行时间
func (ce *CrewEvaluatorImpl) GetExecutionTimes() map[int][]float64 {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	// 深拷贝避免外部修改
	result := make(map[int][]float64)
	for k, v := range ce.runExecutionTimes {
		result[k] = append([]float64{}, v...)
	}
	return result
}

// Reset 重置评估状态
func (ce *CrewEvaluatorImpl) Reset() error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.tasksScores = make(map[int][]float64)
	ce.runExecutionTimes = make(map[int][]float64)
	ce.iteration = 0

	ce.logger.Info("Crew evaluator state reset")
	return nil
}

// SetConfig 设置评估配置
func (ce *CrewEvaluatorImpl) SetConfig(config *EvaluationConfig) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.config = config
}

// GetConfig 获取评估配置
func (ce *CrewEvaluatorImpl) GetConfig() *EvaluationConfig {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.config
}

// ===== 私有辅助方法 =====

// findTaskByOutput 根据任务输出查找对应的任务
func (ce *CrewEvaluatorImpl) findTaskByOutput(taskOutput *TaskOutput) (Task, error) {
	tasks := ce.crew.GetTasks()

	// 优先根据TaskID查找
	if taskOutput.TaskID != "" {
		for _, task := range tasks {
			if task.GetID() == taskOutput.TaskID {
				return task, nil
			}
		}
	}

	// 根据Description查找
	for _, task := range tasks {
		if task.GetDescription() == taskOutput.Description {
			return task, nil
		}
	}

	return nil, NewTaskOutputError(taskOutput.TaskID, "task_lookup",
		fmt.Sprintf("no task found for output with ID '%s' and description '%s'",
			taskOutput.TaskID, taskOutput.Description), ErrTaskNotFound)
}

// createEvaluatorAgent 创建评估代理，对应Python版本的_evaluator_agent()
func (ce *CrewEvaluatorImpl) createEvaluatorAgent(ctx context.Context) (agent.Agent, error) {
	// 这里需要依赖agent.Agent的实现，暂时返回一个简化的Mock Agent
	// 在实际实现中，需要创建一个真正的Agent实例

	/* 对应Python版本的评估代理配置 - 待Agent系统实现后启用
	agentConfig := map[string]interface{}{
		"role": "Task Execution Evaluator",
		"goal": "Your goal is to evaluate the performance of the agents in the crew based on the tasks they have performed using score from 1 to 10 evaluating on completion, quality, and overall performance.",
		"backstory": "Evaluator agent for crew evaluation with precise capabilities to evaluate the performance of the agents in the crew based on the tasks they have performed",
		"llm": ce.llm,
		"verbose": ce.config.EnableVerbose,
		"max_iter": 10,
		"temperature": 0.1, // 保持评估一致性
	}
	*/

	// TODO: 这里需要调用真正的Agent工厂方法创建Agent
	// 目前返回一个错误，提示需要实现Agent系统
	return nil, fmt.Errorf("agent system not fully implemented - cannot create evaluator agent")
}

// createEvaluationTask 创建评估任务，对应Python版本的_evaluation_task()
func (ce *CrewEvaluatorImpl) createEvaluationTask(ctx context.Context, evaluatorAgent agent.Agent, taskToEvaluate Task, taskOutput string) (Task, error) {
	/* 构建评估任务描述，与Python版本保持一致 - 待Task系统实现后启用
		description := fmt.Sprintf(`
	Analyze the task execution and provide an evaluation focusing on completion, quality, and overall performance.

	Task Details:
	- Description: %s
	- Expected Output: %s
	- Agent: %s
	- Actual Output: %s

	Please provide an evaluation with a score from 1 to 10, where:
	- 1-3: Poor performance, significant issues
	- 4-6: Average performance, some issues
	- 7-8: Good performance, minor issues
	- 9-10: Excellent performance, meets or exceeds expectations

	Return the result in JSON format: {"quality": <score>}`,
			taskToEvaluate.GetDescription(),
			taskToEvaluate.GetExpectedOutput(),
			taskToEvaluate.GetAgent().GetRole(),
			taskOutput,
		)

		expectedOutput := `JSON object with quality score: {"quality": <numeric_score>}`
	*/

	// TODO: 这里需要调用真正的Task工厂方法创建Task
	// 目前返回一个错误，提示需要实现Task系统
	return nil, fmt.Errorf("task system not fully implemented - cannot create evaluation task")
}

// parseEvaluationResult 解析评估结果
func (ce *CrewEvaluatorImpl) parseEvaluationResult(taskOutput *TaskOutput) (*TaskEvaluationPydanticOutput, error) {
	if taskOutput == nil {
		return nil, NewTaskOutputError("", "result_parsing", "task output is nil", ErrTaskOutputEmpty)
	}

	result := &TaskEvaluationPydanticOutput{}

	// 优先从JSONDict解析
	if taskOutput.JSONDict != nil {
		if quality, ok := taskOutput.JSONDict["quality"]; ok {
			if qualityFloat, ok := quality.(float64); ok {
				result.Quality = qualityFloat
				return result, nil
			}
		}
	}

	// 从Pydantic模型解析
	if taskOutput.Pydantic != nil {
		if pydanticOutput, ok := taskOutput.Pydantic.(*TaskEvaluationPydanticOutput); ok {
			return pydanticOutput, nil
		}
	}

	// 从Raw文本解析JSON
	if taskOutput.Raw != "" {
		err := result.FromJSON(taskOutput.Raw)
		if err != nil {
			return nil, NewTaskOutputError("", "result_parsing",
				fmt.Sprintf("failed to parse quality score from raw output: %s", taskOutput.Raw), err)
		}
		return result, nil
	}

	return nil, NewTaskOutputError("", "result_parsing", "no valid evaluation result found in task output", ErrTaskOutputInvalid)
}

// recordEvaluationResult 记录评估结果
func (ce *CrewEvaluatorImpl) recordEvaluationResult(quality float64, executionDuration time.Duration) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	// 验证评分范围
	if quality < 0 || quality > 10 {
		ce.logger.Warn("Invalid quality score, clamping to valid range",
			logger.Field{Key: "original_score", Value: quality},
		)
		if quality < 0 {
			quality = 0
		} else if quality > 10 {
			quality = 10
		}
	}

	// 记录评分
	if ce.tasksScores[ce.iteration] == nil {
		ce.tasksScores[ce.iteration] = make([]float64, 0)
	}
	ce.tasksScores[ce.iteration] = append(ce.tasksScores[ce.iteration], quality)

	// 记录执行时间（转换为毫秒）
	executionTimeMs := float64(executionDuration.Nanoseconds()) / 1e6
	if ce.runExecutionTimes[ce.iteration] == nil {
		ce.runExecutionTimes[ce.iteration] = make([]float64, 0)
	}
	ce.runExecutionTimes[ce.iteration] = append(ce.runExecutionTimes[ce.iteration], executionTimeMs)
}

// generateEvaluationResult 生成评估结果
func (ce *CrewEvaluatorImpl) generateEvaluationResult() *CrewEvaluationResult {
	return &CrewEvaluationResult{
		CrewName:       ce.crew.GetName(),
		Iteration:      ce.iteration,
		TasksScores:    ce.copyTasksScores(),
		ExecutionTimes: ce.copyExecutionTimes(),
		ModelUsed:      ce.llm.GetModel(),
		EvaluatedAt:    time.Now(),
	}
}

// copyTasksScores 深拷贝任务评分
func (ce *CrewEvaluatorImpl) copyTasksScores() map[int][]float64 {
	result := make(map[int][]float64)
	for k, v := range ce.tasksScores {
		result[k] = append([]float64{}, v...)
	}
	return result
}

// copyExecutionTimes 深拷贝执行时间
func (ce *CrewEvaluatorImpl) copyExecutionTimes() map[int][]float64 {
	result := make(map[int][]float64)
	for k, v := range ce.runExecutionTimes {
		result[k] = append([]float64{}, v...)
	}
	return result
}

// getGradeFromScore 根据分数获取等级
func (ce *CrewEvaluatorImpl) getGradeFromScore(score float64) string {
	switch {
	case score >= 9.0:
		return "A+"
	case score >= 8.0:
		return "A"
	case score >= 7.0:
		return "B"
	case score >= 6.0:
		return "C"
	case score >= 5.0:
		return "D"
	default:
		return "F"
	}
}
