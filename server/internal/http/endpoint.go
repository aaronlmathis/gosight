package httpserver

import (
	"fmt"
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/gorilla/mux"
)

func HandleEndpointDetail(w http.ResponseWriter, r *http.Request, cfg *config.Config, metricStore store.MetricStore, userStore userstore.UserStore, metaTracker *metastore.MetaTracker) {

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
	// Check if user has permission to view the dashboard
	user, err := userStore.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("❌ Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	// Build Template data based on endpoint_id
	data, err := templates.BuildHostDashboardData(ctx, metricStore, metaTracker, user, endpointID)
	if err != nil {
		utils.Debug("failed to build host dashboard data: %v", err)
	}
	fmt.Printf("🧠 Template Meta: %+v\n", data.Meta)
	// Set breadcrumbs and endpoint id
	data.Title = "Host: " + endpointID
	data.Labels["Breadcrumbs"] = "Endpoints / Host Overview"
	data.Labels["EndpointID"] = endpointID
	data.Labels["status"] = "unknown"

	err = templates.RenderTemplate(w, "dashboard/layout", data)
	if err != nil {
		utils.Error("❌ Template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
