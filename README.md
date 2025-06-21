# Go Concurrent Tree Analyzer

### Overview

I was learning Go’s concurrency and wanted to try it on something practical. So I built a binary tree analyzer to see how fast and efficient tree operations like walking, comparing, and finding differences could be when done in parallel. I also added a live metrics system to track performance and see what’s going on under the hood.

### How it Works

At a high level:

* My main application kicks off various tree analysis tasks.

* These tasks, like walking a tree or comparing two, spawn off **goroutines**, which are Go's lightweight concurrent functions to get things done in parallel.

* **Channels** are the lifelines here. They're how different goroutines talk to each other, passing tree values or `DiffEvent` messages without getting in a tangle.

* For coordinating when all the concurrent work for a task is done, **`sync.WaitGroup`** ensures that no part of the program jumps ahead before its dependencies are complete.

* And because multiple parts of my concurrent code are updating shared performance counters, I'm using a **`sync.Mutex`** (a mutual exclusion lock) to make sure these updates happen safely, preventing any weird data races.

* Finally, there's a small **HTTP server** that exposes these performance metrics which provides a a live window into the application's activity.

**Pros:**

* **Efficient Concurrency:** Operations like tree traversal and comparison run in parallel, making the most of available CPU cores (actually not a limitation in go, since routines are extremely light weight so 1k routines could be spinned up by 1 thread for example bc of runtime scheduler).

* **Robust Error Handling:** Using `context` allows for graceful cancellation and timeouts, preventing runaway goroutines.

* **Observability Built-in:** The metrics system provides valuable insights into the application's runtime behavior, both instantaneously via HTTP and over time through logged data.

* **Clear Design:** The use of channels for communication and mutexes for shared state management keeps the concurrent code organized and less prone to common concurrency bugs.

**Cons:**
* **Verbosity in `Walk`:** The `time.Sleep` calls in `Walk` are there for demonstration purposes (to see context cancellation more clearly), but in a real-world scenario, they would be removed to maximize performance.

### Future Enhancements

If I were to continue developing this, here are a few ideas:
* **Configurable Tree Generation:** Allow users to specify tree size or structure more flexibly.
* **Comparative Benchmarking:** Implement different tree traversal algorithms (e.g., BFS, DFS) and benchmark their performance using the existing metrics system to compare their efficiency.
* **Interactive Web UI:** Instead of just a raw `/metrics` endpoint, build a simple web UI to display metrics in real-time within the browser using websockets or periodic AJAX requests.
