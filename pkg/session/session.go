package session

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)


func (s *session) Found(sessionID string) (status bool)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	if _, found := s.Registry.M[sessionID]; found {
		return true
	}

	return false
}

func (s *session) GetProfile(sessionID string) (profile *models.ProfileData, err error)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	if _, found := s.Registry.M[sessionID]; found {
		prf := s.Registry.M[sessionID].Profile
		profile = &prf
	}

	return profile, err
}

func (s *session) Delete(sessionID string) (err error)  {
	if sessionID == "" {
		return err
	}

	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	delete(s.Registry.M, sessionID)

	return err
}

func (s *session) Set(sessionID string) (err error)  {
	var profile models.ProfileData
	var f = SessionRec{}

	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	if s.Registry.M == nil {
		s.Registry.M = map[string]SessionRec{}
	}

	expiration := time.Now().Add(30 * time.Hour)
	// получем данные из IAM
	b1, err := s.iam.ProfileGet(sessionID)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(b1), &profile)
	if err != nil {
		return err
	}

	// сохраняем значение сессии в локальном хранилище приложения
	f.Profile = profile
	f.DeadTime = expiration.Unix()
	s.Registry.M[sessionID] = f

	return err
}

// список всех токенов для всех пользователей доступных для сервиса
func (s *session) List() (result map[string]SessionRec)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	result = s.Registry.M

	return result
}

//////////////////////////////////
// запускаем очиститель сессий для сервиса
//////////////////////////////////
func (s *session) Cleaner(ctx context.Context) (err error) {
	ticker := time.NewTicker(s.cfg.IntervalCleaner.Value)
	defer ticker.Stop()

	defer func(l lib.Log) {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			l.Warning(fmt.Errorf("%s (Error: %s)", b, rec), "Panic error Balancer")
		}
	}(s.logger)

	for {
		select {
		case <- ctx.Done():
			return
		case <- ticker.C:
			s.CleanSession(ctx)
			ticker = time.NewTicker(s.cfg.IntervalCleaner.Value)
		}
	}

	return
}

func (s *session) CleanSession(ctx context.Context) (err error) {

	listIAM, err := s.iam.ProfileList()		// получаем список актуальных сессий с сервера IAM
	if err != nil {
		return err
	}
	listRegistry := s.List()				// текущий реестр сессий

	// пробегаем свой реестр и если нет в нем ключа из списка сессий с IAM, удаляем
	for key, _ := range listRegistry {
		if !strings.Contains(listIAM, key) {
			s.Delete(key)					// удаляем значение сессии из реестра
		}
	}

	return err
}