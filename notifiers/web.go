package notifiers

import (
	"encoding/json"
	"fmt"
	"sync"

	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/websocket"
	"github.com/rbhz/web_watcher/watcher"
)

// WebNotifier Send notifications for web users
type WebNotifier struct {
	server Server
	mux    sync.Mutex
}

// Notify web users
func (n *WebNotifier) Notify(update watcher.URLUpdate) {
	data, err := json.Marshal(update.New)
	if err != nil {
		return
	}
	n.mux.Lock()
	n.server.Broadcast(data)
	n.mux.Unlock()
}

// Run starts server
func (n *WebNotifier) Run() {
	n.server.Run()
}

// NewWebNotifier initialize web notifier instance
func NewWebNotifier(cfg WebConfig, watcher watcher.Watcher) WebNotifier {
	return WebNotifier{
		server: NewServer(watcher, cfg.Port),
	}
}

// Server with rest api & static
type Server struct {
	watcher  watcher.Watcher
	port     int
	sockets  map[string]*websocket.Conn
	upgrader websocket.Upgrader
}

// Run web server
func (s *Server) Run() {
	http.HandleFunc("/", s.index)
	http.HandleFunc("/api/list", s.list)
	http.HandleFunc("/ws", s.upgrade)
	log.Info().Str("address", fmt.Sprintf("http://0.0.0.0:%v", s.port)).Msg("Starting web server")
	err := http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run web server")
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(indexPageTemplate))
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(s.watcher.GetUrls())
	w.Write(data)
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	remoteAddr := c.RemoteAddr().String()
	defer delete(s.sockets, remoteAddr)
	s.sockets[remoteAddr] = c
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// Broadcast message to all active sockets
func (s *Server) Broadcast(message []byte) {
	wg := sync.WaitGroup{}
	for _, conn := range s.sockets {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Wait()
			conn.WriteMessage(websocket.TextMessage, message)
		}(conn)
	}
	wg.Wait()
}

// NewServer returns new web server
func NewServer(w watcher.Watcher, port int) Server {
	return Server{
		watcher:  w,
		port:     port,
		sockets:  make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{},
	}
}

const indexPageTemplate = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>HTTP checker</title>
    <style>
    .dot {
        height: 25px;
        width: 25px;
        background-color: #bbb;
        border-radius: 50%;
        display: inline-block;
    }
    </style>
  </head>
  <body>
      <div class="container">
          <div class="row">
                <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">#</th>
                            <th scope="col">Url</th>
                            <th scope="col">Last change</th>
                            <th scope="col">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="d-none empty_row">
                            <th scope="row" class="num"></th>
                            <td class="url">
                                <a href=""></a>
                            </td>
                            <td class="change"></td>
                            <td class="status">
                                <span class="dot"></span>
                            </td>

                        </tr>
                    </tbody>
                </table>
          </div>
      </div>

    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.js" integrity="sha256-WpOohJOqMqqyKL9FccASB9O0KwACQJpFTUBLTYOVvVU=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
    <script>
        $(document).ready(function() {
            $.get('/api/list', function(data) {
                let tbody = $('table tbody');
                data = JSON.parse(data);
                for (var idx = 0; idx < data.length; idx++) {
                    tbody.append($('.empty_row').clone());
                    row = $($('.empty_row')[1]);
                    row.attr('class', '');
                    row.find('.num').text(1 + idx);
                    row.find('.url a').text(data[idx].url).attr('href', data[idx].url);
                    let changed = new Date(data[idx].last_change);
                    row.find('.change').text(changed.toLocaleString());
                    if (data[idx]['error'] == "" && data[idx]['status'] == 200) {
                        row.find('.status .dot').css('background-color', 'green');
                    } else {
                        row.find('.status .dot').css('background-color', 'red');
                        let text = "";
                        if (data[idx].error != "") {
                            text = data[idx].error
                        } else {
                            text = 'Status ' + data[idx].status;
                        }
                        row.find('.status .dot').popover({
                            content: text,
                            trigger: 'hover',
                        });
                    }
                }

                ws = new WebSocket('ws://'+window.location.host+'/ws');
                ws.onopen = function(evt) {
                    console.log("ws OPEN");
                }
                ws.onclose = function(evt) {
                    console.log("ws CLOSE");
                }
                ws.onmessage = function(evt) {
                    data = JSON.parse(evt.data);
                    let row = $('a[href="' +data.url+'"]').parents('tr');
                    changed = new Date(data.last_change);
                    row.find('.change').text(changed.toLocaleString());
                    if (data['error'] === "" && data["status"] === 200) {
                        row.find('.status .dot').css('background-color', 'green');
                        row.find('.status .dot').popover('disable');
                    } else {
                        row.find('.status .dot').css('background-color', 'red');
                        let text = "";
                        if (data.error != "") {
                            text = data.error
                        } else {
                            text = 'Status ' + data.status;
                        }
                        row.find('.status .dot').popover({trigger: 'hover'});
                        row.find('.status .dot').popover('enable');
                        row.find('.status .dot').attr('data-content', text);
                    }
                }
                ws.onerror = function(evt) {
                    console.log("ws ERROR: " + evt.data);
                }
            });
        });
    </script>
  </body>
</html>
`
