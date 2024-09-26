package main

import (
	mycommon "github.com/dapr-platform/common"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"websocket-service/api"
	_ "websocket-service/docs"
	"websocket-service/handler"
	_ "websocket-service/prom"
)

var upgrader *websocket.Upgrader

var server common.Service

var initTopic = map[string][]string{}
var messageCenter = handler.NewMessageCenter()
var PORT = 80

func init() {
	if val := os.Getenv("LISTEN_PORT"); val != "" {
		PORT, _ = strconv.Atoi(val)
	}
	log.Println("use PORT ", PORT)
	buf, _ := ioutil.ReadFile("build.time")
	t, _ := strconv.ParseInt(strings.TrimSpace(string(buf))+"000", 10, 64)
	s := time.UnixMilli(t).Format("2006-01-02 15:04:05")
	mycommon.Logger.Debug("build.time:", s, string(buf))

	initTopicStr := os.Getenv("INIT_TOPIC")
	mycommon.Logger.Debug("initTopicStr=" + initTopicStr)
	if initTopicStr == "" {
		initTopicStr = "pubsub:web" //for debug
	}
	if initTopicStr != "" {

		arr := strings.Split(initTopicStr, "#")

		for _, t := range arr {
			tarr := strings.Split(t, ":")
			if len(tarr) == 2 {
				var topics []string
				topics, exists := initTopic[tarr[0]]
				if !exists {
					topics = make([]string, 0)
				}
				topics = append(topics, tarr[1])
				initTopic[tarr[0]] = topics
			}

		}
	}
	mycommon.Logger.Debug("initTopic=", initTopic)
}

// @title websocket-service RESTful API
// @version 1.0
// @description websocket-service API 文档.
// @BasePath /swagger/websocket-service
func main() {
	router := chi.NewRouter()
	api.InitTestRoute(router)
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/swagger*", httpSwagger.WrapHandler)

	upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	router.HandleFunc("/ws/{pubsubName}/{topicName}", func(w http.ResponseWriter, r *http.Request) {
		topicName := chi.URLParam(r, "topicName")
		if topicName == "" {
			mycommon.Logger.Debug("topicName is blank")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("topicName is blank"))
			return
		}
		pubsubName := chi.URLParam(r, "pubsubName")
		if pubsubName == "" {
			mycommon.Logger.Debug("pubsubName is blank")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("pubsubName is blank"))
			return
		}

		mycommon.Logger.Debugf("sub connect. %s %s %s", pubsubName, topicName, r.RemoteAddr)

		unsafeConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			mycommon.Logger.Debugf("upgrade error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("upgrade error:" + err.Error()))
			return
		}
		key := "sub_" + pubsubName + "_" + topicName

		subHandler := handler.NewWsHandler(messageCenter, key, unsafeConn)
		subHandler.Start()
	})

	server = daprd.NewServiceWithMux(":"+strconv.Itoa(PORT), router)
	for k, v := range initTopic {
		for _, t := range v {
			var sub = &common.Subscription{
				PubsubName: k,
				Topic:      t,
				Route:      "/eventHandler",
			}
			err := server.AddTopicEventHandler(sub, messageCenter.EventHandler)
			mycommon.Logger.Debugf("addTopicEventHandler %s,%s", k, t)
			if err != nil {
				panic(err)
			}
		}

	}

	mycommon.Logger.Debug("server start")
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error: %v", err)
	}
}
