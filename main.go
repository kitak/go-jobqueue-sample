package main

import "os"
import "os/signal"
import "fmt"
import "time"

var MaxWorker int = 10
var MaxQueue int = 30

type Job struct{}

func (job *Job) Something() {
	time.Sleep(3 * time.Second)
}

var JobQueue chan Job

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	JobDone    chan struct{}
	quit       chan bool
}

func NewWorker(workerPool chan chan Job, jobDone chan struct{}) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobDone:    jobDone,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				job.Something()
				w.JobDone <- struct{}{}
			case <-w.quit:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	WorkerPool chan chan Job
	JobDone    chan struct{}
}

func NewDispatcher(maxWorkers int, jobDone chan struct{}) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, JobDone: jobDone}
}

func (d *Dispatcher) Run() {
	for i := 0; i < cap(d.WorkerPool); i++ {
		worker := NewWorker(d.WorkerPool, d.JobDone)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	fmt.Println("Start dispatching")
	for {
		job := <-JobQueue
		go func(job Job) {
			jobChannel := <-d.WorkerPool
			jobChannel <- job
		}(job)
	}
}

func main() {
	var work Job
	done := make(chan struct{})
	quit := make(chan struct{})
	jobDone := make(chan struct{}, MaxWorker)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	JobQueue = make(chan Job, MaxQueue)
	dispatcher := NewDispatcher(MaxWorker, jobDone)
	dispatcher.Run()

	go func(quit chan struct{}, done chan struct{}) {
		go func() {
			for i := 0; i < 3; i++ {
				<-jobDone
				fmt.Println("Done job")
			}
			quit <- struct{}{}
		}()

		for i := 0; i < 3; i++ {
			work = Job{}
			JobQueue <- work
		}

		for {
			select {
			case _ = <-quit:
				done <- struct{}{}
			}
		}
	}(quit, done)

	go func() {
		for {
			s := <-sig
			fmt.Println(s)
			done <- struct{}{}
		}
	}()

	<-done
	fmt.Println("Quit")
}
