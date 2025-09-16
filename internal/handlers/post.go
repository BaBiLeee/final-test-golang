package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	repo  *db.Repo
	cache *cache.Client
	es    *es.ES
}

func New(repo *db.Repo, cache *cache.Client, es *es.ES) *Handler {
	return &Handler{repo: repo, cache: cache, es: es}
}

// CreatePost: create in transaction, index to ES
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var payload db.Post
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	created, err := h.repo.CreatePostWithLog(ctx, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Index to ES (best-effort? but spec yêu cầu đồng bộ -> handle error)
	if err := h.es.IndexPost(ctx, created); err != nil {
		// nếu ES indexing thất bại, bạn có thể:
		// - rollback toàn bộ transaction (nếu bắt buộc), hoặc
		// - ghi log và retry async. (đề bài yêu cầu "đồng bộ", nên rollback nếu cần)
		// Ở đây ta sẽ trả lỗi (vì đề bài yêu cầu đồng bộ).
		http.Error(w, "failed to index in elasticsearch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

// GetPost: cache-aside + bonus related
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	ctx := r.Context()

	// check cache
	if cached, err := h.cache.GetPost(ctx, id); err == nil && cached != nil {
		// Optionally fetch related from ES asynchronously or synchronously
		related, _ := h.es.RelatedByTags(ctx, cached.Tags, id, 5)
		resp := map[string]interface{}{"post": cached, "related": related}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// cache miss
	p, err := h.repo.GetPostByID(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.NotFound(w, r)
		return
	}

	// set cache TTL 5m
	_ = h.cache.SetPost(ctx, p, time.Minute*5)

	related, _ := h.es.RelatedByTags(ctx, p.Tags, id, 5)
	resp := map[string]interface{}{"post": p, "related": related}
	json.NewEncoder(w).Encode(resp)
}

// UpdatePost: update DB, invalidate cache, reindex ES
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	var payload db.Post
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := h.repo.UpdatePost(ctx, id, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// invalidate cache
	_ = h.cache.DeletePost(ctx, id)

	// reindex updated post
	p, _ := h.repo.GetPostByID(ctx, id)
	if p != nil {
		_ = h.es.IndexPost(ctx, p)
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchByTag
func (h *Handler) SearchByTag(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "tag required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	posts, err := h.repo.SearchByTag(ctx, tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)
}

// SearchES
func (h *Handler) SearchES(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q required", http.StatusBadRequest)
		return
	}
	res, err := h.es.Search(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
