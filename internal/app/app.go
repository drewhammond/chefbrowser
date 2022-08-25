package app

import (
	"github.com/drewhammond/chefbrowser/internal/chef"
	"github.com/drewhammond/chefbrowser/internal/common/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type AppService struct {
	Log  *logging.Logger
	Chef *chef.Service
}

// let's start with a completely flat package and then refactor later
func nodesEndpoint(c *gin.Context) {

}

type getNodesResponse struct {
}

func (r *AppService) getNodes(c *gin.Context) {

	r.Log.Debug("getting all nodes from chef server")

	var nodes, err = r.Chef.GetNodes()
	if err != nil {
		r.Log.Error("failed to fetch nodes", zap.Error(err))
	}

	c.JSON(http.StatusOK, nodes)
}

func (r *AppService) getRoles(c *gin.Context) {
	r.Log.Debug("getting all roles from chef server")
	roles, err := r.Chef.GetRoles()
	if err != nil {
		r.Log.Error("failed to fetch roles from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, roles)
}

func (r *AppService) getEnvironments(c *gin.Context) {
	r.Log.Debug("getting all environments from chef server")
	environments, err := r.Chef.GetCookbooks()
	if err != nil {
		r.Log.Error("failed to fetch environments from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, environments)
}

func (r *AppService) getCookbooks(c *gin.Context) {
	r.Log.Debug("getting all cookbooks from chef server")
	cookbooks, err := r.Chef.GetCookbooks()
	if err != nil {
		r.Log.Error("failed to fetch cookbooks from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, cookbooks)
}
func (r *AppService) get(c *gin.Context) {
	r.Log.Debug("getting all cookbooks from chef server")
	cookbooks, err := r.Chef.GetCookbooks()
	if err != nil {
		r.Log.Error("failed to fetch cookbooks from server", zap.Error(err))
	}
	c.JSON(http.StatusOK, cookbooks)
}

type HealthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, &HealthResponse{Success: true, Message: "ready"})
}

type ConfigService interface {
	Get()
}

func New() {

	app := AppService{
		Log:  logging.New(),
		Chef: chef.New(),
	}

	app.Log.Info("starting chefbrowser application...")
	app.Log.Debug("here's some debug logging")

	// check health of our services
	//err := app.Chef.CheckHealth()
	//if err != nil {
	//	app.Log.Fatal("failed to initialize connection to chef")
	//}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	// todo: replace with our own logger
	router.Use(gin.Logger(), gin.Recovery())

	_ = router.SetTrustedProxies(nil)

	api := router.Group("/api")
	{
		api.GET("/nodes", app.getNodes)
		api.GET("/roles", app.getRoles)
		api.GET("/environments", app.getEnvironments)
		api.GET("/cookbooks", app.getCookbooks)
		api.GET("/health", getHealth)
	}

	err := router.Run(":8080")
	if err != nil {
		app.Log.Fatal("failed to start web server", zap.Error(err))
	}
}
