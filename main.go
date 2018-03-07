package main

import (
    "encoding/json"
    "flag"
    "log"
    "net/http"

    "github.com/berryhill/mine-ws/services"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:5050", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("upgrade:", err)
        return
    }
    defer c.Close()
    for {
        mt, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            break
        }
        log.Printf("recv: %s", message)
        err = c.WriteMessage(mt, message)
        if err != nil {
            log.Println("write:", err)
            break
        }
    }
}

func main() {

    http.HandleFunc("/", Handler)
    http.HandleFunc("/create", Create)

    http.HandleFunc("/echo", echo)
    log.Fatal(http.ListenAndServe(*addr, nil))

}

func Handler(w http.ResponseWriter, r *http.Request) {

    m := services.Find(1)
    json.NewEncoder(w).Encode(m)
}

func Create(w http.ResponseWriter, r *http.Request) {

    services.Create()
}
