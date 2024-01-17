package gormkit

import (
	"database/sql"
	"time"

	kitlog "github.com/theplant/appkit/log"
	"gorm.io/gorm"
)

type statsLogger struct {
	logger   kitlog.Logger
	db       *gorm.DB
	interval time.Duration

	sqlDB      *sql.DB
	done       chan bool
	ticker     *time.Ticker
	tickerDone chan bool
	prevStats  sql.DBStats
}

var (
	defaultInverval = 5 * time.Second
)

func NewStatsLogger(logger kitlog.Logger, db *gorm.DB, interval time.Duration) *statsLogger {
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	if interval == 0 {
		interval = defaultInverval
	}

	return &statsLogger{
		logger:   logger,
		db:       db,
		interval: interval,

		sqlDB: sqlDB,
	}
}

func (l *statsLogger) Start() {
	if l.ticker != nil {
		return
	}

	l.prevStats = sql.DBStats{}
	l.done = make(chan bool)
	l.ticker = time.NewTicker(l.interval)
	l.tickerDone = make(chan bool)

	go func() {
		for {
			select {
			case <-l.done:
				l.tickerDone <- true
				l.logger.Info().Log("msg", "DB stats logger ticker stopped")
				return
			case <-l.ticker.C:
				l.logDBStats()
			}
		}
	}()
	l.logger.Info().Log("msg", "DB stats logger started")
}

func (l *statsLogger) Stop() {
	if l.ticker == nil {
		return
	}
	l.done <- true
	l.ticker.Stop()
	l.ticker = nil
	<-l.tickerDone

	l.logger.Info().Log("msg", "DB stats logger stopped")
}

func (l *statsLogger) logDBStats() {
	stats := l.sqlDB.Stats()
	l.logger.Info().Log(
		"msg", "DB stats",
		"db.max_open_connections", stats.MaxOpenConnections,
		"db.open_connections", stats.OpenConnections,
		"db.in_use", stats.InUse,
		"db.idle", stats.Idle,
		"db.wait_count", stats.WaitCount-l.prevStats.WaitCount,
		"db.wait_duration_ms", stats.WaitDuration.Milliseconds()-l.prevStats.WaitDuration.Milliseconds(),
		"db.max_idle_closed", stats.MaxIdleClosed-l.prevStats.MaxIdleClosed,
		"db.max_idle_time_closed", stats.MaxIdleTimeClosed-l.prevStats.MaxIdleTimeClosed,
		"db.max_lifetime_closed", stats.MaxLifetimeClosed-l.prevStats.MaxLifetimeClosed,
	)
	l.prevStats = stats
}
