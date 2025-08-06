package scheduler

import (
	"testing"
	"time"

	"github.com/callmebg/cron/internal/types"
)

const (
	job2ID = "job2"
)

func TestNewJobQueue(t *testing.T) {
	queue := NewJobQueue()
	if queue == nil {
		t.Fatal("NewJobQueue returned nil")
	}

	if queue.GetLen() != 0 {
		t.Errorf("New queue length = %d; want 0", queue.GetLen())
	}

	if !queue.IsEmpty() {
		t.Error("New queue should be empty")
	}
}

func TestJobQueueAdd(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	queue.Add(job)

	if queue.GetLen() != 1 {
		t.Errorf("Queue length after add = %d; want 1", queue.GetLen())
	}

	if queue.IsEmpty() {
		t.Error("Queue should not be empty after adding job")
	}
}

func TestJobQueuePeek(t *testing.T) {
	queue := NewJobQueue()

	// Test empty queue
	peeked := queue.Peek()
	if peeked != nil {
		t.Error("Peek on empty queue should return nil")
	}

	// Add job and test peek
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	queue.Add(job)
	peeked = queue.Peek()

	if peeked == nil {
		t.Error("Peek should return the job")
		return
	}

	if peeked.ID != job.ID {
		t.Errorf("Peeked job ID = %q; want %q", peeked.ID, job.ID)
	}

	// Ensure peek doesn't remove the job
	if queue.GetLen() != 1 {
		t.Errorf("Queue length after peek = %d; want 1", queue.GetLen())
	}
}

