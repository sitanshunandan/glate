# Glate: Concurrent Pharmacokinetic Engine
**Glate** (Short for "Regulate" or "Calculated") is a distributed systems simulation designed to model the pharmacokinetics (half-life decay) of nootropics and supplements in real-time.

While most pill-tracking apps are simple CRUD loggers, Glate is a stateful backend engine that simulates the actual plasma concentration of active compounds in the user's bloodstream. It calculates first-order kinetics decay curves dynamically and uses a graph-based dependency engine to predict "clearance windows" for conflicting substances (e.g., preventing Iron intake while Caffeine is still biologically active).

**Link to project:** [TBU]


## How It's Made:

**Tech used:** Go (Golang), RESTful API, Concurrency Patterns (Ticker/Worker), Sync Primitives (RWMutex), Clean Architecture.

The system is architected as a stateful biological simulation:

1.  **The Pharmacopeia (Repository Layer):** A data layer acting as the "Source of Truth." It loads immutable scientific definitions (Half-Life, Tmax​, Bioavailability) and traverses an Interaction Graph to identify potential conflicts (Inhibition vs. Potentiation).
2.  **The Metabolizer (Calculation Engine):** A pure mathematical engine that implements First-Order Kinetics (Ct​=C0​⋅e−kt). It is stateless and decoupled, allowing for varying metabolic models to be injected without breaking the core logic.
3.  **The Session Manager (State Layer):** An in-memory, thread-safe storage engine protected by sync.RWMutex. It manages the chaotic state of multiple concurrent users, ensuring that a "Write" operation (ingesting a pill) never blocks a "Read" operation (checking status).
4. **The Sentinel (Background Monitor)**: A concurrent Goroutine that runs asynchronously alongside the HTTP server. It wakes up on a time.Ticker interval to scan all active bloodstreams, logging alerts when specific compounds drop below their effective threshold (e.g., "Sleep Window Open").
5. **Real-Time Status**: Glate exposes a REST API to query the exact milligram-level concentration of a stack at any given millisecond.

    Start the Server: go run cmd/server/main.go

    Ingest Caffeine: POST /ingest

    Check Blood Levels: GET /status

You will see the current_mg value decay over time as the background engine simulates liver clearance.

## Optimizations

The primary challenge was managing mutable state (user blood levels) in a highly concurrent environment without race conditions.

1.  **Thread-Safe State Mutation:** In a high-throughput scenario, multiple services might try to update a user's stack while the background monitor is reading it. I implemented a sync.RWMutex (Readers-Writer Lock). This allows the background monitor to perform "Cheap Reads" concurrently without blocking, locking the memory only during the brief nanoseconds of a "Write" (Ingestion).
2.  **The "Stateless" Math Core:** Biological simulation can get messy. To keep the code testable, I isolated the math (Ct​=C0​…) into a pure struct with no dependencies. This allows the system to be unit-tested against theoretical values (e.g., "Does 100mg become 50mg after 1 half-life?") without needing a database mock.
3.  **Dependency Injection:** The API Handler does not know how data is stored or calculated; it only knows the Interfaces. This means the InMemoryRepo can be swapped for a PostgresRepo, or the LinearCalculator for a MultiCompartmentCalculator, with zero code changes in the business logic layer.

## Lessons Learned:

Building Glate forced me to treat biology as a state-management problem.

1.  **State Drift":** I learned that unlike a bank ledger, biological data changes even when you don't touch it. A value of "100mg" in the database is wrong the second after it is written. This required shifting from "Stored Values" to "Stored Events" (Event Sourcing)—storing the time of ingestion and calculating the value on-read.
2.  **The Cost of Locking:** Implementing the background monitor taught me the danger of holding locks too long. Initially, the monitor locked the entire store to scan users, which froze the API. I optimized this by creating a "Snapshot" method that copies the data under a read-lock and releases it immediately, allowing the heavy analysis to happen offline.
3.  **Graph Theory in Biology:** Modeling interactions (Vitamin C potentiates Iron; Calcium inhibits Iron) required a Directed Graph. It highlighted that "Safety" isn't binary—it's a time-series problem. A "Dangerous" combination becomes safe if the first substance has cleared sufficiently.

## Usage Examples:

### Docker Deployment (Recommended)

**1. Start the Engine**

```bash
go run cmd/server/main.go
# Output: 🚀 Glate Stateful Server listening on :8080...
# Output: 👁️  Background Monitor online. Scanning every 10s...
```

**2. Ingest a Supplement (Write)**: Simulate taking 200mg of Caffeine.

```bash
# PowerShell / Curl
$body = @{ user_id="dev-1"; substance_id="caffeine"; amount_mg=200 } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/ingest" -Method Post -Body $body -ContentType "application/json"
```

**3. Check Decay (Read)**
Query the engine to see the First-Order Kinetics in action.

```bash
Invoke-RestMethod -Uri "http://localhost:8080/status?user_id=dev-1" -Method Get
# Output: { "substance": "Caffeine", "current_mg": 199.8, "time_elapsed": "1m" }
```

**4. The "Sleep Window" Test**
Wait for the background monitor to detect clearance.

```bash
--- 🏥 System Heartbeat ---
User [dev-1]:
   • Caffeine             | Original: 200mg | Current: 48.2mg (T+240m)
     💤 SLEEP WINDOW OPEN: Caffeine is low enough.
```

## Roadmap

[x] **Stateful Session Store:** In-memory storage with UUID tracking for individual doses. (Completed)

[ ] **Interaction "Lookahead":** Upgrade the /analyze endpoint to check not just current conflicts, but future conflicts (e.g., "If I take Iron now, can I drink Coffee in 1 hour?").

[ ] **Persistent Storage:** Replace the in-memory map with PostgreSQL or SQLite to ensure user history survives server restarts.


