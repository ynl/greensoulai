package garden

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRunGarden_MinimalFlow(t *testing.T) {
	// 检查是否有API key，没有则跳过测试
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("跳过集成测试：需要OPENROUTER_API_KEY环境变量")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := RunGarden(ctx)
	if err != nil {
		t.Fatalf("RunGarden returned error: %v", err)
	}

	if out == nil {
		t.Fatalf("RunGarden returned nil output")
	}

    // 多轮顺序群聊：5（R1）+5（R2）+5（R3）+2 = 17
    if len(out.TasksOutput) != 17 {
        t.Fatalf("expected 17 task outputs, got %d", len(out.TasksOutput))
	}

	// 验证顺序流程合并Raw内容
	if out.Raw == "" {
		t.Fatalf("expected non-empty aggregated raw output")
	}

	// 基本成功标志
	if !out.Success {
		t.Fatalf("expected success=true, got false")
	}
}
