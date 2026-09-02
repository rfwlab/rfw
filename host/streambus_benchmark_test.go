package host

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mirkobrombin/go-warp/v2/streambus"
	"golang.org/x/net/websocket"
)

func BenchmarkStreamBusFanout(b *testing.B) {
	for _, subscribers := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("subscribers-%d", subscribers), func(b *testing.B) {
			bus := streambus.NewInMemory(streambus.Config{DefaultBuffer: 4096, MaxBuffer: 4096})
			defer func() { _ = bus.Close() }()
			topics := make([]string, subscribers)
			for i := 0; i < subscribers; i++ {
				topics[i] = fmt.Sprintf("rfw/connection/%d", i)
				subscription, err := bus.Subscribe(context.Background(), streambus.SubscribeOptions{
					Topic: topics[i], Buffer: 4096, Overflow: streambus.Block,
				})
				if err != nil {
					b.Fatal(err)
				}
				defer func() { _ = subscription.Close() }()
				go func() {
					for range subscription.Frames() {
						continue
					}
				}()
			}
			outbound := Outbound{Component: "Chart", Payload: map[string]any{"value": 42}}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, topic := range topics {
					payload, err := json.Marshal(outbound)
					if err != nil {
						b.Fatal(err)
					}
					if _, err := bus.Publish(context.Background(), streambus.Frame{
						Topic: topic, Payload: payload, Reliability: streambus.Reliable,
						Priority: streambus.PriorityInteractive,
					}); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func BenchmarkWebSocketFanout(b *testing.B) {
	for _, subscribers := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("subscribers-%d", subscribers), func(b *testing.B) {
			accepted := make(chan *websocket.Conn, subscribers)
			stop := make(chan struct{})
			server := httptest.NewServer(websocket.Handler(func(connection *websocket.Conn) {
				accepted <- connection
				<-stop
			}))
			defer server.Close()
			serverConnections := make([]*websocket.Conn, 0, subscribers)
			clientConnections := make([]*websocket.Conn, 0, subscribers)
			for i := 0; i < subscribers; i++ {
				client, err := websocket.Dial("ws"+strings.TrimPrefix(server.URL, "http"), "", server.URL)
				if err != nil {
					b.Fatal(err)
				}
				clientConnections = append(clientConnections, client)
				serverConnections = append(serverConnections, <-accepted)
				go func() {
					for {
						var payload []byte
						if websocket.Message.Receive(client, &payload) != nil {
							return
						}
					}
				}()
			}
			defer func() {
				for _, connection := range clientConnections {
					_ = connection.Close()
				}
				for _, connection := range serverConnections {
					ForgetConnection(connection)
				}
				close(stop)
			}()
			outbound := Outbound{Component: "Chart", Payload: map[string]any{"value": 42}}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, connection := range serverConnections {
					SendOutbound(connection, outbound)
				}
			}
		})
	}
}
