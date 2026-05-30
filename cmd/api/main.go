// @title           Radius Backend API
// @version         1.0
// @description     Radius monolith API — auth and users.
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT token. Format: Bearer {token}
package main

import (
	"fmt"
	"os"

	"github.com/radius/radius-backend/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
