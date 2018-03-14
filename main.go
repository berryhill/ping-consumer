package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "time"

    "github.com/garyburd/redigo/redis"
    "github.com/Shopify/sarama"
)

var addr = flag.String(
    "addr", ":5050", "http service address")

func main() {

    config := sarama.NewConfig()
    config.Consumer.Return.Errors = true

    brokers := []string{"35.193.166.194:9092"}

    master, err := sarama.NewConsumer(brokers, config)
    if err != nil {
        panic(err)
    }

    defer func() {
        if err := master.Close(); err != nil {
            panic(err)
        }
    }()

    topic := "ping"
    consumer, err := master.ConsumePartition(
        topic, 0, sarama.OffsetNewest)
    if err != nil {
        panic(err)
    }

    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt)

    msgCount := 0

    doneCh := make(chan struct{})
    go func() {
        for {
            select {
            case err := <-consumer.Errors():
                fmt.Println(err)
            case msg := <-consumer.Messages():
                msgCount++
                go UpdatePing([]byte(msg.Value))
            case <-signals:
                fmt.Println("Interrupt is detected")
                doneCh <- struct{}{}
            }
        }
    }()

    go StartHttpServer()

    fmt.Println("Up and running")

    <-doneCh
    fmt.Println("Processed", msgCount, "messages")

}

func StartHttpServer() {
    http.HandleFunc("/", Handler)
    http.HandleFunc("/create", CreateRig)
    log.Fatal(http.ListenAndServe(*addr, nil))
}

func Handler(w http.ResponseWriter, r *http.Request) {

    m := Find("100")
    json.NewEncoder(w).Encode(m)
}

func CreateRig(w http.ResponseWriter, r *http.Request) {

    rig := new(Rig)
    Create(*rig)
}

type Rig struct {
    Id 				int 			`json:"id"`
    AccountId 		string 			`json:"account_id"`
    Timestamp 		time.Time		`json:"timestamp"`
    LastPing		Ping 			`json:"last_ping"`
}

var currentPostId int
var currentUserId int

func Create(r Rig) {

    currentPostId += 1
    currentUserId += 1

    r.Id = 100
    r.Timestamp = time.Now()
    c := RedisConnect()
    defer c.Close()

    b, err := json.Marshal(r)
    HandleError(err)

    reply, err := c.Do(
        "SET", "machine:" + strconv.Itoa(r.Id), b)
    HandleError(err)

    fmt.Println("GET ", reply)
}

func Find(id string) Rig {

    var r Rig

    c := RedisConnect()
    defer c.Close()

    reply, err := c.Do("GET", "machine:" + id)
    HandleError(err)

    if err = json.Unmarshal(reply.([]byte), &r); err != nil {
        panic(err)
    }
    return r
}

func HandleError(err error) {

    if err != nil {
        panic(err)
    }
}

func RedisConnect() redis.Conn {

    c, err := redis.Dial("tcp", "redis-master:6379")
    HandleError(err)
    return c
}

func UpdatePing(payload []byte) {

    ping := new(Ping)
    err := json.Unmarshal(payload, ping)
    if err != nil {
        HandleError(err)
    }

    rig := Find("100")
    rig.LastPing = *ping

    Create(rig)
}

type Ping struct {
    Id 			string			`json:"id"`
    Service 	string			`json:"service"`
    Time 		time.Time		`json:"time"`
}

//func FindAll() [][]byte {
//
//	var posts [][]byte
//
//	c := RedisConnect()
//	defer c.Close()
//
//	keys, err := c.Do("KEYS", "post:*")
//	HandleError(err)
//
//	for _, k := range keys.([]interface{}) {
//
//		var post []byte
//
//		reply, err := c.Do("GET", k.([]byte))
//		HandleError(err)
//
//		if err := json.Unmarshal(reply.([]byte), &post); err != nil {
//			panic(err)
//		}
//		posts = append(posts, post)
//	}
//	return posts
//}