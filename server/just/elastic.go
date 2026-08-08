package just

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var RecordingHub EventRecordingHub

type EventRecordingHub struct {
	elasticClient *elasticsearch.Client
}

func (hub *EventRecordingHub) getElasticHeader() map[string]any {
	msg := make(map[string]any)
	msg["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	return msg
}

func (hub *EventRecordingHub) OnPlayerAction(object any, err error) error {
	Logger.Debugf("sending player action update to elastic")
	msg := hub.getElasticHeader()
	msg["event_type"] = "player_action"
	msg["event"] = object
	if err != nil {
		msg["error"] = err.Error()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = hub.elasticClient.Index(
		Env.ElasticIndexRoot,
		bytes.NewReader(body),
		hub.elasticClient.Index.WithContext(context.Background()),
		hub.elasticClient.Index.WithRefresh("true"),
	)

	Logger.Debugf("game state sent to elastic")
	return err
}

func (hub *EventRecordingHub) OnGameUpdate(object any) error {
	Logger.Debugf("sending game state update to elastic")
	msg := hub.getElasticHeader()
	msg["event_type"] = "game_state"
	msg["event"] = object

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = hub.elasticClient.Index(
		Env.ElasticIndexRoot,
		bytes.NewReader(body),
		hub.elasticClient.Index.WithContext(context.Background()),
		hub.elasticClient.Index.WithRefresh("true"),
	)

	Logger.Debugf("game state sent to elastic")
	return err
}
