package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/xoltia/setlink/internal/services"
)

type APIController struct {
	linkSetService *services.LinkSetService
}

func NewAPIController(linkSetService *services.LinkSetService) *APIController {
	return &APIController{
		linkSetService: linkSetService,
	}
}

func (a *APIController) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/linksets/{hash_string}", a.GetLinkSet)
	r.Post("/linksets", a.CreateLinkSet)

	return r
}

func (a *APIController) GetLinkSet(w http.ResponseWriter, r *http.Request) {
	hashString := chi.URLParam(r, "hash_string")

	linkSet, err := a.linkSetService.GetByHash(r.Context(), hashString)

	if err != nil {
		if err == services.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(linkSet)
}

func (a *APIController) CreateLinkSet(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "No request body", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	var urlStrings []string

	err := json.NewDecoder(r.Body).Decode(&urlStrings)

	log.Printf("%+v", urlStrings)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	urls := make([]*url.URL, len(urlStrings))

	for i, urlString := range urlStrings {
		u, err := url.Parse(urlString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		urls[i] = u
	}

	log.Printf("%+v", urls)

	linkSet, err := a.linkSetService.GetOrCreateSet(r.Context(), urls)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(linkSet)
}
