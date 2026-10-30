package recommendations

import (
	"sort"
	"strings"
)

type SimilarAppsService struct{}

func NewSimilarAppsService() *SimilarAppsService {
	return &SimilarAppsService{}
}

type RecommendedApp struct {
	AppID          string  `json:"app_id"`
	Slug           string  `json:"slug"`
	PackageName    string  `json:"package_name"`
	Title          string  `json:"title"`
	Category       string  `json:"category"`
	BayesianRating float64 `json:"bayesian_rating"`
	SimilarityScore float64 `json:"similarity_score"`
}

// FindSimilarApps computes similarity between target app and catalog candidates
func (s *SimilarAppsService) FindSimilarApps(target CatalogAppMetric, candidates []CatalogAppMetric, limit int) []RecommendedApp {
	if limit <= 0 {
		limit = 10
	}

	type ScoredSimilar struct {
		App   CatalogAppMetric
		Score float64
	}

	var matches []ScoredSimilar

	targetTags := make(map[string]bool)
	for _, tag := range target.Tags {
		targetTags[strings.ToLower(tag)] = true
	}

	for _, cand := range candidates {
		if cand.AppID == target.AppID || (cand.PackageName != "" && cand.PackageName == target.PackageName) {
			continue // Do not recommend self
		}
		if cand.RightsLevel == "LEVEL_5_RESTRICTED_PROHIBITED" {
			continue
		}

		baseScore := 0.0

		// 1. Category match (Weight: 3.0)
		if strings.EqualFold(cand.Category, target.Category) {
			baseScore += 3.0
		}

		// 2. Tag Jaccard similarity (Weight: up to 2.0)
		if len(targetTags) > 0 && len(cand.Tags) > 0 {
			sharedTags := 0
			for _, t := range cand.Tags {
				if targetTags[strings.ToLower(t)] {
					sharedTags++
				}
			}
			union := len(targetTags) + len(cand.Tags) - sharedTags
			if union > 0 {
				jaccard := float64(sharedTags) / float64(union)
				baseScore += jaccard * 2.0
			}
		}

		if baseScore <= 0.0 {
			continue
		}

		bayesian := ComputeBayesianRating(cand.RawRating, cand.ReviewCount)
		finalScore := baseScore * (bayesian / 5.0)

		matches = append(matches, ScoredSimilar{
			App:   cand,
			Score: finalScore,
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	results := make([]RecommendedApp, 0, len(matches))
	for i, m := range matches {
		if i >= limit {
			break
		}
		results = append(results, RecommendedApp{
			AppID:           m.App.AppID,
			Slug:            m.App.Slug,
			PackageName:     m.App.PackageName,
			Title:           m.App.Title,
			Category:        m.App.Category,
			BayesianRating:  ComputeBayesianRating(m.App.RawRating, m.App.ReviewCount),
			SimilarityScore: m.Score,
		})
	}

	return results
}
