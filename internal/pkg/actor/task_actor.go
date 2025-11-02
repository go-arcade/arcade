package actor

import (
	"sync"
	"time"

	streamv1 "github.com/go-arcade/arcade/api/stream/v1"
	"github.com/go-arcade/arcade/internal/engine/model/entity"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/pkg/sse"
)

type logBatch struct {
	TaskID  string
	AgentID string
	Logs    []*streamv1.LogChunk
}

type taskActor struct {
	taskID   string
	mailbox  chan logBatch
	repo     *repo.LogRepository
	hub      *sse.SSEHub
	shutdown chan struct{}
	wg       sync.WaitGroup
}

func newTaskActor(taskID string, repo *repo.LogRepository, hub *sse.SSEHub, mailbox int) *taskActor {
	a := &taskActor{
		taskID:   taskID,
		mailbox:  make(chan logBatch, mailbox),
		repo:     repo,
		hub:      hub,
		shutdown: make(chan struct{}),
	}
	a.wg.Add(1)
	go a.loop()
	return a
}

func (a *taskActor) loop() {
	defer a.wg.Done()
	batchBuf := make([]entity.TaskLog, 0, 256)
	flush := func() {
		if len(batchBuf) == 0 {
			return
		}
		// _ = a.repo.InsertMany(context.Background(), batchBuf)
		// for _, doc := range batchBuf {
		// 	a.hub.Broadcast(doc.TaskID, doc)
		// }
		batchBuf = batchBuf[:0]
	}

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case b := <-a.mailbox:

			// 聚合写 Mongo，并实时广播给 SSE 与 gRPC 订阅者（SSE 使用 hub，gRPC 下方会利用 hub 同步获取）
			for _, l := range b.Logs {
				doc := entity.TaskLog{
					TaskID:  b.TaskID,
					AgentID: b.AgentID,
					Logs:    []*streamv1.LogChunk{l},
				}
				batchBuf = append(batchBuf, doc)
			}
			// 若积累较大，立即 flush
			if len(batchBuf) >= 512 {
				flush()
			}

			// 指标更新

		case <-ticker.C:
			flush()
		case <-a.shutdown:
			flush()
			return
		}
	}
}

func (a *taskActor) Tell(b logBatch) {
	select {
	case a.mailbox <- b:
	default:
		<-a.mailbox
		a.mailbox <- b
	}
}

func (a *taskActor) Stop() {
	close(a.shutdown)
	a.wg.Wait()
}
