package aliyundrive

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type TokenManager interface {
	AccessToken(ctx context.Context) (string, error)
}

type staticTokenManager struct {
	accessToken string
}

func NewStaticTokenManager(accessToken string) *staticTokenManager {
	return &staticTokenManager{accessToken: accessToken}
}

func (m *staticTokenManager) AccessToken(ctx context.Context) (string, error) {
	return m.accessToken, nil
}

type refreshTokenManager struct {
	drive                 *Drive
	refreshToken          string
	accessToken           string
	accessTokenExpireTime time.Time
	lock                  *sync.Mutex
}

func NewRefreshTokenManager(drive *Drive, refreshToken string) *refreshTokenManager {
	return &refreshTokenManager{
		drive:                 drive,
		refreshToken:          refreshToken,
		accessTokenExpireTime: time.Unix(0, 0),
		lock:                  new(sync.Mutex),
	}
}

func (m *refreshTokenManager) AccessToken(ctx context.Context) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	now := time.Now()
	if now.Before(m.accessTokenExpireTime) {
		return m.accessToken, nil
	}

	err := m.refresh(ctx)
	if err != nil {
		return "", err
	}
	return m.accessToken, nil
}

func (m *refreshTokenManager) refresh(ctx context.Context) error {
	now := time.Now()
	api := "https://api.aliyundrive.com/token/refresh"
	params := Object{
		"refresh_token": m.refreshToken,
	}
	respData, err := m.drive.requestWithoutCredit(ctx, api, params)
	if err != nil {
		return err
	}

	result := &struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}{}
	err = json.Unmarshal(respData, result)
	if err != nil {
		return err
	}

	m.refreshToken = result.RefreshToken
	m.accessToken = result.AccessToken
	m.accessTokenExpireTime = now.Add(time.Second * time.Duration(result.ExpiresIn-60))
	return nil
}

type keepAliveTokenManager struct {
	tokenManager TokenManager
	wg           *sync.WaitGroup
}

func NewKeepAliveTokenManager(tokenManager TokenManager) *keepAliveTokenManager {
	return &keepAliveTokenManager{
		tokenManager: tokenManager,
		wg:           new(sync.WaitGroup),
	}
}

func (m *keepAliveTokenManager) AccessToken(ctx context.Context) (string, error) {
	return m.tokenManager.AccessToken(ctx)
}

func (m *keepAliveTokenManager) KeepAlive(ctx context.Context, t time.Duration) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(t)
	keepaliveLoop:
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				break keepaliveLoop
			}
			m.AccessToken(ctx)
		}
		ticker.Stop()
	}()
}

func (m *keepAliveTokenManager) WaitStop() {
	m.wg.Wait()
}
