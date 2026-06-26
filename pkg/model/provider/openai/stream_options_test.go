package openai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/docker/docker-agent/pkg/chat"
	"github.com/docker/docker-agent/pkg/config/latest"
	"github.com/docker/docker-agent/pkg/environment"
)

// captureChatCompletionBody runs a Chat Completions stream against a fake
// server and returns the raw request body it received. trackUsage is wired
// to the model's track_usage setting.
func captureChatCompletionBody(t *testing.T, trackUsage *bool) []byte {
	t.Helper()

	var (
		receivedBody []byte
		mu           sync.Mutex
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedBody = body
		mu.Unlock()
		writeSSEResponse(w)
	}))
	defer server.Close()

	cfg := &latest.ModelConfig{
		Provider:   "custom",
		Model:      "test",
		BaseURL:    server.URL,
		TokenKey:   "MY_TOKEN",
		TrackUsage: trackUsage,
		ProviderOpts: map[string]any{
			"api_type": "openai_chatcompletions",
		},
	}

	env := environment.NewMapEnvProvider(map[string]string{
		"MY_TOKEN": "secret",
	})

	client, err := NewClient(t.Context(), cfg, env)
	require.NoError(t, err)

	stream, err := client.CreateChatCompletionStream(
		t.Context(),
		[]chat.Message{{Role: chat.MessageRoleUser, Content: "hi"}},
		nil,
	)
	require.NoError(t, err)
	defer stream.Close()

	for {
		if _, err := stream.Recv(); err != nil {
			break
		}
	}

	mu.Lock()
	defer mu.Unlock()
	return receivedBody
}

// TestChatCompletions_StreamOptionsOmittedWhenUsageDisabled verifies that
// setting track_usage:false drops the stream_options field from the request
// entirely, rather than sending stream_options:{include_usage:false}. Some
// OpenAI-compatible servers reject any stream_options object.
func TestChatCompletions_StreamOptionsOmittedWhenUsageDisabled(t *testing.T) {
	t.Parallel()

	disabled := false
	body := captureChatCompletionBody(t, &disabled)

	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &payload))

	_, ok := payload["stream_options"]
	assert.False(t, ok, "stream_options must not be present when track_usage is false")
}

// TestChatCompletions_StreamOptionsPresentByDefault verifies that the default
// behaviour (usage tracking enabled) still sends stream_options with
// include_usage:true.
func TestChatCompletions_StreamOptionsPresentByDefault(t *testing.T) {
	t.Parallel()

	body := captureChatCompletionBody(t, nil)

	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &payload))

	raw, ok := payload["stream_options"]
	require.True(t, ok, "stream_options must be present by default")

	var so struct {
		IncludeUsage bool `json:"include_usage"`
	}
	require.NoError(t, json.Unmarshal(raw, &so))
	assert.True(t, so.IncludeUsage, "include_usage must be true by default")
}
