package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouterGroup string

type MethodName string

const (
	Any    MethodName = "Any"
	GET    MethodName = "GET"
	POST   MethodName = "POST"
	DELETE MethodName = "DELETE"
	PATCH  MethodName = "PATCH"
	PUT    MethodName = "PUT"
)

var GroupRoter map[RouterGroup][]EndPointGroup

type HandlerFunc func(c *gin.Context) any

type Url struct {
	Path   string
	Method MethodName
}
type EndPointGroup interface {
	Urls() []Url
	Router(string) HandlerFunc
}

func RegisterRouter(group RouterGroup, param EndPointGroup) {
	if param == nil {
		return
	}
	if GroupRoter == nil {
		GroupRoter = make(map[RouterGroup][]EndPointGroup)
	}
	_, ok := GroupRoter[group]
	if !ok {
		GroupRoter[group] = make([]EndPointGroup, 0)
	}
	GroupRoter[group] = append(GroupRoter[group], param)
}

func Package(g *gin.Engine) {
	for key, roues := range GroupRoter {
		group := g.Group(string(key))
		for idx := range roues {
			h := roues[idx]
			urls := h.Urls()
			for _, val := range urls {
				funcHandler := h.Router(val.Path)
				if funcHandler != nil {
					do := func(ctx *gin.Context) {
						data := funcHandler(ctx)
						ctx.SecureJSON(http.StatusOK, data)
					}
					switch val.Method {
					case Any:
						group.Any(val.Path, do)
					case GET:
						group.GET(val.Path, do)
					case POST:
						group.POST(val.Path, do)
					case DELETE:
						group.DELETE(val.Path, do)
					case PATCH:
						group.PATCH(val.Path, do)
					case PUT:
						group.PUT(val.Path, do)
					}
				}
			}
		}
	}
}
