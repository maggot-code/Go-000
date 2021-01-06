package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// SlidingCounter 秒级为单位的滑动窗口计数器
type SlidingCounter struct {
	buckets  map[int64]*bucket // 秒级为单位的桶
	interval int64             // 时间周期
	mu       *sync.RWMutex
}

type bucket struct {
	Value float64
}

// NewSlidingCounter 创建一个滑动窗口计数器
func NewSlidingCounter(interval int64) *SlidingCounter {
	c := &SlidingCounter{
		buckets:  make(map[int64]*bucket),
		interval: interval,
		mu:       &sync.RWMutex{},
	}
	return c
}

func (c *SlidingCounter) currentBucket() *bucket {
	now := time.Now().Unix()

	// 当前时间有桶存在 直接返回
	if b, ok := c.buckets[now]; ok {
		return b
	}

	// 否则创建新的桶
	b := &bucket{}
	c.buckets[now] = b
	return b
}

func (c *SlidingCounter) removeOldBuckets() {
	t := time.Now().Unix() - c.interval
	for timestamp := range c.buckets {
		if timestamp <= t {
			delete(c.buckets, timestamp)
		}
	}
}

// Incr 累加
func (c *SlidingCounter) Incr(i float64) {
	if i == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.currentBucket()
	b.Value += i
	c.removeOldBuckets()
}

// Sum 累计
func (c *SlidingCounter) Sum() float64 {
	t := time.Now().Unix() - c.interval

	sum := float64(0)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for timestamp, bucket := range c.buckets {
		if timestamp >= t {
			sum += bucket.Value
		}
	}

	return sum
}

// Max 最大值
func (c *SlidingCounter) Max() float64 {
	t := time.Now().Unix() - c.interval

	var max float64

	c.mu.RLock()
	defer c.mu.RUnlock()

	for timestamp, bucket := range c.buckets {
		if timestamp >= t {
			if bucket.Value > max {
				max = bucket.Value
			}
		}
	}

	return max
}

// Min 最小值
func (c *SlidingCounter) Min() float64 {
	t := time.Now().Unix() - c.interval

	var min float64

	c.mu.RLock()
	defer c.mu.RUnlock()

	for timestamp, bucket := range c.buckets {
		if timestamp >= t {
			if min == 0 {
				min = bucket.Value
				continue
			}
			if bucket.Value < min {
				min = bucket.Value
			}
		}
	}

	return min
}

// Avg 平均值
func (c *SlidingCounter) Avg() float64 {
	return c.Sum() / float64(c.interval)
}

func main() {
	// 窗口周期为10秒
	c := NewSlidingCounter(10)

	// 统计
	go func() {
		tick := time.Tick(1 * time.Second)
		for range tick {
			m := make(map[int64]float64)
			for t, v := range c.buckets {
				m[t] = v.Value
			}
			fmt.Println("buckets:", m)
			fmt.Println("max:", c.Max())
			fmt.Println("min:", c.Min())
			fmt.Println("sum:", c.Sum())
			fmt.Println("avg:", c.Avg())

		}
	}()

	// 每500ms累加一次数据
	for {
		n := rand.Intn(100)
		c.Incr(float64(n))
		time.Sleep(500 * time.Millisecond)
	}
}
