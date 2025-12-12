package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRedisClient 模拟 Redis 客户端
type mockRedisClient struct {
	redis.UniversalClient
}

// 创建测试用的配置
func createTestConfig() *Config {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用不同的 DB 避免冲突
	})

	return &Config{
		RedisClient:      redisClient,
		MongoDB:          nil, // 测试时不使用 MongoDB
		Concurrency:      2,
		StrictPriority:   false,
		Queues:           map[string]int{Critical: 6, Default: 3, Low: 1},
		DefaultQueue:     Default,
		LogLevel:         "info",
		ShutdownTimeout:  10,
		GroupGracePeriod: 5,
		GroupMaxDelay:    20,
		GroupMaxSize:     100,
	}
}

func TestNewTaskQueue(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "queue config is required",
		},
		{
			name:    "nil redis client",
			cfg:     &Config{RedisClient: nil},
			wantErr: true,
			errMsg:  "redis client is required",
		},
		{
			name:    "valid config",
			cfg:     createTestConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue, err := NewTaskQueue(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, queue)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, queue)
				assert.NotNil(t, queue.client)
				assert.NotNil(t, queue.server)
				assert.NotNil(t, queue.mux)
				assert.NotNil(t, queue.config)
				assert.NotNil(t, queue.handlers)
			}
		})
	}
}

