package result

import (
	"context"
	"errors"
	"testing"
)

func TestMatch(t *testing.T) {
	t.Parallel()
	t.Run("ok result calls ok function", func(t *testing.T) {
		t.Parallel()
		result := Ok(42)
		matched := Match(result,
			func(v int) string { return "success: " + string(rune(v)) },
			func(e error) string { return "error: " + e.Error() },
		)
		if matched != "success: *" {
			t.Errorf("Match() = %v, want success callback", matched)
		}
	})

	t.Run("error result calls err function", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("test error")
		result := Err[int](testErr)
		matched := Match(result,
			func(v int) string { return "success" },
			func(e error) string { return "error: " + e.Error() },
		)
		if matched != "error: test error" {
			t.Errorf("Match() = %v, want error callback", matched)
		}
	})
}

func TestTapErr(t *testing.T) {
	t.Parallel()
	t.Run("calls function on error", func(t *testing.T) {
		t.Parallel()
		var tappedError error
		testErr := errors.New("tapped error")
		result := Err[int](testErr).TapErr(func(e error) {
			tappedError = e
		})

		if !errors.Is(tappedError, testErr) {
			t.Errorf("TapErr() did not call function with correct error")
		}
		if !errors.Is(result.Error(), testErr) {
			t.Errorf("TapErr() should return original result")
		}
	})

	t.Run("does not call function on success", func(t *testing.T) {
		t.Parallel()
		called := false
		result := Ok(42).TapErr(func(e error) {
			called = true
		})

		if called {
			t.Errorf("TapErr() should not call function on success")
		}
		if result.Value() != 42 {
			t.Errorf("TapErr() should return original result")
		}
	})
}

func TestWhen(t *testing.T) {
	t.Parallel()
	t.Run("calls function on success", func(t *testing.T) {
		t.Parallel()
		var tappedValue int
		result := Ok(42).When(func(v int) {
			tappedValue = v
		})

		if tappedValue != 42 {
			t.Errorf("When() did not call function with correct value")
		}
		if result.Value() != 42 {
			t.Errorf("When() should return original result")
		}
	})

	t.Run("does not call function on error", func(t *testing.T) {
		t.Parallel()
		called := false
		result := Err[int](errors.New("test")).When(func(v int) {
			called = true
		})

		if called {
			t.Errorf("When() should not call function on error")
		}
		if result.IsOk() {
			t.Errorf("When() should return error result")
		}
	})
}

func TestUnless(t *testing.T) {
	t.Parallel()
	t.Run("calls function on error", func(t *testing.T) {
		t.Parallel()
		var tappedError error
		testErr := errors.New("test error")
		result := Err[int](testErr).Unless(func(e error) {
			tappedError = e
		})

		if !errors.Is(tappedError, testErr) {
			t.Errorf("Unless() did not call function with correct error")
		}
		if !errors.Is(result.Error(), testErr) {
			t.Errorf("Unless() should return original result")
		}
	})

	t.Run("does not call function on success", func(t *testing.T) {
		t.Parallel()
		called := false
		result := Ok(42).Unless(func(_ error) {
			called = true
		})

		if called {
			t.Errorf("Unless() should not call function on success")
		}
		if result.Value() != 42 {
			t.Errorf("Unless() should return original result")
		}
	})
}

func TestFilter(t *testing.T) {
	t.Parallel()
	t.Run("passes through on valid predicate", func(t *testing.T) {
		t.Parallel()
		result := Ok(42).Filter(func(v int) bool { return v > 0 }, "must be positive")
		if result.IsErr() {
			t.Errorf("Filter() should pass through on valid predicate")
		}
		if result.Value() != 42 {
			t.Errorf("Filter() should return original value")
		}
	})

	t.Run("returns error on invalid predicate", func(t *testing.T) {
		t.Parallel()
		result := Ok(-5).Filter(func(v int) bool { return v > 0 }, "must be positive")
		if result.IsOk() {
			t.Errorf("Filter() should return error on invalid predicate")
		}
		if result.Error().Error() != "must be positive" {
			t.Errorf("Filter() should return correct error message")
		}
	})

	t.Run("passes through existing error", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("original error")
		result := Err[int](testErr).Filter(func(v int) bool { return v > 0 }, "must be positive")
		if result.IsOk() {
			t.Errorf("Filter() should pass through existing error")
		}
		if !errors.Is(result.Error(), testErr) {
			t.Errorf("Filter() should preserve original error")
		}
	})
}

