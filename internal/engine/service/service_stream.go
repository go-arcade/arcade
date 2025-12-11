package service

import (
	"context"
	"fmt"
	"io"
	"sync"

	streamv1 "github.com/go-arcade/arcade/api/stream/v1"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

// StreamServiceImpl Stream 服务实现
type StreamServiceImpl struct {
	streamv1.UnimplementedStreamServiceServer
	logAggregator *LogAggregator
	redis         *redis.Client
	mongo         *mongo.Database
	mu            sync.RWMutex
	subscribers   map[string][]*LogSubscriber // taskID -> subscribers
}

// LogSubscriber 日志订阅者
type LogSubscriber struct {
	TaskID string
	Stream grpc.ServerStreamingServer[streamv1.StreamTaskLogResponse]
	Cancel context.CancelFunc
}

// NewStreamService 创建Stream服务实例
func NewStreamService(redis *redis.Client, mongo *mongo.Database) *StreamServiceImpl {
	return &StreamServiceImpl{
		logAggregator: NewLogAggregator(redis, mongo),
		redis:         redis,
		mongo:         mongo,
		subscribers:   make(map[string][]*LogSubscriber),
	}
}

// GetLogAggregator 获取日志聚合器
func (s *StreamServiceImpl) GetLogAggregator() *LogAggregator {
	return s.logAggregator
}

// UploadTaskLog Agent端流式上报日志给Server
func (s *StreamServiceImpl) UploadTaskLog(stream grpc.ClientStreamingServer[streamv1.UploadTaskLogRequest, streamv1.UploadTaskLogResponse]) error {
	var taskID, agentID string
	receivedLines := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// 客户端关闭流，返回响应
			log.Infow("log upload stream closed", "taskId", taskID, "agentId", agentID, "receivedLines", receivedLines)
			return stream.SendAndClose(&streamv1.UploadTaskLogResponse{
				Success:       true,
				Message:       "logs uploaded successfully",
				ReceivedLines: receivedLines,
			})
		}
		if err != nil {
			log.Errorw("failed to receive log upload", "error", err)
			return err
		}

		// 记录任务ID和AgentID
		if taskID == "" {
			taskID = req.TaskId
			agentID = req.AgentId
			log.Infow("start receiving logs for task", "taskId", taskID, "agentId", agentID)
		}

		// 转换日志条目并推送到聚合器
		for _, logChunk := range req.Logs {
			entry := &LogEntry{
				TaskID:     req.TaskId,
				Timestamp:  logChunk.Timestamp,
				LineNumber: logChunk.LineNumber,
				Level:      logChunk.Level,
				Content:    logChunk.Content,
				Stream:     logChunk.Stream,
				AgentID:    req.AgentId,
			}

			if err := s.logAggregator.PushLog(entry); err != nil {
				log.Errorw("failed to push log to aggregator", "taskId", req.TaskId, "error", err)
			}
			receivedLines++

			// 通知订阅者
			s.notifySubscribers(req.TaskId, logChunk)
		}
	}
}

