package main

import (
	"context"
	"fmt"
	"time"

	garden "github.com/ynl/greensoulai/examples/garden"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	out, err := garden.RunGarden(ctx)
	if err != nil {
		fmt.Println("❌ RunGarden error:", err)
		return
	}

	fmt.Printf("\n✅ Garden run success: tasks=%d\n\n", len(out.TasksOutput))
	fmt.Println("--- Aggregated Output (Raw) ---")
	fmt.Println(out.Raw)
}
