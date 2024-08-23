package server

import (
	apiV1 "SolProject/api/v1"
	"SolProject/docs"
	"SolProject/internal/handler"
	"SolProject/internal/middleware"
	"SolProject/pkg/jwt"
	"SolProject/pkg/log"
	"SolProject/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(
	logger *log.Logger,
	conf *viper.Viper,
	jwt *jwt.JWT,
	userHandler *handler.UserHandler,
) *http.Server {
	gin.SetMode(gin.DebugMode)
	s := http.NewServer(
		gin.Default(),
		logger,
		http.WithServerHost(conf.GetString("http.host")),
		http.WithServerPort(conf.GetInt("http.port")),
	)

	// swagger doc
	docs.SwaggerInfo.BasePath = "/v1"
	s.GET("/v1/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		//ginSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", conf.GetInt("app.http.port"))),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	))

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)
	s.GET("/", func(ctx *gin.Context) {
		logger.WithContext(ctx).Info("hello")
		apiV1.HandleSuccess(ctx, map[string]interface{}{
			":)": "Thank you for using nunu!",
		})
	})

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/login", userHandler.Login)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/user").Use(middleware.StrictAuth(jwt, logger))
		{
			strictAuthRouter.POST("/bind/evmaddress", userHandler.BindEvmAddress)
			strictAuthRouter.GET("/select", userHandler.Select)
			strictAuthRouter.POST("/Invitecode", userHandler.Invitecode)
			strictAuthRouter.POST("/Solsearch", userHandler.Solsearch)
			strictAuthRouter.POST("/horsh/transfer", userHandler.HorshTransfer)
			strictAuthRouter.POST("/usdt/transfer", userHandler.USDTTransfer)
			strictAuthRouter.POST("/claim/reward", userHandler.ClaimReward)
		}

		adminRouter := v1.Group("/admin")
		{
			adminRouter.POST("/login", userHandler.AdminLogin)
			adminRouter.POST("/search", middleware.StrictAuth(jwt, logger), userHandler.AdminSearch)
			adminRouter.GET("/allCount", middleware.StrictAuth(jwt, logger), userHandler.AdminAllCount)
			adminRouter.GET("/CreationCount", middleware.StrictAuth(jwt, logger), userHandler.CreationCount)
			adminRouter.POST("/export-record", middleware.StrictAuth(jwt, logger), userHandler.ExportRecord)
			adminRouter.POST("/export-record-team", userHandler.ExportRecordTeam)
			adminRouter.POST("export-record-regis", userHandler.ExportRecordRegis)
		}
	}

	return s
}
