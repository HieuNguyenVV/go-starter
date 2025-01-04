package main

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"google.golang.org/api/option"
	"log"
)

func main() {
	ctx := context.Background()
	client, err := NewFcm(ctx)
	if err != nil {
		log.Fatal(err)
	}
	invalidTokens, err := client.MultiPush(ctx, Notification{
		Title:     "Hello client",
		Body:      "This is an specific notification.",
		AppID:     "abcd-1234",
		ChannelID: "1234",
		Message:   "Hello",
		Type:      "text",
		Data:      nil,
	}, "token-device")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(invalidTokens)
}

type Notification struct {
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	AppID     string                 `json:"app_id"`
	ChannelID string                 `json:"channel_id"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}
type Fcm struct {
	client *messaging.Client
}

func NewFcm(ctx context.Context) (*Fcm, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting client: %v", err)
	}
	return &Fcm{client: client}, nil
}

func (f *Fcm) Push(ctx context.Context, token string, notification Notification) error {
	msg := &messaging.Message{
		Token:   token,
		Webpush: f.buildWebMessage(ctx, notification),
		APNS:    f.buildIosMessage(ctx, notification),
		Android: f.buildAndroidMessage(ctx, token, notification),
	}
	_, err := f.client.Send(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

func (f *Fcm) MultiPush(ctx context.Context, notification Notification, tokens ...string) ([]string, error) {
	invalidTokens := make([]string, 0)
	if len(tokens) == 0 {
		return invalidTokens, nil
	}
	if len(tokens) > 500 {
		return invalidTokens, fmt.Errorf("too many tokens")
	}

	messages := make([]*messaging.Message, len(tokens))
	for i, token := range tokens {
		messages[i] = &messaging.Message{
			Token:   token,
			Webpush: f.buildWebMessage(ctx, notification),
			APNS:    f.buildIosMessage(ctx, notification),
			Android: f.buildAndroidMessage(ctx, token, notification),
		}
	}

	responses, err := f.client.SendEach(ctx, messages)
	if err != nil {
		return invalidTokens, err
	}
	for i, resp := range responses.Responses {
		if resp.Success {
			continue
		}
		if messaging.IsUnregistered(resp.Error) {
			invalidTokens = append(invalidTokens, tokens[i])
		} else {
			log.Fatalln(resp.Error)
		}
	}
	return invalidTokens, nil
}

func (f *Fcm) buildWebMessage(ctx context.Context, notification Notification) *messaging.WebpushConfig {
	message := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}{
		Title: notification.Title,
		Body:  notification.Body,
	}

	return &messaging.WebpushConfig{
		Data: map[string]string{
			"message": StructToJson(message),
			"data":    StructToJson(notification),
		},
	}
}

func (f *Fcm) buildIosMessage(ctx context.Context, notification Notification) *messaging.APNSConfig {
	header := make(map[string]string)
	header["apns-collapse-id"] = "123"
	return &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				AlertString: notification.Body,
			},
			CustomData: map[string]interface{}{
				"data": notification,
			},
		},
		Headers: header,
	}
}

func (f *Fcm) buildAndroidMessage(ctx context.Context, token string, notification Notification) *messaging.AndroidConfig {
	var message struct {
		Token        string `json:"token"`
		Notification struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"notification"`
	}
	message.Token = token
	message.Notification.Title = notification.Title
	message.Notification.Body = notification.Body

	return &messaging.AndroidConfig{
		Data: map[string]string{
			"message": StructToJson(message),
			"data":    StructToJson(notification),
		},
	}
}

func StructToJson(v interface{}) string {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(jsonData)
}
