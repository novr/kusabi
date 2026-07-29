package runner_test

import (
	"sync"
	"testing"

	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

func TestRun_Hooks(t *testing.T) {
	repos := map[string]manifest.Repository{
		"a": {Path: "a"},
		"b": {Path: "b"},
		"c": {Path: "c"},
	}

	var mu sync.Mutex
	started := make([]string, 0, 3)
	done := make([]string, 0, 3)

	hooks := &runner.Hooks{
		OnStart: func(name string, index, total int) {
			mu.Lock()
			defer mu.Unlock()
			if index < 1 || index > total || total != 3 {
				t.Errorf("OnStart(%q, %d, %d): bad index/total", name, index, total)
			}
			started = append(started, name)
		},
		OnDone: func(r runner.Result) {
			mu.Lock()
			defer mu.Unlock()
			done = append(done, r.RepoName)
		},
	}

	results := runner.Run([]string{"a", "b", "c"}, repos, 0, hooks, func(name string, repo manifest.Repository) runner.Result {
		return runner.Result{RepoName: name, Output: repo.Path}
	})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if len(started) != 3 || len(done) != 3 {
		t.Fatalf("expected 3 starts and 3 done, got %d starts and %d done", len(started), len(done))
	}
}
