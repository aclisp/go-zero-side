package delayq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageEncodeDecode(t *testing.T) {
	msg := Message{
		Carrier: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Action: "action1",
		Body:   []byte("body1"),
	}
	data, err := msg.Encode()
	assert.NoError(t, err)

	newMsg := Message{}
	err = newMsg.Decode(data)
	assert.NoError(t, err)

	assert.Exactly(t, msg, newMsg)
}
