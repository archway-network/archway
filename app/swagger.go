package app

import (
	"fmt"
	"net/http"

	"io/fs"

	"github.com/archway-network/archway/docs"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"

	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router, swaggerEnabled bool) error {
	if !swaggerEnabled {
		return nil
	}

	// rtr.Handle("/docs/client/swagger.yaml", http.FileServer(http.FS(docs.Docs)))
	// rtr.HandleFunc("/", docs.SwaggerHandler("Archway API", "/docs/client/swagger.yaml"))

	staticSubDir, err := fs.Sub(docs.Docs, "static")
	if err != nil {
		return fmt.Errorf("failed to create filesystem: %w", err)
	}

	staticServer := http.FileServer(http.FS(staticSubDir))
	//rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))

	return nil
}
