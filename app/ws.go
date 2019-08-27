package app

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
)

var (
	upgrader = websocket.Upgrader{}
)



func NewWsApp() WsApp {

	return &implWsApp {

	}
}

type WsApp interface {
	Run() error
	Close()
}

type implWsApp struct {
	e *echo.Echo
}

func (self *implWsApp) Run() error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	pwd , _ := os.Getwd()
	fmt.Println(pwd)

	e.Static("/", pwd + "/public")
	e.GET("/ws", hello)

	self.e = e

	e.Logger.Fatal(e.Start(":1323"))

	return nil
}

func (self *implWsApp) Close() {

	self.e.Close()
}


func hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// Write
		err := ws.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
		if err != nil {
			c.Logger().Error(err)
		}

		// Read
		_, msg, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error(err)
		}
		fmt.Printf("%s\n", msg)
	}
}

func WS() {

}