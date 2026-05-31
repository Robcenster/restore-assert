# Restore-Assert

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Restore-Assert is a CLI tool for automated backup integrity verification (Burn-testing). It doesn't just check for file existence; it restores the backup into an isolated Docker container and runs a suite of tests (Assertions), ensuring your data is truly recoverable.

## 🛠 Features

* **Isolated Environment:** Automated launch of temporary containers via [Testcontainers](https://testcontainers.com/).
* **Smart detector:** Automatic detection of dump formats (Custom, Directory, Tar, Plain SQL).
* **Deep Verification (Assertions):**
    * Presence of tables, extensions, and schemas.
    * Metrics: Row count, Table size.
    * Freshness: Data relevance check (ensures the backup is not too old).
    * Null Ratio: Data quality control (checks for anomalous amounts of empty fields).
* **Inspection:** View dump structure without running tests.

## 📥 Installation

```bash
go install github.com/Robcenster/restore-assert@latest
```

## 🚀 Quick Start
<img src="docs/demo.gif" alt="Restore-assert cli usage demonstration" width="100%">


#### 1. Initialization
Create a configuration file template in the current folder:
```bash
restore-assert init --name "configname.yaml" --path .
```
#### 2. Run Burn-test
Launch the full cycle: container creation -> restoration -> assertion execution:
```bash
restore-assert check ./prod_backup.sql --config restore-config.yaml
```

## Project Structure
```
├── cmd/                # Entry point (main.go)
├── internal/
│   ├── app/            # Pipeline logic (RunCheck)
│   ├── cli/            # Commands and interface (Cobra)
│   ├── config/         # YAML parsing and validation
│   ├── container/      # Docker orchestration and dump type detection
│   ├── repository/     # SQL queries for the restored DB
│   ├── verifier/       # Assertion execution engine
│   └── formatter/      # Report and DB tree visualization
└── config/
    └── template/       # Configuration file templates
```

## FAQ (Frequently Asked Questions)

<details>
<summary><b>Is it safe to run this against my production database?</b></summary>

<b>No.</b> This tool is designed to test <b>backup files</b> (dumps), not live databases. It restores the dump into an isolated Docker container. Never point it at your production connection strings if you are using automated cleanup features.
</details>

<details>
<summary><b>Why use Docker instead of just local psql/pg_restore?</b></summary>

Docker ensures a clean, isolated environment. It prevents "it works on my machine" issues caused by different local PostgreSQL versions, installed extensions, or conflicting environment variables. Once the test is done, the container is destroyed, leaving your system clean.
</details>

<details>
<summary><b>Does it support MySQL, MariaDB or SQL Server?</b></summary>

Currently, the primary focus is <b>PostgreSQL</b>.
</details>

<details>
<summary><b>How long does a typical "Burn-test" take?</b></summary>

It depends on your backup size and hardware. For a 1GB dump, it usually takes 1-3 minutes (including container startup, restore, and assertions). Using `fsync: "off"` in the config significantly accelerates this process.
</details>

<details>
<summary><b>Can I run this in CI/CD (GitHub Actions, GitLab CI)?</b></summary>

Yes! Since it uses Docker, you just need a runner with Docker-in-Docker (DinD) support. It's a perfect way to verify your backups daily.
</details>

<details>
<summary><b>Why does the program exit with code 1 (os.Exit(1)) if anything fails?</b></summary>

This is an intentional characteristic of the program's logic. If absolutely anything fails during the restoration process or the subsequent logical checks (including the assertions themselves), the entire backup verification is considered unsuccessful, and the program will terminate with `os.Exit(1)`. Knowing this beforehand prevents initial confusion: any single failure means the recovery test has failed.
</details>

##  Troubleshooting

Common issues and how to fix them:

### Docker & Connection Issues

<details>
<summary><code>port already allocated</code></summary>

* **Reason:** Another database or service is using the port `restore-assert` is trying to bind to.
* **Fix:** Change the port in your config or stop the conflicting container: `docker stop $(docker ps -q)`.
</details>

<details>
<summary><code>context deadline exceeded</code> (during restore)</summary>

* **Reason:** The backup is too large, and Docker/Postgres couldn't finish the job in time.
* **Fix:** Increase resources in the `docker` section of your config (`cpu_limit`, `memory_limit`) or check your disk I/O.
</details>

### 🐘 Postgres Specifics
<details>
<summary><code>role "xyz" does not exist</code></summary>

* **Reason:** The dump contains objects owned by a user that wasn't created.
* **Fix:** Ensure the role is listed in the `database.roles` section of your `restore-config.yaml` OR enable `no_owner: true` in the restore settings OR make other dump with CREATE ROLE.
</details>

<details>
<summary><code>role "abc" already exists</code></summary>

* **Reason:** You are trying to create multiple existing roles. `restore-assert` can create roles from different sources: `restore-config.yaml` in `database.roles` OR `database.user` OR dump that creates own roles.
* **Fix:** Use `no-owner` OR/AND change `database.roles` OR/AND `database.user` in different cases.
</details>

<details>
<summary><code>extension "xyz" does not exist</code></summary>

* **Reason:** The environment (Postgres image) doesn't have the required extension binaries, or it wasn't enabled.
* **Fix:** Ensure the extension is listed in the `database.extensions` section of your `restore-config.yaml` OR its creating in dump file OR ensure you are using a Docker image that includes these extensions (e.g., PostGIS).
</details>

<details>
<summary><code>database "xyz" already exists</code></summary>

* **Reason:** Conflict between `database.db_name` in config and the database name inside the dump/restore commands.
* **Fix:** Ensure `database.db_name` is unique OR change dump settings for not creating database.
</details>

<details>
<summary><code>assertions execute incorrectly during multi-database or cluster restoration</code></summary>

* **Reason:** Assertions connect to only a single database context. If your backup contains multiple databases (e.g., when restoring from a cluster dump), there is a high probability that the asserts will be executed incorrectly.
* **Fix:** It is highly recommended to use a single database for asserts. If an empty database is created at the beginning of the restoration process to serve as an entry point, it should not cause any issues.
</details>

### Tool Errors

<details>
<summary><code>failed to detect dump format</code></summary>

* **Reason:** The file is corrupted or in a format the tool doesn't recognize yet.
* **Fix:** Run `file your_backup.sql` to check the actual type. Ensure it's a valid `pg_dump` output (Plain, Custom, or Tar). If so, you can create issue to discuss.
</details>

<details>
<summary><code>could not be restored due to an extension error</code></summary>

* **Reason:** Restoring multiple databases at the same time may cause errors when installing PostgreSQL extensions.
* **Fix:** Enable the `modify_template` variable to replace `template0` during the process. However, for optimal stability and test accuracy, it is recommended to restore databases individually.
</details>

##
Developed to ensure your backups actually work when the fire starts.
