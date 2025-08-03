package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

// JobQueue implements a priority queue for jobs based on their next execution time
type JobQueue struct {
	jobs []*Job
	mu   sync.RWMutex
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	jq := &JobQueue{
		jobs: make([]*Job, 0),
	}
	heap.Init(jq)
	return jq
}

// Len returns the number of jobs in the queue (heap interface - no lock needed)
func (jq *JobQueue) Len() int {
	return len(jq.jobs)
}

// Less compares two jobs based on their next execution time (heap interface - no lock needed)
func (jq *JobQueue) Less(i, j int) bool {
	return jq.jobs[i].NextExecution.Before(jq.jobs[j].NextExecution)
}

// Swap swaps two jobs in the queue (heap interface - no lock needed)
func (jq *JobQueue) Swap(i, j int) {
	jq.jobs[i], jq.jobs[j] = jq.jobs[j], jq.jobs[i]
}

// Push adds a job to the queue (heap interface - no lock needed)
func (jq *JobQueue) Push(x interface{}) {
	jq.jobs = append(jq.jobs, x.(*Job))
}

// Pop removes and returns the job with the earliest execution time (heap interface - no lock needed)
func (jq *JobQueue) Pop() interface{} {
	old := jq.jobs
	n := len(old)
	job := old[n-1]
	old[n-1] = nil // avoid memory leak
	jq.jobs = old[0 : n-1]
	return job
}

// GetLen returns the number of jobs in the queue (thread-safe public method)
func (jq *JobQueue) GetLen() int {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	return len(jq.jobs)
}

// Add adds a job to the queue
func (jq *JobQueue) Add(job *Job) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	heap.Push(jq, job)
}

// Remove removes a job from the queue by ID
func (jq *JobQueue) Remove(jobID string) bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	for i, job := range jq.jobs {
		if job.ID == jobID {
			heap.Remove(jq, i)
			return true
		}
	}
	return false
}

// Peek returns the job with the earliest execution time without removing it
func (jq *JobQueue) Peek() *Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	if len(jq.jobs) == 0 {
		return nil
	}
	return jq.jobs[0]
}

// PopReady removes and returns all jobs that are ready to run at the given time
func (jq *JobQueue) PopReady(now time.Time) []*Job {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	var readyJobs []*Job

	for len(jq.jobs) > 0 && !jq.jobs[0].NextExecution.After(now) {
		job := heap.Pop(jq).(*Job)
		readyJobs = append(readyJobs, job)
	}

	return readyJobs
}

// Update updates a job's position in the queue after its next execution time changes
func (jq *JobQueue) Update(job *Job) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	// Remove and re-add the job to update its position
	for i, j := range jq.jobs {
		if j.ID == job.ID {
			heap.Remove(jq, i)
			break
		}
	}
	heap.Push(jq, job)
}

// GetAll returns a copy of all jobs in the queue
func (jq *JobQueue) GetAll() []*Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	jobs := make([]*Job, len(jq.jobs))
	copy(jobs, jq.jobs)
	return jobs
}

// GetByID returns a job by its ID
func (jq *JobQueue) GetByID(jobID string) *Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	for _, job := range jq.jobs {
		if job.ID == jobID {
			return job
		}
	}
	return nil
}

// GetByName returns a job by its name
func (jq *JobQueue) GetByName(name string) *Job {
	jq.mu.RLock()
	defer jq.mu.RUnlock()

	for _, job := range jq.jobs {
		if job.Name == name {
			return job
		}
	}
	return nil
}

// Clear removes all jobs from the queue
func (jq *JobQueue) Clear() {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.jobs = jq.jobs[:0]
}

// Size returns the current size of the queue
func (jq *JobQueue) Size() int {
	return jq.GetLen()
}

// IsEmpty returns true if the queue is empty
func (jq *JobQueue) IsEmpty() bool {
	return jq.GetLen() == 0
}
