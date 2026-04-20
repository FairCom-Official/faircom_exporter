/* FaircomDB SNAPSHOT api wrapper for cgo */
#include <stdio.h>
#include "ctreep.h"
#include "snapshot.h"

static int isInitialized = 0;

/*
* @brief Terminate snapshot connection and clean up resources.
* 
*/
void TermSnapshot(void)
{
	if (isInitialized)
	{
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

/*
* @brief Collect data FaircomDB using snapshot api.
* 
* @param data [OUT] Pointer  struct to be filled with snapshot data. 
*
* @param datasize [IN] size of data
* 
* @return 0 on success, or a non-zero error code on failure.
*/
int GetSnapshotData(struct SnapshotDataC* data, size_t datasize)
{
	union
	{
		ctGSMS system;
		ctSQLS sql;
	} stats;

	if (!isInitialized)
	{
		printf("FaircomDB not initialized.\n");
		return -1;
	}
	if (!data)
	{
		printf("Invalid data pointer.\n");
		return -1;
	}
	if (datasize != sizeof(struct SnapshotDataC))
	{
		printf("Invalid SnapshotData size: expected %zu, got %zu\n", sizeof(struct SnapshotDataC), datasize);
		return -1;
	}
	memset(&stats, 0, sizeof(stats));

	/* FairComDB SnapShot API expects a buffer and size */
	int rc = SnapShot(ctPSSsystem,NULL,&stats.system, sizeof(stats.system));
	if (rc)
	{
		printf("SnapShot failed: error %d\n", rc);
		return rc;
	}
	/* Map fields from ServerStats to SnapshotData */ 
	data->DataCacheHit = (uint64_t)stats.system.sct_dbhit;
	data->DataCacheMiss = (uint64_t)(stats.system.sct_dbrqs - stats.system.sct_dbhit);
	data->IndexCacheHit = (uint64_t)stats.system.sct_ibhit;
	data->IndexCacheMiss = (uint64_t)(stats.system.sct_ibrqs - stats.system.sct_ibhit);
	data->FileReadOps = (uint64_t)stats.system.sct_rdops;
	data->FileReadBytes = (uint64_t)stats.system.sct_rdbyt;
	data->FileWriteOps = (uint64_t)stats.system.sct_wrops;
	data->FileWriteBytes = (uint64_t)stats.system.sct_wrbyt;
	data->CommReadOps = (uint64_t)stats.system.sct_rcops;
	data->CommReadBytes = (uint64_t)stats.system.sct_rcbyt;
	data->CommWriteOps = (uint64_t)stats.system.sct_wcops;
	data->CommWriteBytes = (uint64_t)stats.system.sct_wcbyt;
	data->TranLogReadOps = (uint64_t)stats.system.sctrlgops;
	data->TranLogReadBytes = (uint64_t)stats.system.sctrlgbyt;
	data->TranLogWriteOps = (uint64_t)stats.system.sctwlgops;
	data->TranLogWriteBytes = (uint64_t)stats.system.sctwlgbyt;

	data->CtreeCallCount = (uint64_t)stats.system.scttot_call;
	if(stats.system.scthrtimbas) 
	{
		data->CtreeCallTime = (uint64_t)stats.system.scttot_work / stats.system.scthrtimbas;
		data->CommIdleTime = (uint64_t)stats.system.scttot_recv  / stats.system.scthrtimbas;
		data->CommSendTime = (uint64_t)stats.system.scttot_send  / stats.system.scthrtimbas;
		data->TotalTransactionTime = (uint64_t)stats.system.scttrntim / stats.system.scthrtimbas;
	} 
	else 
	{
		data->CtreeCallTime = 0;
		data->CommIdleTime = 0;
		data->CommSendTime = 0;
		data->TotalTransactionTime = 0;
	}

	data->TranBegins = (uint64_t)stats.system.sct_trbeg;
	data->TranCommits = (uint64_t)stats.system.sct_trend;
	data->TranAborts = (uint64_t)stats.system.sct_trabt;
	data->TranSavepoints = (uint64_t)stats.system.sct_trsav;
	data->TranRestores = (uint64_t)stats.system.sct_trrst;
	data->TranLogFlush= (uint64_t)stats.system.sct_trfls;
	data->TotalTransactions = (uint64_t)stats.system.scttrncnt;

	data->CurrentSystemFilesOpen = (uint64_t)stats.system.sctactfil;
	data->MaxSystemFilesOpen = (uint64_t)stats.system.sctactfilx;
	data->CurrentFilesOpen = (uint64_t)stats.system.scttotblk;
	data->MaxFilesOpen = (uint64_t)stats.system.scttotblkx;
	data->UsersActive = (uint64_t)stats.system.sctnusers;
	data->MaxUsersActive = (uint64_t)stats.system.sctnusersx;

	data->CurrentLocksHeld = (uint64_t)stats.system.sctlokcur;
	data->CurrentLockWaits = (uint64_t)stats.system.sctblkcur;
	
	data->LockMiss = (uint64_t)(stats.system.sctlokdny + stats.system.sctlokblk);
	data->LockHit = (uint64_t)(stats.system.sctloktry - data->LockMiss);
	data->LockWaits = (uint64_t)stats.system.sctlokblk;
	data->Deadlocks = (uint64_t)stats.system.sctlokdlk;

	data->IsamAdds = (uint64_t)stats.system.sctismaddcnt;
	data->IsamDeletes = (uint64_t)stats.system.sctismdelcnt;
	data->IsamUpdates = (uint64_t)stats.system.sctismupdcnt;
	data->IsamReads = (uint64_t)stats.system.sctismredcnt;

	data->SystemFileOpens = (uint64_t)stats.system.sphyopncnt;
	data->SystemFileCloses = (uint64_t)stats.system.sphyclscnt;
	data->FileOpens = (uint64_t)stats.system.slogopncnt;
	data->FileCloses = (uint64_t)stats.system.slogclscnt;
	data->FileCreates = (uint64_t)stats.system.sfilcrecnt;
	data->FileRenames = (uint64_t)stats.system.sfilrencnt;
	data->FileDeletes = (uint64_t)stats.system.sfildelcnt;

	data->TotalMemory = (uint64_t)stats.system.sctmemsum;

#if ctSQLSvern > 2
	/* FairComDB SnapShot API expects a buffer and size */
	rc = SnapShot(ctPSSsqlSystem,NULL,&stats.sql, sizeof(stats.sql));
	if (rc)
	{
		printf("SnapShot(ctPSSsqlSystem) error %d\n", rc);
		return rc;
	}
	if (stats.sql.server_ver > 2)
	{
		data->SQLSelect = stats.sql.select;
		data->SQLInsert = stats.sql.insert;
		data->SQLUpdate = stats.sql.update;
		data->SQLDelete = stats.sql.deletes;
		data->SQLCommit = stats.sql.commit;
		data->SQLRollback = stats.sql.rollback;
	}
#endif
	return rc;
}
