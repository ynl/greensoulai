package garden

import (
	"context"
	"testing"
	"time"
)

func TestRunGarden_MinimalFlow(t *testing.T) {
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
