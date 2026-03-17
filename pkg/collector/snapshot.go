package collector

// #cgo linux LDFLAGS: -L${SRCDIR} -lsnapshot -L/usr/lib64/faircom -Wl,-rpath,/usr/lib64/faircom -lmtclient
// #cgo windows LDFLAGS: -L${SRCDIR} -lsnapshot -lmtclient
// #cgo linux CFLAGS: -fPIC
// #include <stdlib.h>
// #include <c/snapshot.h>
//
import "C"
import (
	"fmt"
	"unsafe"
)

// SnapshotData holds comprehensive FairCom metrics from multiple ctstat reports

type SnapshotData struct {
	DataCacheHit      uint64
	DataCacheMiss     uint64
	IndexCacheHit     uint64
	IndexCacheMiss    uint64
	FileReadOps       uint64
	FileReadBytes     uint64
	FileWriteOps      uint64
	FileWriteBytes    uint64
	CommReadOps       uint64
	CommReadBytes     uint64
	CommWriteOps      uint64
	CommWriteBytes    uint64
	TranLogReadOps    uint64
	TranLogReadBytes  uint64
	TranLogWriteOps   uint64
	TranLogWriteBytes uint64

	CtreeCallCount uint64
	CtreeCallTime  uint64
	CommIdleTime   uint64
	CommSendTime   uint64

	TranBegins           uint64
	TranCommits          uint64
	TranAborts           uint64
	TranSavepoints       uint64
	TranRestores         uint64
	TranLogFlush         uint64
	TotalTransactions    uint64
	TotalTransactionTime uint64

	CurrentSystemFilesOpen uint64
	MaxSystemFilesOpen     uint64
	CurrentFilesOpen       uint64
	MaxFilesOpen           uint64
	UsersActive            uint64
	MaxUsersActive         uint64

	CurrentLocksHeld uint64
	CurrentLockWaits uint64
	LockHit          uint64
	LockMiss         uint64
	LockWaits        uint64
	Deadlocks        uint64

	IsamAdds    uint64
	IsamDeletes uint64
	IsamUpdates uint64
	IsamReads   uint64

	SystemFileOpens  uint64
	SystemFileCloses uint64
	FileOpens        uint64
	FileCloses       uint64
	FileCreates      uint64
	FileRenames      uint64
	FileDeletes      uint64
	TotalMemory      uint64
}

// initSnapshot initializes the faircomDB connection by calling the C function. Returns 0 on error, or owner on success
func initSnapshot(servername, username, password, clientcert, clientkey, passphrase, cafile string) (int, error) {
	// treat these empty go strings as C NULL values
	var cert *C.char
	if clientcert != "" {
		cert = C.CString(clientcert)
		defer C.free(unsafe.Pointer(cert))
	}
	var key *C.char
	if clientkey != "" {
		key = C.CString(clientkey)
		defer C.free(unsafe.Pointer(key))
	}
	var phrase *C.char
	if passphrase != "" {
		phrase = C.CString(passphrase)
		defer C.free(unsafe.Pointer(phrase))
	}
	var ca *C.char
	if cafile != "" {
		ca = C.CString(cafile)
		defer C.free(unsafe.Pointer(ca))
	}
	server := C.CString(servername)
	user := C.CString(username)
	pw := C.CString(password)
	defer C.free(unsafe.Pointer(server))
	defer C.free(unsafe.Pointer(user))
	defer C.free(unsafe.Pointer(pw))
	ret := C.InitSnapshot(server, user, pw, cert, key, phrase, ca)
	if ret != 0 {
		return 0, fmt.Errorf("FaircomDB error %d", ret)
	}
	return int(C.ctOWNER()), nil
}

// termSnapshot terminates and cleans up the faircomDb connection by calling the C function
func termSnapshot(owner int) {
	cowner := C.int(owner)
	C.ctSetOWNER(cowner)
	C.TermSnapshot()
}

