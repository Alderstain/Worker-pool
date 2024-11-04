package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	id   int
	pool *WorkerPool
	stop chan struct{}
}

type WorkerPool struct {
	tasks      chan string
	workers    []*Worker
	minWorkers int
	wg         sync.WaitGroup
	mut        sync.Mutex
}

func NewWorkerPool(taskChannel chan string) *WorkerPool {
	return &WorkerPool{
		tasks:      taskChannel,
		workers:    make([]*Worker, 0),
		minWorkers: 2,
	}
}

func (wp *WorkerPool) addWorker() {
	wp.mut.Lock()
	defer wp.mut.Unlock()
	worker := &Worker{
		id:   len(wp.workers) + 1,
		pool: wp,
		stop: make(chan struct{}),
	}
	wp.workers = append(wp.workers, worker)
	go worker.Start()
	fmt.Printf("Добавлен воркер %d\n", worker.id)

}

func (w *Worker) Stop() {
	close(w.stop)
}

func (wp *WorkerPool) deleteWorker() {
	wp.mut.Lock()
	defer wp.mut.Unlock()
	if len(wp.workers) <= 0 {
		return
	}
	lastWorker := wp.workers[len(wp.workers)-1]
	lastWorker.Stop()
	wp.workers = wp.workers[:len(wp.workers)-1]
	fmt.Printf("Удален воркер %d\n", lastWorker.id)
}

func (w *Worker) Start() {
	w.pool.wg.Add(1)
	defer w.pool.wg.Done()
	for {
		select {
		case task, ok := <-w.pool.tasks:
			if !ok {
				return
			}
			fmt.Printf("Воркер %d: принял %s\n", w.id, task)
		case <-w.stop:
			fmt.Printf("Воркер %d: остановлен\n", w.id)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (wp *WorkerPool) monitor() { //раз в 2 секунды проверяем количество воркеров и количество задач
	for {
		wp.mut.Lock()
		numWorkers := len(wp.workers)
		wp.mut.Unlock()
		switch {
		case len(wp.tasks) > 0:
			if numWorkers < wp.minWorkers {
				wp.addWorker()
			}
		case len(wp.tasks) <= 0:
			if numWorkers > 0 {
				wp.deleteWorker()
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func (wp *WorkerPool) AddTask(task string) {
	wp.tasks <- task
	fmt.Println("Задача добавлена:", task)
}

func (wp *WorkerPool) Close() {
	close(wp.tasks)
	wp.wg.Wait()
	fmt.Println("Все воркеры завершили работу")
}

func main() {
	taskChannel := make(chan string, 10)
	pool := NewWorkerPool(taskChannel)
	go pool.monitor()

	for i := 1; i <= 10; i++ {
		task := fmt.Sprintf("Задача %d", i)
		pool.AddTask(task)
	}
	time.Sleep(15 * time.Second)
	pool.Close()
}

//Создаём пул
//Кидаем задачи в канал
//Проверяем канал на наличие задач, если задачи есть - создаём воркеров для их решения
//Воркеры выводят задачи
//если задачи кончаются, закрываем воркеров
