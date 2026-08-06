package authentication

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginAndProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}
	service := NewService(string(hash), false)
	router := gin.New()
	router.POST("/login", service.Login)
	router.GET("/protected", service.RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("未认证请求状态码 = %d，期望 %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	failedLogin := httptest.NewRecorder()
	failedLoginRequest := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"password":"wrong"}`))
	failedLoginRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(failedLogin, failedLoginRequest)
	if failedLogin.Code != http.StatusUnauthorized {
		t.Fatalf("错误密码登录状态码 = %d，期望 %d", failedLogin.Code, http.StatusUnauthorized)
	}

	successfulLogin := httptest.NewRecorder()
	successfulLoginRequest := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"password":"test-password"}`))
	successfulLoginRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(successfulLogin, successfulLoginRequest)
	if successfulLogin.Code != http.StatusOK {
		t.Fatalf("正确密码登录状态码 = %d，期望 %d", successfulLogin.Code, http.StatusOK)
	}
	cookies := successfulLogin.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("登录响应 Cookie 数量 = %d，期望 1", len(cookies))
	}
	if !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Path != "/" {
		t.Fatal("登录 Cookie 未使用预期安全属性")
	}

	authenticated := httptest.NewRecorder()
	authenticatedRequest := httptest.NewRequest(http.MethodGet, "/protected", nil)
	authenticatedRequest.AddCookie(cookies[0])
	router.ServeHTTP(authenticated, authenticatedRequest)
	if authenticated.Code != http.StatusNoContent {
		t.Fatalf("认证后请求状态码 = %d，期望 %d", authenticated.Code, http.StatusNoContent)
	}
}
