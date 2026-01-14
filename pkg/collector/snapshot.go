package collector

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SnapshotData holds comprehensive FairCom metrics from multiple ctstat reports
type SnapshotData struct {
	// From -vas (Admin System) - expanded to capture ALL fields
	DataCacheHitPct   float64
	DataCacheMissPct  float64
	IndexCacheHitPct  float64
	IndexCacheMissPct float64
	ReadOpsPerSec     int64
	WriteOpsPerSec    int64
	FilesOpen         int64
	FilesMax          int64
	UsersActive       int64
	UsersMax          int64
	LocksHeld         int64
	LockHitPct        float64
	LockMissPct       float64
	Deadlocks         int64
	TransActive       int64
	TransPerSec       int64
	ReadTranTime      int64
	WriteTranTime     int64

	// From -vat (Admin Transaction)
	TranBegins     int64
	TranCommits    int64
	TranAborts     int64
	TranSavepoints int64
	TranRestores   int64
	TranLogWrites  int64
	TranLogBytes   int64

	// From -isam (ISAM Activity)
	IsamAdds    int64
	IsamDeletes int64
	IsamUpdates int64
	IsamReads   int64
	IsamFirst   int64
	IsamLast    int64
	IsamNext    int64
	IsamPrev    int64

	// From -sql (SQL Activity)
	SQLSelects   int64
	SQLInserts   int64
	SQLUpdates   int64
	SQLDeletes   int64
	SQLCommits   int64
	SQLRollbacks int64

	// From -fileops (File Operations)
	FileOpens      int64
	FileCloses     int64
	FileCreates    int64
	FileDeletes    int64
	FileRenames    int64
	PhysicalReads  int64
	PhysicalWrites int64

	// From -userinfox (aggregate across all users)
	TotalMemoryKB      int64
	TotalReadOps       int64
	TotalReadBytes     int64
	TotalWriteOps      int64
	TotalWriteBytes    int64
	TotalDataRequests  int64
	TotalDataHits      int64
	TotalIndexRequests int64
	TotalIndexHits     int64
	TotalActiveUsers   int64
	TotalIdleUsers     int64
}

// GetSnapshot retrieves comprehensive metrics using multiple ctstat reports
func GetSnapshot(username, password, ctstatPath string, timeout time.Duration) (*SnapshotData, error) {
	data := &SnapshotData{}

	// Get -vas (Admin System Report) - single snapshot
	vas, err := runCtstat("-vas", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseVas(vas, data)
	}

	// Get -vat (Admin Transaction Report) - single snapshot
	vat, err := runCtstat("-vat", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseVat(vat, data)
	}

	// Get -isam (ISAM Activity Report) - single snapshot
	isam, err := runCtstat("-isam", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseIsam(isam, data)
	}

	// Get -sql (SQL Activity Report) - single snapshot
	sql, err := runCtstat("-sql", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseSQL(sql, data)
	}

	// Get -fileops (File Operations Report) - single snapshot
	fileops, err := runCtstat("-fileops", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseFileops(fileops, data)
	}

	// Get -userinfox (Extended User Report) - single snapshot to aggregate user metrics
	userinfox, err := runCtstat("-userinfox", username, password, ctstatPath, timeout, true)
	if err == nil {
		parseUserinfox(userinfox, data)
	}

	return data, nil
}

