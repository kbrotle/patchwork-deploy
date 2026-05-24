// Package signal provides graceful shutdown handling for patchwork-deploy.
//
// It listens for OS-level termination signals (SIGINT, SIGTERM by default)
// and invokes a chain of registered ShutdownFunc hooks in order, allowing
// components such as the SSH pool, scheduler, and lock manager to cleanly
// release resources before the process exits.
//
// Basic usage:
//
//	h := signal.NewHandler(cfg,
//		func(ctx context.Context) error { return pool.CloseAll() },
//		func(ctx context.Context) error { return scheduler.Stop() },
//	)
//	if err := h.Wait(ctx); err != nil {
//		log.Println("shutdown error:", err)
//	}
package signal
