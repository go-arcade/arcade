package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	agentmodel "github.com/go-arcade/arcade/internal/engine/model"
	agentrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

type AgentService struct {
	agentRepo              agentrepo.IAgentRepository
	generalSettingsService *GeneralSettingsService
}

func NewAgentService(agentRepo agentrepo.IAgentRepository, generalSettingsService *GeneralSettingsService) *AgentService {
	return &AgentService{
		agentRepo:              agentRepo,
		generalSettingsService: generalSettingsService,
	}
}

func (al *AgentService) CreateAgent(createAgentReq *agentmodel.CreateAgentReq) (*agentmodel.CreateAgentResp, error) {
	agentId := id.ShortId()
	agent := &agentmodel.Agent{
		AgentId:   agentId,
		AgentName: createAgentReq.AgentName,
		Address:   "0.0.0.0",
		Port:      "8080",
		OS:        "Linux",
		Arch:      "amd64",
		Version:   "0.0.0",
		Status:    createAgentReq.Status,
		Labels:    createAgentReq.Labels,
		IsEnabled: 1,
		Metrics:   "/metrics",
	}

	// Create Agent
	if err := al.agentRepo.CreateAgent(agent); err != nil {
		log.Errorw("create agent failed", "error", err)
		return nil, err
	}

	// Generate token for agent communication based on agentId
	token, err := al.generateAgentToken(agentId)
	if err != nil {
		log.Errorw("generate agent token failed", "error", err)
		return nil, err
	}

	// Return created agent with token
	resp := &agentmodel.CreateAgentResp{
		Agent: *agent,
		Token: token,
	}
	return resp, nil
}

// agentSecretConfig represents the structure of agent secret key configuration
type agentSecretConfig struct {
	Salt      string `json:"salt"`
	SecretKey string `json:"secret_key"`
}

// generateAgentToken generates a token based on agentId for agent communication
// The token is generated using HMAC-SHA256 with secret key and salt from database
func (al *AgentService) generateAgentToken(agentId string) (string, error) {
	// Get agent secret key configuration from database
	settings, err := al.generalSettingsService.GetGeneralSettingsByName("system", "agent_secret_key")
	if err != nil {
		log.Errorw("failed to get agent secret key configuration", "error", err)
		return "", err
	}

	// Parse JSON data
	var config agentSecretConfig
	if err := sonic.Unmarshal(settings.Data, &config); err != nil {
		log.Errorw("failed to parse agent secret key configuration", "error", err)
		return "", err
	}

	// Validate configuration
	if config.SecretKey == "" {
		log.Errorw("agent secret key is empty")
		return "", fmt.Errorf("agent secret key is empty")
	}
	if config.Salt == "" {
		log.Errorw("agent salt is empty")
		return "", fmt.Errorf("agent salt is empty")
	}

	// Generate token using HMAC-SHA256
	h := hmac.New(sha256.New, []byte(config.SecretKey))
	h.Write([]byte(agentId))
	h.Write([]byte(config.Salt))
	signature := h.Sum(nil)

	token := base64.URLEncoding.EncodeToString(signature)
	return token, nil
}

func (al *AgentService) GetAgentById(id uint64) (*agentmodel.AgentDetail, error) {
	detail, err := al.agentRepo.GetAgentDetailById(id)
	if err != nil {
		log.Errorw("get agent detail by id failed", "id", id, "error", err)
		return nil, err
	}
	return detail, nil
}

func (al *AgentService) UpdateAgent(id uint64, updateReq *agentmodel.UpdateAgentReq) error {
	// Check if agent exists
	_, err := al.agentRepo.GetAgentById(id)
	if err != nil {
		log.Errorw("get agent by id failed", "id", id, "error", err)
		return err
	}

	// Build and update Agent fields
	updates := buildAgentUpdateMap(updateReq)
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := al.agentRepo.UpdateAgentById(id, updates); err != nil {
			log.Errorw("update agent failed", "id", id, "error", err)
			return err
		}
	}

	return nil
}

// buildAgentUpdateMap builds update map for Agent fields
func buildAgentUpdateMap(req *agentmodel.UpdateAgentReq) map[string]any {
	updates := make(map[string]any)
	setIfNotNil(updates, "agent_name", req.AgentName)
	setIfNotNil(updates, "address", req.Address)
	setIfNotNil(updates, "port", req.Port)
	setIfNotNil(updates, "os", req.OS)
	setIfNotNil(updates, "arch", req.Arch)
	setIfNotNil(updates, "version", req.Version)
	setIfNotNil(updates, "status", req.Status)
	setIfNotNil(updates, "is_enabled", req.IsEnabled)
	if req.Labels != nil {
		updates["labels"] = req.Labels
	}
	if req.Metrics != nil {
		updates["metrics"] = req.Metrics
	}
	return updates
}

// setIfNotNil sets the value in the map if the pointer is not nil
func setIfNotNil[T any](m map[string]any, key string, ptr *T) {
	if ptr != nil {
		m[key] = *ptr
	}
}

func (al *AgentService) DeleteAgent(id uint64) error {
	if err := al.agentRepo.DeleteAgent(id); err != nil {
		log.Errorw("delete agent failed", "id", id, "error", err)
		return err
	}
	return nil
}

func (al *AgentService) ListAgent(pageNum, pageSize int) ([]agentmodel.Agent, int64, error) {
	offset := (pageNum - 1) * pageSize
	agents, count, err := al.agentRepo.ListAgent(offset, pageSize)

	if err != nil {
		log.Errorw("list agent failed", "error", err)
		return nil, 0, err
	}
	return agents, count, err
}