func TestFold(t *testing.T) {
	t.Parallel()
	t.Run("reduces all results successfully", func(t *testing.T) {
		t.Parallel()
		results := []Result[int]{Ok(1), Ok(2), Ok(3)}
		sum := Fold(results, 0, func(acc, val int) int { return acc + val })

		if sum.IsErr() {
			t.Errorf("Fold() should not return error")
		}
		if sum.Value() != 6 {
			t.Errorf("Fold() = %v, want 6", sum.Value())
		}
	})

	t.Run("short-circuits on error", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("fold error")
		results := []Result[int]{Ok(1), Err[int](testErr), Ok(3)}
		sum := Fold(results, 0, func(acc, val int) int { return acc + val })

		if sum.IsOk() {
			t.Errorf("Fold() should return error on short-circuit")
		}
		if !errors.Is(sum.Error(), testErr) {
			t.Errorf("Fold() should return first error")
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		t.Parallel()
		results := []Result[int]{}
		sum := Fold(results, 10, func(acc, val int) int { return acc + val })

		if sum.IsErr() {
			t.Errorf("Fold() should not error on empty slice")
		}
		if sum.Value() != 10 {
			t.Errorf("Fold() = %v, want initial value 10", sum.Value())
		}
	})
}

func TestFoldAll(t *testing.T) {
	t.Parallel()
	t.Run("collects all values", func(t *testing.T) {
		t.Parallel()
		results := []Result[int]{Ok(1), Ok(2), Ok(3)}
		all := FoldAll(results)

		if all.IsErr() {
			t.Errorf("FoldAll() should not return error")
		}
		values := all.Value()
		if len(values) != 3 {
			t.Errorf("FoldAll() len = %v, want 3", len(values))
		}
	})

	t.Run("short-circuits on error", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("fold all error")
		results := []Result[int]{Ok(1), Err[int](testErr), Ok(3)}
		all := FoldAll(results)

		if all.IsOk() {
			t.Errorf("FoldAll() should return error")
		}
	})
}

func TestPartition(t *testing.T) {
	t.Parallel()
	t.Run("separates ok and error results", func(t *testing.T) {
		t.Parallel()
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		results := []Result[int]{Ok(1), Err[int](err1), Ok(2), Err[int](err2), Ok(3)}
		ok, errs := Partition(results)

		if len(ok) != 3 {
			t.Errorf("Partition() ok len = %v, want 3", len(ok))
		}
		if len(errs) != 2 {
			t.Errorf("Partition() errs len = %v, want 2", len(errs))
		}
	})

	t.Run("handles all ok", func(t *testing.T) {
		t.Parallel()
		results := []Result[int]{Ok(1), Ok(2), Ok(3)}
		ok, errs := Partition(results)

		if len(ok) != 3 {
			t.Errorf("Partition() ok len = %v, want 3", len(ok))
		}
		if len(errs) != 0 {
			t.Errorf("Partition() errs len = %v, want 0", len(errs))
		}
	})
}

func TestSequence(t *testing.T) {
	t.Parallel()
	t.Run("sequences all results", func(t *testing.T) {
		t.Parallel()
		results := []Result[int]{Ok(1), Ok(2), Ok(3)}
		seq := Sequence(results)

		if seq.IsErr() {
			t.Errorf("Sequence() should not return error")
		}
		values := seq.Value()
		if len(values) != 3 {
			t.Errorf("Sequence() len = %v, want 3", len(values))
		}
	})
}

func TestTraverse(t *testing.T) {
	t.Parallel()
	t.Run("applies function and sequences results", func(t *testing.T) {
		t.Parallel()
		items := []int{1, 2, 3}
		result := Traverse(items, func(i int) Result[int] {
			return Ok(i * 2)
		})

		if result.IsErr() {
			t.Errorf("Traverse() should not return error")
		}
		values := result.Value()
		if len(values) != 3 {
			t.Errorf("Traverse() len = %v, want 3", len(values))
		}
	})
}

