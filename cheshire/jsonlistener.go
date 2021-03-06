package cheshire

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	// "github.com/trendrr/cheshire-golang/dynmap"
)

type JsonConnection struct {
	serverConfig *ServerConfig
	conn         net.Conn
	writerLock   sync.Mutex
}

func (conn JsonConnection) Write(response *Response) (int, error) {
	json, err := json.Marshal(response)
	if err != nil {
		//TODO: uhh, do something..
		log.Print(err)
	}
	defer conn.writerLock.Unlock()
	conn.writerLock.Lock()
	bytes, err := conn.conn.Write(json)
	return bytes, err
}

func JsonListen(port int, config *ServerConfig) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// handle error
		log.Println(err)
		return err
	}
	log.Println("Json Listener on port: ", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			// handle error
			continue
		}
		go handleConnection(JsonConnection{serverConfig: config, conn: conn})
	}
	return nil
}

func handleConnection(conn JsonConnection) {
	defer conn.conn.Close()
	// log.Print("CONNECT!")

	dec := json.NewDecoder(bufio.NewReader(conn.conn))

	for {
		var req Request
		err := dec.Decode(&req)

		if err == io.EOF {
			log.Print(err)
			break
		} else if err != nil {
			log.Print(err)
			break
		}
		//request
		controller := conn.serverConfig.Router.Match(req.Method(), req.Uri())
		go controller.HandleRequest(&req, conn)
	}

	log.Print("DISCONNECT!")
}
