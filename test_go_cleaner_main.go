package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/testhelper"
)

func main() {
	ctx := context.Background()

	err := testhelper.GoCleanerTest(ctx, "Go Cache Cleaner Test")
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}
