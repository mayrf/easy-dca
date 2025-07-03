// Package scheduler provides interfaces and implementations for scheduling DCA operations.
package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
	"github.com/mayrf/easy-dca/internal/config"
)

// DCARunner defines the interface for running a DCA operation.
type DCARunner interface {
	RunDCA() error
}

// Scheduler defines the interface for scheduling DCA operations.
type Scheduler interface {
	// Start begins the scheduling process. This should block until the scheduler is stopped.
	Start(ctx context.Context) error
	// Stop gracefully stops the scheduler.
	Stop() error
}

// CronScheduler implements Scheduler using cron expressions.
type CronScheduler struct {
	runner DCARunner
	cron   *cron.Cron
	expr   string
}

// NewCronScheduler creates a new cron-based scheduler.
func NewCronScheduler(runner DCARunner, cronExpr string) *CronScheduler {
	return &CronScheduler{
		runner: runner,
		cron:   cron.New(),
		expr:   cronExpr,
	}
}

// Start begins the cron scheduler.
func (cs *CronScheduler) Start(ctx context.Context) error {
	if cs.expr == "" {
		return fmt.Errorf("cron expression is required")
	}

	_, err := cs.cron.AddFunc(cs.expr, func() {
		if err := cs.runner.RunDCA(); err != nil {
			log.Printf("DCA run failed: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	log.Printf("Starting cron scheduler with expression: %s", cs.expr)
	cs.cron.Start()

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Stop stops the cron scheduler.
func (cs *CronScheduler) Stop() error {
	ctx := cs.cron.Stop()
	<-ctx.Done()
	return nil
}

// OneTimeScheduler implements Scheduler for single execution.
type OneTimeScheduler struct {
	runner DCARunner
}

// NewOneTimeScheduler creates a new one-time scheduler.
func NewOneTimeScheduler(runner DCARunner) *OneTimeScheduler {
	return &OneTimeScheduler{
		runner: runner,
	}
}

// Start runs the DCA operation once and returns.
func (ots *OneTimeScheduler) Start(ctx context.Context) error {
	log.Print("Running DCA operation once")
	return ots.runner.RunDCA()
}

// Stop is a no-op for one-time scheduler.
func (ots *OneTimeScheduler) Stop() error {
	return nil
}

// SystemdScheduler implements Scheduler for systemd timer integration.
// This is designed to work with systemd's OnCalendar expressions.
type SystemdScheduler struct {
	runner DCARunner
}

// NewSystemdScheduler creates a new systemd-compatible scheduler.
func NewSystemdScheduler(runner DCARunner) *SystemdScheduler {
	return &SystemdScheduler{
		runner: runner,
	}
}

// Start runs the DCA operation once (systemd handles the scheduling).
func (ss *SystemdScheduler) Start(ctx context.Context) error {
	log.Print("Running DCA operation (scheduled by systemd)")
	return ss.runner.RunDCA()
}

// Stop is a no-op for systemd scheduler.
func (ss *SystemdScheduler) Stop() error {
	return nil
}

// CreateScheduler creates the appropriate scheduler based on configuration.
func CreateScheduler(runner DCARunner, cfg config.Config) (Scheduler, error) {
	switch cfg.SchedulerMode {
	case "cron":
		if cfg.CronExpr == "" {
			return nil, fmt.Errorf("cron scheduler mode requires EASY_DCA_CRON to be set")
		}
		return NewCronScheduler(runner, cfg.CronExpr), nil
	case "systemd":
		return NewSystemdScheduler(runner), nil
	case "manual":
		return NewOneTimeScheduler(runner), nil
	default:
		return nil, fmt.Errorf("unknown scheduler mode: %s (supported: cron, systemd, manual)", cfg.SchedulerMode)
	}
} 