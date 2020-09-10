package main

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/vishal1132/cafebucks/eventbus"
)

// eventHandler is the handler for events
func (s *server) eventHandler(msg kafka.Message, produce bool) {
	switch string(msg.Key) {
	case string(eventbus.OrderAccept), string(eventbus.OrderReceived):
		var event eventbus.EventC
		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			s.Logger.
				Error().
				Err(err).
				Msg("error unmarshaling order accept event")
			return
		}
		var order eventbus.Order
		order = event.Order
		orderMap[event.Order.OrderID] = order
	case string(eventbus.OrderProcessed):
		var event eventbus.EventC
		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			s.Logger.
				Error().
				Err(err).
				Msg("error unmarshaling order accept event")
			return
		}
		event.Event = eventbus.OrderDelivered
		event.Order.Status = eventbus.OrderDelivered
		var order eventbus.Order
		order = event.Order
		b, err := json.Marshal(event)
		if err != nil {
			s.Logger.
				Error().
				Err(err).
				Msg("error marshaling order accept event")
			return
		}
		if produce {
			err = s.EventBus.Publish(context.Background(), eventbus.OrderDelivered, b)
		}
		if err != nil {
			s.Logger.Error().Err(err).Msg("error pushing event to kafka")
		}
		orderMap[event.Order.OrderID] = order
	}
}
