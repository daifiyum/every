package every

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Task 定义一个定时任务结构体
type Task struct {
	duration time.Duration  // 任务执行的间隔
	taskFunc func()         // 定时执行的任务
	timer    *time.Timer    // 定时器 timer，用于周期性执行任务
	stopChan chan struct{}  // 停止任务的信号通道
	wg       sync.WaitGroup // 用于等待任务完成
}

// NewTask 创建一个新的定时任务
func NewTask(interval string, task func()) (*Task, error) {
	duration, err := parseDuration(interval)
	if err != nil {
		return nil, err
	}

	return &Task{
		duration: duration,
		taskFunc: task,
		stopChan: make(chan struct{}),
	}, nil
}

// parseDuration 解析时间格式
func parseDuration(interval string) (time.Duration, error) {
	var totalDuration time.Duration

	// 解析分钟、小时、天和秒
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

		t.timer = time.NewTimer(t.duration) // 创建 Timer

		for {
			select {
			case <-t.stopChan:
				// 停止任务
				t.timer.Stop()
				return
			case <-t.timer.C:
				// 执行任务
				t.taskFunc()

				// 激活定时器，并使用新的间隔
				t.timer.Reset(t.duration)
			}
		}
	}()
}

// Stop 停止定时任务
func (t *Task) Stop() {
	close(t.stopChan)
	t.wg.Wait()
}

// UpdateInterval 动态更新定时任务的间隔
func (t *Task) UpdateInterval(interval string) error {
	// 解析新的间隔时间
	duration, err := parseDuration(interval)
	if err != nil {
		return err
	}

	// 更新任务的间隔
	t.duration = duration
	return nil
}
