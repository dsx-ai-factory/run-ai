//go:build integration
// +build integration

package tests

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
)

// leaderGateRunnable signals when it passes the election gate.
type leaderGateRunnable struct {
	elected <-chan struct{}
	started chan<- string
	id      string
}

func (r leaderGateRunnable) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.elected:
		r.started <- r.id
		<-ctx.Done()
		return nil
	}
}

func (r leaderGateRunnable) NeedLeaderElection() bool { return false }

func startManager(t *testing.T, mgr manager.Manager, ctx context.Context) {
	t.Helper()
	go func() {
		_ = mgr.Start(ctx)
	}()
}

func TestLeaderElectionSingleLeaderAndFailover(t *testing.T) {
	g := NewWithT(t)

	// Set up envtest environment for this test
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}
	cfg, err := testEnv.Start()
	g.Expect(err).NotTo(HaveOccurred())
	defer func() { _ = testEnv.Stop() }()

	// mgrA with election enabled
	mgrA, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:           scheme_init.Scheme(),
		LeaderElection:   true,
		LeaderElectionID: "runai-leader-election-test",
	})
	g.Expect(err).NotTo(HaveOccurred())

	// mgrB with the same election ID
	mgrB, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:           scheme_init.Scheme(),
		LeaderElection:   true,
		LeaderElectionID: "runai-leader-election-test",
	})
	g.Expect(err).NotTo(HaveOccurred())

	started := make(chan string, 2)
	_ = mgrA.Add(leaderGateRunnable{elected: mgrA.Elected(), started: started, id: "A"})
	_ = mgrB.Add(leaderGateRunnable{elected: mgrB.Elected(), started: started, id: "B"})

	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()

	startManager(t, mgrA, ctxA)
	startManager(t, mgrB, ctxB)

	// Expect only one leader within timeout
	var first string
	select {
	case first = <-started:
	case <-time.After(20 * time.Second):
		t.Fatalf("timed out waiting for initial leader")
	}

	// Drain any unexpected second start quickly
	select {
	case second := <-started:
		t.Fatalf("expected single leader, but got two: %s then %s", first, second)
	case <-time.After(1 * time.Second):
		// ok, only one leader
	}

	// Trigger failover by canceling the current leader's context
	if first == "A" {
		cancelA()
	} else {
		cancelB()
	}

	// Expect the other manager to take leadership
	select {
	case next := <-started:
		if next == first {
			t.Fatalf("failover did not switch leader; got same leader %s", next)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("timed out waiting for leader failover")
	}
}

func TestLeaderElectionDisabledStartsImmediately(t *testing.T) {
	g := NewWithT(t)

	// Set up envtest environment for this test
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}
	cfg, err := testEnv.Start()
	g.Expect(err).NotTo(HaveOccurred())
	defer func() { _ = testEnv.Stop() }()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme_init.Scheme(),
		LeaderElection: false,
	})
	g.Expect(err).NotTo(HaveOccurred())

	started := make(chan string, 1)
	// Closed channel to simulate immediate election pass
	closed := make(chan struct{})
	close(closed)
	_ = mgr.Add(leaderGateRunnable{elected: closed, started: started, id: "X"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startManager(t, mgr, ctx)

	var seen atomic.Bool
	select {
	case <-time.After(5 * time.Second):
		// if not started within 5s, fail
		if !seen.Load() {
			t.Fatalf("manager did not start runnable promptly with election disabled")
		}
	case id := <-started:
		seen.Store(true)
		if id != "X" {
			t.Fatalf("unexpected runnable id: %s", id)
		}
	}
}
