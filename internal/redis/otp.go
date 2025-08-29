package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type LoginState int

const (
	LoginStateWaiting LoginState = iota
	LoginStateSuccess
	LoginStateCorrupted
)

type UserLoginSession struct {
	Tries       uint       `json:"tries"`
	OtpID       uint       `json:"otp_id"`
	Code        string     `json:"code"`
	PhoneNumber string     `json:"phone_number"`
	State       LoginState `json:"state"`
}

func (c *Config) SetUserLoginSession(ctx context.Context, key string, session *UserLoginSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// Set with 3-minute expiration
	return c.client.Set(ctx, key, data, 3*time.Minute).Err()
}

func (c *Config) CreateUserLoginSession(ctx context.Context, otpID uint, code string) (string, error) {
	key := uuid.New().String()

	session := &UserLoginSession{
		Tries: 0,
		OtpID: otpID,
		Code:  code,
		State: LoginStateWaiting,
	}

	return key, c.SetUserLoginSession(ctx, key, session)
}

const luaIncrementTries = `
local key = KEYS[1]
local session = redis.call('GET', key)
if not session then
    return nil
end

local data = cjson.decode(session)
data.tries = data.tries + 1

if data.tries > 3 then
    data.state = 2
end

local updated = cjson.encode(data)
redis.call('SET', key, updated, 'EX', 180)
return updated
`

func (c *Config) IncreaseUserLoginTries(ctx context.Context, key string) (*UserLoginSession, error) {
	result, err := c.client.Eval(ctx, luaIncrementTries, []string{key}).Result()
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("session not found")
	}

	var session UserLoginSession
	resultStr, ok := result.(string)
	if !ok {
		return nil, errors.New("unexpected result type from Redis Lua script")
	}
	if err = json.Unmarshal([]byte(resultStr), &session); err != nil {
		return nil, err
	}

	switch session.State {
	case LoginStateSuccess, LoginStateCorrupted:
		return &session, errors.New("invalid code")
	default:
	}

	return &session, nil
}

func (c *Config) CheckUserLoginCode(ctx context.Context, token, code string) (uint, error) {
	session, err := c.IncreaseUserLoginTries(ctx, token)
	if err != nil {
		return 0, err
	}

	if session.Code != code {
		return 0, errors.New("invalid code")
	}

	session.State = LoginStateSuccess
	return session.OtpID, c.SetUserLoginSession(ctx, token, session)
}
