package services

import (
	"fmt"
	"encoding/json"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Mine struct {
	Id 				int 			`json:"id"`
	AccountId 		string 			`json:"account_id"`
	Timestamp 		time.Time		`json:"timestamp"`
	LastPing		string 			`json:"last_ping"`
}

var currentPostId int
var currentUserId int

func Create() {

	m := new(Mine)

	currentPostId += 1
	currentUserId += 1

	m.Id = currentPostId
	m.Timestamp = time.Now()
	c := RedisConnect()
	defer c.Close()

	b, err := json.Marshal(m)
	HandleError(err)

	reply, err := c.Do("SET", "machine:" + strconv.Itoa(m.Id), b)
	HandleError(err)

	fmt.Println("GET ", reply)
}

func Find(id int) Mine {

	var m Mine

	c := RedisConnect()
	defer c.Close()

	reply, err := c.Do("GET", "machine:" + strconv.Itoa(id))
	HandleError(err)

	if err = json.Unmarshal(reply.([]byte), &m); err != nil {
		panic(err)
	}
	return m
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func RedisConnect() redis.Conn {
	c, err := redis.Dial("tcp", "localhost:6379")
	HandleError(err)
	return c
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
