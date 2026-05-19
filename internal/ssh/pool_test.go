package ssh_test

import (
	"sync"
	"testing"

	sshclient "github.com/yourorg/patchwork-deploy/internal/ssh"
)

func TestPool_GetReturnsError_OnBadConfig(t *testing.T) {
	pool := sshclient.NewPool()
	defer pool.CloseAll()

	_, err := pool.Get("host1", sshclient.Config{
		Host:    "127.0.0.1",
		Port:    22,
		User:    "root",
		KeyPath: "/no/such/key",
	})
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestPool_CloseAll_EmptyPool(t *testing.T) {
	pool := sshclient.NewPool()
	// Should not panic on empty pool.
	pool.CloseAll()
}

func TestPool_Remove_NonExistent(t *testing.T) {
	pool := sshclient.NewPool()
	// Should not panic when removing a key that was never added.
	pool.Remove("ghost-host")
}

func TestPool_Concurrent_Get(t *testing.T) {
	pool := sshclient.NewPool()
	defer pool.CloseAll()

	var wg sync.WaitGroup
	errs := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = pool.Get("shared", sshclient.Config{
				Host:    "127.0.0.1",
				KeyPath: "/no/such/key",
			})
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		if err == nil {
			t.Error("expected error from concurrent Get, got nil")
		}
	}
}
