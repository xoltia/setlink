package controllers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/xoltia/setlink/internal/models"
	"github.com/xoltia/setlink/internal/services"
	"github.com/xoltia/setlink/web"
)

type StaticController struct {
	linkSetService *services.LinkSetService
	setTemplate    *template.Template
	staticAssets   http.FileSystem
}

func NewStaticController(linkSetService *services.LinkSetService) *StaticController {
	setTemplate := template.Must(template.ParseFS(web.Templates, "template/set.html"))
	staticAssets := http.FS(web.StaticAssets)

	return &StaticController{
		linkSetService: linkSetService,
		setTemplate:    setTemplate,
		staticAssets:   staticAssets,
	}
}

func (s *StaticController) Router() chi.Router {
	r := chi.NewRouter()
	fileServer := http.FileServer(s.staticAssets)

	// r.Get("/", s.Index)
	r.Get("/{hash_string}", s.GetLinkSet)
	r.Get("/", s.GetLinkSetFromQuery)
	r.Get("/favicon.ico", fileServer.ServeHTTP)
	r.Get("/static/*", fileServer.ServeHTTP)

	return r
}

func (s *StaticController) GetLinkSetFromQuery(w http.ResponseWriter, r *http.Request) {
	hashString := r.URL.Query().Get("hash_string")
	idString := r.URL.Query().Get("id")

	var linkSet *models.LinkSet
	var err error

	if hashString != "" {
		linkSet, err = s.linkSetService.GetByHash(r.Context(), hashString)
		if err != nil {
			if err == services.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if idString != "" {
		id, err := strconv.Atoi(idString)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		linkSet, err = s.linkSetService.GetByID(r.Context(), id)

		if err != nil {
			if err == services.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No hash_string or id provided", http.StatusBadRequest)
		return
	}

	err = s.setTemplate.Execute(w, linkSet)

	if err != nil {
		log.Println(err)
		return
	}
}

func (s *StaticController) GetLinkSet(w http.ResponseWriter, r *http.Request) {
	hashString := chi.URLParam(r, "hash_string")

	linkSet, err := s.linkSetService.GetByHash(r.Context(), hashString)

	if err != nil {
		if err == services.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Found link set: %+v", linkSet)

	err = s.setTemplate.Execute(w, linkSet)

	if err != nil {
		log.Println(err)
		return
	}
}
