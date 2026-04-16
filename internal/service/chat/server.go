package chat

//
//import (
//	"IMChat/pkg/constants"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/go-redis/redis/v8"
//	"github.com/gorilla/websocket"
//	"log"
//	"strings"
//	"sync"
//	"time"
//)
//
//type Server struct {
//	//Clients map[string] *Client;
//	mutex    *sync.Mutex
//	Transmit chan []byte
//	//Login chan *Client;
//	//Logout chan *Client;
//}
//
//var ChatServer *Server
//
//func init() {
//	if ChatServer == nil {
//		ChatServer = &Server{
//			mutex:    new(sync.Mutex),
//			Transmit: make(chan []byte, constants.CHANNEL_SIZE),
//			//...
//		}
//	}
//}
