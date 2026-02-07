package testing

import "testing"

func TestRunTestCases(t *testing.T) {
	tests := []TestCase[int, string]{
		{"one", 1, "1"},
		{"two", 2, "2"},
		{"three", 3, "3"},
	}

	RunTestCases(t, tests, func(tc TestCase[int, string]) string {
		return tc.Expected
	})
}

func TestRunTestCasesDifferentTypes(t *testing.T) {
	tests := []TestCase[bool, int]{
		{"true is 1", true, 1},
		{"false is 0", false, 0},
	}

	RunTestCases(t, tests, func(tc TestCase[bool, int]) int {
		if tc.Input {
			return 1
		}
		return 0
	})
}

func TestRunValueTestCases(t *testing.T) {
	tests := []ValueTestCase[int, string]{
		{1, "one"},
		{2, "two"},
		{3, "three"},
	}

	RunValueTestCases(t, tests, func(tc ValueTestCase[int, string]) string {
		switch tc.Value {
		case 1:
			return "one"
		case 2:
			return "two"
		case 3:
			return "three"
		default:
			return "unknown"
		}
	})
}
