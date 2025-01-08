package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarRepository interface {
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	CreateEvent(token *oauth2.Token, event *calendar.Event) (*calendar.Event, error)
}

type GoogleCalendarRepository struct {
	oauthConfig *oauth2.Config
}

func NewGoogleCalendarRepository(config *oauth2.Config) *GoogleCalendarRepository {
	return &GoogleCalendarRepository{oauthConfig: config}
}

func (repo *GoogleCalendarRepository) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := repo.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, errors.New("failed to exchange authorization code for token")
	}
	return token, nil
}

func (repo *GoogleCalendarRepository) CreateEvent(token *oauth2.Token, event *calendar.Event) (*calendar.Event, error) {
	client := repo.oauthConfig.Client(context.Background(), token)
	calendarService, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create calendar service: %w", err)
	}

	createdEvent, err := calendarService.Events.Insert("primary", event).Do()
	if err != nil {
		log.Printf("Google Calendar API error details: %+v", err)
		return nil, fmt.Errorf("unable to create calendar event: %w", err)
	}

	return createdEvent, nil
}
