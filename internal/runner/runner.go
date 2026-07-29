package runner

import (
	"context"
	"runtime"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/novr/kusabi/internal/manifest"
)

// Result holds the output from a single repository operation.
type Result struct {
	RepoName string
	Output   string
	Err      error
	Skipped  bool // true when the operation was gracefully skipped (not a failure)
	Duration time.Duration
}

// Run executes fn for each repository in parallel, bounded by maxConcurrency.
// Results are returned sorted by repository name for stable output ordering.
func Run(
	repos map[string]manifest.Repository,
	maxConcurrency int,
	fn func(name string, repo manifest.Repository) Result,
) []Result {
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU() * 2
	}

	sem := semaphore.NewWeighted(int64(maxConcurrency))
	ctx := context.Background()

	results := make([]Result, 0, len(repos))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, repo := range repos {
		wg.Add(1)
		go func(n string, r manifest.Repository) {
			defer wg.Done()
			_ = sem.Acquire(ctx, 1)
			defer sem.Release(1)

			res := fn(n, r)
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(name, repo)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].RepoName < results[j].RepoName
	})
	return results
}

// RunTyped is like Run but the callback returns an arbitrary type T.
// Results are returned in stable alphabetical order by repository name.
func RunTyped[T any](
	repos map[string]manifest.Repository,
	maxConcurrency int,
	fn func(name string, repo manifest.Repository) T,
) []T {
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU() * 2
	}

	type indexed struct {
		name string
		val  T
	}

	sem := semaphore.NewWeighted(int64(maxConcurrency))
	ctx := context.Background()

	items := make([]indexed, 0, len(repos))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, repo := range repos {
		wg.Add(1)
		go func(n string, r manifest.Repository) {
			defer wg.Done()
			_ = sem.Acquire(ctx, 1)
			defer sem.Release(1)

			v := fn(n, r)
			mu.Lock()
			items = append(items, indexed{name: n, val: v})
			mu.Unlock()
		}(name, repo)
	}

	wg.Wait()

	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	out := make([]T, len(items))
	for i, it := range items {
		out[i] = it.val
	}
	return out
}
