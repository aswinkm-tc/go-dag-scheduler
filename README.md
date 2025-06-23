# go-dag-scheduler

`go-dag-scheduler` is a lightweight Directed Acyclic Graph (DAG)-based task scheduler implemented in Go. It allows you to define tasks with dependencies and ensures that tasks are executed in the correct order while detecting potential deadlocks or cyclic dependencies.

## Design Overview

The scheduler is designed to process tasks in a topological order based on their dependencies. Below is a breakdown of the key components:

### Components

1. **Task (`task` struct)**:
   - Represents a unit of work.
   - Contains:
     - `name`: A unique identifier for the task.
     - `dependsOn`: A list of task names that this task depends on.

2. **Scheduler (`scheduler` struct)**:
   - Manages the execution of tasks.
   - Contains:
     - `tasks`: A map of task names to their corresponding `task` structs.
     - `adjacencyList`: A map representing the dependency graph (edges between tasks).
     - `inDegree`: A map tracking the number of dependencies (in-degree) for each task.
     - `readyQueue`: A channel used to queue tasks that are ready for execution.
     - `isStopped`: A flag indicating whether the scheduler has been stopped.

3. **Dependency Graph**:
   - The scheduler uses an adjacency list to represent the task dependencies.
   - Tasks with zero in-degree (no dependencies) are added to the `readyQueue` for immediate processing.

4. **Concurrency**:
   - The scheduler uses a `sync.Mutex` to ensure thread-safe access to shared data structures.
   - Tasks are processed concurrently using a `readyQueue`.

5. **Deadlock Detection**:
   - A periodic check ensures that tasks are being processed.
   - If the `readyQueue` is empty but there are still tasks in the scheduler, it indicates a potential deadlock or cyclic dependency.

### Workflow

1. **Adding Tasks**:
   - Tasks are added using the `addTask` method.
   - Dependencies are updated in the adjacency list and in-degree map.
   - Tasks with zero in-degree are immediately added to the `readyQueue`.

2. **Processing Tasks**:
   - The `process` method runs in a separate goroutine.
   - Tasks are dequeued from the `readyQueue` and processed.
   - After processing a task, its dependent tasks' in-degrees are decremented. If a dependent task's in-degree reaches zero, it is added to the `readyQueue`.

3. **Graceful Shutdown**:
   - The scheduler listens for termination signals (e.g., `SIGTERM`).
   - Upon receiving a signal, the scheduler stops processing and logs any remaining tasks.

### Example

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    s := newScheduler()
    s.addTask(&task{"Task 1", []string{"Task 2", "Task 3"}})
    s.addTask(&task{"Task 2", nil})
    s.addTask(&task{"Task 3", nil})
    s.addTask(&task{"Task 4", nil})

    go s.process(ctx)
    <-ctx.Done()
}
```
