package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	datastar "github.com/starfederation/datastar/sdk/go"

	"stackfoundry.co.uk/components"
)

//go:embed public/*
var embeddedFiles embed.FS

func main() {
	mux := http.NewServeMux()

	// 1. Static Files (CSS)
	publicFS, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/css/", http.FileServer(http.FS(publicFS)))

	// 2. Home Page
	mux.Handle("/", templ.Handler(components.Home()))

	// 3. Contact API (The only interactivity)
	mux.HandleFunc("/api/contact", func(w http.ResponseWriter, r *http.Request) {
		// In a real app, you would parse the form and email it here.
		// For now, we just simulate a delay and swap the form with the success message.

		time.Sleep(500 * time.Millisecond) // Simulate network

		sse := datastar.NewSSE(w, r)

		// Swap the form <div id="contact_target"> with the Success Message
		if err := sse.MergeFragmentTempl(components.ContactSuccess()); err != nil {
			sse.ConsoleError(err)
		}
	})

	// 4. Start Server (Lambda or Local)
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("☁️  Running in AWS Lambda Mode")
		adapter := httpadapter.New(mux)

		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.ProxyWithContext(ctx, req)
		})
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		fmt.Printf("⚡ StackFoundry running on http://localhost:%s\n", port)
		http.ListenAndServe(":"+port, mux)
	}
}
