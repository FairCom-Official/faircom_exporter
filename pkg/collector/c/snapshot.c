/* FaircomDB SNAPSHOT api wrapper for cgo */
#include <stdio.h>
#include "ctreep.h"
#include "snapshot.h"

static int isInitialized = 0;

/*
* @brief Terminate snapshot connection and clean up resources.
* 
*/
void TermSnapshot(void) {
	if (isInitialized) {
		STPUSR();
		ctThrdTerm();
		isInitialized = 0;
	}
}

/*
* @brief Initialize FaircomDB connection options for snapshot retrieval.
* 
* @param servername [IN] Pointer to the faircomDB server name or IP address to connect to.
*
* @param username [IN] Pointer to the username for authentication with the faircomDB server.(can be NULL if clientauthfile is provided).
* 
* @param password [IN] Pointer to the password for authentication with the faircomDB server.(can be NULL if clientauthfile is provided).
* 
* @param clientauthfile [IN] Optional pointer to a client certificate file for TLS authentication (can be NULL if not using TLS).
* 
* @param clientkeyfile [IN] Optional pointer to a client private key file for TLS authentication (can be NULL if not using TLS).
* 
* @param passphrase [IN] Optional pointer to a passphrase for the client private key (can be NULL if not using TLS or if the key is not encrypted).
* 
* @param cafile [IN] Optional pointer to a CA certificate file for TLS authentication (can be NULL if not using TLS).
* 
* @return 0 on success, or a non-zero error code on failure.
*/
int InitSnapshot(const char * servername, const char * username, const char * password, const char * clientauthfile,
		const char * clientkeyfile,const char * passphrase,const char * cafile)
{
	int rc;

	if(!servername)
	{
		printf("No servername provided.\n");
		return -1;
	}
	
	/* initialize FaircomDB library */
	rc = ctThrdInit(1, 0, NULL);
	if (rc)
	{
		printf("Could not initialize c-tree threading: error %d\n", rc);
		return rc;
	}
	/* set optional TLS information */
	if (cafile)
	{
		rc = ctSetCommProtocolOption(ctCOMMOPT_FSSLTCP_SERVER_CERTIFICATE, cafile);
		if (rc)
		{
			printf("Could not set CA file: error %d\n", rc);
			goto err_ret;
		}
	}
	if (clientauthfile)
	{
		rc = ctSetCommProtocolOption(ctCOMMOPT_FSSLTCP_CLIENT_CERTIFICATE, clientauthfile);
		if (rc)
		{
			printf("Could not set client auth file: error %d\n", rc);
			goto err_ret;
		}
	}
	if (clientkeyfile)
	{
		rc = ctSetCommProtocolOption(ctCOMMOPT_FSSLTCP_CLIENT_KEY, clientkeyfile);
		if (rc)
		{
			printf("Could not set client key file: error %d\n", rc);
			goto err_ret;
		}
	}
	if(passphrase)
	{
		rc = ctSetCommProtocolOption(ctCOMMOPT_FSSLTCP_CLIENT_PASSPHRASE, passphrase);
		if (rc)
		{
			printf("Could not set client key passphrase: error %d\n", rc);
			goto err_ret;
		}
	}
	/* initialize ISAM connection */
	rc = INTISAMX(0, 0, 512, 0, 0, username, password, servername);
	if(rc)
	{
		ctrt_printf("Could not initialize FaircomDB: error %d\n", rc);
		goto err_ret;
	}
	isInitialized = 1;
	return 0;
err_ret:
	ctThrdTerm();
	return rc;
}

