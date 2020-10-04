package usersrv

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

const localfile = "/static"

// Set sets the routing for the gin engine
func SetRoute(server *Server) error {

	// Access Auth service from default app
	defaultClient, err := server.App.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	r := server.Engine
	r.Use(static.Serve("/", static.LocalFile(localfile, false)))

	apiRouter := r.Group("/api")

	apiRouter.GET("/verifyToken", func(c *gin.Context) {
		apiLogger := log.WithFields(log.Fields{
			"path": c.FullPath(),
		})

		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		idToken := authHeader[len(BEARER_SCHEMA):]

		token, err := defaultClient.VerifyIDToken(c, idToken)
		if err != nil {
			apiLogger.Infof("error verifying ID token: %v", err)
			// apiLogger.Infof("token: %v", idToken)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		apiLogger.Infof("Verified ID token: %v", token)
		c.Status(http.StatusOK)
	})

	apiRouter.GET("/users/:userID", func(c *gin.Context) {
		apiLogger := log.WithFields(log.Fields{
			"path": c.FullPath(),
		})

		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		idToken := authHeader[len(BEARER_SCHEMA):]

		_, err := defaultClient.VerifyIDToken(c, idToken)
		if err != nil {
			apiLogger.Infof("error verifying ID token: %v", err)
			// apiLogger.Infof("token: %v", idToken)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		firebaseID := c.Param("userID")

		type User struct {
			ID                    int64
			FirebaseId            string
			Email                 string
			Name                  *string
			Nickname              *string
			Bio                   *string
			State                 int
			Birthday              *time.Time
			ImageId               int64
			Gender                int
			Phone                 *string
			AddressId             int64
			Point                 int
			CreatedAt             time.Time
			UpdatedAt             time.Time
			MembershipValidBefore *time.Time
			MembershipType        int
			MembershipValidAfter  *time.Time
			CreatedByOperator     int64
		}

		db, err := NewDB()
		if err != nil {
			apiLogger.Infof("db open error: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var user User
		db.Where("firebase_id = ?", firebaseID).First(&user)
		apiLogger.Infof("firebase_id(%s):%+v", firebaseID, user)
		if user.ID == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.JSON(http.StatusOK, user)
	})

	return nil
}
