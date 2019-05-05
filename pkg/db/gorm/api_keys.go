package gorm

import (
	"fmt"
	"strconv"
	"time"

	apiG "github.com/tomogoma/go-api-guard"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/api"
)

// InsertAPIKey inserts an API key for the userID.
func (g *Gorm) InsertAPIKey(userID string, key []byte) (apiG.Key, error) {
	if err := g.InitDBIfNot(); err != nil {
		return nil, err
	}
	k := api.Key{UserID: userID, Val: key}
	intUserId, err := strconv.Atoi(userID)
	if err != nil {
		return nil, err
	}

	var apiKey ApiKey
	gormErr := g.db.FirstOrCreate(&apiKey, ApiKey{UserId: intUserId, Key: string(key), UpdatedAt: time.Now()})
	if gormErr.Error != nil {
		return nil, gormErr.Error
	}
	k.LastUpdated = apiKey.UpdatedAt
	return k, nil
}

// APIKeyByUserIDVal returns API keys for the provided userID/key combination.
func (g *Gorm) APIKeyByUserIDVal(userID string, key []byte) (apiG.Key, error) {
	if err := g.InitDBIfNot(); err != nil {
		return nil, err
	}
	var apiKey ApiKey
	err := g.db.Where(ColUserID+" = ?", userID).Where(ColKey+" = ?", key).First(&apiKey)
	if err.Error != nil {
		if err.RecordNotFound() {
			return nil, errors.NewNotFound("API key not found")
		}
		return nil, err.Error
	}

	k := api.Key{
		ID:          fmt.Sprint(apiKey.ID),
		UserID:      strconv.Itoa(apiKey.UserId),
		Val:         []byte(apiKey.Key),
		Created:     apiKey.CreatedAt,
		LastUpdated: apiKey.UpdatedAt,
	}
	return k, nil
}
