package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "gopkg.in/redis.v3"

    "github.com/gorilla/websocket"
)

var addr = flag.String(
    "addr", "localhost:5050", "http service address")

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
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })

    pong, err := client.Ping().Result()
    if err != nil {
        panic(err)
    }
    fmt.Println(pong)
    //http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    //    fmt.Fprintf(w, "Hello, %q", pong)
    //})

    http.HandleFunc("/echo", echo)
    //http.HandleFunc("/", home)
    log.Fatal(http.ListenAndServe(*addr, nil))

}

