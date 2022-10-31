// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var addr = flag.String("addr", ":8081", "http service address")

// CreateCorsObject creates a cors object with the required config
func createCorsObject() *cors.Cors {
	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowOriginFunc: func(s string) bool {
			return true
		},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		ExposedHeaders: []string{"Authorization", "Content-Type"},
	})
}

func main() {
	flag.Parse()

	hub := newHub()
	go hub.run()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})
	router.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		servePublish(hub, w, r)
	})
	corsObj := createCorsObject()
	handler := corsObj.Handler(router)

	err := http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
