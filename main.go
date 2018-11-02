package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
	"chessSever/program/player"
	"chessSever/program/game"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"chessSever/program/model"
	"chessSever/program/game/msg"
)

var addr = flag.String("addr", "localhost:8888", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type TempUser struct{
	userID int
	sync.RWMutex
}

var user TempUser
func echo(w http.ResponseWriter, r *http.Request) {

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer con.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	//暂时用变量模拟用户登陆，后续从数据库读取用户信息，实例化用户，游戏过程中用redis来暂存游戏信息，用户推出后持久化到数据库
	user.Lock()
	user.userID++
	nowId := user.userID
	user.Unlock()

	currUser := &model.User{
		Id:nowId,
		NickName:"玩家"+strconv.Itoa(nowId),
		Avatar:"no_avatar",
	}
	currPlayer := player.NewPlayer(currUser,con)

	player.SendMsgToPlayer(currPlayer,msg.MSG_TYPE_OF_LOGIN,"用户登陆")

	if currPlayer.User.Id%3 == 1{
		currPlayer.CreateTable(game.GAME_ID_OF_DOUDOZHU)
	}else{
		currPlayer.JoinTable(player.GetRoom().GetAllTable()[len(player.GetRoom().GetAllTable())-1])
	}

	wg.Add(1)
	//启动一个goroutine监听该客户端发来的消息
	go player.HandlerUserMsg(&wg,con,currPlayer)

	wg.Wait()
}

func home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w,r,"/pages/index.html",301)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	db,err := gorm.Open("mysql", "root:root@/chessgame?charset=utf8&parseTime=True&loc=Local")
	if err != nil{
		panic(err.Error())
	}
	db.AutoMigrate(&model.User{})

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./views/static"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("./views/pages"))))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