func TestTaskQueue_Enqueue(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	tests := []struct {
		name      string
		payload   *TaskPayload
		queueName string
		wantErr   bool
	}{
		{
			name: "valid payload with default queue",
			payload: &TaskPayload{
				TaskID:     "test-task-1",
				TaskType:   TaskTypeJob,
				Priority:   5,
				RetryCount: 3,
				Timeout:    3600,
			},
			queueName: "",
			wantErr:   false,
		},
		{
			name: "valid payload with specific queue",
			payload: &TaskPayload{
				TaskID:     "test-task-2",
				TaskType:   TaskTypePipeline,
				Priority:   6,
				RetryCount: 3,
				Timeout:    3600,
			},
			queueName: Critical,
			wantErr:   false,
		},
		{
			name: "payload with timeout",
			payload: &TaskPayload{
				TaskID:     "test-task-3",
				TaskType:   TaskTypeStep,
				Priority:   3,
				RetryCount: 2,
				Timeout:    1800,
			},
			queueName: Default,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queue.Enqueue(tt.payload, tt.queueName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskQueue_EnqueueWithPriority(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	tests := []struct {
		name           string
		payload        *TaskPayload
		priorityWeight int
		expectedQueue  string
		wantErr        bool
	}{
		{
			name: "high priority should use critical queue",
			payload: &TaskPayload{
				TaskID:     "test-high-priority",
				TaskType:   TaskTypeJob,
				RetryCount: 3,
				Timeout:    3600,
			},
			priorityWeight: 6,
			expectedQueue:  Critical,
			wantErr:        false,
		},
		{
			name: "medium priority should use default queue",
			payload: &TaskPayload{
				TaskID:     "test-medium-priority",
				TaskType:   TaskTypeJob,
				RetryCount: 3,
				Timeout:    3600,
			},
			priorityWeight: 3,
			expectedQueue:  Default,
			wantErr:        false,
		},
		{
			name: "low priority should use low queue",
			payload: &TaskPayload{
				TaskID:     "test-low-priority",
				TaskType:   TaskTypeJob,
				RetryCount: 3,
				Timeout:    3600,
			},
			priorityWeight: 1,
			expectedQueue:  Low,
			wantErr:        false,
		},
		{
			name: "priority higher than max should use highest queue",
			payload: &TaskPayload{
				TaskID:     "test-very-high-priority",
				TaskType:   TaskTypeJob,
				RetryCount: 3,
				Timeout:    3600,
			},
			priorityWeight: 10,
			expectedQueue:  Critical,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queue.EnqueueWithPriority(tt.payload, tt.priorityWeight)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskQueue_EnqueueDelayed(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	payload := &TaskPayload{
		TaskID:     "test-delayed-task",
		TaskType:   TaskTypeJob,
		Priority:   5,
		RetryCount: 3,
		Timeout:    3600,
	}

	delay := 5 * time.Second
	err = queue.EnqueueDelayed(payload, delay, "")
	assert.NoError(t, err)
}

func TestTaskQueue_RegisterHandler(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	handler := TaskHandlerFunc(func(ctx context.Context, payload *TaskPayload) error {
		assert.Equal(t, "test-task", payload.TaskID)
		return nil
	})

	queue.RegisterHandler("test-type", handler)
	assert.NotNil(t, queue.handlers["test-type"])
}

func TestTaskQueue_RegisterHandlerFunc(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	handlerFunc := func(ctx context.Context, payload *TaskPayload) error {
		return nil
	}

	queue.RegisterHandlerFunc("test-func-type", handlerFunc)
	assert.NotNil(t, queue.handlers["test-func-type"])
}

func TestTaskQueue_GetClient(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	client := queue.GetClient()
	assert.NotNil(t, client)
	assert.IsType(t, &asynq.Client{}, client)
}

func TestTaskQueue_GetServer(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	server := queue.GetServer()
	assert.NotNil(t, server)
	assert.IsType(t, &asynq.Server{}, server)
}

func TestTaskQueue_GetRedisConnOpt(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	redisOpt := queue.GetRedisConnOpt()
	assert.NotNil(t, redisOpt)
}

func TestNewTaskManager(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	manager := NewTaskManager(queue)
	assert.NotNil(t, manager)
	assert.Equal(t, queue, manager.queue)
}

func TestTaskManager_EnqueueTask(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	manager := NewTaskManager(queue)

	tests := []struct {
		name           string
		payload        *TaskPayload
		priorityWeight int
		wantErr        bool
	}{
		{
			name: "valid payload",
			payload: &TaskPayload{
				TaskID:     "test-manager-task",
				TaskType:   TaskTypeJob,
				RetryCount: 3,
				Timeout:    3600,
			},
			priorityWeight: 5,
			wantErr:        false,
		},
		{
			name:           "nil payload",
			payload:        nil,
			priorityWeight: 5,
			wantErr:        true,
		},
		{
			name: "payload with default timeout and retry",
			payload: &TaskPayload{
				TaskID:   "test-defaults",
				TaskType: TaskTypeJob,
			},
			priorityWeight: 3,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := manager.EnqueueTask(ctx, tt.payload, tt.priorityWeight)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.payload != nil {
					assert.Equal(t, tt.priorityWeight, tt.payload.Priority)
					if tt.payload.Timeout == 0 {
						assert.Equal(t, 3600, tt.payload.Timeout)
					}
					if tt.payload.RetryCount == 0 {
						assert.Equal(t, 3, tt.payload.RetryCount)
					}
				}
			}
		})
	}
}

func TestTaskManager_EnqueueDelayedTask(t *testing.T) {
	cfg := createTestConfig()
	queue, err := NewTaskQueue(cfg)
	require.NoError(t, err)
	defer queue.Shutdown()

	manager := NewTaskManager(queue)

	payload := &TaskPayload{
		TaskID:     "test-delayed-manager-task",
		TaskType:   TaskTypeJob,
		RetryCount: 3,
		Timeout:    3600,
	}

	delay := 10 * time.Second
	ctx := context.Background()
	err = manager.EnqueueDelayedTask(ctx, payload, delay)
	assert.NoError(t, err)
}

func TestTaskPayload_MarshalUnmarshal(t *testing.T) {
	original := &TaskPayload{
		TaskID:        "test-task",
		TaskType:      TaskTypePipeline,
		Priority:      5,
		PipelineID:    "pipeline-1",
		PipelineRunID: "run-1",
		StageID:       "stage-1",
		Stage:         1,
		AgentID:       "agent-1",
		Name:          "test-job",
		Commands:      []string{"echo", "hello"},
		Env:           map[string]string{"KEY": "value"},
		Workspace:     "/tmp",
		Timeout:       3600,
		RetryCount:    3,
		LabelSelector: map[string]any{"label": "value"},
		Data:          map[string]any{"data": "value"},
	}

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled TaskPayload
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.TaskID, unmarshaled.TaskID)
	assert.Equal(t, original.TaskType, unmarshaled.TaskType)
	assert.Equal(t, original.Priority, unmarshaled.Priority)
	assert.Equal(t, original.PipelineID, unmarshaled.PipelineID)
	assert.Equal(t, original.Commands, unmarshaled.Commands)
	assert.Equal(t, original.Env, unmarshaled.Env)
}

func TestTaskHandlerFunc(t *testing.T) {
	called := false
	handler := TaskHandlerFunc(func(ctx context.Context, payload *TaskPayload) error {
		called = true
		return nil
	})

	ctx := context.Background()
	payload := &TaskPayload{TaskID: "test"}
	err := handler.HandleTask(ctx, payload)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "pipeline", TaskTypePipeline)
	assert.Equal(t, "job", TaskTypeJob)
	assert.Equal(t, "step", TaskTypeStep)
	assert.Equal(t, "custom", TaskTypeCustom)

	assert.Equal(t, "critical", Critical)
	assert.Equal(t, "default", Default)
	assert.Equal(t, "low", Low)

	assert.Equal(t, "pending", TaskRecordStatusPending)
	assert.Equal(t, "running", TaskRecordStatusRunning)
	assert.Equal(t, "completed", TaskRecordStatusCompleted)
	assert.Equal(t, "failed", TaskRecordStatusFailed)
}
