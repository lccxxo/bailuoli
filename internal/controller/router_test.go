package controller

import (
	"fmt"
	"github.com/lccxxo/bailuoli/internal/model"
	"net/http"
	"testing"
)

// 测试路由匹配规则
func TestRouter_MatchRoute(t *testing.T) {
	routes := []*model.Route{
		{
			Method:    "GET",
			Path:      "/api/user/login/ccb",
			MatchType: "prefix",
		},
		{
			Method:    "GET",
			Path:      "/api/user/login",
			MatchType: "exact",
		},
		{
			Method:    "GET",
			Path:      "/api/user/login/c",
			MatchType: "prefix",
		},
	}

	router := NewRouter(routes)

	req, _ := http.NewRequest("GET", "/api/user/login", nil)

	route, _ := router.MatchRoute(req)

	fmt.Println(route)
}
