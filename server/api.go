package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/seapvnk/qokl/core"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func (server *Server) setupApi(r chi.Router) error {
	apiPath := filepath.Join(server.baseDir, apiDir)
	return filepath.Walk(apiPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, _ := filepath.Rel(apiPath, path)
		parts := strings.Split(rel, string(filepath.Separator))

		method := strings.ToUpper(parts[len(parts)-1])
		method = strings.Replace(method, ".LISP", "", -1)

		route := buildRoutePath(parts[:len(parts)-1])

		handler := wrapApiHandler(path)
		handler = injectRouteVars(handler, parts[1:len(parts)-1])

		switch method {
		case "GET":
			r.Get(route, handler)
		case "POST":
			r.Post(route, handler)
		case "PUT":
			r.Put(route, handler)
		case "PATCH":
			r.Patch(route, handler)
		case "DELETE":
			r.Delete(route, handler)
		default:
			log.Printf("Unknown HTTP method %s for %s", method, path)
		}

		return nil
	})
}

func wrapApiHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// route vars
		vars := GetRouteVars(r.Context())

		// headers
		headers := map[string]string{}
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// body if it exists
		var body interface{}
		if r.Body != nil && r.ContentLength > 0 {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
		}

		input := map[string]any{
			"method":  r.Method,
			"params":  vars,
			"headers": headers,
			"body":    body,
		}

		result, err := core.ExecuteScript(path, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if result.Error == nil {
			response, _ := core.SexpToGo(result.Value)
			json.NewEncoder(w).Encode(response)
		} else {
			response := ErrorResponse{Error: result.Error.Error()}
			json.NewEncoder(w).Encode(response)
		}
	}
}

func injectRouteVars(next http.HandlerFunc, parts []string) http.HandlerFunc {
	paramKeys := []string{}
	for _, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			param := strings.TrimPrefix(p, "{")
			param = strings.TrimSuffix(param, "}")
			paramKeys = append(paramKeys, param)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := map[string]string{}
		for _, key := range paramKeys {
			vars[key] = chi.URLParam(r, key)
		}
		ctx := withRouteVars(r.Context(), vars)
		next(w, r.WithContext(ctx))
	}
}
