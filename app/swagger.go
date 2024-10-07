package app

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/server/api"

	"github.com/archway-network/archway/app/appconst"
	"github.com/archway-network/archway/docs"
	"github.com/archway-network/archway/pkg/openapiconsole"
)

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(apiSvr *api.Server) error {
	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/swagger.min.json", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(appconst.AppName+" Swagger UI", "/static/swagger.min.json"))
	return nil
}
