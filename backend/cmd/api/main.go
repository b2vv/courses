package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	coursehttp "courses/internal/course/adapters/http"
	coursepostgres "courses/internal/course/adapters/postgres"
	courseapplication "courses/internal/course/application"
	coursedomain "courses/internal/course/domain"
	discussionhttp "courses/internal/discussion/adapters/http"
	discussionpostgres "courses/internal/discussion/adapters/postgres"
	discussionapplication "courses/internal/discussion/application"
	passwords "courses/internal/identity/adapters/argon2id"
	identityhttp "courses/internal/identity/adapters/http"
	identitymail "courses/internal/identity/adapters/mail"
	identitypostgres "courses/internal/identity/adapters/postgres"
	identityredis "courses/internal/identity/adapters/redis"
	identityapplication "courses/internal/identity/application"
	mediaasynq "courses/internal/media/adapters/asynq"
	mediahttp "courses/internal/media/adapters/http"
	mediapostgres "courses/internal/media/adapters/postgres"
	mediaapplication "courses/internal/media/application"
	ordershttp "courses/internal/orders/adapters/http"
	orderspostgres "courses/internal/orders/adapters/postgres"
	wayforpay "courses/internal/orders/adapters/wayforpay"
	ordersapplication "courses/internal/orders/application"
	ordersdomain "courses/internal/orders/domain"
	ordersports "courses/internal/orders/ports"
	playbackhttp "courses/internal/playback/adapters/http"
	playbackpostgres "courses/internal/playback/adapters/postgres"
	playbackredis "courses/internal/playback/adapters/redis"
	playbackapplication "courses/internal/playback/application"
	transcriptasynq "courses/internal/transcript/adapters/asynq"
	transcripthttp "courses/internal/transcript/adapters/http"
	transcriptpostgres "courses/internal/transcript/adapters/postgres"
	transcriptapplication "courses/internal/transcript/application"

	"courses/platform/config"
	"courses/platform/httpserver"
	platformmail "courses/platform/mail"
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
	mailConfig, err := config.MailConfig(os.Getenv)
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
		ordersRepository := orderspostgres.NewRepository(dependencies.Database)
		wayforpayConfig, wayforpayConfigured := wayforpay.FromEnvironment(os.Getenv)
		// Keep these interface-typed and assign only when configured: a nil *wayforpay.Provider
		// placed in an interface is a non-nil interface, which would make the service and handler
		// treat an unconfigured provider as present (nil deref instead of the 503/404 contract).
		var paymentProvider ordersports.PaymentProvider
		var callbackVerifier ordersports.CallbackVerifier
		var publicBaseURL string
		if wayforpayConfigured {
			provider := wayforpay.New(wayforpayConfig)
			paymentProvider, callbackVerifier, publicBaseURL = provider, provider, wayforpayConfig.PublicBaseURL
		}
		orders := ordersapplication.New(ordersRepository, ordersRepository, ordersRepository, time.Now, newUUID, paymentProvider)
		authoring := courseapplication.NewAuthoring(courseRepository, time.Now, courseGrandfatherer{orders})
		courseHandler := coursehttp.NewHandler(catalog, authoring)
		ordersHandler := ordershttp.NewHandler(orders, callbackVerifier, publicBaseURL)
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
		loginLimiter, err := identityredis.NewRateLimiter(dependencies.Redis, 5, 15*time.Minute, logger)
		if err != nil {
			return err
		}
		pendingTwoFactor, err := identityredis.NewPendingTwoFactorStore(dependencies.Redis)
		if err != nil {
			return err
		}
		twoFactor := identityapplication.NewTwoFactor(
			identityRepository, identityRepository, pendingTwoFactor, identityRepository, time.Now,
			identityapplication.TwoFactorDependency{Passwords: identityRepository, Limiter: loginLimiter, Logger: logger},
		)
		identity := identityapplication.NewIdentity(identityRepository, identityRepository, time.Now, loginLimiter).WithTwoFactor(twoFactor)
		identityMailer, err := identitymail.New(platformmail.NewSender(mailConfig), os.Getenv("PUBLIC_ORIGIN"))
		if err != nil {
			return err
		}
		accounts := identityapplication.NewAccounts(identityRepository, identityRepository, passwords.NewVerifier(), identityRepository, identityMailer, identityRepository, time.Now, newUUID, identityapplication.AccountsDependency{Limiter: loginLimiter, Credentials: identityRepository})
		administration := identityapplication.NewAdministration(identityRepository, time.Now)
		identityHandler := identityhttp.NewHandler(identity, identityRepository, administration, identityConfig, accounts).WithTwoFactor(twoFactor)
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
				ordersHandler.Register(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				ordersHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin, identityHandler.RequireAdministrator)
				ordersHandler.RegisterTeach(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				identityHandler.Register(mux)
				learningHandler.Register(mux, identityHandler.Authenticate)
				transcriptHandler.Register(mux, identityHandler.Authenticate)
				discussionHandler.Register(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				discussionHandler.RegisterTeach(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
				discussionHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin, identityHandler.RequireAdministrator)
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

type courseGrandfatherer struct{ service *ordersapplication.Service }

func (adapter courseGrandfatherer) GrandfatherOnPriceSet(ctx context.Context, courseID string, expectedUpdatedAt time.Time, priceMinor int64, currency string, updatedAt time.Time) error {
	err := adapter.service.GrandfatherOnPriceSet(ctx, courseID, expectedUpdatedAt, priceMinor, currency, updatedAt)
	if errors.Is(err, ordersdomain.ErrCourseModified) {
		return coursedomain.ErrCourseModified
	}
	return err
}

func newUUID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(fmt.Sprintf("generate uuid: %v", err))
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}
