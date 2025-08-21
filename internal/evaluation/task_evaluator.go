package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// TaskEvaluatorImpl TaskEvaluator接口的实现，对应Python版本的TaskEvaluator类
// 负责评估单个任务的执行质量和性能
type TaskEvaluatorImpl struct {
	originalAgent agent.Agent       // 原始执行任务的agent
	llm           llm.LLM           // 评估用的LLM
	config        *EvaluationConfig // 评估配置
	eventBus      events.EventBus   // 事件总线
	logger        logger.Logger     // 日志器
	mu            sync.RWMutex      // 并发安全锁
}

// NewTaskEvaluator 创建新的TaskEvaluator实例，对应Python版本的__init__()
func NewTaskEvaluator(
	originalAgent agent.Agent,
	evalLLM llm.LLM,
	config *EvaluationConfig,
	eventBus events.EventBus,
	logger logger.Logger,
) *TaskEvaluatorImpl {
	// 如果配置为空，使用默认配置
	if config == nil {
		config = DefaultEvaluationConfig()
	}

	// 如果没有提供评估LLM，使用原始agent的LLM
	if evalLLM == nil && originalAgent != nil {
		evalLLM = originalAgent.GetLLM()
	}

	return &TaskEvaluatorImpl{
		originalAgent: originalAgent,
		llm:           evalLLM,
		config:        config,
		eventBus:      eventBus,
		logger:        logger,
	}
}

// Evaluate 评估任务执行结果，对应Python版本的evaluate()方法
func (te *TaskEvaluatorImpl) Evaluate(ctx context.Context, task Task, output string) (*TaskEvaluation, error) {
	startTime := time.Now()

	if task == nil {
		return nil, NewTaskOutputError("", "validation", "task cannot be nil", ErrTaskNotFound)
	}
	if output == "" {
		return nil, NewTaskOutputError(task.GetID(), "validation", "output cannot be empty", ErrTaskOutputEmpty)
	}

	te.logger.Debug("Starting task evaluation",
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "task_description", Value: task.GetDescription()},
		logger.Field{Key: "output_length", Value: len(output)},
	)

	// 发射任务评估开始事件
	if te.eventBus != nil {
		te.eventBus.Emit(ctx, te, NewTaskEvaluationStartedEvent(
			te, task.GetID(), task.GetDescription(), task.GetAgent().GetRole(), fmt.Sprintf("eval_%d", time.Now().Unix()),
		))
	}

	// 构建评估查询，对应Python版本的evaluation_query
	evaluationQuery := te.buildEvaluationQuery(task, output)

	// 构建指令
	instructions := "Convert all responses into valid JSON output."

	// 检查LLM是否支持函数调用
	supportsFunctionCalling := false
	if te.llm != nil {
		supportsFunctionCalling = te.llm.SupportsFunctionCalling()
	}

	if !supportsFunctionCalling {
		// 如果不支持函数调用，添加JSON模式说明
		schema := te.getTaskEvaluationSchema()
		instructions = fmt.Sprintf("%s\n\nReturn only valid JSON with the following schema:\n```json\n%s\n```", instructions, schema)
	}

	// 执行LLM调用进行评估
	evaluation, err := te.executeEvaluationLLMCall(ctx, evaluationQuery, instructions)
	if err != nil {
		executionTime := float64(time.Since(startTime).Milliseconds())
		te.logger.Error("Task evaluation failed",
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "execution_time_ms", Value: executionTime},
		)

		if te.eventBus != nil {
			te.eventBus.Emit(ctx, te, NewEvaluationFailedEvent(
				te, "task_evaluation", task.GetID(), task.GetDescription(), fmt.Sprintf("eval_%d", time.Now().Unix()),
				err.Error(), "llm_execution", executionTime,
			))
		}

		return nil, NewLLMResponseError(te.llm.GetModel(), evaluationQuery, "", "llm_call", err)
	}

	// 设置评估时间戳和元数据
	evaluation.Timestamp = time.Now()
	evaluation.ExecutionTimeMs = float64(time.Since(startTime).Milliseconds())
	evaluation.EvaluatorVersion = "1.0.0"
	if evaluation.Metadata == nil {
		evaluation.Metadata = make(map[string]interface{})
	}
	evaluation.Metadata["task_id"] = task.GetID()
	evaluation.Metadata["agent_role"] = task.GetAgent().GetRole()
	evaluation.Metadata["evaluation_model"] = te.llm.GetModel()

	te.logger.Info("Task evaluation completed successfully",
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "score", Value: evaluation.GetOverallScore()},
		logger.Field{Key: "grade", Value: evaluation.GetGrade()},
		logger.Field{Key: "execution_time_ms", Value: evaluation.ExecutionTimeMs},
	)

	// 发射任务评估完成事件
	if te.eventBus != nil {
		te.eventBus.Emit(ctx, te, NewEvaluationCompletedEvent(
			te, "task_evaluation", task.GetID(), task.GetDescription(), fmt.Sprintf("eval_%d", time.Now().Unix()),
			evaluation.GetOverallScore(), evaluation.GetGrade(), evaluation.ExecutionTimeMs, true,
		))
	}

	return evaluation, nil
}

