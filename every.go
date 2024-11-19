package every

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	duration   time.Duration      // 当前任务的执行间隔
	taskFunc   func()             // 定时任务的逻辑
	timer      *time.Timer        // 用于控制定时任务的 Timer
	stopChan   chan struct{}      // 停止信号
	updateChan chan time.Duration // 更新间隔信号
	wg         sync.WaitGroup     // 用于等待任务退出
}

// NewTask 创建一个新的定时任务
func NewTask(interval string, task func()) (*Task, error) {
	duration, err := parseDuration(interval)
	if err != nil {
		return nil, err
	}

	return &Task{
		duration:   duration,
		taskFunc:   task,
		stopChan:   make(chan struct{}),
		updateChan: make(chan time.Duration),
	}, nil
}

// parseDuration 解析时间间隔
func parseDuration(interval string) (time.Duration, error) {
	var totalDuration time.Duration

	for _, part := range strings.Split(interval, ",") {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}

		unit := part[len(part)-1]
		value, err := strconv.Atoi(part[:len(part)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid time value: %s", part)
		}

		switch unit {
		case 's': // 秒
			totalDuration += time.Duration(value) * time.Second
		case 'm': // 分钟
			totalDuration += time.Duration(value) * time.Minute
		case 'h': // 小时
			totalDuration += time.Duration(value) * time.Hour
		case 'd': // 天
			totalDuration += time.Duration(value) * 24 * time.Hour
		default:
			return 0, fmt.Errorf("unsupported time unit: %c", unit)
		}
	}

	return totalDuration, nil
}

// Start 启动定时任务
func (t *Task) Start() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		// 初始化定时器
		t.timer = time.NewTimer(t.duration)

		for {
			select {
			case <-t.stopChan:
				// 停止任务
				t.timer.Stop()
				return
			case newDuration := <-t.updateChan:
				// 更新间隔时间并重置定时器
				t.timer.Stop()
				t.duration = newDuration
				t.timer.Reset(t.duration)
			case <-t.timer.C:
				// 执行任务
				t.taskFunc()

				// 重置定时器
				t.timer.Reset(t.duration)
			}
		}
	}()
}

// Stop 停止任务
func (t *Task) Stop() {
	close(t.stopChan)
	t.wg.Wait()
}

// UpdateInterval 更新任务间隔时间
func (t *Task) UpdateInterval(interval string) error {
	newDuration, err := parseDuration(interval)
	if err != nil {
		return err
	}

	// 发送更新信号
	t.updateChan <- newDuration
	return nil
}
