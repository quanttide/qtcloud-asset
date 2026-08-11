package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/quanttide/qtcloud-asset-provider-example/internal/model"
	"github.com/quanttide/qtcloud-asset-provider-example/internal/store"
)

type ResourceHandler struct {
	store store.Store
}

func NewResourceHandler(st store.Store) *ResourceHandler {
	return &ResourceHandler{store: st}
}

// --- QtCloud Resources ---

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	data, err := h.store.List("qtcloud/resources")
	if err != nil {
		slog.Error("list resources", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to list resources", http.StatusInternalServerError)
		return
	}
	var items []model.QtCloudResource
	if err := json.Unmarshal(data, &items); err != nil {
		slog.Error("parse resources", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to parse resources", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, items, http.StatusOK)
}

func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var item model.QtCloudResource
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		WriteError(w, "INVALID_INPUT", "invalid request body", http.StatusBadRequest)
		return
	}
	if item.Name == "" {
		WriteError(w, "VALIDATION_ERROR", "name is required", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		slog.Error("encode resource", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to encode data", http.StatusInternalServerError)
		return
	}

	id, err := h.store.Create("qtcloud/resources", data)
	if err != nil {
		slog.Error("create resource", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to create resource", http.StatusInternalServerError)
		return
	}

	item.ID = id
	data, err = json.Marshal(item)
	if err != nil {
		slog.Error("encode resource with id", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to encode data", http.StatusInternalServerError)
		return
	}
	if err := h.store.Update("qtcloud/resources", id, data); err != nil {
		slog.Error("persist resource id", "error", err)
	}

	WriteJSON(w, item, http.StatusCreated)
}

func (h *ResourceHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, err := h.store.Get("qtcloud/resources", id)
	if err != nil {
		WriteError(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
		return
	}
	var item model.QtCloudResource
	if err := json.Unmarshal(data, &item); err != nil {
		slog.Error("parse resource", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to parse resource", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, item, http.StatusOK)
}

func (h *ResourceHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var item model.QtCloudResource
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		WriteError(w, "INVALID_INPUT", "invalid request body", http.StatusBadRequest)
		return
	}
	item.ID = id

	data, err := json.Marshal(item)
	if err != nil {
		slog.Error("encode resource", "error", err)
		WriteError(w, "INTERNAL_ERROR", "failed to encode data", http.StatusInternalServerError)
		return
	}
	if err := h.store.Update("qtcloud/resources", id, data); err != nil {
		WriteError(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
		return
	}
	WriteJSON(w, item, http.StatusOK)
}

func (h *ResourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete("qtcloud/resources", id); err != nil {
		WriteError(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
