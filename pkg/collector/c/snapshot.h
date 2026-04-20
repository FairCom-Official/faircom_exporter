#include <stdint.h>
#include <stddef.h>

// SnapshotData holds comprehensive FairCom metrics from multiple snapshot reports
struct SnapshotDataC {
    uint64_t DataCacheHit;
    uint64_t DataCacheMiss;
    uint64_t IndexCacheHit;
    uint64_t IndexCacheMiss;
    uint64_t FileReadOps;
    uint64_t FileReadBytes;
    uint64_t FileWriteOps;
    uint64_t FileWriteBytes;
    uint64_t CommReadOps;
    uint64_t CommReadBytes;
    uint64_t CommWriteOps;
    uint64_t CommWriteBytes;
    uint64_t TranLogReadOps;
    uint64_t TranLogReadBytes;
    uint64_t TranLogWriteOps;
    uint64_t TranLogWriteBytes;

    uint64_t CtreeCallCount;
    uint64_t CtreeCallTime;
    uint64_t CommIdleTime;
    uint64_t CommSendTime;

    uint64_t TranBegins;
    uint64_t TranCommits;
    uint64_t TranAborts;
    uint64_t TranSavepoints;
    uint64_t TranRestores;
    uint64_t TranLogFlush;
    uint64_t TotalTransactions;
    uint64_t TotalTransactionTime;
   
    uint64_t CurrentSystemFilesOpen;
    uint64_t MaxSystemFilesOpen;
    uint64_t CurrentFilesOpen;
    uint64_t MaxFilesOpen;
    uint64_t UsersActive;
    uint64_t MaxUsersActive;

    uint64_t CurrentLocksHeld;
    uint64_t CurrentLockWaits;
    uint64_t LockHit;
    uint64_t LockMiss;
    uint64_t LockWaits;
    uint64_t Deadlocks;

    uint64_t IsamAdds;
    uint64_t IsamDeletes;
    uint64_t IsamUpdates;
    uint64_t IsamReads;

    uint64_t SystemFileOpens;
    uint64_t SystemFileCloses;
    uint64_t FileOpens;
    uint64_t FileCloses;
    uint64_t FileCreates;
    uint64_t FileRenames;
    uint64_t FileDeletes;
    uint64_t TotalMemory;

    /* SQL activity */
    uint64_t SQLSelect;
    uint64_t SQLInsert;
    uint64_t SQLUpdate;
    uint64_t SQLDelete;
    uint64_t SQLCommit;
    uint64_t SQLRollback;
};

int InitSnapshot(const char * servername, const char * username, const char * password, const char * clientauthfile, const char * clientkeyfile,const char * passphrase,const char * cafile);
int GetSnapshotData(struct SnapshotDataC* data, size_t datasize);
void TermSnapshot(void);
int ctOWNER(void);
void ctSetOWNER(int);