// EvaluateTrainingData 评估训练数据，对应Python版本的evaluate_training_data()方法
func (te *TaskEvaluatorImpl) EvaluateTrainingData(ctx context.Context, trainingData map[string]interface{}, agentID string) (*TrainingTaskEvaluation, error) {
	startTime := time.Now()

	if trainingData == nil {
		return nil, NewTaskOutputError("", "validation", "training data cannot be nil", ErrInvalidDataFormat)
	}
	if agentID == "" {
		return nil, NewAgentEvaluationError("", "", "", "validation", "agent ID cannot be empty", ErrAgentNotFound)
	}

	te.logger.Debug("Starting training data evaluation",
		logger.Field{Key: "agent_id", Value: agentID},
		logger.Field{Key: "data_keys", Value: te.getMapKeys(trainingData)},
	)

	// 提取训练数据字段
	taskDescription, _ := trainingData["task_description"].(string)
	expectedOutput, _ := trainingData["expected_output"].(string)
	actualOutput, _ := trainingData["actual_output"].(string)
	iterationRaw := trainingData["iteration"]
	iteration := 0
	if iterationInt, ok := iterationRaw.(int); ok {
		iteration = iterationInt
	} else if iterationFloat, ok := iterationRaw.(float64); ok {
		iteration = int(iterationFloat)
	}

	// 构建训练数据评估查询
	evaluationQuery := te.buildTrainingDataEvaluationQuery(taskDescription, expectedOutput, actualOutput)

	// 执行评估
	evaluation, err := te.executeTrainingEvaluationLLMCall(ctx, evaluationQuery)
	if err != nil {
		// executionTime := float64(time.Since(startTime).Milliseconds())
		te.logger.Error("Training data evaluation failed",
			logger.Field{Key: "agent_id", Value: agentID},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, NewLLMResponseError(te.llm.GetModel(), evaluationQuery, "", "training_evaluation", err)
	}

	// 构建训练任务评估结果
	result := &TrainingTaskEvaluation{
		TaskID:          fmt.Sprintf("training_%s_%d", agentID, time.Now().Unix()),
		AgentRole:       te.originalAgent.GetRole(),
		TaskDescription: taskDescription,
		ExpectedOutput:  expectedOutput,
		ActualOutput:    actualOutput,
		Score:           evaluation.GetOverallScore(),
		Feedback:        evaluation.Feedback,
		Improvements:    evaluation.Suggestions,
		ExecutionTimeMs: float64(time.Since(startTime).Milliseconds()),
		Iteration:       iteration,
		Timestamp:       time.Now(),
		ModelUsed:       te.llm.GetModel(),
		Metadata:        make(map[string]interface{}),
	}

	// 设置元数据
	result.Metadata["agent_id"] = agentID
	result.Metadata["evaluation_type"] = "training_data"
	result.Metadata["completion_score"] = evaluation.CompletionScore
	result.Metadata["quality_score"] = evaluation.QualityScore
	result.Metadata["performance_score"] = evaluation.PerformanceScore

	te.logger.Info("Training data evaluation completed",
		logger.Field{Key: "agent_id", Value: agentID},
		logger.Field{Key: "task_id", Value: result.TaskID},
		logger.Field{Key: "score", Value: result.Score},
		logger.Field{Key: "iteration", Value: result.Iteration},
	)

	return result, nil
}

// SetOriginalAgent 设置原始agent
func (te *TaskEvaluatorImpl) SetOriginalAgent(agent agent.Agent) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.originalAgent = agent
}

// GetOriginalAgent 获取原始agent
func (te *TaskEvaluatorImpl) GetOriginalAgent() agent.Agent {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.originalAgent
}

// SetLLM 设置评估用的LLM
func (te *TaskEvaluatorImpl) SetLLM(llm llm.LLM) error {
	if llm == nil {
		return NewEvaluationConfigError("llm", "", "LLM cannot be nil")
	}

	te.mu.Lock()
	defer te.mu.Unlock()
	te.llm = llm
	return nil
}

