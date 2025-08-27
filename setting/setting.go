package setting

import (
	"context"
	"github.com/MoSed3/otp-server/db"
	"log"
	"sync"
	"time"
)

var mutex sync.RWMutex
var secretKey []byte
var accessTokenExpire uint

func Update(s *db.Setting) {
	mutex.Lock()
	defer mutex.Unlock()

	accessTokenExpire = s.AccessTokenExpire
	secretKey = []byte(s.SecretKey)
}

func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx := db.GetTransaction(ctx)

	s, err := db.GetSetting(tx)
	if err != nil {
		log.Fatalf("Error while getting settings from db: %v", err)
	}

	Update(s)
}

func SecretKey() []byte {
	mutex.RLock()
	defer mutex.RUnlock()
	return secretKey
}

func AccessTokenExpire() uint {
	mutex.RLock()
	defer mutex.RUnlock()
	return accessTokenExpire
}
