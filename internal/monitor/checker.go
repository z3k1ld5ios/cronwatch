package monitor

import (
	"log"
	"time"

	"github.com/yourorg/cronwatch/internal/schedule"
	"github.com/yourorg/cronwatch/internal/webhook"
)

// JobConfig holds the configuration for a monitored cron job.
type JobConfig struct {
	Name            string
	CronExpr        string
	DriftThreshold  time.Duration
	MissedThreshold time.Duration
}

// Checker periodically checks tracked jobs for drift or missed runs.
type Checker struct {
	configs   []JobConfig
	tracker   Tracker
	notifier  *webhook.Notifier
	interval  time.Duration
	stopCh    chan struct{}
}

// Tracker defines the interface for retrieving job run state.
type Tracker interface {
	LastRun(name string) (time.Time, bool)
}

// NewChecker creates a new Checker with the given job configs and dependencies.
func NewChecker(configs []JobConfig, tracker Tracker, notifier *webhook.Notifier, interval time.Duration) *Checker {
	return &Checker{
		configs:  configs,
		tracker:  tracker,
		notifier: notifier,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the checking loop in a goroutine.
func (c *Checker) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.checkAll(time.Now())
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop halts the checking loop.
func (c *Checker) Stop() {
	close(c.stopCh)
}

// checkAll evaluates all configured jobs against the current time.
func (c *Checker) checkAll(now time.Time) {
	for _, cfg := range c.configs {
		c.checkJob(cfg, now)
	}
}

// checkJob evaluates a single job configuration against the current time,
// sending an alert if the job has been missed or has drifted beyond its
// configured thresholds.
func (c *Checker) checkJob(cfg JobConfig, now time.Time) {
	sched, err := schedule.Parse(cfg.CronExpr)
	if err != nil {
		log.Printf("[checker] invalid cron expr for job %q: %v", cfg.Name, err)
		return
	}
	lastRun, ok := c.tracker.LastRun(cfg.Name)
	if !ok {
		return
	}
	nextExpected := sched.Next(lastRun)
	if now.After(nextExpected.Add(cfg.MissedThreshold)) {
		c.notifier.Send(buildAlert(cfg.Name, "missed", lastRun, nextExpected, now))
	} else if now.After(nextExpected.Add(cfg.DriftThreshold)) {
		c.notifier.Send(buildAlert(cfg.Name, "drift", lastRun, nextExpected, now))
	}
}
