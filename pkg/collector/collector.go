package collector

import (
	"github.com/faircom/prometheus-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// FairComCollector collects metrics from FairCom using ctstat
type FairComCollector struct {
	config config.FairComConfig
	logger *logrus.Logger
}

// NewFairComCollector creates a new FairCom collector
func NewFairComCollector(cfg config.FairComConfig) *FairComCollector {
	return &FairComCollector{
		config: cfg,
		logger: logrus.StandardLogger(),
	}
}

// Describe implements prometheus.Collector
func (c *FairComCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil)
}

// Collect implements prometheus.Collector
func (c *FairComCollector) Collect(ch chan<- prometheus.Metric) {
	snapshot, err := GetSnapshot(c.config.Username, c.config.Password, c.config.CtstatPath, c.config.GetTimeout())
	if err != nil {
		c.logger.Errorf("Failed to get snapshot: %v", err)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil),
			prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil),
		prometheus.GaugeValue, 1)

	// Cache metrics - hit AND miss percentages
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_cache_hit_percent", "Cache hit percentage", []string{"type"}, nil),
		prometheus.GaugeValue, snapshot.DataCacheHitPct, "data")
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_cache_hit_percent", "Cache hit percentage", []string{"type"}, nil),
		prometheus.GaugeValue, snapshot.IndexCacheHitPct, "index")
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_cache_miss_percent", "Cache miss percentage", []string{"type"}, nil),
		prometheus.GaugeValue, snapshot.DataCacheMissPct, "data")
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_cache_miss_percent", "Cache miss percentage", []string{"type"}, nil),
		prometheus.GaugeValue, snapshot.IndexCacheMissPct, "index")

	// I/O metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_read_ops_per_sec", "Read operations per second", nil, nil),
		prometheus.GaugeValue, float64(snapshot.ReadOpsPerSec))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_write_ops_per_sec", "Write operations per second", nil, nil),
		prometheus.GaugeValue, float64(snapshot.WriteOpsPerSec))

	// File metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_files_open", "Files currently open", nil, nil),
		prometheus.GaugeValue, float64(snapshot.FilesOpen))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_files_max", "Maximum files", nil, nil),
		prometheus.GaugeValue, float64(snapshot.FilesMax))

	// Connection metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_users_active", "Active users", nil, nil),
		prometheus.GaugeValue, float64(snapshot.UsersActive))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_users_max", "Maximum users", nil, nil),
		prometheus.GaugeValue, float64(snapshot.UsersMax))

	// Lock metrics - hit AND miss percentages
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_locks_held", "Locks currently held", nil, nil),
		prometheus.GaugeValue, float64(snapshot.LocksHeld))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_lock_hit_percent", "Lock hit percentage", nil, nil),
		prometheus.GaugeValue, snapshot.LockHitPct)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_lock_miss_percent", "Lock miss percentage", nil, nil),
		prometheus.GaugeValue, snapshot.LockMissPct)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_deadlocks_total", "Total deadlocks", nil, nil),
		prometheus.CounterValue, float64(snapshot.Deadlocks))

	// Transaction metrics - with timing
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transactions_active", "Active transactions", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TransActive))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transactions_per_sec", "Transactions per second", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TransPerSec))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_read_time", "Transaction read time", nil, nil),
		prometheus.GaugeValue, float64(snapshot.ReadTranTime))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_write_time", "Transaction write time", nil, nil),
		prometheus.GaugeValue, float64(snapshot.WriteTranTime))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_begins_total", "Transaction begins", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranBegins))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_commits_total", "Transaction commits", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranCommits))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_aborts_total", "Transaction aborts", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranAborts))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_savepoints_total", "Transaction savepoints", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranSavepoints))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_restores_total", "Transaction restores", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranRestores))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_log_writes_total", "Transaction log writes", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranLogWrites))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_transaction_log_bytes_total", "Transaction log bytes", nil, nil),
		prometheus.CounterValue, float64(snapshot.TranLogBytes))

	// ISAM metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_adds_total", "ISAM add operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamAdds))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_deletes_total", "ISAM delete operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamDeletes))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_updates_total", "ISAM update operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamUpdates))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_reads_total", "ISAM read operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamReads))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_first_total", "ISAM first operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamFirst))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_last_total", "ISAM last operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamLast))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_next_total", "ISAM next operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamNext))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_isam_prev_total", "ISAM previous operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.IsamPrev))

	// SQL metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_selects_total", "SQL SELECT operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLSelects))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_inserts_total", "SQL INSERT operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLInserts))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_updates_total", "SQL UPDATE operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLUpdates))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_deletes_total", "SQL DELETE operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLDeletes))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_commits_total", "SQL COMMIT operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLCommits))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_sql_rollbacks_total", "SQL ROLLBACK operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.SQLRollbacks))

	// File operation metrics
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_file_opens_total", "File open operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.FileOpens))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_file_closes_total", "File close operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.FileCloses))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_file_creates_total", "File create operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.FileCreates))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_file_deletes_total", "File delete operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.FileDeletes))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_file_renames_total", "File rename operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.FileRenames))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_physical_reads_total", "Physical read operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.PhysicalReads))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_physical_writes_total", "Physical write operations", nil, nil),
		prometheus.CounterValue, float64(snapshot.PhysicalWrites))

	// Aggregated user metrics from userinfox
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_memory_kb", "Total memory usage across all users (KB)", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TotalMemoryKB))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_active_users", "Total active (busy) users", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TotalActiveUsers))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_idle_users", "Total idle users", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TotalIdleUsers))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_read_ops", "Total read operations across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalReadOps))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_read_bytes", "Total read bytes across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalReadBytes))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_write_ops", "Total write operations across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalWriteOps))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_write_bytes", "Total write bytes across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalWriteBytes))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_data_requests", "Total data buffer requests across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalDataRequests))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_data_hits", "Total data buffer hits across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalDataHits))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_index_requests", "Total index buffer requests across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalIndexRequests))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_index_hits", "Total index buffer hits across all users", nil, nil),
		prometheus.CounterValue, float64(snapshot.TotalIndexHits))
}
