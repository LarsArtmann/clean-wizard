package format

import (
	"encoding/json/v2"
	"strings"
	"testing"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanOutcomesToSARIFCleanableFinding(t *testing.T) {
	t.Parallel()

	outcomes := []ScanOutcome{
		{
			Name:           "Nix",
			RegistryName:   "nix",
			Description:    "Nix store generations",
			ItemsCount:     12,
			BytesCleanable: 2_500_000_000,
		},
	}

	data, err := ScanOutcomesToSARIF(outcomes)
	require.NoError(t, err)

	doc := parseSARIF(t, data)

	assert.Equal(t, "2.1.0", doc["version"])

	results := sarifResults(t, doc)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "nix", result["ruleId"])
	assert.Equal(t, "note", result["level"])
	assert.Contains(t, messageText(t, result), "12 items")

	props := properties(t, result)
	assert.Equal(t, "unused", props["go-finding/category"])
	assert.Equal(t, "suggest", props["go-finding/fixStrategy"])
	assert.Equal(t, "12", props["go-finding/meta/items"])
	assert.Equal(t, "2500000000", props["go-finding/meta/bytes"])
	assert.Equal(t, "Nix store generations", props["go-finding/meta/description"])
	assert.Equal(t, "Run 'clean-wizard clean --dry-run' to preview reclaiming this space", props["go-finding/suggestion"])

	uri := sarifLocationURI(t, result)
	assert.Equal(t, "cleaner://nix", uri)
}

func TestScanOutcomesToSARIFFailureFinding(t *testing.T) {
	t.Parallel()

	outcomes := []ScanOutcome{
		{
			Name:         "Cargo",
			RegistryName: "cargo",
			Err:          errorfamily.NewTransient("scan.exec", "du command failed"),
		},
	}

	data, err := ScanOutcomesToSARIF(outcomes)
	require.NoError(t, err)

	results := sarifResults(t, parseSARIF(t, data))
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "cargo", result["ruleId"])
	assert.Equal(t, "error", result["level"])
	assert.Equal(t, "[transient:scan.exec] du command failed", messageText(t, result))

	props := properties(t, result)
	assert.Equal(t, "transient", props["go-finding/meta/family"])
	assert.Equal(t, "scan.exec", props["go-finding/meta/code"])
	assert.Equal(t, "true", props["go-finding/meta/retryable"])
}

func TestScanOutcomesToSARIFOmitsNonFindings(t *testing.T) {
	t.Parallel()

	outcomes := []ScanOutcome{
		{Name: "Go", RegistryName: "go", ItemsCount: 3, BytesCleanable: 0},
		{Name: "Cargo", RegistryName: "cargo", Err: cleanerNotAvailableErr("cargo")},
	}

	data, err := ScanOutcomesToSARIF(outcomes)
	require.NoError(t, err)

	results := sarifResults(t, parseSARIF(t, data))
	assert.Empty(t, results)
}

func TestScanOutcomesToSARIFEmpty(t *testing.T) {
	t.Parallel()

	data, err := ScanOutcomesToSARIF(nil)
	require.NoError(t, err)

	doc := parseSARIF(t, data)
	assert.Equal(t, "2.1.0", doc["version"])
	assert.Empty(t, sarifResults(t, doc))
}

func TestScanOutcomesToSARIFDeterministic(t *testing.T) {
	t.Parallel()

	outcomes := []ScanOutcome{
		{Name: "Nix", RegistryName: "nix", ItemsCount: 1, BytesCleanable: 100},
		{Name: "Go", RegistryName: "go", ItemsCount: 2, BytesCleanable: 200},
	}

	first, err := ScanOutcomesToSARIF(outcomes)
	require.NoError(t, err)

	second, err := ScanOutcomesToSARIF(outcomes)
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second))
}

func TestScanOutcomesToFindingsAllValid(t *testing.T) {
	t.Parallel()

	outcomes := []ScanOutcome{
		{Name: "Nix", RegistryName: "nix", ItemsCount: 1, BytesCleanable: 100},
		{Name: "Cargo", RegistryName: "cargo", Err: errorfamily.NewRejection("scan.config", "bad settings")},
		{Name: "Docker", RegistryName: "docker", Err: cleanerNotAvailableErr("docker")},
	}

	findings, err := scanOutcomesToFindings(outcomes)
	require.NoError(t, err)
	require.Len(t, findings, 2)

	for _, f := range findings {
		require.NoError(t, f.Validate())
		assert.Equal(t, finding.ToolName("clean-wizard"), f.ToolName)
		assert.True(t, strings.HasPrefix(string(f.Position.File), "cleaner://"))
	}
}

// cleanerNotAvailableErr avoids importing the cleaner package (import cycle).
func cleanerNotAvailableErr(name string) error {
	return errorfamily.NewInfrastructure("cleaner."+name+".not_available", name+" is not installed")
}

func parseSARIF(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var doc map[string]any
	require.NoError(t, json.Unmarshal(data, &doc))

	return doc
}

func sarifResults(t *testing.T, doc map[string]any) []map[string]any {
	t.Helper()

	runs, ok := doc["runs"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, runs)

	run, ok := runs[0].(map[string]any)
	require.True(t, ok)

	rawResults, ok := run["results"].([]any)
	if !ok {
		return nil
	}

	results := make([]map[string]any, 0, len(rawResults))
	for _, raw := range rawResults {
		result, ok := raw.(map[string]any)
		require.True(t, ok)
		results = append(results, result)
	}

	return results
}

func messageText(t *testing.T, result map[string]any) string {
	t.Helper()

	message, ok := result["message"].(map[string]any)
	require.True(t, ok)

	text, ok := message["text"].(string)
	require.True(t, ok)

	return text
}

func properties(t *testing.T, result map[string]any) map[string]any {
	t.Helper()

	props, ok := result["properties"].(map[string]any)
	require.True(t, ok)

	return props
}

func sarifLocationURI(t *testing.T, result map[string]any) string {
	t.Helper()

	locations, ok := result["locations"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, locations)

	location, ok := locations[0].(map[string]any)
	require.True(t, ok)

	physical, ok := location["physicalLocation"].(map[string]any)
	require.True(t, ok)

	artifact, ok := physical["artifactLocation"].(map[string]any)
	require.True(t, ok)

	uri, ok := artifact["uri"].(string)
	require.True(t, ok)

	return uri
}
