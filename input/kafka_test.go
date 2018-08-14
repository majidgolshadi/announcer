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

func TestDefaultPersistableMessageTrue(t *testing.T) {
	messages := []*sarama.ConsumerMessage{
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U="}`),
		},
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U=",
				"persist": true}`),
		},
	}

	for i := 0; i < len(messages); i++ {
		req := &kafkaMsg{
			Persist: true,
		}
		if err := saramMessageUnmarshal(messages[i], req); err != nil {
			t.Log(i, err.Error())
			t.Fail()
		}

		if req.Persist != true {
			t.Fail()
		}
	}
}

func TestPersistableMessageFalse(t *testing.T) {
	messages := []*sarama.ConsumerMessage{
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U=",
				"persist": false}`),
		},
		{
			Value: []byte(`{ "usernames":["ub1j8ribz"], "message":"PG1lc3NhZ2UgeG1sOmxhbmc9J2VuJyB0bz0nJXMnIGZyb209J3V5M3duMm1odUBzL2Fubm91bmNlcicgdHlwZT0nY2hhdCcgaWQ9JzE1MjU4OTAxMzMwODVkMDE0ZjM1Y0F4TSc+PGJvZHk+IDwvYm9keT48Ym9keSB4bWw6bGFuZz0nRURJVEVEX1RFWFQnPtin2YLYpyDbjNmHINmG2qnYqtmHINin24wg24zaqduMINin2LIg2K/ZiNiz2KrYp9mGINiv2KfZhti02q/Yp9mHINqv2YHYqiDaqdmHINio2Ycg2LTZhdinINmF2YbYqtmC2YQg2qnZhtmFPC9ib2R5Pjxib2R5IHhtbDpsYW5nPSdNQUpPUl9UWVBFJz5DT05UUk9MX01FU1NBR0U8L2JvZHk+PGJvZHkgeG1sOmxhbmc9J01FU1NBR0VfSUQnPjE1MjU4OTAxMjgxOTVkMDE0ZjM1T1RIZTwvYm9keT48Ym9keSB4bWw6bGFuZz0nTUlOT1JfVFlQRSc+RURJVF9NRVNTQUdFPC9ib2R5Pjxib2R5IHhtbDpsYW5nPSdTRU5EX1RJTUVfSU5fR01UJz4xNTI1ODkwMTQwMjM3PC9ib2R5Pjxib2R5IHhtbDpsYW5nPSdVUERBVEVfVElNRVNUQU1QJz4xNTI1ODkwMTM5OTkxPC9ib2R5PjwvbWVzc2FnZT4=", "persist":false}`),
		},
	}

	for i := 0; i < len(messages); i++ {
		req := &kafkaMsg{
			Persist: true,
		}
		if err := saramMessageUnmarshal(messages[i], req); err != nil {
			t.Log(i, err.Error())
			t.Fail()
		}

		if req.Persist != false {
			t.Fail()
		}
	}
}

func TestPersistableMessageFalseTrueFalse(t *testing.T) {
	messages := []*sarama.ConsumerMessage{
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U=",
				"persist": false}`),
		},
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U=",
				"persist": true}`),
		},
		{
			Value: []byte(`{
				"usernames": ["user_1", "user_2", "user_3"],
				"message": "dGVzdE1lc3NhZ2U=",
				"persist": false}`),
		},
	}

	for i := 0; i < len(messages); i++ {
		req := &kafkaMsg{
			Persist: true,
		}
		if err := saramMessageUnmarshal(messages[i], req); err != nil {
			t.Log(i, err.Error())
			t.Fail()
		}

		if i == 2 && req.Persist != false {
			t.Fail()
		}
	}
}