// GetLLM 获取评估用的LLM
func (te *TaskEvaluatorImpl) GetLLM() llm.LLM {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.llm
}

// SetConfig 设置评估配置
func (te *TaskEvaluatorImpl) SetConfig(config *EvaluationConfig) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.config = config
}

// GetConfig 获取评估配置
func (te *TaskEvaluatorImpl) GetConfig() *EvaluationConfig {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.config
}

// ===== 私有辅助方法 =====

// buildEvaluationQuery 构建评估查询，对应Python版本的evaluation_query
func (te *TaskEvaluatorImpl) buildEvaluationQuery(task Task, output string) string {
	return fmt.Sprintf(`Assess the quality of the task completed based on the description, expected output, and actual results.

Task Description:
%s

Expected Output:
%s

Actual Output:
%s

Please provide:
- Bullet points suggestions to improve future similar tasks
- A score from 0 to 10 evaluating on completion, quality, and overall performance
- Entities extracted from the task output, if any, their type, description, and relationships

Please be thorough in your evaluation and provide constructive feedback.`,
		task.GetDescription(),
		task.GetExpectedOutput(),
		output,
	)
}

// buildTrainingDataEvaluationQuery 构建训练数据评估查询
func (te *TaskEvaluatorImpl) buildTrainingDataEvaluationQuery(taskDescription, expectedOutput, actualOutput string) string {
	return fmt.Sprintf(`Evaluate this training data sample for quality and effectiveness.

Task Description:
%s

Expected Output:
%s

Actual Output:
%s

Please provide:
- An overall score from 0 to 10
- Detailed feedback on the quality of the output
- Specific suggestions for improvement
- Assessment of how well the actual output meets the expected output

Focus on training effectiveness and learning potential.`,
		taskDescription,
		expectedOutput,
		actualOutput,
	)
}

// executeEvaluationLLMCall 执行评估LLM调用
func (te *TaskEvaluatorImpl) executeEvaluationLLMCall(ctx context.Context, query, instructions string) (*TaskEvaluation, error) {
	if te.llm == nil {
		return nil, NewEvaluationConfigError("llm", "", "LLM not configured for task evaluator")
	}

	// 构建完整的prompt
	fullPrompt := fmt.Sprintf("%s\n\nInstructions:\n%s", query, instructions)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are an expert task evaluator. Provide comprehensive, objective evaluations of task outputs.",
		},
		{
			Role:    "user",
			Content: fullPrompt,
		},
	}

	// 添加超时控制
	var cancel context.CancelFunc
	if te.config.TimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(te.config.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// 调用LLM
	response, err := te.llm.Call(ctx, messages, &llm.CallOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM for task evaluation: %w", err)
	}

	if response.Content == "" {
		return nil, NewLLMResponseError(te.llm.GetModel(), fullPrompt, "", "empty_response", ErrLLMResponseEmpty)
	}

	// 解析LLM响应为TaskEvaluation
	evaluation, err := te.parseLLMResponseToTaskEvaluation(response.Content)
	if err != nil {
		return nil, NewLLMResponseError(te.llm.GetModel(), fullPrompt, response.Content, "response_parsing", err)
	}

	return evaluation, nil
}

// executeTrainingEvaluationLLMCall 执行训练评估LLM调用
func (te *TaskEvaluatorImpl) executeTrainingEvaluationLLMCall(ctx context.Context, query string) (*TaskEvaluation, error) {
	instructions := "Provide a JSON response with score, feedback, and suggestions for training data evaluation."
	// evaluation, err := te.executeEvaluationLLMCall(ctx, query, instructions)
	return te.executeEvaluationLLMCall(ctx, query, instructions)
}

// parseLLMResponseToTaskEvaluation 解析LLM响应为TaskEvaluation
func (te *TaskEvaluatorImpl) parseLLMResponseToTaskEvaluation(response string) (*TaskEvaluation, error) {
	// evaluation := &TaskEvaluation{
	// 	Suggestions: make([]string, 0),
	// 	Entities:    make([]EntityExtraction, 0),
	// 	Metadata:    make(map[string]interface{}),
	// }

	// 尝试解析为JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
		// 成功解析为JSON
		return te.parseJSONToTaskEvaluation(jsonData)
	}

	// 如果不是JSON，尝试从文本中提取信息
	return te.parseTextToTaskEvaluation(response)
}