// GetSnapshot retrieves comprehensive metrics using Snapshot call.
func GetSnapshot(owner int) (*SnapshotData, error) {
	data := &SnapshotData{}
	Cdata := C.struct_SnapshotDataC{}
	// set the current owner
	cowner := C.int(owner)
	C.ctSetOWNER(cowner)
	// Call the C function to populate the C struct with all metrics.
	// NOTE: This is a small subset of available C metrics.
	ret := C.GetSnapshotData(&Cdata, C.size_t(unsafe.Sizeof(Cdata)))
	if ret != 0 {
		return nil, fmt.Errorf("failed to get snapshot data: %d", ret)
	}

	// Map the C struct fields to the Go struct
	data.DataCacheHit = uint64(Cdata.DataCacheHit)
	data.DataCacheMiss = uint64(Cdata.DataCacheMiss)
	data.IndexCacheHit = uint64(Cdata.IndexCacheHit)
	data.IndexCacheMiss = uint64(Cdata.IndexCacheMiss)
	data.FileReadOps = uint64(Cdata.FileReadOps)
	data.FileReadBytes = uint64(Cdata.FileReadBytes)
	data.FileWriteOps = uint64(Cdata.FileWriteOps)
	data.FileWriteBytes = uint64(Cdata.FileWriteBytes)
	data.CommReadOps = uint64(Cdata.CommReadOps)
	data.CommReadBytes = uint64(Cdata.CommReadBytes)
	data.CommWriteOps = uint64(Cdata.CommWriteOps)
	data.CommWriteBytes = uint64(Cdata.CommWriteBytes)
	data.TranLogReadOps = uint64(Cdata.TranLogReadOps)
	data.TranLogReadBytes = uint64(Cdata.TranLogReadBytes)
	data.TranLogWriteOps = uint64(Cdata.TranLogWriteOps)
	data.TranLogWriteBytes = uint64(Cdata.TranLogWriteBytes)

	data.CtreeCallCount = uint64(Cdata.CtreeCallCount)
	data.CtreeCallTime = uint64(Cdata.CtreeCallTime)
	data.CommIdleTime = uint64(Cdata.CommIdleTime)
	data.CommSendTime = uint64(Cdata.CommSendTime)

	data.TranBegins = uint64(Cdata.TranBegins)
	data.TranCommits = uint64(Cdata.TranCommits)
	data.TranAborts = uint64(Cdata.TranAborts)
	data.TranSavepoints = uint64(Cdata.TranSavepoints)
	data.TranRestores = uint64(Cdata.TranRestores)
	data.TranLogFlush = uint64(Cdata.TranLogFlush)
	data.TotalTransactions = uint64(Cdata.TotalTransactions)
	data.TotalTransactionTime = uint64(Cdata.TotalTransactionTime)

	data.CurrentSystemFilesOpen = uint64(Cdata.CurrentSystemFilesOpen)
	data.MaxSystemFilesOpen = uint64(Cdata.MaxSystemFilesOpen)
	data.CurrentFilesOpen = uint64(Cdata.CurrentFilesOpen)
	data.MaxFilesOpen = uint64(Cdata.MaxFilesOpen)
	data.UsersActive = uint64(Cdata.UsersActive)
	data.MaxUsersActive = uint64(Cdata.MaxUsersActive)

	data.CurrentLocksHeld = uint64(Cdata.CurrentLocksHeld)
	data.CurrentLockWaits = uint64(Cdata.CurrentLockWaits)
	data.LockHit = uint64(Cdata.LockHit)
	data.LockMiss = uint64(Cdata.LockMiss)
	data.LockWaits = uint64(Cdata.LockWaits)
	data.Deadlocks = uint64(Cdata.Deadlocks)

	data.IsamAdds = uint64(Cdata.IsamAdds)
	data.IsamDeletes = uint64(Cdata.IsamDeletes)
	data.IsamUpdates = uint64(Cdata.IsamUpdates)
	data.IsamReads = uint64(Cdata.IsamReads)

	data.SystemFileOpens = uint64(Cdata.SystemFileOpens)
	data.SystemFileCloses = uint64(Cdata.SystemFileCloses)
	data.FileOpens = uint64(Cdata.FileOpens)
	data.FileCloses = uint64(Cdata.FileCloses)
	data.FileCreates = uint64(Cdata.FileCreates)
	data.FileRenames = uint64(Cdata.FileRenames)
	data.FileDeletes = uint64(Cdata.FileDeletes)
	data.TotalMemory = uint64(Cdata.TotalMemory)

	return data, nil
}
