package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type task struct {
	name      string
	dependsOn []string
}

type scheduler struct {
	mu            sync.Mutex
	tasks         map[string]*task
	adjacencyList map[string][]string
	inDegree      map[string]int
	readyQueue    chan *task
	isStopped     bool
}

func (s *scheduler) addTask(t *task) {
	if s.isStopped {
		log.Println("Scheduler is stopped, cannot add new tasks")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.name] = t
	if _, exists := s.adjacencyList[t.name]; !exists {
		s.adjacencyList[t.name] = make([]string, 0)
	}
	s.inDegree[t.name] = 0
	for _, dep := range t.dependsOn {
		if dep == "" {
			continue
		}
		s.adjacencyList[dep] = append(s.adjacencyList[dep], t.name)
		s.inDegree[t.name]++
	}

	if s.inDegree[t.name] == 0 {
		s.readyQueue <- t
	}
}

func (s *scheduler) process(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			defer s.mu.Unlock()
			s.isStopped = true
			close(s.readyQueue)
			log.Println("Context done, stopping scheduler")
			return
		case <-ticker.C:
			if len(s.tasks) > 0 && len(s.readyQueue) == 0 {
				log.Println("Possible deadlock or cyclic dependency detected")
				return
			}
		case t := <-s.readyQueue:
			if t == nil {
				continue
			}
			// Simulate task processing
			println("Processing task:", t.name)
			time.Sleep(1 * time.Second) // Simulate work
			// After processing, update the in-degree of dependent tasks
			for _, dep := range s.adjacencyList[t.name] {
				s.mu.Lock()
				s.inDegree[dep]--
				if s.inDegree[dep] == 0 {
					s.readyQueue <- s.tasks[dep]
				}
				s.mu.Unlock()
			}
			s.mu.Lock()
			delete(s.tasks, t.name) // Remove task after processing
			delete(s.adjacencyList, t.name)
			delete(s.inDegree, t.name)
			s.mu.Unlock()
			println("Completed task:", t.name)
		}
	}
}

func newScheduler() *scheduler {
	return &scheduler{
		tasks:         make(map[string]*task),
		adjacencyList: make(map[string][]string),
		inDegree:      make(map[string]int),
		readyQueue:    make(chan *task, 100),
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	s := newScheduler()
	s.addTask(&task{"Task 1", []string{"Task 2", "Task 3"}})
	s.addTask(&task{"Task 2", nil})
	s.addTask(&task{"Task 3", nil})
	s.addTask(&task{"Task 4", nil})
	defer cancel()
	go s.process(ctx)
	<-ctx.Done()
	if len(s.tasks) > 0 {
		log.Println("Some tasks were not completed before shutdown.")
		log.Println("Remaining tasks:", func() []string {
			s.mu.Lock()
			defer s.mu.Unlock()
			var remaining []string
			for name := range s.tasks {
				remaining = append(remaining, name)
			}
			return remaining
		}())
	} else {
		log.Println("All tasks completed successfully.")
	}
}
