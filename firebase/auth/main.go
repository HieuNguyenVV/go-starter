package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"google.golang.org/api/option"
	"log"
)

func main() {
	ctx := context.Background()
	tokenID := ""
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalln(fmt.Errorf("auth error: %w", err))
	}

	token, err := client.VerifySessionCookieAndCheckRevoked(ctx, tokenID)
	if err != nil {
		log.Fatalln(err)
	}

	user, err := client.GetUser(ctx, token.UID)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(user)
}
