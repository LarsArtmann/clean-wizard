package cleaner

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// toGoCacheType converts GoPackagesSettings bools to type-safe GoCacheType.
func toGoCacheType(settings *domain.GoPackagesSettings) GoCacheType {
	if settings == nil {
		return GoCacheNone
	}

	cacheType := GoCacheNone

	if settings.CleanCache {
		cacheType |= GoCacheGOCACHE
	}
	if settings.CleanTestCache {
		cacheType |= GoCacheTestCache
	}
	if settings.CleanModCache {
		cacheType |= GoCacheModCache
	}
	if settings.CleanBuildCache {
		cacheType |= GoCacheBuildCache
	}
	if settings.CleanLintCache {
		cacheType |= GoCacheLintCache
	}

	return cacheType
}

// fromGoCacheType converts GoCacheType back to individual bools for backward compatibility.
func fromGoCacheType(cacheType GoCacheType) (cleanCache, cleanTestCache, cleanModCache, cleanBuildCache, cleanLintCache bool) {
	cleanCache = cacheType.Has(GoCacheGOCACHE)
	cleanTestCache = cacheType.Has(GoCacheTestCache)
	cleanModCache = cacheType.Has(GoCacheModCache)
	cleanBuildCache = cacheType.Has(GoCacheBuildCache)
	cleanLintCache = cacheType.Has(GoCacheLintCache)
	return cleanCache, cleanTestCache, cleanModCache, cleanBuildCache, cleanLintCache
}
