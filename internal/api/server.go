package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/forgeflow/forgeflow/internal/analytics"
	"github.com/forgeflow/forgeflow/internal/auth"
	"github.com/forgeflow/forgeflow/internal/projects"
	"github.com/forgeflow/forgeflow/internal/repositories"
	"github.com/forgeflow/forgeflow/internal/tasks"
	"github.com/forgeflow/forgeflow/internal/workflows"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Health interface{Ping(context.Context)error}
type Dependencies struct{Auth *auth.Service;Projects *projects.Service;Tasks *tasks.Service;Workflows *workflows.Service;Repositories *repositories.Service;Insights *analytics.Service;Database Health;Queue Health;Logger *slog.Logger;AllowedOrigins []string}

func NewServer(dependencies Dependencies)*http.Server{
	metrics:=NewMetrics();router:=chi.NewRouter();router.Use(requestContext);router.Use(func(next http.Handler)http.Handler{return recoverer(dependencies.Logger,next)});router.Use(func(next http.Handler)http.Handler{return accessLog(dependencies.Logger,metrics,next)});router.Use(cors.Handler(cors.Options{AllowedOrigins:dependencies.AllowedOrigins,AllowedMethods:[]string{"GET","POST","PATCH","DELETE","OPTIONS"},AllowedHeaders:[]string{"Accept","Authorization","Content-Type","X-Request-ID"},MaxAge:300}))
	router.Get("/health",func(w http.ResponseWriter,_ *http.Request){writeJSON(w,http.StatusOK,map[string]string{"status":"ok"})});router.Get("/ready",readiness(dependencies.Database,dependencies.Queue));router.Handle("/metrics",metrics)
	authRoutes:=authHandler{service:dependencies.Auth};projectsRoutes:=projectHandler{projects:dependencies.Projects,tasks:dependencies.Tasks};workflowRoutes:=workflowHandler{service:dependencies.Workflows}
	repositoryRoutes:=repositoryHandler{service:dependencies.Repositories};insightRoutes:=insightHandler{service:dependencies.Insights}
	router.Route("/api/v1",func(api chi.Router){api.Post("/auth/register",authRoutes.register);api.Post("/auth/login",authRoutes.login);api.Group(func(protected chi.Router){protected.Use(func(next http.Handler)http.Handler{return authenticate(dependencies.Auth,next)});protected.Post("/auth/logout",authRoutes.logout);protected.Get("/auth/me",authRoutes.me);protected.Get("/projects",projectsRoutes.list);protected.Post("/projects",projectsRoutes.create);protected.Get("/projects/{projectID}",projectsRoutes.get);protected.Get("/projects/{projectID}/tasks",projectsRoutes.listTasks);protected.Post("/projects/{projectID}/tasks",projectsRoutes.createTask);protected.Get("/projects/{projectID}/repositories",repositoryRoutes.list);protected.Post("/projects/{projectID}/repositories",repositoryRoutes.create);protected.Post("/projects/{projectID}/workflows",workflowRoutes.create);protected.Post("/workflows/{workflowID}/runs",workflowRoutes.run);protected.Get("/runs/{runID}",workflowRoutes.getRun);protected.Get("/runs/{runID}/logs",workflowRoutes.logs);protected.Get("/analytics",insightRoutes.analytics);protected.Get("/audit",insightRoutes.audit)})})
	return &http.Server{Addr:":8080",Handler:router,ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:2*time.Minute}
}

func readiness(database,queue Health)http.HandlerFunc{return func(w http.ResponseWriter,r *http.Request){ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel();checks:=map[string]string{"database":"ok","queue":"ok"};status:=http.StatusOK;if database==nil||database.Ping(ctx)!=nil{checks["database"]="unavailable";status=http.StatusServiceUnavailable};if queue==nil||queue.Ping(ctx)!=nil{checks["queue"]="unavailable";status=http.StatusServiceUnavailable};writeJSON(w,status,map[string]any{"status":map[bool]string{true:"ok",false:"degraded"}[status==http.StatusOK],"checks":checks})}
