package app

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/server/api"

	"github.com/archway-network/archway/docs"
	"github.com/archway-network/archway/pkg/openapiconsole"
)

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(apiSvr *api.Server) error {
	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.yml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(appName+" Swagger UI", "/static/openapi.yml"))
	return nil
}
