package app

import (
	"net/http"

	"github.com/archway-network/archway/docs"
	"github.com/archway-network/archway/pkg/openapiconsole"
	"github.com/cosmos/cosmos-sdk/server/api"

	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(apiSvr *api.Server) error {
	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.yml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(appName, "/static/openapi.yml"))
	return nil
}