func TestJobQueueOrdering(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	now := time.Now()

	// Create jobs with different execution times
	job1, _ := NewJob("job1", "job1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job1.NextExecution = now.Add(time.Hour * 2) // Later

	job2, _ := NewJob("job2", "job2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2.NextExecution = now.Add(time.Hour * 1) // Earlier

	job3, _ := NewJob("job3", "job3", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job3.NextExecution = now.Add(time.Hour * 3) // Latest

	// Add in random order
	queue.Add(job1)
	queue.Add(job3)
	queue.Add(job2)

	// Peek should return the earliest (job2)
	peeked := queue.Peek()
	if peeked.ID != job2ID {
		t.Errorf("Peek returned job %q; want job2", peeked.ID)
	}
}

func TestJobQueuePopReady(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	now := time.Now()

	// Create jobs - some ready, some not
	job1, _ := NewJob("ready1", "ready1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job1.NextExecution = now.Add(-time.Minute) // Ready (past)

	job2, _ := NewJob("ready2", "ready2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2.NextExecution = now.Add(-time.Second * 30) // Ready (past)

	job3, _ := NewJob("future", "future", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job3.NextExecution = now.Add(time.Minute) // Not ready (future)

	queue.Add(job1)
	queue.Add(job2)
	queue.Add(job3)

	readyJobs := queue.PopReady(now)

	// Should return 2 ready jobs
	if len(readyJobs) != 2 {
		t.Errorf("PopReady returned %d jobs; want 2", len(readyJobs))
	}

	// Queue should still have the future job
	if queue.GetLen() != 1 {
		t.Errorf("Queue length after PopReady = %d; want 1", queue.GetLen())
	}

	remaining := queue.Peek()
	if remaining.ID != "future" {
		t.Errorf("Remaining job ID = %q; want future", remaining.ID)
	}
}

func TestJobQueueRemove(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job1, _ := NewJob("job1", "job1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2, _ := NewJob("job2", "job2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)

	queue.Add(job1)
	queue.Add(job2)

	// Remove existing job
	removed := queue.Remove("job1")
	if !removed {
		t.Error("Remove should return true for existing job")
	}

	if queue.GetLen() != 1 {
		t.Errorf("Queue length after remove = %d; want 1", queue.GetLen())
	}

	// Try to remove non-existing job
	removed = queue.Remove("non-existing")
	if removed {
		t.Error("Remove should return false for non-existing job")
	}

	if queue.GetLen() != 1 {
		t.Errorf("Queue length should remain 1 after failed remove")
	}
}

func TestJobQueueGetByID(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job, _ := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	queue.Add(job)

	// Test existing job
	found := queue.GetByID("test-id")
	if found == nil {
		t.Error("GetByID should find existing job")
		return
	}

	if found.ID != "test-id" {
		t.Errorf("Found job ID = %q; want test-id", found.ID)
	}

	// Test non-existing job
	notFound := queue.GetByID("non-existing")
	if notFound != nil {
		t.Error("GetByID should return nil for non-existing job")
	}
}

func TestJobQueueGetByName(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job, _ := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	queue.Add(job)

	// Test existing job
	found := queue.GetByName("test-job")
	if found == nil {
		t.Error("GetByName should find existing job")
		return
	}

	if found.Name != "test-job" {
		t.Errorf("Found job Name = %q; want test-job", found.Name)
	}

	// Test non-existing job
	notFound := queue.GetByName("non-existing")
	if notFound != nil {
		t.Error("GetByName should return nil for non-existing job")
	}
}

func TestJobQueueGetAll(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job1, _ := NewJob("job1", "job1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2, _ := NewJob("job2", "job2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)

	queue.Add(job1)
	queue.Add(job2)

	all := queue.GetAll()

	if len(all) != 2 {
		t.Errorf("GetAll returned %d jobs; want 2", len(all))
	}

	// Check that it's a copy (modifying the slice shouldn't affect the queue)
	all[0] = nil
	if queue.GetLen() != 2 {
		t.Error("Modifying GetAll result should not affect the queue")
	}
}

func TestJobQueueUpdate(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	now := time.Now()

	job1, _ := NewJob("job1", "job1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job1.NextExecution = now.Add(time.Hour * 2)

	job2, _ := NewJob("job2", "job2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2.NextExecution = now.Add(time.Hour * 3)

	queue.Add(job1)
	queue.Add(job2)

	// job1 should be at the front (earlier execution time)
	first := queue.Peek()
	if first.ID != "job1" {
		t.Errorf("First job ID = %q; want job1", first.ID)
	}

	// Update job1 to have a later execution time
	job1.NextExecution = now.Add(time.Hour * 4)
	queue.Update(job1)

	// Now job2 should be at the front
	first = queue.Peek()
	if first.ID != "job2" {
		t.Errorf("First job after update ID = %q; want job2", first.ID)
	}
}

func TestJobQueueClear(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	job1, _ := NewJob("job1", "job1", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	job2, _ := NewJob("job2", "job2", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)

	queue.Add(job1)
	queue.Add(job2)

	if queue.GetLen() != 2 {
		t.Errorf("Queue length before clear = %d; want 2", queue.GetLen())
	}

	queue.Clear()

	if queue.GetLen() != 0 {
		t.Errorf("Queue length after clear = %d; want 0", queue.GetLen())
	}

	if !queue.IsEmpty() {
		t.Error("Queue should be empty after clear")
	}
}

func TestJobQueueConcurrency(t *testing.T) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	// Test concurrent adds
	jobs := make([]*Job, 100)
	for i := 0; i < 100; i++ {
		job, _ := NewJob(string(rune('a'+i)), string(rune('a'+i)), "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
		jobs[i] = job
	}

	// Add jobs concurrently
	done := make(chan bool, 100)
	for _, job := range jobs {
		go func(j *Job) {
			queue.Add(j)
			done <- true
		}(job)
	}

	// Wait for all adds to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	if queue.GetLen() != 100 {
		t.Errorf("Queue length after concurrent adds = %d; want 100", queue.GetLen())
	}
}

func BenchmarkJobQueueAdd(b *testing.B) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	jobs := make([]*Job, b.N)
	for i := 0; i < b.N; i++ {
		job, _ := NewJob(string(rune(i)), string(rune(i)), "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
		jobs[i] = job
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Add(jobs[i])
	}
}

func BenchmarkJobQueuePeek(b *testing.B) {
	queue := NewJobQueue()

	jobFunc := func() {
		// Test job function
	}

	// Add some jobs
	for i := 0; i < 1000; i++ {
		job, _ := NewJob(string(rune(i)), string(rune(i)), "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
		queue.Add(job)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Peek()
	}
}
