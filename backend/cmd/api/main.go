package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	coursehttp "courses/internal/course/adapters/http"
	coursepostgres "courses/internal/course/adapters/postgres"
	courseapplication "courses/internal/course/application"
	discussionhttp "courses/internal/discussion/adapters/http"
	discussionpostgres "courses/internal/discussion/adapters/postgres"
	discussionapplication "courses/internal/discussion/application"
	passwords "courses/internal/identity/adapters/argon2id"
	identityhttp "courses/internal/identity/adapters/http"
	identitypostgres "courses/internal/identity/adapters/postgres"
	identityapplication "courses/internal/identity/application"
	mediaasynq "courses/internal/media/adapters/asynq"
	mediahttp "courses/internal/media/adapters/http"
	mediapostgres "courses/internal/media/adapters/postgres"
	mediaapplication "courses/internal/media/application"
	playbackhttp "courses/internal/playback/adapters/http"
	playbackpostgres "courses/internal/playback/adapters/postgres"
	playbackredis "courses/internal/playback/adapters/redis"
	playbackapplication "courses/internal/playback/application"
	transcriptasynq "courses/internal/transcript/adapters/asynq"
	transcripthttp "courses/internal/transcript/adapters/http"
	transcriptpostgres "courses/internal/transcript/adapters/postgres"
	transcriptapplication "courses/internal/transcript/application"

	"courses/platform/httpserver"
	"courses/platform/observability"
	"courses/platform/runtime"
)

func main() {
	logger := observability.NewLogger("api")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	identityConfig, err := identityhttp.ConfigFromEnvironment(os.Getenv)
	if err != nil {
		logger.Error("startup failed", "reason", err.Error())
		os.Exit(1)
	}
	playbackConfig, err := playbackhttp.ConfigFromEnvironment(os.Getenv)
	if err != nil {
		logger.Error("startup failed", "reason", err.Error())
		os.Exit(1)
	}
	dependencies, err := runtime.Open(ctx, "api", logger)
	if err != nil {
		logger.Error("startup failed", "reason", err.Error())
		os.Exit(1)
	}
	err = func() error {
		defer dependencies.Close()
		courseRepository := coursepostgres.NewRepository(dependencies.Database)
		catalog := courseapplication.NewCatalog(courseRepository)
		learning := courseapplication.NewLearning(courseRepository)
		authoring := courseapplication.NewAuthoring(courseRepository, time.Now)
		courseHandler := coursehttp.NewHandler(catalog, authoring)
		learningHandler := coursehttp.NewLearningHandler(learning)
		transcriptReading := transcriptapplication.NewReading(transcriptpostgres.NewRepository(dependencies.Database))
		transcriptHandler := transcripthttp.NewHandler(transcriptReading)
		discussionRepository := discussionpostgres.NewRepository(dependencies.Database)
		discussionHandler := discussionhttp.NewHandler(discussionapplication.New(discussionRepository, time.Now))
		transcriptRepository := transcriptpostgres.NewRepository(dependencies.Database)
		transcriptAdminHandler := transcripthttp.NewAdminHandler(
			transcriptapplication.NewAuthoring(transcriptRepository),
			transcriptapplication.NewLibrary(transcriptRepository),
		)
		recognitionQueue := transcriptasynq.NewQueue(dependencies.Redis)
		drafting := transcriptapplication.NewDrafting(transcriptRepository, transcriptRepository, transcriptRepository, recognitionQueue, time.Now, logger)
		draftHandler := transcripthttp.NewDraftHandler(drafting)
		identityRepository := identitypostgres.NewRepository(dependencies.Database, passwords.NewVerifier())
		identity := identityapplication.NewIdentity(identityRepository, identityRepository, time.Now)
		administration := identityapplication.NewAdministration(identityRepository, time.Now)
		identityHandler := identityhttp.NewHandler(identity, identityRepository, administration, identityConfig)
		mediaRepository := mediapostgres.NewRepository(dependencies.Database)
		mediaQueue := mediaasynq.NewQueue(dependencies.Redis)
		mediaTickets := mediaapplication.NewTickets(mediaRepository, mediaQueue, time.Now)
		mediaHandler := mediahttp.NewHandler(mediaTickets)
		playbackOwnership := playbackredis.NewOwnershipReader(dependencies.Redis)
		playbackOwnershipWriter := playbackredis.NewOwnershipWriter(dependencies.Redis)
		playbackAccess := playbackpostgres.NewVideoAccessReader(dependencies.Database, dependencies.Metrics.PlaybackDatabaseChecks)
		playbackLessonReader := playbackpostgres.NewLessonPlaybackReader(dependencies.Database)
		playbackProgress := playbackpostgres.NewProgressWriter(dependencies.Database)
		playbackSessions := playbackapplication.NewSessions(
			playbackLessonReader,
			playbackOwnershipWriter,
			playbackProgress,
			time.Duration(playbackConfig.OwnershipTTLSeconds())*time.Second,
			time.Now,
		)
		playbackAuthorizer := playbackapplication.NewAuthorizer(playbackOwnership, playbackAccess)
		playbackHandler := playbackhttp.NewHandler(playbackAuthorizer, playbackSessions, playbackConfig, func(result string) {
			dependencies.Metrics.PlaybackAV1Probe.WithLabelValues(result).Inc()
		})
		handler := httpserver.NewHandler(httpserver.Options{
			Service: "api", Ready: dependencies.Ready, Logger: logger, Metrics: dependencies.Metrics,
			Tracer: dependencies.Tracing.Tracer("http"),
			Register: func(mux *http.ServeMux) {
				courseHandler.Register(mux)
				identityHandler.Register(mux)
				learningHandler.Register(mux, identityHandler.Authenticate)
				transcriptHandler.Register(mux, identityHandler.Authenticate)
				discussionHandler.Register(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				transcriptAdminHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				draftHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				courseHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				mediaHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				playbackHandler.Register(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
			},
		})
		logger.Info("server starting", "address", runtime.Address(":8080"))
		return httpserver.Run(ctx, runtime.Address(":8080"), handler)
	}()
	if err != nil {
		logger.Error("server stopped with an error")
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}
