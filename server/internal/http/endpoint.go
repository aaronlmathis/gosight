package httpserver

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func HandleEndpointDetail(w http.ResponseWriter, r *http.Request, cfg *config.Config, metricStore store.MetricStore, userStore userstore.UserStore) {

	vars := mux.Vars(r)
	endpointID := vars["endpoint_id"]
	ctx := r.Context()

	// Check for forbidden access first
	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if user is authenticated
	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := userStore.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("❌ Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}
	// Use VictoriaStore to fetch latest metrics for this host
	hostMetrics := []string{
		"system.host.uptime",
		"system.host.procs",
		"system.host.users_loggedin",
		"system.host.info",
	}
	metrics := map[string]float64{}
	for _, metric := range hostMetrics {
		rows, err := metricStore.QueryInstant(metric, map[string]string{"endpoint_id": endpointID})
		if err == nil && len(rows) > 0 {
			metrics[metric] = rows[0].Value
		}
	}

	data := map[string]any{
		"User":        user,
		"Breadcrumbs": "Endpoints / Host Overview",
		"EndpointID":  endpointID,
		"Metrics":     metrics,
	}

	err = templates.RenderTemplate(w, "layout", data)
	if err != nil {
		utils.Error("❌ Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
