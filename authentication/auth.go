package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"ginMeterBox/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "ginmeterbox_session"
	defaultSessionTTL = 8 * time.Hour
)

type session struct {
	expiresAt time.Time
}

type Service struct {
	passwordHash []byte
	cookieSecure bool
	ttl          time.Duration
	now          func() time.Time

	mu       sync.Mutex
	sessions map[string]session
}

type loginRequest struct {
	Password string `json:"password"`
}

func NewService(passwordHash string, cookieSecure bool) *Service {
	return NewServiceWithTTL(passwordHash, cookieSecure, defaultSessionTTL)
}

func NewServiceWithTTL(passwordHash string, cookieSecure bool, ttl time.Duration) *Service {
	return &Service{
		passwordHash: []byte(passwordHash),
		cookieSecure: cookieSecure,
		ttl:          ttl,
		now:          time.Now,
		sessions:     make(map[string]session),
	}
}

func (s *Service) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Password == "" {
		response.BadRequest(c, "请输入密码")
		return
	}
	if bcrypt.CompareHashAndPassword(s.passwordHash, []byte(request.Password)) != nil {
		response.Unauthorized(c, "密码错误")
		return
	}

	token, err := s.createSession()
	if err != nil {
		response.ServerError(c, "创建会话失败")
		return
	}
	s.setSessionCookie(c, token)
	response.OK(c, gin.H{"authenticated": true})
}

func (s *Service) Logout(c *gin.Context) {
	if token, err := c.Cookie(SessionCookieName); err == nil {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
	}
	s.clearSessionCookie(c)
	response.OK(c, gin.H{"authenticated": false})
}

func (s *Service) Session(c *gin.Context) {
	response.OK(c, gin.H{"authenticated": true})
}

func (s *Service) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil || !s.isValid(token) {
			response.Unauthorized(c, "未登录或会话已过期")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Service) Authenticate() gin.HandlerFunc {
	return s.RequireAuth()
}

func (s *Service) createSession() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	now := s.now()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeExpiredLocked(now)
	s.sessions[token] = session{expiresAt: now.Add(s.ttl)}
	return token, nil
}

func (s *Service) isValid(token string) bool {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeExpiredLocked(now)

	stored, ok := s.sessions[token]
	return ok && now.Before(stored.expiresAt)
}

func (s *Service) removeExpiredLocked(now time.Time) {
	for token, stored := range s.sessions {
		if !now.Before(stored.expiresAt) {
			delete(s.sessions, token)
		}
	}
}

func (s *Service) setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.cookieSecure,
	})
}

func (s *Service) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.cookieSecure,
	})
}
