package auth

import (
	"time"
	"sync"
)

type Tokens interface {
	addToken(hash int64, id int, ctx map[string]interface{}) int64
	rmToken(bearer string) error
}

const tokenExpires = 60*60*24

type mapToken struct {
	accessToken     int64
	userId          int
	expiresIn       time.Duration
	isAdmin         bool
	ctxRoute        map[string]interface{}	
	lock            *sync.RWMutex
}


func (m *mapToken) SetValue(key string, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ctxRoute[key] = val
}

func (m *mapToken) Value(key interface{}) interface{} {
	if key, ok := key.(string); ok {
		return m.ctxRoute[key]
	}

	return nil
}

func (m *mapToken) GetUserID() int {
	return m.userId
}

type mapTokens struct {
	tokens map[int64] *mapToken
	lock sync.RWMutex
}


func (m *mapTokens) addToken(hash int64, id int, ctx map[string]interface{}) int64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.tokens == nil {
		m.tokens = make(map[int64] *mapToken, 0)
	}
	
		m.tokens[hash] = &mapToken{
			accessToken: hash,
			userId     : id,
			ctxRoute   : ctx,
			expiresIn  : time.Now(),
			lock	   : &sync.RWMutex{},
	}
	//	todo: решить вопрос про уникальность токена
	return m.tokens[hash].accessToken
}

func (m *mapTokens) getToken(bearer string) int64 {
	hash, err := strconv.ParseInt(bearer, 10, 64)
	if err != nil {
		logs.ErrorLog(err, "ParseInt( bearer %s", string(b))
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	token, ok := m.tokens[hash]
	if ok {
		return token.accessToken
	} 
		
	return -1
}

func (m *mapTokens) rmToken(bearer string) error {
	hash, err := strconv.ParseInt( bearer, 10, 64 )
	if err != nil {
		logs.ErrorLog(err, "ParseInt( bearer %s", string(b))
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	
	_, ok := m.tokens[hash]
	if !ok {
		return errors.New("not found user in active")
	}
	
	delete(m.tokens, hash)

	return nil
}