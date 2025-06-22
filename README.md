# Go Concurrent Tree Analyzer

### Overview

![Untitled design (9)](https://github.com/user-attachments/assets/1239c4ad-de0e-4f50-8882-bd375fbc30af)

While learning Go concurrency, I wanted a fun project to apply it to. So I built a tool that analyzes walking binary trees, comparing them, and finding differences all in parallel. It also exposes live performance metrics so I can see how well everything runs under the hood.

### What it does

- Spawns goroutines to walk and compare trees concurrently.

- Uses channels to pass values and diff messages between routines.

- Relies on sync.WaitGroup to coordinate task completion.

- Applies sync.Mutex to safely update shared performance counters.

- Runs a simple HTTP server to expose runtime metrics.

⚡ **Pros:**

- High Performance: Makes full use of Go's lightweight goroutines and scheduler.

- Smart Error Handling: Graceful exits with context cancellation & timeouts.

- Built in Observability: Exposes real-time metrics through an HTTP endpoint.

- Clean Concurrency: Uses channels and mutexes to avoid common race bugs.

📦 **Cons:**
* **Verbosity in `Walk`:** The `time.Sleep` calls in `Walk` are there for demonstration purposes (to see context cancellation more clearly), but in a real-world scenario, they would be removed to maximize performance.

### Future Enhancements

If I was to continue developing this:
* **Configurable Tree Generation:** Allow users to specify tree size or structure more flexibly.
* **Comparative Benchmarking:** Implement different tree traversal algorithms (e.g., BFS, DFS) and benchmark their performance using the existing metrics system to compare their efficiency.
* **Interactive Web UI:** Instead of just a raw `/metrics` endpoint, build a simple web UI to display metrics in real-time within the browser using websockets or periodic AJAX requests.
