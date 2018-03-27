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
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
    "github.com/streadway/amqp"
)

var rabbitAddr = flag.String(
    "addr",
    "amqp://guest:guest@localhost:5672/",
    "rabbitmq service address",
    )

var redisAddr = flag.String(
    "redis",
    "localhost:6379",
    "redis service address",
    )

func main() {

    flag.Parse()

    conn, err := amqp.Dial(*rabbitAddr)
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "ping",
        false,
        false,
        false,
        false,
        nil,
    )
    failOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        q.Name, // queue
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    failOnError(err, "Failed to register a consumer")

    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt)

    doneCh := make(chan struct{})
    go func() {
        for {
            select {
            case d := <- msgs:
                if string(d.Body) == "" {
                    //fmt.Println(d.Body)
                } else {
                    log.Printf("Received a message: %s", d.Body)

                    message := new(Message)
                    err := json.Unmarshal(d.Body, message)
                    if err != nil {
                        log.Printf("Error marshalling json", err)
                    }

                    if message.Type == "ping" {
                        go UpdatePing(
                            message.Payload.(map[string]interface{}))
                    }
                }
            case <-signals:
                fmt.Println("Interrupt is detected")
                doneCh <- struct{}{}
            }
        }
    }()

    go StartHttpServer()
    fmt.Println("Up and running")

    <-doneCh
    fmt.Println("Finished")

}

type Message struct {
    SenderId 		string					`json:"mac"`
    UserHash 		string					`json:"mac"`
    Type 			string					`json:"type"`
    Payload 		interface{}				`json:"payload"`
}

func StartHttpServer() {

    e := echo.New()
    e.Use(middleware.CORS())

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    e.GET("/", Handler)
    e.GET("/create", CreateRig)
    e.Logger.Fatal(e.Start(":5051"))
}

func Handler(c echo.Context) error {

    m := Find("1")
    return c.JSON(http.StatusOK, m)
}

func CreateRig(c echo.Context) error {

    rig := new(Rig)
    Create(*rig)

    return nil
}

type Rig struct {
    Id 				int 			`json:"id"`
    AccountId 		string 			`json:"account_id"`
    Timestamp 		time.Time		`json:"timestamp"`
    LastPing		Ping 			`json:"last_ping"`
    TotalHashRate   string          `json:"total_hash_rate"`
    Gpus            []*Gpu          `json:"gpus"`
}

type Gpu struct {
    Id              int             `json:"id"`
    HashRate        string          `json:"hash_rate"`
}


type Ping struct {
    Time 		time.Time		`json:"time"`
}


func Create(r Rig) {

    r.Id = 1
    r.Timestamp = time.Now()
    c := RedisConnect()
    defer c.Close()

    b, err := json.Marshal(r)
    handlerError(err)

    reply, err := c.Do(
        "SET", "machine:" + strconv.Itoa(r.Id), b)
    handlerError(err)

    fmt.Println("SET ", reply)
}

func Find(id string) Rig {

    var r Rig
    c := RedisConnect()
    defer c.Close()

    reply, err := c.Do("GET", "machine:" + id)
    handlerError(err)

    err = json.Unmarshal(reply.([]byte), &r)
    handlerError(err)

    return r
}
func RedisConnect() redis.Conn {

    c, err := redis.Dial("tcp", *redisAddr)
    handlerError(err)
    return c
}

func UpdatePing(payload map[string]interface{}) {

    payloadJson, err := json.Marshal(payload)
    ping := new(Ping)
    err = json.Unmarshal(payloadJson, ping)
    handlerError(err)

    rig := Find("1")
    rig.LastPing = *ping
    UpdateRig(rig)
}

func UpdateRig(r Rig) {

    r.Timestamp = time.Now()
    c := RedisConnect()
    defer c.Close()

    b, err := json.Marshal(r)
    handlerError(err)

    reply, err := c.Do(
        "SET", "machine:" + strconv.Itoa(r.Id), b)
    handlerError(err)

    fmt.Println("GET ", reply)
}

func failOnError(err error, msg string) {

    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

func handlerError(err error) {

    if err != nil {
        panic(err)
    }
}

