package actor

import (
	"sync"

	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/pkg/sse"
)

type ActorSystem struct {
	mu     sync.RWMutex
	actors map[string]*taskActor
	repo   *repo.LogRepository
	hub    *sse.SSEHub
}

func NewActorSystem(repo *repo.LogRepository, hub *sse.SSEHub) *ActorSystem {
	return &ActorSystem{
		actors: make(map[string]*taskActor),
		repo:   repo,
		hub:    hub,
	}
}

func (s *ActorSystem) getOrCreate(taskId string) *taskActor {
	s.mu.RLock()
	a := s.actors[taskId]
	s.mu.RUnlock()
	if a != nil {
		return a
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if a = s.actors[taskId]; a != nil {
		return a
	}
	a = newTaskActor(taskId, s.repo, s.hub, 1024)
	s.actors[taskId] = a
	return a
}

func (s *ActorSystem) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, a := range s.actors {
		a.Stop()
		delete(s.actors, id)
	}
}
