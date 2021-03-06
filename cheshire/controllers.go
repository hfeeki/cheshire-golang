package cheshire

import (
	"fmt"
	"github.com/hoisie/mustache"
	"log"
	"net/http"
)

type HtmlConnection struct {
	*HttpConnection
}

func (this *HtmlConnection) Render(path string, context map[string]interface{}) {
	viewsPath := this.ServerConfig.MustString("http.html.view_directory", "")
	templatePath := fmt.Sprintf("%s%s", viewsPath, path)
	this.WriteResponse("text/html", mustache.RenderFile(templatePath, context))
}

func (this *HtmlConnection) WriteResponse(contentType string, value interface{}) {

	this.Writer.Header().Set("Content-Type", contentType)
	this.Writer.WriteHeader(200)
	this.WriteContent(value)
}

//write out an object 
//this assumes the header has been written already
func (this *HtmlConnection) WriteContent(value interface{}) {
	switch v := value.(type) {
	case string:
		this.Writer.Write([]byte(v))
	case []byte:
		this.Writer.Write(v)
	default:
		log.Println("Dont know how to write :", value)
		//TODO: response object, dynmap, map ect.
	}
}

type HtmlController struct {
	Handlers map[string]func(*Request, *HtmlConnection)
	Conf     *ControllerConfig
}

func NewHtmlController(route string, methods []string, handler func(*Request, *HtmlConnection)) *HtmlController {
	def := &HtmlController{Handlers: make(map[string]func(*Request, *HtmlConnection)), Conf: NewControllerConfig(route)}
	for _, m := range methods {
		def.Handlers[m] = handler
	}
	return def
}

func (this *HtmlController) Config() *ControllerConfig {
	return this.Conf
}

func (this *HtmlController) HandleRequest(request *Request, conn Connection) {
	handler := this.Handlers[request.Method()]
	if handler == nil {
		handler = this.Handlers["ALL"]
	}
	if handler == nil {
		log.Println("Error, not found ", request.Uri())
		//not found!
		//TODO: send 404 page.
		return
	}

	connection, ok := conn.(*HttpConnection)
	if !ok {
		log.Println("not an http connection")
		//not an http connect
		//TODO: send error
		return
	}

	htmlconn := &HtmlConnection{connection}

	handler(request, htmlconn)
}

type StaticFileController struct {
	Route   string
	Path    string
	Conf    *ControllerConfig
	Handler http.Handler
}

// initial the handler via http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp")))
func NewStaticFileController(route string, path string) *StaticFileController {
	handler := http.StripPrefix(route, http.FileServer(http.Dir(path)))
	def := &StaticFileController{Handler: handler, Path: path, Route: route, Conf: NewControllerConfig(route)}
	return def
}

func (this *StaticFileController) Config() *ControllerConfig {
	return this.Conf
}

func (this StaticFileController) HandleRequest(*Request, Connection) {
	//Empty method, this is never called because we have the HttpHijack method in place
}

func (this StaticFileController) HttpHijack(writer http.ResponseWriter, req *http.Request) {
	this.Handler.ServeHTTP(writer, req)
}
