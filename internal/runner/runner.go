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

// Hooks are optional callbacks invoked during parallel execution.
// OnStart and OnDone may run concurrently from multiple goroutines.
type Hooks struct {
	OnStart func(name string, index, total int)
	OnDone  func(result Result)
}

func resolveOrder(order []string, repos map[string]manifest.Repository) []string {
	if len(order) > 0 {
		out := make([]string, 0, len(order))
		for _, name := range order {
			if _, ok := repos[name]; ok {
				out = append(out, name)
			}
		}
		return out
	}
	names := make([]string, 0, len(repos))
	for name := range repos {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func runParallel[T any](
	order []string,
	repos map[string]manifest.Repository,
	maxConcurrency int,
	hooks *Hooks,
	fn func(name string, repo manifest.Repository) T,
) []T {
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU() * 2
	}

	names := resolveOrder(order, repos)
	sem := semaphore.NewWeighted(int64(maxConcurrency))
	ctx := context.Background()

	type indexed struct {
		pos int
		val T
	}

	items := make([]indexed, len(names))
	var wg sync.WaitGroup

	for i, name := range names {
		wg.Add(1)
		go func(pos int, n string) {
			defer wg.Done()
			_ = sem.Acquire(ctx, 1)
			defer sem.Release(1)
			if hooks != nil && hooks.OnStart != nil {
				hooks.OnStart(n, pos+1, len(names))
			}
			val := fn(n, repos[n])
			if hooks != nil && hooks.OnDone != nil {
				hooks.OnDone(toResult(val))
			}
			items[pos] = indexed{pos: pos, val: val}
		}(i, name)
	}

	wg.Wait()

	out := make([]T, len(items))
	for _, it := range items {
		out[it.pos] = it.val
	}
	return out
}

func toResult[T any](v T) Result {
	if r, ok := any(v).(Result); ok {
		return r
	}
	return Result{}
}

// Run executes fn for each repository in parallel.
// order selects which repos to run and in what sequence; pass nil to run all repos sorted by name.
func Run(
	order []string,
	repos map[string]manifest.Repository,
	maxConcurrency int,
	hooks *Hooks,
	fn func(name string, repo manifest.Repository) Result,
) []Result {
	return runParallel(order, repos, maxConcurrency, hooks, fn)
}

// RunTyped is like Run but the callback returns an arbitrary type T.
func RunTyped[T any](
	order []string,
	repos map[string]manifest.Repository,
	maxConcurrency int,
	fn func(name string, repo manifest.Repository) T,
) []T {
	return runParallel(order, repos, maxConcurrency, nil, fn)
}
