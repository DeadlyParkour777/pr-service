package handler

import (
	"errors"
	"net/http"

	"github.com/DeadlyParkour777/pr-service/internal/service"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type APIErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Handler struct {
	teamService  TeamService
	userService  UserService
	prService    PullRequestService
	statsService StatsService

	validate *validator.Validate
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{
		teamService:  s.Team,
		userService:  s.User,
		prService:    s.PR,
		statsService: s.Stats,

		validate: validator.New(),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Route("/stats", func(r chi.Router) {
		r.Get("/user", h.getUserStats)
	})

	router.Route("/team", func(r chi.Router) {
		r.Post("/add", h.createTeam)
		r.Get("/get", h.getTeam)
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.setUserIsActive)
		r.Get("/getReview", h.getReviewsForUser)
	})

	router.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.createPullRequest)
		r.Post("/merge", h.mergePullRequest)
		r.Post("/reassign", h.reassignReviewer)
	})

	return router
}

func (h *Handler) writeBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	resp := APIErrorResponse{}
	resp.Error.Code = "BAD_REQUEST"
	resp.Error.Message = message

	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, resp)
}

func (h *Handler) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	resp := APIErrorResponse{}
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, service.ErrNotFound):
		status = http.StatusNotFound
		resp.Error.Code = "NOT_FOUND"
		resp.Error.Message = "resource not found"

	case errors.Is(err, service.ErrTeamExists):
		status = http.StatusBadRequest
		resp.Error.Code = "TEAM_EXISTS"
		resp.Error.Message = "team name already exists"

	case errors.Is(err, service.ErrPRExists):
		status = http.StatusConflict
		resp.Error.Code = "PR_EXISTS"
		resp.Error.Message = "PR id already exists"

	case errors.Is(err, service.ErrPRMerged):
		status = http.StatusConflict
		resp.Error.Code = "PR_MERGED"
		resp.Error.Message = "cannot reassign on merged PR"

	case errors.Is(err, service.ErrNotAssigned):
		status = http.StatusConflict
		resp.Error.Code = "NOT_ASSIGNED"
		resp.Error.Message = "reviewer is not assigned to this PR"

	case errors.Is(err, service.ErrNoCandidates):
		status = http.StatusConflict
		resp.Error.Code = "NO_CANDIDATE"
		resp.Error.Message = "no active replacement candidate in team"

	default:
		resp.Error.Code = "INTERNAL_ERROR"
		resp.Error.Message = "internal server error"
	}

	render.Status(r, status)
	render.JSON(w, r, resp)
}
