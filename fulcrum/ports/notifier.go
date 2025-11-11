package ports

import "net/http"

type Notifier interface {
	Run()
	HandleWebSocket(w http.ResponseWriter, r *http.Request)
	BroadcastJson(message any) error
}
