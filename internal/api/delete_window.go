package api

import (
	"net/http"
)

// handleDeleteWindow removes a maintenance window by label.
func handleDeleteWindow(wm windowManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		label := r.URL.Query().Get("label")
		if label == "" {
			http.Error(w, "label query param required", http.StatusBadRequest)
			return
		}

		wm.RemoveWindow(label)
		w.WriteHeader(http.StatusNoContent)
	}
}
