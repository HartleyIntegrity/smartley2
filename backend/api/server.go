package api

import (
	"fmt"
	"net/http"

	"smartley-contracts/blockchain"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var bc *blockchain.Blockchain

func Start() {
	bc = blockchain.NewBlockchain()
	blockchain.BlockchainInstance = &blockchain.BlockchainWrapper{Blockchain: bc}

	PORT := "8080"
	fmt.Println("Starting server on port", PORT)

	// Initialize CORS middleware with default options
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})

	handler := c.Handler(routes(bc))

	logrus.Fatal(http.ListenAndServe(":"+PORT, handler))
}
