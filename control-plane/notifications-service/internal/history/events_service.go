/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history

type EventsRepository interface {
	GetEvents(ef EventFilter) ([]*Event, error)
	DeleteEvent(eventId string) error
	DeleteEvents(ef EventFilter) error
	MarkRead(eventId string) error
	MarkUnread(eventId string) error
	Cleanup(cf CleanupFilter) error
}

type EventsHistoryService struct {
	repository EventsRepository
}

func NewEventsHistoryService(repository EventsRepository) *EventsHistoryService {
	return &EventsHistoryService{repository: repository}
}

func (s *EventsHistoryService) GetEvents(ef EventFilter) ([]*Event, error) {
	return s.repository.GetEvents(ef)
}

func (s *EventsHistoryService) DeleteEvent(eventId string) error {
	return s.repository.DeleteEvent(eventId)
}

func (s *EventsHistoryService) DeleteEvents(ef EventFilter) error {
	return s.repository.DeleteEvents(ef)
}

func (s *EventsHistoryService) UpdateEventRead(eventId string, read bool) error {
	if read {
		return s.repository.MarkRead(eventId)
	}

	return s.repository.MarkUnread(eventId)
}