// StreamTaskLog 实时获取任务日志流
func (s *StreamServiceImpl) StreamTaskLog(req *streamv1.StreamTaskLogRequest, stream grpc.ServerStreamingServer[streamv1.StreamTaskLogResponse]) error {
	ctx := stream.Context()
	taskID := req.JobId

	log.Infow("client requesting log stream", "taskId", taskID, "fromLine", req.FromLine, "follow", req.Follow)

	// 先从MongoDB获取历史日志
	historicalLogs, err := s.logAggregator.GetLogsByTaskID(taskID, req.FromLine, 1000)
	if err != nil {
		log.Errorw("failed to get historical logs", "taskId", taskID, "error", err)
		return err
	}

	// 发送历史日志
	for _, entry := range historicalLogs {
		logChunk := &streamv1.LogChunk{
			Timestamp:  entry.Timestamp,
			LineNumber: entry.LineNumber,
			Level:      entry.Level,
			Content:    entry.Content,
			Stream:     entry.Stream,
		}

		if err := stream.Send(&streamv1.StreamTaskLogResponse{
			TaskId:     taskID,
			LogChunk:   logChunk,
			IsComplete: false,
		}); err != nil {
			log.Errorw("failed to send log", "taskId", taskID, "error", err)
			return err
		}
	}

	// 如果不需要持续跟踪，发送完成标记并返回
	if !req.Follow {
		return stream.Send(&streamv1.StreamTaskLogResponse{
			TaskId:     taskID,
			IsComplete: true,
		})
	}

	// 订阅实时日志
	subscriber := &LogSubscriber{
		TaskID: taskID,
		Stream: stream,
	}
	ctx2, cancel := context.WithCancel(ctx)
	subscriber.Cancel = cancel
	defer cancel()

	// 注册订阅者
	s.mu.Lock()
	s.subscribers[taskID] = append(s.subscribers[taskID], subscriber)
	s.mu.Unlock()

	// 清理订阅者
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		subs := s.subscribers[taskID]
		for i, sub := range subs {
			if sub == subscriber {
				s.subscribers[taskID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(s.subscribers[taskID]) == 0 {
			delete(s.subscribers, taskID)
		}
	}()

	// 订阅实时日志channel
	logChan := s.logAggregator.Subscribe(ctx2, taskID)
	log.Infow("subscribed to real-time logs for task", "taskId", taskID)

	// 持续发送日志直到上下文取消
	for {
		select {
		case <-ctx.Done():
			log.Infow("log stream cancelled for task", "taskId", taskID)
			return ctx.Err()
		case entry, ok := <-logChan:
			if !ok {
				// 频道关闭
				return stream.Send(&streamv1.StreamTaskLogResponse{
					TaskId:     taskID,
					IsComplete: true,
				})
			}

			logChunk := &streamv1.LogChunk{
				Timestamp:  entry.Timestamp,
				LineNumber: entry.LineNumber,
				Level:      entry.Level,
				Content:    entry.Content,
				Stream:     entry.Stream,
			}

			if err := stream.Send(&streamv1.StreamTaskLogResponse{
				TaskId:     taskID,
				LogChunk:   logChunk,
				IsComplete: false,
			}); err != nil {
				log.Errorw("failed to send real-time log", "taskId", taskID, "error", err)
				return err
			}
		}
	}
}

// StreamTaskStatus 实时获取任务状态流
func (s *StreamServiceImpl) StreamTaskStatus(req *streamv1.StreamTaskStatusRequest, stream grpc.ServerStreamingServer[streamv1.StreamTaskStatusResponse]) error {
	// TODO: 实现任务状态流
	return fmt.Errorf("not implemented")
}

// StreamPipelineStatus 实时获取流水线状态流
func (s *StreamServiceImpl) StreamPipelineStatus(req *streamv1.StreamPipelineStatusRequest, stream grpc.ServerStreamingServer[streamv1.StreamPipelineStatusResponse]) error {
	// TODO: 实现流水线状态流
	return fmt.Errorf("not implemented")
}

// AgentChannel Agent与Server双向通信流
func (s *StreamServiceImpl) AgentChannel(stream grpc.BidiStreamingServer[streamv1.AgentChannelRequest, streamv1.AgentChannelResponse]) error {
	// TODO: 实现Agent通道
	return fmt.Errorf("not implemented")
}

// StreamAgentStatus 实时监控Agent状态流
func (s *StreamServiceImpl) StreamAgentStatus(req *streamv1.StreamAgentStatusRequest, stream grpc.ServerStreamingServer[streamv1.StreamAgentStatusResponse]) error {
	// TODO: 实现Agent状态流
	return fmt.Errorf("not implemented")
}

// StreamEvents 实时事件流
func (s *StreamServiceImpl) StreamEvents(req *streamv1.StreamEventsRequest, stream grpc.ServerStreamingServer[streamv1.StreamEventsResponse]) error {
	// TODO: 实现事件流
	return fmt.Errorf("not implemented")
}

// notifySubscribers 通知订阅者
func (s *StreamServiceImpl) notifySubscribers(taskID string, logChunk *streamv1.LogChunk) {
	s.mu.RLock()
	subs := s.subscribers[taskID]
	s.mu.RUnlock()

	if len(subs) == 0 {
		return
	}

	// 异步通知所有订阅者
	for _, sub := range subs {
		go func(subscriber *LogSubscriber) {
			err := subscriber.Stream.Send(&streamv1.StreamTaskLogResponse{
				TaskId:     taskID,
				LogChunk:   logChunk,
				IsComplete: false,
			})
			if err != nil {
				log.Errorw("failed to send log to subscriber", "taskId", taskID, "error", err)
			}
		}(sub)
	}
}
