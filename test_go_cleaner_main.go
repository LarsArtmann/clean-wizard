package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LarsArtmann/clean-wizard/internal/testhelper"
)

func main() {
	ctx := context.Background()

	if err := testhelper.GoCleanerTest(ctx, "Go Cache Cleaner Test"); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}
