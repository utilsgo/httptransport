package routes

import (
	"github.com/go-courier/courier"
	"github.com/utilsgo/httptransport"
	"github.com/utilsgo/httptransport/openapi"
)

var RootRouter = courier.NewRouter(httptransport.BasePath("/demo"))

func init() {
	RootRouter.Register(openapi.OpenAPIRouter)
}