func TestBranchFlow(t *testing.T) {
	t.Parallel()
	t.Run("executes matching branch", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			Branch(func() bool { return false }, func() Result[int] { return Ok(100) }).
			Branch(func() bool { return true }, func() Result[int] { return Ok(42) }).
			Branch(func() bool { return true }, func() Result[int] { return Ok(99) })

		result := flow.Execute()

		if result.IsErr() {
			t.Errorf("BranchFlow() should not error")
		}
		if result.Value() != 42 {
			t.Errorf("BranchFlow() = %v, want 42 (first matching branch)", result.Value())
		}
	})

	t.Run("executes fallback when no branch matches", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			Branch(func() bool { return false }, func() Result[int] { return Ok(100) }).
			Fallback(func() Result[int] { return Ok(999) })

		result := flow.Execute()

		if result.IsErr() {
			t.Errorf("BranchFlow() should use fallback")
		}
		if result.Value() != 999 {
			t.Errorf("BranchFlow() = %v, want 999 (fallback)", result.Value())
		}
	})

	t.Run("returns error when no branch matches and no fallback", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			Branch(func() bool { return false }, func() Result[int] { return Ok(100) })

		result := flow.Execute()

		if result.IsOk() {
			t.Errorf("BranchFlow() should return error")
		}
	})

	t.Run("applies finalizer", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			Branch(func() bool { return true }, func() Result[int] { return Ok(42) }).
			Finalize(func(r Result[int]) Result[int] {
				if r.IsOk() {
					return Ok(r.Value() * 2)
				}
				return r
			})

		result := flow.Execute()

		if result.Value() != 84 {
			t.Errorf("BranchFlow() with finalizer = %v, want 84", result.Value())
		}
	})

	t.Run("BranchWithValue executes when predicate matches", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			BranchWithValue(10, func(v int) bool { return v > 5 }, func(v int) Result[int] { return Ok(v * 2) }).
			FallbackValue(0)

		result := flow.Execute()

		if result.IsErr() {
			t.Errorf("BranchWithValue() should not error")
		}
		if result.Value() != 20 {
			t.Errorf("BranchWithValue() = %v, want 20", result.Value())
		}
	})

	t.Run("BranchWithValue falls back when predicate does not match", func(t *testing.T) {
		t.Parallel()
		flow := NewBranchFlow[int]().
			BranchWithValue(3, func(v int) bool { return v > 5 }, func(v int) Result[int] { return Ok(v * 2) }).
			FallbackValue(100)

		result := flow.Execute()

		if result.IsErr() {
			t.Errorf("BranchWithValue() should not error")
		}
		if result.Value() != 100 {
			t.Errorf("BranchWithValue() = %v, want 100 (fallback)", result.Value())
		}
	})
}

func TestSwitchFlow(t *testing.T) {
	t.Parallel()
	t.Run("matches first predicate", func(t *testing.T) {
		t.Parallel()
		value := 5
		cases := []Case[int, string]{
			{
				Predicate: func(i int) bool { return i < 0 },
				Execute:   func() Result[string] { return Ok("negative") },
			},
			{
				Predicate: func(i int) bool { return i == 0 },
				Execute:   func() Result[string] { return Ok("zero") },
			},
			{
				Predicate: func(i int) bool { return i > 0 },
				Execute:   func() Result[string] { return Ok("positive") },
			},
		}
		defaultCase := func() Result[string] { return Ok("unknown") }

		result := SwitchFlow(value, cases, defaultCase)

		if result.IsErr() {
			t.Errorf("SwitchFlow() should not error")
		}
		if result.Value() != "positive" {
			t.Errorf("SwitchFlow() = %v, want positive", result.Value())
		}
	})

	t.Run("SwitchFlowWithResult handles error result", func(t *testing.T) {
		t.Parallel()
		testErr := errors.New("test error")
		result := Err[int](testErr)
		cases := []Case[int, string]{
			{
				Predicate: func(i int) bool { return i > 0 },
				Execute:   func() Result[string] { return Ok("positive") },
			},
		}
		defaultCase := func() Result[string] { return Ok("unknown") }

		switchResult := SwitchFlowWithResult(result, cases, defaultCase)

		if switchResult.IsOk() {
			t.Errorf("SwitchFlowWithResult() should return error for error input")
		}
	})
}

