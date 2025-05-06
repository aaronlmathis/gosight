package debugtools

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/aaronlmathis/gosight/server/internal/cache"
)

type CacheAuditResult struct {
	LabelKey        string   `json:"label_key"`
	UniqueValues    int      `json:"unique_values"`
	ExampleValues   []string `json:"example_values"`
	PossibleProblem string   `json:"possible_problem,omitempty"`
	Source          string   `json:"source"` // tag or meta
}

type CacheAuditReport struct {
	TagCache            []CacheAuditResult `json:"tag_cache"`
	MetricCache         []CacheAuditResult `json:"metric_cache"`
	RedundantKeys       []string           `json:"redundant_keys"`
	TagCacheMemoryKB    int                `json:"tag_cache_estimated_kb"`
	MetricCacheMemoryKB int                `json:"metric_cache_estimated_kb"`
}

func AuditCaches(tagCache cache.TagCache, metricCache cache.MetricCache) CacheAuditReport {
	tagResults := AuditTagCache(tagCache)
	metricResults := AuditMetricCache(metricCache)

	tagKeys := map[string]struct{}{}
	for _, r := range tagResults {
		tagKeys[r.LabelKey] = struct{}{}
	}
	redundant := []string{}
	for _, r := range metricResults {
		if _, exists := tagKeys[r.LabelKey]; exists {
			redundant = append(redundant, r.LabelKey)
		}
	}

	// Estimate memory usage
	tagMem := EstimateTagCacheMemory(tagCache)
	metricMem := EstimateMetricCacheMemory(metricCache)

	return CacheAuditReport{
		TagCache:            tagResults,
		MetricCache:         metricResults,
		RedundantKeys:       redundant,
		TagCacheMemoryKB:    tagMem / 1024,
		MetricCacheMemoryKB: metricMem / 1024,
	}
}

func AuditTagCache(tc cache.TagCache) []CacheAuditResult {
	results := []CacheAuditResult{}
	tagKeys := tc.GetTagKeys()

	for _, key := range tagKeys {
		values := tc.GetTagValues(key)
		count := len(values)

		examples := make([]string, 0, 5)
		for val := range values {
			examples = append(examples, val)
			if len(examples) >= 5 {
				break
			}
		}

		problem := ""
		if count > 1000 {
			problem = "Too many values for tag key"
		}

		results = append(results, CacheAuditResult{
			LabelKey:        key,
			UniqueValues:    count,
			ExampleValues:   examples,
			PossibleProblem: problem,
			Source:          "tag",
		})
	}

	return results
}

func AuditMetricCache(mc cache.MetricCache) []CacheAuditResult {
	results := []CacheAuditResult{}
	labelCardinality := make(map[string]map[string]struct{})

	for _, entry := range mc.GetAllEntries() {
		for k, set := range entry.Labels {
			if _, ok := labelCardinality[k]; !ok {
				labelCardinality[k] = make(map[string]struct{})
			}
			for v := range set {
				labelCardinality[k][v] = struct{}{}
			}
		}
	}

	for k, values := range labelCardinality {
		valList := make([]string, 0, 5)
		for v := range values {
			valList = append(valList, v)
			if len(valList) >= 5 {
				break
			}
		}
		count := len(values)

		problem := ""
		if count > 1000 {
			problem = "High label cardinality"
		}

		results = append(results, CacheAuditResult{
			LabelKey:        k,
			UniqueValues:    count,
			ExampleValues:   valList,
			PossibleProblem: problem,
			Source:          "meta",
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UniqueValues > results[j].UniqueValues
	})

	return results
}

func EstimateTagCacheMemory(tc cache.TagCache) int {
	total := 0
	endpoints := tc.GetAllEndpoints()
	for _, tags := range endpoints {
		for k, set := range tags {
			total += utf8.RuneCountInString(k)
			for v := range set {
				total += utf8.RuneCountInString(v)
			}
		}
	}
	return total
}

func EstimateMetricCacheMemory(mc cache.MetricCache) int {
	total := 0
	for _, entry := range mc.GetAllEntries() {
		for k, set := range entry.Labels {
			total += utf8.RuneCountInString(k)
			for v := range set {
				total += utf8.RuneCountInString(v)
			}
		}
	}
	return total
}
func PrintAuditReport(title string, results []CacheAuditResult) {
	fmt.Printf("=== %s ===\n", title)
	for _, r := range results {
		fmt.Printf("- %s (%d values)", r.LabelKey, r.UniqueValues)
		if r.PossibleProblem != "" {
			fmt.Printf(" ⚠️  %s", r.PossibleProblem)
		}
		fmt.Println()
		if len(r.ExampleValues) > 0 {
			fmt.Printf("    Examples: %v\n", r.ExampleValues)
		}
	}
	fmt.Println()
}