int GetSnapshotData(struct SnapshotDataC* data, size_t datasize)
{
	ctGSMS ServerStats;

	if (!isInitialized) {
		printf("FaircomDB not initialized.\n");
		return -1;
	}
	if (!data) {
		printf("Invalid data pointer.\n");
		return -1;
	}
	if (datasize != sizeof(struct SnapshotDataC)) {
		printf("Invalid SnapshotData size: expected %zu, got %zu\n", sizeof(struct SnapshotDataC), datasize);
		return -1;
	}

	/* FairComDB SnapShot API expects a buffer and size */
	int rc = SnapShot(ctPSSsystem,NULL,&ServerStats, sizeof(ServerStats));
	if (rc) {
		printf("SnapShot failed: error %d\n", rc);
		return rc;
	}
	/* Map fields from ServerStats to SnapshotData */ 
	data->DataCacheHit = (uint64_t)ServerStats.sct_dbhit;
	data->DataCacheMiss = (uint64_t)(ServerStats.sct_dbrqs - ServerStats.sct_dbhit);
	data->IndexCacheHit = (uint64_t)ServerStats.sct_ibhit;
	data->IndexCacheMiss = (uint64_t)(ServerStats.sct_ibrqs - ServerStats.sct_ibhit);
	data->FileReadOps = (uint64_t)ServerStats.sct_rdops;
	data->FileReadBytes = (uint64_t)ServerStats.sct_rdbyt;
	data->FileWriteOps = (uint64_t)ServerStats.sct_wrops;
	data->FileWriteBytes = (uint64_t)ServerStats.sct_wrbyt;
	data->CommReadOps = (uint64_t)ServerStats.sct_rcops;
	data->CommReadBytes = (uint64_t)ServerStats.sct_rcbyt;
	data->CommWriteOps = (uint64_t)ServerStats.sct_wcops;
	data->CommWriteBytes = (uint64_t)ServerStats.sct_wcbyt;
	data->TranLogReadOps = (uint64_t)ServerStats.sctrlgops;
	data->TranLogReadBytes = (uint64_t)ServerStats.sctrlgbyt;
	data->TranLogWriteOps = (uint64_t)ServerStats.sctwlgops;
	data->TranLogWriteBytes = (uint64_t)ServerStats.sctwlgbyt;

	data->CtreeCallCount = (uint64_t)ServerStats.scttot_call;
	data->CtreeCallTime = (uint64_t)ServerStats.scttot_work / ServerStats.scthrtimbas;
	data->CommIdleTime = (uint64_t)ServerStats.scttot_recv  / ServerStats.scthrtimbas;
	data->CommSendTime = (uint64_t)ServerStats.scttot_send  / ServerStats.scthrtimbas;




	data->TranBegins = (uint64_t)ServerStats.sct_trbeg;
	data->TranCommits = (uint64_t)ServerStats.sct_trend;
	data->TranAborts = (uint64_t)ServerStats.sct_trabt;
	data->TranSavepoints = (uint64_t)ServerStats.sct_trsav;
	data->TranRestores = (uint64_t)ServerStats.sct_trrst;
	data->TranLogFlush= (uint64_t)ServerStats.sct_trfls;
	data->TotalTransactions = (uint64_t)ServerStats.scttrncnt;
	data->TotalTransactionTime = (uint64_t)ServerStats.scttrntim / ServerStats.scthrtimbas;

	data->CurrentSystemFilesOpen = (uint64_t)ServerStats.sctactfil;
	data->MaxSystemFilesOpen = (uint64_t)ServerStats.sctactfilx;
	data->CurrentFilesOpen = (uint64_t)ServerStats.scttotblk;
	data->MaxFilesOpen = (uint64_t)ServerStats.scttotblkx;
	data->UsersActive = (uint64_t)ServerStats.sctnusers;
	data->MaxUsersActive = (uint64_t)ServerStats.sctnusersx;

	data->CurrentLocksHeld = (uint64_t)ServerStats.sctlokcur;
	data->CurrentLockWaits = (uint64_t)ServerStats.sctblkcur;
	
	data->LockMiss = (uint64_t)(ServerStats.sctlokdny + ServerStats.sctlokblk);
	data->LockHit = (uint64_t)(ServerStats.sctloktry - data->LockMiss);
	data->LockWaits = (uint64_t)ServerStats.sctlokblk;
	data->Deadlocks = (uint64_t)ServerStats.sctlokdlk;

	data->IsamAdds = (uint64_t)ServerStats.sctismaddcnt;
	data->IsamDeletes = (uint64_t)ServerStats.sctismdelcnt;
	data->IsamUpdates = (uint64_t)ServerStats.sctismupdcnt;
	data->IsamReads = (uint64_t)ServerStats.sctismredcnt;

	data->SystemFileOpens = (uint64_t)ServerStats.sphyopncnt;
	data->SystemFileCloses = (uint64_t)ServerStats.sphyclscnt;
	data->FileOpens = (uint64_t)ServerStats.slogopncnt;
	data->FileCloses = (uint64_t)ServerStats.slogclscnt;
	data->FileCreates = (uint64_t)ServerStats.sfilcrecnt;
	data->FileRenames = (uint64_t)ServerStats.sfilrencnt;
	data->FileDeletes = (uint64_t)ServerStats.sfildelcnt;

	data->TotalMemory = (uint64_t)ServerStats.sctmemsum;

	return rc;
}
