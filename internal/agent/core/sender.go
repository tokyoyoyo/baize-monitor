package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Sender 消息发送器
type Sender struct {
	serverURL    string
	client       *http.Client
	messageQueue chan *Message
	rateLimiter  *rate.Limiter
	maxRetries   int
	retryDelay   time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Message 消息结构
type Message struct {
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	Timestamp  time.Time   `json:"timestamp"`
	RetryCount int         `json:"retry_count"`
}

// NewSender 创建消息发送器
func NewSender(serverURL string) *Sender {
	return &Sender{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		messageQueue: make(chan *Message, 1000),
		rateLimiter:  rate.NewLimiter(rate.Every(100*time.Millisecond), 10), // 限制发送频率
		maxRetries:   3,
		retryDelay:   5 * time.Second,
	}
}

// Start 启动发送器
func (s *Sender) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.wg.Add(1)
	go s.messageProcessor()

	fmt.Println("Message sender started")
}

// Stop 停止发送器
func (s *Sender) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	// 关闭消息队列
	close(s.messageQueue)

	// 等待处理完成
	s.wg.Wait()

	fmt.Println("Message sender stopped")
}

// SendMetric 发送指标数据
func (s *Sender) SendMetric(data interface{}) error {
	message := &Message{
		Type:       "metric",
		Data:       data,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	return s.sendMessage(message)
}

// SendAlert 发送告警数据
func (s *Sender) SendAlert(data interface{}) error {
	message := &Message{
		Type:       "alert",
		Data:       data,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	return s.sendMessage(message)
}

// sendMessage 发送消息到队列
func (s *Sender) sendMessage(message *Message) error {
	select {
	case s.messageQueue <- message:
		return nil
	case <-s.ctx.Done():
		return fmt.Errorf("sender is shutting down")
	default:
		return fmt.Errorf("message queue is full")
	}
}

// messageProcessor 消息处理器
func (s *Sender) messageProcessor() {
	defer s.wg.Done()

	for {
		select {
		case message := <-s.messageQueue:
			if message != nil {
				s.processMessage(message)
			}
		case <-s.ctx.Done():
			// 处理剩余消息
			s.flushQueue()
			return
		}
	}
}

// processMessage 处理单个消息
func (s *Sender) processMessage(message *Message) {
	// 限流控制
	if err := s.rateLimiter.Wait(s.ctx); err != nil {
		fmt.Printf("Rate limiter error: %v\n", err)
		return
	}

	// 发送消息
	if err := s.sendToServer(message); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)

		// 重试逻辑
		if message.RetryCount < s.maxRetries {
			message.RetryCount++
			go s.scheduleRetry(message)
		} else {
			fmt.Printf("Message failed after %d retries, dropping: %s\n", s.maxRetries, message.Type)
		}
	} else {
		fmt.Printf("Message sent successfully: %s\n", message.Type)
	}
}

// scheduleRetry 安排重试
func (s *Sender) scheduleRetry(message *Message) {
	timer := time.NewTimer(s.retryDelay * time.Duration(message.RetryCount))
	defer timer.Stop()

	select {
	case <-timer.C:
		select {
		case s.messageQueue <- message:
			// 重新入队成功
		case <-s.ctx.Done():
			// 发送器已停止
		}
	case <-s.ctx.Done():
		// 发送器已停止
	}
}

// sendToServer 发送到服务端
func (s *Sender) sendToServer(message *Message) error {
	// 序列化消息
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 构造请求URL
	url := fmt.Sprintf("%s/api/v1/messages", s.serverURL)

	// 创建请求
	req, err := http.NewRequestWithContext(s.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Baize-Agent/1.0")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// flushQueue 刷新队列中的剩余消息
func (s *Sender) flushQueue() {
	fmt.Println("Flushing remaining messages...")

	// 处理队列中剩余的消息
	for {
		select {
		case message := <-s.messageQueue:
			if message != nil {
				// 尝试最后一次发送
				if err := s.sendToServer(message); err != nil {
					fmt.Printf("Failed to send final message %s: %v\n", message.Type, err)
				}
			}
		default:
			// 队列为空
			return
		}
	}
}

// GetQueueStats 获取队列统计信息
func (s *Sender) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(s.messageQueue),
		"max_retries":  s.maxRetries,
		"retry_delay":  s.retryDelay.String(),
	}
}
