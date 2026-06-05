package main

import (
	_ "github.com/joho/godotenv/autoload"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"


	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Host string
	Port int
	Auth struct {
		AccessSecret string
	}
	Services []struct {
		Prefix string
		Target string
	}
}

func main() {
	var configFile = flag.String("f", "etc/gateway.yaml", "config file")
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 构建反向代理映射
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, s := range c.Services {
		targetUrl, _ := url.Parse(s.Target)
		proxies[s.Prefix] = httputil.NewSingleHostReverseProxy(targetUrl)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// 1. JWT 校验 (排除注册/登录等路径)
		if !isPublicPath(r.URL.Path) {
			tokenStr := r.Header.Get("Authorization")
			tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}
			
			claims := make(jwt.MapClaims)
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(c.Auth.AccessSecret), nil
			})

			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Unauthorized")
				return
			}

			// 2. 身份透传：将 userId 注入 Header
			userId := claims["userId"].(string)
			r.Header.Set("X-User-ID", userId)
		}

		// 3. 动态路由转发
		for prefix, proxy := range proxies {
			if strings.HasPrefix(r.URL.Path, prefix) {
				proxy.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Printf("Global Gateway starting at %s:%d...\n", c.Host, c.Port)
	logx.Must(http.ListenAndServe(fmt.Sprintf("%s:%d", c.Host, c.Port), http.HandlerFunc(handler)))
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/user/login",
		"/api/v1/user/register",
	}
	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}
	return false
}
