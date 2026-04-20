package collector

import (
	"github.com/faircom/prometheus-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"runtime"
	"sync"
)

// FairComCollector collects metrics from FairCom using ctstat
type FairComCollector struct {
	config       config.FairComConfig
	metric       config.CollectorsConfig
	logger       *logrus.Logger
	isConnected  bool
	faircomOwner int
	FCcleanup    runtime.Cleanup
	mutex        *sync.Mutex //serializes faircomDB access
}

// NewFairComCollector creates a new FairCom collector
func NewFairComCollector(cfg *config.Config) *FairComCollector {
	obj := &FairComCollector{
		config:       cfg.FairCom,
		metric:       cfg.Collectors,
		logger:       logrus.StandardLogger(),
		isConnected:  false,
		faircomOwner: 0,
		mutex:        &sync.Mutex{},
	}
	return obj
}

// Describe implements prometheus.Collector
func (c *FairComCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil)
}

// Collect implements prometheus.Collector
func (c *FairComCollector) Collect(ch chan<- prometheus.Metric) {
	//Ensure only a single caller on this object uses the FaircomDb connection
	var err error
	var tid int
	c.mutex.Lock()
	if c.isConnected == false {
		// try to connect
		var certFile, keyFile, caFile string
		if c.config.TLS.Enabled {
			certFile = c.config.TLS.CertFile
			keyFile = c.config.TLS.KeyFile
			caFile = c.config.TLS.CAFile
		}
		tid, err = initSnapshot(c.config.Servername, c.config.BasicAuth.Username, c.config.BasicAuth.Password, certFile, keyFile, "", caFile)
		if tid == 0 && err != nil {
			c.mutex.Unlock()
			c.logger.Errorf("Failed to initialize snapshot: %v", err)
			c.logger.Errorf("Server: %s User: %s", c.config.Servername, c.config.BasicAuth.Username)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil),
				prometheus.GaugeValue, 0)
			return
		}
		// cleanup database connection when GC reclaims
		c.FCcleanup = runtime.AddCleanup(c, termSnapshot, tid)
		c.faircomOwner = tid
		c.isConnected = true
	}
	snapshot, err := GetSnapshot(c.faircomOwner)
	if err != nil {
		c.logger.Errorf("Failed to get snapshot: %v", err)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil),
			prometheus.GaugeValue, 0)
		termSnapshot(c.faircomOwner)
		// cancel the cleanup, we did it manually
		c.FCcleanup.Stop()
		c.isConnected = false
		c.faircomOwner = 0
		c.mutex.Unlock()
		return
	}
	c.mutex.Unlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_up", "FairCom server is up", nil, nil),
		prometheus.GaugeValue, 1)

	// Cache metrics - hit AND miss
	if c.metric.Cache == true {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_cache_hit", "Cache hit count", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.DataCacheHit), "data")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_cache_hit", "Cache hit count", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.IndexCacheHit), "index")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_cache_miss", "Cache miss count", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.DataCacheMiss), "data")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_cache_miss", "Cache miss count", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.IndexCacheMiss), "index")
	}
	// I/O metrics
	if c.metric.IO == true {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_ops", "Read operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.FileReadOps), "disk")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_bytes", "Read bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.FileReadBytes), "disk")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_ops", "Write operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.FileWriteOps), "disk")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_bytes", "Write bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.FileWriteBytes), "disk")

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_ops", "Read operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.CommReadOps), "ipc")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_bytes", "Read bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.CommReadBytes), "ipc")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_ops", "Write operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.CommWriteOps), "ipc")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_bytes", "Write bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.CommWriteBytes), "ipc")

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_ops", "Read operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.TranLogReadOps), "transaction log")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_read_bytes", "Read bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.TranLogReadBytes), "transaction log")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_ops", "Write operations", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.TranLogWriteOps), "transaction log")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_write_bytes", "Write bytes", []string{"type"}, nil),
			prometheus.CounterValue, float64(snapshot.TranLogWriteBytes), "transaction log")
	}

	if c.metric.Transactions == true {
		// Transaction metrics
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
			prometheus.NewDesc("faircom_transaction_log_flush_total", "Transaction log flushes", nil, nil),
			prometheus.CounterValue, float64(snapshot.TranLogFlush))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_transaction_total", "Transaction total", nil, nil),
			prometheus.CounterValue, float64(snapshot.TotalTransactions))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_transaction_total_time", "Transaction total time(seconds)", nil, nil),
			prometheus.CounterValue, float64(snapshot.TotalTransactionTime))
	}
	if c.metric.Users == true {
		// Connection metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_users_current", "Users currently connected", nil, nil),
			prometheus.GaugeValue, float64(snapshot.UsersActive))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_users_max", "Maximum users connected", nil, nil),
			prometheus.CounterValue, float64(snapshot.MaxUsersActive))
	}
	if c.metric.Locks == true {
		// Lock metrics - hit AND miss
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_lock_held_current", "Locks currently held", nil, nil),
			prometheus.GaugeValue, float64(snapshot.CurrentLocksHeld))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_lock_wait_current", "Locks currently waiting", nil, nil),
			prometheus.GaugeValue, float64(snapshot.CurrentLockWaits))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_lock_hit_total", "Lock hit count", nil, nil),
			prometheus.CounterValue, float64(snapshot.LockHit))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_lock_miss_total", "Lock miss count", nil, nil),
			prometheus.CounterValue, float64(snapshot.LockMiss))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_lock_wait_total", "Lock wait count", nil, nil),
			prometheus.CounterValue, float64(snapshot.LockWaits))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_deadlocks_total", "Deadlock count", nil, nil),
			prometheus.CounterValue, float64(snapshot.Deadlocks))
	}
	if c.metric.ISAM == true {
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
	}
	if c.metric.SQL == true {
		// SQL metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_select_total", "SQL select operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLSelect))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_insert_total", "SQL insert operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLInsert))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_delete_total", "SQL delete operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLDelete))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_update_total", "SQL update operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLUpdate))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_commit_total", "SQL commit operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLCommit))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_sql_rollback_total", "SQL rollback operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SQLRollback))

	}
	if c.metric.CallTime == true {
		// work timing
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_isam_ipc_call_count", "Ctree call count", nil, nil),
			prometheus.CounterValue, float64(snapshot.CtreeCallCount))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_isam_ipc_call_time", "Ctree call time(seconds)", nil, nil),
			prometheus.CounterValue, float64(snapshot.CtreeCallTime))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_isam_ipc_idle_time", "Ipc idle time(seconds)", nil, nil),
			prometheus.CounterValue, float64(snapshot.CommIdleTime))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_isam_ipc_send_time", "Ipc send time(seconds)", nil, nil),
			prometheus.CounterValue, float64(snapshot.CommSendTime))
	}
	if c.metric.Files == true {
		// File metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_physical_files_open", "System files currently open", nil, nil),
			prometheus.GaugeValue, float64(snapshot.CurrentSystemFilesOpen))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_physical_files_open_max", "Maximum System files open", nil, nil),
			prometheus.CounterValue, float64(snapshot.MaxSystemFilesOpen))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_files_open_current", "Files currently open", nil, nil),
			prometheus.GaugeValue, float64(snapshot.CurrentFilesOpen))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_files_open_max", "Maximum files open", nil, nil),
			prometheus.CounterValue, float64(snapshot.MaxFilesOpen))

		// File operation metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_physical_file_opens_total", "System file open operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SystemFileOpens))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_physical_file_closes_total", "System file close operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.SystemFileCloses))
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
			prometheus.NewDesc("faircom_file_renames_total", "File rename operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.FileRenames))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("faircom_file_deletes_total", "File delete operations", nil, nil),
			prometheus.CounterValue, float64(snapshot.FileDeletes))

	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("faircom_total_memory_bytes", "Total heap memory usage", nil, nil),
		prometheus.GaugeValue, float64(snapshot.TotalMemory))
	return
}