func TestFlowBuilder(t *testing.T) {
	t.Parallel()
	t.Run("executes steps in order", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		flow := NewFlowBuilder[int]().
			Step("step1", func(ctx context.Context) Result[int] { return Ok(1) }).
			Step("step2", func(ctx context.Context) Result[int] { return Ok(2) }).
			Step("step3", func(ctx context.Context) Result[int] { return Ok(3) })

		result := flow.Execute(ctx)

		if result.IsErr() {
			t.Errorf("FlowBuilder() should not error")
		}
		if result.Value() != 3 {
			t.Errorf("FlowBuilder() = %v, want 3 (last step)", result.Value())
		}
	})

	t.Run("stops on error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		testErr := errors.New("step 2 failed")
		flow := NewFlowBuilder[int]().
			Step("step1", func(ctx context.Context) Result[int] { return Ok(1) }).
			Step("step2", func(ctx context.Context) Result[int] { return Err[int](testErr) }).
			Step("step3", func(ctx context.Context) Result[int] { return Ok(99) })

		result := flow.Execute(ctx)

		if result.IsOk() {
			t.Errorf("FlowBuilder() should return error")
		}
	})

	t.Run("StepWithRetry retries on failure", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		attempts := 0
		flow := NewFlowBuilder[int]().
			StepWithRetry("unreliable", 3, func(ctx context.Context) Result[int] {
				attempts++
				if attempts < 3 {
					return Err[int](errors.New("temporary failure"))
				}
				return Ok(42)
			})

		result := flow.Execute(ctx)

		if result.IsErr() {
			t.Errorf("StepWithRetry() should succeed after retries")
		}
		if result.Value() != 42 {
			t.Errorf("StepWithRetry() = %v, want 42", result.Value())
		}
		if attempts != 3 {
			t.Errorf("StepWithRetry() attempts = %v, want 3", attempts)
		}
	})

	t.Run("Then chains steps with previous result", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		flow := NewFlowBuilder[int]().
			Step("initial", func(ctx context.Context) Result[int] { return Ok(10) }).
			Then("double", func(ctx context.Context, prev int) Result[int] { return Ok(prev * 2) })

		result := flow.Execute(ctx)

		if result.IsErr() {
			t.Errorf("Then() should not error")
		}
		if result.Value() != 20 {
			t.Errorf("Then() = %v, want 20", result.Value())
		}
	})
}

func TestParallelFlow(t *testing.T) {
	t.Parallel()
	t.Run("executes all steps concurrently", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		parallel := NewParallelFlow[int]().
			Add("step1", func(ctx context.Context) Result[int] { return Ok(1) }).
			Add("step2", func(ctx context.Context) Result[int] { return Ok(2) }).
			Add("step3", func(ctx context.Context) Result[int] { return Ok(3) })

		results := parallel.Execute(ctx)

		if len(results) != 3 {
			t.Errorf("ParallelFlow() results len = %v, want 3", len(results))
		}

		successful := parallel.Successful()
		if len(successful) != 3 {
			t.Errorf("ParallelFlow() successful len = %v, want 3", len(successful))
		}
	})

	t.Run("separates successful and failed", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		testErr := errors.New("step 2 failed")
		parallel := NewParallelFlow[int]().
			Add("step1", func(ctx context.Context) Result[int] { return Ok(1) }).
			Add("step2", func(ctx context.Context) Result[int] { return Err[int](testErr) }).
			Add("step3", func(ctx context.Context) Result[int] { return Ok(3) })

		parallel.Execute(ctx)

		successful := parallel.Successful()
		if len(successful) != 2 {
			t.Errorf("ParallelFlow() successful len = %v, want 2", len(successful))
		}

		failed := parallel.Failed()
		if len(failed) != 1 {
			t.Errorf("ParallelFlow() failed len = %v, want 1", len(failed))
		}
	})

	t.Run("Results returns empty before execution", func(t *testing.T) {
		t.Parallel()
		parallel := NewParallelFlow[int]()

		results := parallel.Results()

		if len(results) != 0 {
			t.Errorf("ParallelFlow() Results() before Execute() should be empty")
		}
	})
}
