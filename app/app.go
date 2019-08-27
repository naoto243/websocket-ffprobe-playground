package app

import (
	"fmt"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"net/http"
)

func NewServerApp() ServerApp {

	return &implServerApp {

	}
}

type ServerApp interface {
	Run() error
	Close()
}

type implServerApp struct {
	e *echo.Echo
}

func (self *implServerApp) Run() error {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/ffprobe", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/ws", ws)

	self.e = e

	e.Logger.Fatal(e.Start(":1323"))

	return nil
}

func (self *implServerApp) Close() {
	self.e.Close()
}


func ws(c echo.Context) error {




	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			// Write
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}

			// Read
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
			}
			fmt.Printf("%s\n", msg)
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