// parseJSONToTaskEvaluation 从JSON解析TaskEvaluation
func (te *TaskEvaluatorImpl) parseJSONToTaskEvaluation(jsonData map[string]interface{}) (*TaskEvaluation, error) {
	evaluation := &TaskEvaluation{
		Suggestions: make([]string, 0),
		Entities:    make([]EntityExtraction, 0),
		Metadata:    make(map[string]interface{}),
	}

	// 提取分数字段
	if score, ok := jsonData["score"].(float64); ok {
		evaluation.Score = score
	}
	if completionScore, ok := jsonData["completion_score"].(float64); ok {
		evaluation.CompletionScore = completionScore
	}
	if qualityScore, ok := jsonData["quality_score"].(float64); ok {
		evaluation.QualityScore = qualityScore
	}
	if performanceScore, ok := jsonData["performance_score"].(float64); ok {
		evaluation.PerformanceScore = performanceScore
	}

	// 提取反馈
	if feedback, ok := jsonData["feedback"].(string); ok {
		evaluation.Feedback = feedback
	}

	// 提取建议
	if suggestions, ok := jsonData["suggestions"].([]interface{}); ok {
		for _, suggestion := range suggestions {
			if suggestionStr, ok := suggestion.(string); ok {
				evaluation.Suggestions = append(evaluation.Suggestions, suggestionStr)
			}
		}
	}

	// 如果没有设置整体评分，计算一个默认值
	if evaluation.Score == 0 && (evaluation.CompletionScore > 0 || evaluation.QualityScore > 0 || evaluation.PerformanceScore > 0) {
		evaluation.Score = evaluation.GetOverallScore()
	} else if evaluation.Score == 0 {
		// 如果所有分数都为0，设置一个默认的中等分数
		evaluation.Score = 5.0
		evaluation.CompletionScore = 5.0
		evaluation.QualityScore = 5.0
		evaluation.PerformanceScore = 5.0
	}

	return evaluation, nil
}

// parseTextToTaskEvaluation 从文本解析TaskEvaluation
func (te *TaskEvaluatorImpl) parseTextToTaskEvaluation(response string) (*TaskEvaluation, error) {
	evaluation := &TaskEvaluation{
		Feedback:    response, // 将整个响应作为反馈
		Score:       5.0,      // 默认中等分数
		Suggestions: te.extractSuggestionsFromText(response),
		Entities:    make([]EntityExtraction, 0),
		Metadata:    make(map[string]interface{}),
	}

	// 尝试从文本中提取数值分数
	if score := te.extractScoreFromText(response); score > 0 {
		evaluation.Score = score
		// 设置子分数为相同值
		evaluation.CompletionScore = score
		evaluation.QualityScore = score
		evaluation.PerformanceScore = score
	} else {
		evaluation.CompletionScore = 5.0
		evaluation.QualityScore = 5.0
		evaluation.PerformanceScore = 5.0
	}

	return evaluation, nil
}

// extractSuggestionsFromText 从文本中提取建议
func (te *TaskEvaluatorImpl) extractSuggestionsFromText(text string) []string {
	suggestions := make([]string, 0)

	// 查找以 "-", "*", "•" 开头的行作为建议
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•") {
			suggestion := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*"), "•"))
			if suggestion != "" {
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

// extractScoreFromText 从文本中提取分数
func (te *TaskEvaluatorImpl) extractScoreFromText(text string) float64 {
	// 查找形如 "score: X" 或 "Score: X/10" 的模式
	text = strings.ToLower(text)

	// 简单的模式匹配
	patterns := []string{
		"score:",
		"rating:",
		"evaluation:",
		"quality:",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(text, pattern); idx != -1 {
			// 在找到的位置后寻找数字
			remaining := text[idx+len(pattern):]
			for i, r := range remaining {
				if r >= '0' && r <= '9' {
					// 找到数字，尝试解析
					var score float64
					n, err := fmt.Sscanf(remaining[i:], "%f", &score)
					if err == nil && n == 1 && score >= 0 && score <= 10 {
						return score
					}
				}
			}
		}
	}

	return 0 // 未找到有效分数
}

// getTaskEvaluationSchema 获取TaskEvaluation的JSON模式
func (te *TaskEvaluatorImpl) getTaskEvaluationSchema() string {
	schema := map[string]interface{}{
		"score":             "number (0-10)",
		"completion_score":  "number (0-10)",
		"quality_score":     "number (0-10)",
		"performance_score": "number (0-10)",
		"feedback":          "string",
		"suggestions":       []string{"suggestion1", "suggestion2"},
		"entities": []map[string]interface{}{
			{
				"name":        "entity_name",
				"type":        "entity_type",
				"description": "entity_description",
			},
		},
	}

	bytes, _ := json.MarshalIndent(schema, "", "  ")
	return string(bytes)
}

// getMapKeys 获取map的所有键
func (te *TaskEvaluatorImpl) getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