// runCtstat executes ctstat command with timeout
func runCtstat(report, username, password, ctstatPath string, timeout time.Duration, singleSnapshot bool) (string, error) {
	args := []string{report, "-u", username, "-p", password}
	if singleSnapshot {
		args = append(args, "-i", "1", "1", "-h", "1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, ctstatPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// parseVas parses -vas (Admin System Report)
// Format:      cache     disk i/o   files   connect        locks         transactions
//
//	d%h %m i%h %m  r/s  w/s  cur/max  cur/max   cur l%h %m  dead act t/s  r/t  w/t
func parseVas(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return
	}

	dataLine := getLastNonEmptyLine(lines)
	fields := strings.Fields(dataLine)
	if len(fields) < 13 {
		return
	}

	// Cache stats - capture hit AND miss percentages
	data.DataCacheHitPct = parseFloat(fields[0])
	data.DataCacheMissPct = parseFloat(fields[1])
	data.IndexCacheHitPct = parseFloat(fields[2])
	data.IndexCacheMissPct = parseFloat(fields[3])

	// Disk I/O
	data.ReadOpsPerSec = parseInt(fields[4])
	data.WriteOpsPerSec = parseInt(fields[5])

	// Parse files: cur/max
	filesParts := strings.Split(fields[6], "/")
	if len(filesParts) == 2 {
		data.FilesOpen = parseInt(filesParts[0])
		data.FilesMax = parseInt(filesParts[1])
	}

	// Parse connections: cur/max
	connParts := strings.Split(fields[7], "/")
	if len(connParts) == 2 {
		data.UsersActive = parseInt(connParts[0])
		data.UsersMax = parseInt(connParts[1])
	}

	// Lock stats - capture hit AND miss percentages
	data.LocksHeld = parseInt(fields[8])
	data.LockHitPct = parseFloat(fields[9])
	data.LockMissPct = parseFloat(fields[10])
	data.Deadlocks = parseInt(fields[11])

	// Transaction stats
	data.TransActive = parseInt(fields[12])
	if len(fields) > 13 {
		data.TransPerSec = parseInt(fields[13])
	}
	if len(fields) > 14 {
		data.ReadTranTime = parseInt(fields[14])
	}
	if len(fields) > 15 {
		data.WriteTranTime = parseInt(fields[15])
	}
}

// parseVat parses -vat (Admin Transaction Report)
func parseVat(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")
	dataLine := getLastNonEmptyLine(lines)
	fields := strings.Fields(dataLine)

	if len(fields) >= 6 {
		data.TranBegins = parseInt(fields[0])
		data.TranCommits = parseInt(fields[1])
		data.TranAborts = parseInt(fields[2])
		data.TranSavepoints = parseInt(fields[3])
		data.TranRestores = parseInt(fields[4])
	}
	if len(fields) >= 8 {
		data.TranLogWrites = parseInt(fields[6])
		data.TranLogBytes = parseInt(fields[7])
	}
}

// parseIsam parses -isam (ISAM Activity Report)
func parseIsam(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")
	dataLine := getLastNonEmptyLine(lines)
	fields := strings.Fields(dataLine)

	if len(fields) >= 8 {
		data.IsamAdds = parseInt(fields[0])
		data.IsamDeletes = parseInt(fields[1])
		data.IsamUpdates = parseInt(fields[2])
		data.IsamReads = parseInt(fields[3])
		data.IsamFirst = parseInt(fields[4])
		data.IsamLast = parseInt(fields[5])
		data.IsamNext = parseInt(fields[6])
		data.IsamPrev = parseInt(fields[7])
	}
}

// parseSQL parses -sql (SQL Activity Report)
func parseSQL(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")
	dataLine := getLastNonEmptyLine(lines)
	fields := strings.Fields(dataLine)

	if len(fields) >= 6 {
		data.SQLSelects = parseInt(fields[0])
		data.SQLInserts = parseInt(fields[1])
		data.SQLUpdates = parseInt(fields[2])
		data.SQLDeletes = parseInt(fields[3])
		data.SQLCommits = parseInt(fields[4])
		data.SQLRollbacks = parseInt(fields[5])
	}
}

// parseFileops parses -fileops (File Operations Report)
func parseFileops(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")
	dataLine := getLastNonEmptyLine(lines)
	fields := strings.Fields(dataLine)

	if len(fields) >= 7 {
		data.FileOpens = parseInt(fields[0])
		data.FileCloses = parseInt(fields[1])
		data.FileCreates = parseInt(fields[2])
		data.FileDeletes = parseInt(fields[3])
		data.FileRenames = parseInt(fields[4])
		data.PhysicalReads = parseInt(fields[5])
		data.PhysicalWrites = parseInt(fields[6])
	}
}

// parseUserinfox parses -userinfox (Extended User Report) and aggregates across all users
// Format: status lastrequest trntime mem fils time uid/tid/nodename commprotocol readops readbytes writeops writebytes datarqsts datahits indexrqsts indexhits
func parseUserinfox(output string, data *SnapshotData) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		// Skip headers and lines without enough fields
		if len(fields) < 18 || fields[0] == "status" {
			continue
		}

		// fields[0] = status (idle/busy)
		// fields[3] = mem (e.g., "137K")
		// fields[8] = readops
		// fields[9] = readbytes
		// fields[10] = writeops
		// fields[11] = writebytes
		// fields[12] = datarqsts
		// fields[13] = datahits
		// fields[14] = indexrqsts
		// fields[15] = indexhits

		// Track active vs idle users
		if fields[0] == "busy" {
			data.TotalActiveUsers++
		} else if fields[0] == "idle" {
			data.TotalIdleUsers++
		}

		// Aggregate memory (parse KB value like "137K")
		memStr := fields[3]
		if strings.HasSuffix(memStr, "K") {
			memKB := parseInt(strings.TrimSuffix(memStr, "K"))
			data.TotalMemoryKB += memKB
		}

		// Aggregate I/O operations
		data.TotalReadOps += parseInt(fields[8])
		data.TotalReadBytes += parseInt(fields[9])
		data.TotalWriteOps += parseInt(fields[10])
		data.TotalWriteBytes += parseInt(fields[11])

		// Aggregate cache statistics
		data.TotalDataRequests += parseInt(fields[12])
		data.TotalDataHits += parseInt(fields[13])
		data.TotalIndexRequests += parseInt(fields[14])
		data.TotalIndexHits += parseInt(fields[15])
	}
}

// Helper functions
func getLastNonEmptyLine(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "-") && !strings.Contains(line, "cache") {
			return line
		}
	}
	return ""
}

func parseInt(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
