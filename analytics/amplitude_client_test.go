package analytics

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAmplitudeServerURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		useBatch bool
		want     string
	}{
		{"host only, HTTP API", "https://api.eu.amplitude.com", false, "https://api.eu.amplitude.com/2/httpapi"},
		{"host only, batch", "https://api.eu.amplitude.com", true, "https://api.eu.amplitude.com/batch"},
		{"trailing slash, HTTP API", "https://api.eu.amplitude.com/", false, "https://api.eu.amplitude.com/2/httpapi"},
		{"trailing slash, batch", "https://api.eu.amplitude.com/", true, "https://api.eu.amplitude.com/batch"},
		{"host and port", "http://127.0.0.1:8080", false, "http://127.0.0.1:8080/2/httpapi"},
		{"full HTTP API URL kept", "https://api.eu.amplitude.com/2/httpapi", false, "https://api.eu.amplitude.com/2/httpapi"},
		{"full batch URL kept", "https://api2.amplitude.com/batch", true, "https://api2.amplitude.com/batch"},
		{"relay path kept", "https://relay.example.com/amplitude/events", false, "https://relay.example.com/amplitude/events"},
		{"query kept", "https://relay.example.com?team=1", true, "https://relay.example.com/batch?team=1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := amplitudeServerURL(tc.endpoint, tc.useBatch)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAmplitudeServerURLInvalid(t *testing.T) {
	for _, endpoint := range []string{
		"api.eu.amplitude.com",
		"ftp://api.eu.amplitude.com",
		"https://",
		"https://api.eu.amplitude.com/%zz",
	} {
		t.Run(endpoint, func(t *testing.T) {
			_, err := amplitudeServerURL(endpoint, false)
			assert.Error(t, err)

			_, err = NewClient(amplitudeTestConfig(endpoint, false))
			assert.Error(t, err, "NewClient must refuse an invalid endpoint")
		})
	}
}

// A local HTTP sink standing in for Amplitude. It records the path and the events of each request.
type amplitudeSink struct {
	server *httptest.Server

	mu       sync.Mutex
	paths    []string
	events   []string
	received chan struct{}
}

func newAmplitudeSink(t *testing.T) *amplitudeSink {
	s := &amplitudeSink{received: make(chan struct{}, 100)}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var payload struct {
			APIKey string `json:"api_key"`
			Events []struct {
				EventType string `json:"event_type"`
			} `json:"events"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		s.paths = append(s.paths, r.URL.Path)
		for _, e := range payload.Events {
			s.events = append(s.events, e.EventType)
		}
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"events_ingested":1,"payload_size_bytes":1,"server_upload_time":1}`))
		s.received <- struct{}{}
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *amplitudeSink) waitForRequest(t *testing.T) {
	select {
	case <-s.received:
	case <-time.After(10 * time.Second):
		t.Fatal("the sink received no request")
	}
}

func amplitudeTestConfig(endpoint string, useBatch bool) Config {
	return Config{
		IsAmplitudeEnabled: true,
		AmplitudeConfig: AmplitudeConfig{
			AmplitudeAPIKey:   "test-key",
			AmplitudeEndpoint: endpoint,
			IsBatchingEnabled: useBatch,
			FlushQueueSize:    1,
		},
		App: AppInfo{Name: "amplitude-client-test", Version: "1.0.0"},
	}
}

// Events reach the configured endpoint, on the API path of the selected mode, with or without batching.
func TestAmplitudeEndpointReceivesEvents(t *testing.T) {
	for _, tc := range []struct {
		name     string
		useBatch bool
		path     string
	}{
		{"HTTP API", false, "/2/httpapi"},
		{"batch", true, "/batch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sink := newAmplitudeSink(t)

			client, err := NewClient(amplitudeTestConfig(sink.server.URL, tc.useBatch))
			require.NoError(t, err)

			client.Track("user-1", "Endpoint Test", map[string]any{"someKey": "value"})
			sink.waitForRequest(t)
			require.NoError(t, client.Close())

			sink.mu.Lock()
			defer sink.mu.Unlock()
			assert.Equal(t, []string{tc.path}, sink.paths)
			assert.Equal(t, []string{"Insights - Agent - Endpoint Test"}, sink.events)
		})
	}
}
