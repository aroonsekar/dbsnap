# dbSnap 📸

A production-grade, background scheduled MongoDB backup and restore CLI utility. 

dbSnap provides an interactive terminal UI to easily configure automated, compressed local backups for your MongoDB clusters. It safely locks your database credentials in the native OS vault (Windows Credential Manager / macOS Keychain) and hooks directly into the OS task scheduler to run silently in the background.

## Features
* **Interactive UI:** Select target databases and collections using a sleek terminal interface.
* **Multi-Database Support:** Target specific collections across multiple databases in a single run.
* **Native Scheduling:** Automatically registers background tasks (`schtasks` for Windows).
* **High Security:** Passwords never touch a config file. They are locked in the OS Keyring.
* **Data Purging (Optional):** Safely drop records from live collections immediately after a verified backup.
* **Automated Retention:** Enforces a rolling backup limit (default: 6 years) and rotates background logs (30 days).
* **Interactive Restore:** Arrow-key navigated menus to quickly recover data from historical timestamps.

## Prerequisites
* Go 1.21+
* [MongoDB Database Tools](https://www.mongodb.com/try/download/database-tools) (`mongodump` and `mongorestore` must be in your system PATH).

## Installation

Clone the repository and compile the executable:

```bash
git clone [https://github.com/yourusername/dbsnap.git](https://github.com/yourusername/dbsnap.git)
cd dbsnap
go mod tidy
go build -ldflags="-s -w" -o dbsnap main.go
```
## Usage
dbSnap is operated entirely through the CLI.

### 1. Initial Setup
Run the setup wizard to configure your target paths, database connections, and schedule the background task.

```bash
dbsnap setup
```

(Note: On Windows, you must run your terminal as Administrator to allow schtasks to register the background job).

### 2. Manual Backup (Dry Run)
You can manually trigger the backup sequence to verify network connections and folder generation without waiting for the scheduled time.

```bash
dbsnap backup
```

### 3. Restore Data
If you need to recover data, use the restore command to pull up the interactive timestamp menu. You can choose to inject the data alongside existing records or drop the live collections first for a clean slate.

```bash
dbsnap restore
```

### 4. Teardown
To cleanly unregister the background OS task and wipe your credentials from the OS vault (without deleting your backed-up data):

```bash
dbsnap revert
```

## Security & Architecture

### Credentials: Managed via github.com/zalando/go-keyring.

### Logging: Background runs append to a dbsnap.log file in your target backup directory, rotating every 30 days.

### Purge Safety: Uses an "Isolate & Continue" architecture. If mongodump fails to export a collection, the subsequent destructive deleteMany operation for that specific collection is automatically aborted to prevent data loss.

### Disclaimer
If you enable the "Purge after backup" feature during setup, dbSnap will execute a destructive deleteMany({}) operation on the selected live collections once the backup is verified. Use this feature with caution and ensure proper database user privileges.