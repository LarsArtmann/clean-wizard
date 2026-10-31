package finding

import (
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/larsartmann/go-finding/lockutil"
)

// LinterRegistry maps linter/analyzer names to Categories.
// The zero value is ready to use. All methods are safe for concurrent use.
//
// Use DefaultLinterRegistry for the global registry with built-in mappings,
// or create isolated registries for testing or custom tool chains.
type LinterRegistry struct {
	mu    sync.RWMutex
	items map[string]Category
}

// NewLinterRegistry creates a registry pre-populated with the given mappings.
func NewLinterRegistry(items map[string]Category) *LinterRegistry {
	r := &LinterRegistry{items: make(map[string]Category, len(items))}
	for k, v := range items {
		r.items[strings.ToLower(k)] = v
	}

	return r
}

// Lookup returns the Category for a linter name (case-insensitive).
// Returns the provided fallback if the name is not registered.
func (r *LinterRegistry) Lookup(name string, fallback Category) Category {
	return readLinter(r, func() Category {
		if cat, ok := r.items[strings.ToLower(name)]; ok {
			return cat
		}

		return fallback
	})
}

// Register adds or overrides a Category mapping for a linter name.
func (r *LinterRegistry) Register(name string, cat Category) {
	writeLinter(r, func() struct{} {
		if r.items == nil {
			r.items = make(map[string]Category)
		}

		r.items[strings.ToLower(name)] = cat

		return struct{}{}
	})
}

// Names returns all registered linter names in sorted order.
func (r *LinterRegistry) Names() []string {
	return readLinter(r, func() []string {
		return slices.Sorted(maps.Keys(r.items))
	})
}

// Clone returns a deep copy of the registry.
func (r *LinterRegistry) Clone() *LinterRegistry {
	return readLinter(r, func() *LinterRegistry {
		return &LinterRegistry{items: maps.Clone(r.items)}
	})
}

// readLinter runs fn while holding r.mu.RLock and returns its result.
func readLinter[T any](r *LinterRegistry, fn func() T) T {
	return lockutil.RLocked(&r.mu, fn)
}

// writeLinter runs fn while holding r.mu.Lock and returns its result.
func writeLinter[T any](r *LinterRegistry, fn func() T) T {
	return lockutil.Locked(&r.mu, fn)
}

// DefaultLinterRegistry is the global registry with built-in linter→category mappings.
var DefaultLinterRegistry = NewLinterRegistry(map[string]Category{
	// Security
	"gosec":         CategorySecurity,
	"noctx":         CategorySecurity,
	"errchkjson":    CategorySecurity,
	"bidichk":       CategorySecurity,
	"nosprintfhost": CategorySecurity,

	// Correctness
	"govet":         CategoryCorrectness,
	"staticcheck":   CategoryCorrectness,
	"errcheck":      CategoryCorrectness,
	"nilerr":        CategoryCorrectness,
	"ineffassign":   CategoryCorrectness,
	"unconvert":     CategoryCorrectness,
	"bodyclose":     CategoryCorrectness,
	"contextcheck":  CategoryCorrectness,
	"durationcheck": CategoryCorrectness,
	"typecheck":     CategoryCorrectness,
	"gosimple":      CategoryCorrectness,
	"deadcode":      CategoryCorrectness,
	"varcheck":      CategoryCorrectness,
	"nakedret":      CategoryCorrectness,
	"rowserrcheck":  CategoryCorrectness,
	"sqlclosecheck": CategoryCorrectness,
	"wastedassign":  CategoryCorrectness,
	"exportloopref": CategoryCorrectness,
	"nilnesserr":    CategoryCorrectness,
	"recvcheck":     CategoryCorrectness,

	// Performance
	"prealloc":   CategoryPerformance,
	"perfsprint": CategoryPerformance,
	"unparam":    CategoryPerformance,

	// Complexity
	"gocyclo":        CategoryComplexity,
	"cyclop":         CategoryComplexity,
	"gocognit":       CategoryComplexity,
	"maintidx":       CategoryComplexity,
	"funlen":         CategoryComplexity,
	"nestif":         CategoryComplexity,
	"interfacebloat": CategoryComplexity,
	"gocritic":       CategoryComplexity,

	// Duplication
	"dupl":    CategoryDuplication,
	"goconst": CategoryDuplication,

	// Error handling
	"wrapcheck": CategoryErrorHandling,
	"errorlint": CategoryErrorHandling,
	"errname":   CategoryErrorHandling,
	"nilnil":    CategoryErrorHandling,

	// Style
	"misspell":       CategoryStyle,
	"revive":         CategoryStyle,
	"gofmt":          CategoryStyle,
	"gofumpt":        CategoryStyle,
	"goimports":      CategoryStyle,
	"gci":            CategoryStyle,
	"wsl":            CategoryStyle,
	"wsl_v5":         CategoryStyle,
	"dupword":        CategoryStyle,
	"godot":          CategoryStyle,
	"lll":            CategoryStyle,
	"whitespace":     CategoryStyle,
	"nlreturn":       CategoryStyle,
	"golint":         CategoryStyle,
	"dogsled":        CategoryStyle,
	"nolintlint":     CategoryStyle,
	"forbidigo":      CategoryStyle,
	"gomnd":          CategoryStyle,
	"varnamelen":     CategoryStyle,
	"tagalign":       CategoryStyle,
	"nonamedreturns": CategoryStyle,

	// Testing
	"paralleltest":     CategoryTesting,
	"thelper":          CategoryTesting,
	"testifylint":      CategoryTesting,
	"ginkgolinter":     CategoryTesting,
	"tparallel":        CategoryTesting,
	"testpackage":      CategoryTesting,
	"testableexamples": CategoryTesting,

	// Type safety
	"exhaustive":      CategoryTypeSafety,
	"exhaustruct":     CategoryTypeSafety,
	"forcetypeassert": CategoryTypeSafety,
	"musttag":         CategoryTypeSafety,
	"gochecksumtype":  CategoryTypeSafety,
	"copyloopvar":     CategoryTypeSafety,
	"intrange":        CategoryTypeSafety,
	"tagliatelle":     CategoryTypeSafety,

	// Structure
	"sloglint":    CategoryStructure,
	"loggercheck": CategoryStructure,
	"unused":      CategoryUnused,

	// Best practice
	"depguard": CategoryBestPractice,
	"mirror":   CategoryBestPractice,

	// Configuration
	"gomodguard_v2": CategoryConfiguration,
})

// CategoryForLinter returns the default Category for a well-known linter or analyzer name.
// The lookup is case-insensitive. If the name is not registered, it returns CategoryCorrectness.
// Register custom mappings with RegisterLinterCategory.
func CategoryForLinter(name string) Category {
	return DefaultLinterRegistry.Lookup(name, CategoryCorrectness)
}

// RegisterLinterCategory registers or overrides the Category for a linter name.
// The name is stored in lowercase for case-insensitive lookup.
// Safe for concurrent use.
func RegisterLinterCategory(name string, cat Category) {
	DefaultLinterRegistry.Register(name, cat)
}
