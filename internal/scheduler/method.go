package scheduler

import (
	"fmt"
	"time"
)

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: []*info{},
	}
}

func (s *Scheduler) AddJob(name string, interval time.Duration, task func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &info{
		name: &data{
			Interval: interval,
			Task:     task,
			stop:     make(chan struct{}),
		},
	}
	s.jobs = append(s.jobs, job)
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	jobsCopy := make([]*info, len(s.jobs))
	copy(jobsCopy, s.jobs)
	s.mu.Unlock()

	for _, jobMap := range jobsCopy {
		for name, job := range *jobMap {
			s.wg.Add(1)
			go run(s, name, job)
		}
	}
}

func run(s *Scheduler, name string, job *data) {
	defer s.wg.Done()
	for {
		start := time.Now()

		go job.Task()

		select {
		case <-job.stop:
			fmt.Println("Stopped:", name)
			return
		case <-time.After(job.Interval - time.Since(start)):
		}
	}
}

func (s *Scheduler) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, jobMap := range s.jobs {
		for name, job := range *jobMap {
			select {
			case <-job.stop:
			default:
				close(job.stop)
				fmt.Println("Stopped job:", name)
			}
		}
	}

	// Wait for all running goroutines to finish
	s.wg.Wait()
}

func (s *Scheduler) StopByName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, jobMap := range s.jobs {
		if job, exists := (*jobMap)[name]; exists {
			// Close stop channel safely
			select {
			case <-job.stop:
				// already closed
			default:
				close(job.stop)
				fmt.Println("Stopped job:", name)
			}

			// Remove job from the map
			delete(*jobMap, name)

			// If map becomes empty, remove it from the slice
			if len(*jobMap) == 0 {
				s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			}
			break
		}
	}
}
