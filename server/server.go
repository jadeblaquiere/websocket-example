// Copyright (c) 2016, Joseph deBlaquiere <jadeblaquiere@yahoo.com>
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of ciphrtxt nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/kataras/iris.v5"
)

type IndexResponse struct {
	RequestIP string `json:"request_ip"`
	Time      int64  `json:"unix_time"`
}

type WSClient struct {
	con iris.WebsocketConnection
	wss *WSServer
}

func (wsc *WSClient) Receive(message []byte) {
	fmt.Println("recv :", string(message))
	wsc.con.EmitMessage(message)
}

func (wsc *WSClient) Disconnect() {
	fmt.Println("client disconnect @ ", time.Now().Format("2006-01-02 15:04:05.000000"))
	wsc.wss.Disconnect(wsc)
}

type WSServer struct {
	clients   []*WSClient
	listMutex sync.Mutex
}

func (wss *WSServer) Connect(con iris.WebsocketConnection) {
	wss.listMutex.Lock()
	defer wss.listMutex.Unlock()

	c := &WSClient{con: con, wss: wss}
	wss.clients = append(wss.clients, c)

	fmt.Printf("Connect # active clients : %d\n", len(wss.clients))

	con.OnMessage(c.Receive)

	con.OnDisconnect(c.Disconnect)
}

func (wss *WSServer) Disconnect(wsc *WSClient) {
	wss.listMutex.Lock()
	defer wss.listMutex.Unlock()

	l := len(wss.clients)

	if l == 0 {
		panic("WSS:trying to delete client from empty list")
	}

	for p, v := range wss.clients {
		if v == wsc {
			wss.clients[p] = wss.clients[l-1]
			wss.clients = wss.clients[:l-1]

			fmt.Printf("Disconnect # active clients : %d\n", len(wss.clients))

			return
		}
	}
	panic("WSS:trying to delete client not in list")
}

var wss WSServer

func main() {
	i := iris.New()
	i.Get("/", index)
	i.Config.Websocket.Endpoint = "/ws"
	i.Config.Websocket.BinaryMessages = true
	i.Config.Websocket.WriteTimeout = 60 * time.Second
	i.Config.Websocket.ReadTimeout = 60 * time.Second
	i.Websocket.OnConnection(wss.Connect)

	i.Listen(":8080")
}

func index(ctx *iris.Context) {
	t := time.Now().Unix()
	ctx.JSON(iris.StatusOK, IndexResponse{RequestIP: ctx.RequestIP(), Time: t})
}
