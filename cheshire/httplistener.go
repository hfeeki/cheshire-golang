package cheshire

import (
	"encoding/json"
	"fmt"
	"github.com/trendrr/cheshire-golang/dynmap"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type HttpConnection struct {
	Writer        http.ResponseWriter
	request       *http.Request
	ServerConfig  *ServerConfig
	headerWritten sync.Once
}

func (conn *HttpConnection) Write(response *Response) (int, error) {

	json, err := json.Marshal(response)
	if err != nil {
		//TODO: uhh, do something..
		log.Print(err)
	}
	conn.headerWritten.Do(func() {
		conn.Writer.Header().Set("Content-Type", "application/json")
		conn.Writer.WriteHeader(response.StatusCode())
	})
	conn.Writer.Write(json)
	conn.Writer.Write([]byte("\n"))
	v, ok := conn.Writer.(http.Flusher)
	if ok {
		v.Flush()
	}

	return 200, err
}

type httpHandler struct {
	serverConfig *ServerConfig
}

// Implement this interface for a controller to skip the normal cheshire life cycle
// This should be only used in special cases (static file serving, websockets, ect)
// controllers that implement this interface will skip the HandleRequest function alltogether
type HttpHijacker interface {
	HttpHijack(writer http.ResponseWriter, req *http.Request)
}

func (this *httpHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	controller := this.serverConfig.Router.Match(req.Method, req.URL.Path)

	//check if controller is the special HttpHijacker.
	h, hijack := controller.(HttpHijacker)
	if hijack {
		h.HttpHijack(writer, req)
		return
	}

	//we are already in a go routine, so no need to start another one.
	request := ToStrestRequest(req)

	//TODO: filters here..
	conn := HttpConnection{Writer: writer, request: req, ServerConfig: this.serverConfig}

	controller.HandleRequest(request, &conn)
}

func ToStrestRequest(req *http.Request) *Request {
	var request = new(Request)
	request.SetUri(req.URL.Path)
	request.SetMethod(req.Method)
	request.SetTxnId(req.Header.Get("Strest-Txn-Id"))
	request.SetTxnAccept(req.Header.Get("Strest-Txn-Accept"))
	if len(request.TxnAccept()) == 0 {
		request.SetTxnAccept("single")
	}

	if req.Method == "POST" || req.Method == "PUT" {
		req.ParseForm()
		pms, _ := dynmap.ToDynMap(parseValues(req.Form))
		request.SetParams(pms)
	} else {
		//parse the query params
		values := req.URL.Query()
		pms, _ := dynmap.ToDynMap(parseValues(values))
		request.SetParams(pms)
	}
	return request
}

func parseValues(values url.Values) map[string]interface{} {
	params := map[string]interface{}{}
	for k := range values {
		var v = values[k]
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}
	return params
}

func HttpListen(port int, serverConfig *ServerConfig) error {
	handler := &httpHandler{serverConfig}

	log.Println("HTTP Listener on port: ", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
