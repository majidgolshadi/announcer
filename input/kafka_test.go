package input

import (
	"github.com/Shopify/sarama"
	"strings"
	"testing"
)

func TestGenerateGroupName(t *testing.T) {
	groupName := generateGroupName(5)

	if !strings.Contains(groupName, "AnnouncerConsumer_") {
		t.Fail()
	}

	if len(groupName) != len("AnnouncerConsumer_")+5 {
		t.Fail()
	}
}

func TestValidSaramMessageUnmarshal(t *testing.T) {
	messages := []*sarama.ConsumerMessage{
		{
			Value: []byte(`{
				"channel_id": "testChannel",
				"message": "dGVzdE1lc3NhZ2U="}`),
		},
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U="}`),
		},
	}

	for i := 0; i < len(messages); i++ {
		req := &kafkaMsg{}
		if err := saramMessageUnmarshal(messages[i], req); err != nil {
			t.Log(i, err.Error())
			t.Fail()
		}

		if string(req.template) != "testMessage" {
			t.Fail()
		}
	}
}

func TestInValidSaramMessageUnmarshal(t *testing.T) {
	messages := []*sarama.ConsumerMessage{
		{
			Value: []byte(`{
				"channel_id": "testChannel", 
				"message": ""}`),
		},
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"], 
				"message": "testMessage"}`),
		},
		{
			Value: []byte(`{
				"channel_id": "",
				"usernames": [],
				"message": "dGVzdE1lc3NhZ2U="}`),
		},
		{
			Value: []byte(`{
				"channel_id": "testChannel",
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U="}`),
		},
	}

	for i := 0; i < len(messages); i++ {
		req := &kafkaMsg{}
		if err := saramMessageUnmarshal(messages[i], req); err == nil {
			t.Log(i)
			t.Fail()
		}
	}
}
