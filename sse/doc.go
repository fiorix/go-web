// Server-Sent events (SSE)
// http://dev.w3.org/html5/eventsource/
//
// Usage example:
//
//	func SSEHandler(w http.ResponseWriter, req *http.Request) {
//	        conn, err := sse.ServeEvents(w)
//	        if err != nil {
//	                http.Error(w, err.Error(), http.StatusInternalServerError)
//	                return
//	        }
//	        defer conn.Close()
//	        for i := 0; i < 10; i++ {
//	                sse.SendEvent(conn, &sse.MessageEvent{Data: "Hello, world"})
//	                time.Sleep(1 * time.Second)
//	        }
//	}
//
// These connections are never logged by Server's Logger.
// See http.Server for details.

package sse
