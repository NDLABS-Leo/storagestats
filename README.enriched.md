# StorageStats

**StorageStats** is a modular Filecoin data auditing and analytics suite. It features multiple pluggable integration pipelines for tasks such as Fil+ sampling, RepDAO analysis, and storage provider statistics collection.

---

## 📦 Modules Overview

| Module         | Description |
|----------------|-------------|
| `filplus`      | Performs Fil+ verified deal sampling and task generation. |
| `spcoverage`   | Analyzes storage provider coverage statistics. |
| `repdao`       | Processes RepDAO retrieval auditing. |
| `repdao_dp`    | DP-based RepDAO module for differential proof. |
| `spadev0`      | Tests specific storage provider datasets. |
| `statemarketdeals` | Processes state market deals from chain data. |
| `oneoff`       | One-time deal checks or manual sampling scripts. |

---

## 🧱 Architecture Diagram

```mermaid
flowchart TD
    A[MongoDB: state_market_deals] -->|reads| B(filplus / spcoverage / repdao)
    B -->|generates| C[Task Queue (MongoDB)]
    C --> D[RetrievalBot / Executor]
    D --> E[Results DB (MongoDB)]
    F[Location/IP Info] --> B
    G[Lotus API / Chain] --> B
```

---

## 🚀 Deployment

### Requirements

- Go 1.20+
- Docker (for container-based deployment)
- MongoDB
- IPInfo token
- Optional: Lotus node (read-only) access

### Local Build

```bash
make build
go run integration/filplus/main.go
```

### Docker Build

```bash
docker build -t storagestats -f Dockerfile .
```

For AWS-optimized build:

```bash
docker build -t storagestats:aws -f aws.Dockerfile .
```

---

## ⚙️ Configuration

Set environment variables via `.env` file or system:

| Variable | Description |
|----------|-------------|
| `QueueMongoURI` | MongoDB URI for task queue |
| `QueueMongoDatabase` | DB name for task queue |
| `StatemarketdealsMongoURI` | MongoDB for state market deals |
| `StatemarketdealsMongoDatabase` | DB name for state market deals |
| `ResultMongoURI` | MongoDB URI for result output |
| `ResultMongoDatabase` | DB for result output |
| `IPInfoToken` | Token for IP geolocation |
| `LotusAPIUrl` | (optional) Lotus node API |
| `LotusAPIToken` | (optional) Token for Lotus API |

---

## 🧪 Example Run (FilPlus)

```bash
go run integration/filplus/main.go
```

This module:
- Samples verified deals (top 30% volume per provider)
- Randomly selects up to 100
- Queues retrieval tasks
- Logs summary to results DB

---

## 📁 Project Tree (Key Parts)

```
.
├── integration/
│   ├── filplus/
│   ├── spcoverage/
│   ├── repdao/
│   ├── repdao_dp/
│   ├── spadev0/
│   ├── oneoff/
│   └── statemarketdeals/
├── Dockerfile
├── aws.Dockerfile
├── Makefile
├── go.mod
├── .env.example (recommended)
```

---

## 📜 License

See [LICENSE](./LICENSE).


---

## ⚙️ Workers

Workers are the components that consume tasks from the queue and perform data retrieval. There are currently four types:

### 🟦 Bitswap Worker
- Looks up the storage provider's libp2p protocols
- If using Boost market, queries supported retrieval protocols
- Uses Bitswap to fetch a **single block** from the provider

### 🟩 Graphsync Worker
- Performs Graphsync retrieval with a selector that only matches the **root block**

### 🟧 HTTP Worker
- Retrieves the **first few MiB** of the piece data
- Also relies on libp2p protocol lookup and boost market protocol info

### 🧪 Stub Worker
- Doesn't perform actual retrieval
- Pushes mock results to the result database
- Useful for testing queue and DB connectivity

---

## 🔗 Integrations

Integrations push work items to the retrieval task queue or manage long-running data flows.

### 📄 StateMarketDeals Integration
- Periodically pulls `statemarketdeals.json` from GLIP API
- Saves deal info to the database

### 📦 FILPLUS Integration
- Pulls **random active deals** from the `state_market_deals` DB
- Pushes `Bitswap`, `Graphsync`, and `HTTP` tasks to the task queue

---

## 🚀 Getting Started

1. 🛠 Setup a MongoDB instance
2. 🔐 Create an IPInfo account and get a token
3. 🔧 Run `make build`
4. ✅ Launch the three programs (native or Docker):
   - `statemarketdeals`: fetches deals from GLIP API  
     → uses env from `.env.statemarketdeals`
   - `filplus_integration`: generates retrieval tasks  
     → uses `.env.filplus`
   - `retrieval_worker`: processes retrieval tasks  
     → uses `.env.retrievalworker`

5. 📦 Copy correct env files to `.env` in working directory for each
6. 🧩 Ensure the following binaries are available:
   - `bitswap_worker`, `graphsync_worker`, `http_worker` (needed by `retrieval_worker`)

