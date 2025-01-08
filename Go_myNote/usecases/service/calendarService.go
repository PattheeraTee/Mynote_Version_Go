package service

import (
	"context"
	"log"
	"miw/entities"
	"miw/usecases/repository"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type CalendarService interface {
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	CreateEvent(token *oauth2.Token, event *entities.EventGoogle) (*calendar.Event, error)
}

type DefaultCalendarService struct {
	repo repository.CalendarRepository
}

func NewCalendarService(repo repository.CalendarRepository) *DefaultCalendarService {
	return &DefaultCalendarService{repo: repo}
}

func (s *DefaultCalendarService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.repo.ExchangeCode(ctx, code)
}

func (s *DefaultCalendarService) CreateEvent(token *oauth2.Token, event *entities.EventGoogle) (*calendar.Event, error) {
	// Map entities.Event to calendar.Event
	calEvent := &calendar.Event{
		Summary:     event.Summary,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.Start,
			TimeZone: "Asia/Bangkok",
		},
		End: &calendar.EventDateTime{
			DateTime: event.End,
			TimeZone: "Asia/Bangkok",
		},
	}

	log.Printf("Creating event from form data: %+v\n", calEvent)

	return s.repo.CreateEvent(token, calEvent)
}
