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

// This example based on https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gorilla/websocket"
	"net/http"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 60 * time.Second

	// Time allowed to write a message to the peer.
	readWait = 60 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func main() {
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("get:", string(body))

	var dialer *websocket.Dialer

	con, _, err := dialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		panic(err)
	}

	wchan := make(chan []byte)

	pchan := make(chan []byte)

	go func(con *websocket.Conn) {
		defer con.Close()
		con.SetReadLimit(maxMessageSize)
		con.SetReadDeadline(time.Now().Add(readWait))
		con.SetPongHandler(func(s string) error {
			fmt.Printf("received PONG from %s\n", con.UnderlyingConn().RemoteAddr().String())
			con.SetReadDeadline(time.Now().Add(readWait))
			return nil
		})
		con.SetPingHandler(func(s string) error {
			fmt.Printf("received PING (%s) from %s\n", s, con.UnderlyingConn().RemoteAddr().String())
			con.SetReadDeadline(time.Now().Add(readWait))
			pchan <- []byte(s)
			return nil
		})
		for {
			_, message, err := con.ReadMessage()
			if err != nil {
				fmt.Println("panic @ ", time.Now().Format("2006-01-02 15:04:05.000000"))
				panic(err)
			}
			con.SetReadDeadline(time.Now().Add(readWait))

			fmt.Println("recv:", string(message), "@", time.Now().Format("2006-01-02 15:04:05.000000"))
		}
	}(con)

	go func(con *websocket.Conn, wchan chan []byte) {
		pingtimer := time.NewTicker(pingPeriod)
		defer con.Close()
		con.SetWriteDeadline(time.Now().Add(writeWait))
		for {
			select {
			case wmsg := <-wchan:
				w, err := con.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(wmsg)

				if err := w.Close(); err != nil {
					return
				}

			case pmsg := <-pchan:
				con.SetWriteDeadline(time.Now().Add(writeWait))
				fmt.Printf("sending PONG to %s\n", con.UnderlyingConn().RemoteAddr().String())
				if err := con.WriteControl(websocket.PongMessage, pmsg, time.Now().Add(writeWait)); err != nil {
					return
				}

			case <-pingtimer.C:
				con.SetWriteDeadline(time.Now().Add(writeWait))
				fmt.Printf("sending PING to %s\n", con.UnderlyingConn().RemoteAddr().String())
				if err := con.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
					return
				}
			}
		}
	}(con, wchan)

	go func(con *websocket.Conn) {
		time.Sleep(time.Second * 10)
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond * 500)
			wchan <- []byte(time.Now().Format("2006-01-02 15:04:05.000000"))
		}
		for i := 0; i < 100; i++ {
			time.Sleep(time.Second * 1)
			fmt.Println("(sleeping) time := ", time.Now().Format("2006-01-02 15:04:05.000000"))
		}
		for {
			time.Sleep(time.Millisecond * 500)
			wchan <- []byte(time.Now().Format("2006-01-02 15:04:05.000000"))
		}

	}(con)

	time.Sleep(time.Second * 600)
}
