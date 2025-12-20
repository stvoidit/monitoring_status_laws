package scheduler

import (
	"context"
	"sync"
	"time"
)

const (
	oneDayPeriod = time.Hour * 24
)

// Job - задача для планировщика
type Job struct {
	f           func(ctx context.Context) // функция, которую необходимо выполнять по таймеру
	timer       *time.Timer               // таймер по которому выполняется функция
	stop        chan struct{}             // канал для сообщение об об остановке задачи
	isStarted   bool                      // флаг запущена задача или нет, чтобы избежать повторного запуска
	lock        sync.RWMutex              // на всякий случай мбютекс, т.к. много асинхронных операций
	startPeriod struct {
		Hours   int // в какой час дня выполнять задачу
		Minutes int // в какую минуту часа нужно выполнять
	} // периоды времени для выполнения задачи
}

// Stop - остановить задачу
func (j *Job) Stop() {
	if j.isStarted {
		j.lock.Lock()
		j.stop <- struct{}{}
		close(j.stop)
		j.isStarted = false
		j.lock.Unlock()
	}
}

// Run - запустить задачу
func (j *Job) Run(ctx context.Context) {
	if j.isStarted {
		return
	}
	// инициализация таймера или перезапуск, если уже инициализирован (на всякий случай)
	if j.timer == nil {
		j.timer = time.NewTimer(j.diffDuration())
	} else {
		j.resetTimer()
	}
	go func() {
		j.isStarted = true
		for {
			defer j.timer.Stop()
			select {
			case <-j.timer.C:
				j.lock.RLock()
				j.f(ctx)
				j.resetTimer()
				j.lock.RUnlock()
			case <-ctx.Done():
				j.Stop()
				return
			case <-j.stop:
				return
			}
		}
	}()
}

// diffDuration - расчет time.Duration для time.Timer, когда выполнять функцию в следующий раз.
// time.Duration всегда > 0, т.к. если расчетное время следующего выполнения функции < time.Now(), то добавляется 24 часа
func (j *Job) diffDuration() time.Duration {
	var now = time.Now()
	var next = time.Date(now.Year(), now.Month(), now.Day(), j.startPeriod.Hours, j.startPeriod.Minutes, 0, 0, time.Local)
	if next.Before(now) {
		next = next.Add(oneDayPeriod)
	}
	return next.Sub(now)
}

func (j *Job) resetTimer() { j.timer.Reset(j.diffDuration()) }

// ChangeTime - изменить время
func (j *Job) ChangeTime(hours, minutes int) {
	j.lock.Lock()
	j.startPeriod.Hours = hours
	j.startPeriod.Minutes = minutes
	j.lock.Unlock()
	j.resetTimer()
}

// NewJob - создание новой задачи
func NewJob(startHour, startMinute int, f func(ctx context.Context)) (j *Job) {
	j = &Job{
		stop: make(chan struct{}, 1),
		f:    f,
	}
	j.startPeriod.Hours = startHour
	j.startPeriod.Minutes = startMinute
	return
}

// JobsScheduler - планировщик
type JobsScheduler struct {
	ctx  context.Context
	lock sync.RWMutex
	jobs map[string]*Job
}

// GetJob - получить задачу из пула
func (js *JobsScheduler) GetJob(jobID string) *Job { return js.jobs[jobID] }

// ChangeJobTime - изменить время выполнения задачи
func (js *JobsScheduler) ChangeJobTime(jobID string, hours, minutes int) {
	if job, has := js.jobs[jobID]; has {
		job.ChangeTime(hours, minutes)
	}
}

// AddJob - добавить задачу и запустить
func (js *JobsScheduler) AddJob(jobID string, job *Job) {
	if job == nil {
		return
	}
	js.lock.Lock()
	js.jobs[jobID] = job
	js.lock.Unlock()
	js.jobs[jobID].Run(js.ctx)
}

// StopAndRemoveJob - остановить и удалить задачу
func (js *JobsScheduler) StopAndRemoveJob(jobID string) {
	js.lock.RLock()
	job, has := js.jobs[jobID]
	js.lock.RUnlock()
	if !has || job == nil {
		return
	}
	job.Stop()
	delete(js.jobs, jobID)
}

// Close - корректное завершение работы всех задач, удаление всез задач
func (js *JobsScheduler) Close() {
	for k := range js.jobs {
		go js.StopAndRemoveJob(k)
	}
}

// Stats - статистика планироващика - работы и их следующий запуск
func (js *JobsScheduler) Stats() map[string]time.Time {
	js.lock.RLock()
	var now = time.Now()
	var stat = make(map[string]time.Time, len(js.jobs))
	for k, j := range js.jobs {
		stat[k] = now.Add(j.diffDuration())
	}
	js.lock.RUnlock()
	return stat
}

// NewJobsScheduler - инициализация нового планировщика
func NewJobsScheduler(ctx context.Context) *JobsScheduler {
	return &JobsScheduler{ctx: ctx, jobs: make(map[string]*Job)}
}
