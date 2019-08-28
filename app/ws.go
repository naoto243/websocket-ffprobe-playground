package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	upgrader = websocket.Upgrader{}
)



func NewWsApp() WsApp {

	return &implWsApp {

		wsPool: make( map[string]*websocket.Conn , 0),
		fileInfoPool:make(map[string]*FileInfo , 0),
	}
}

type WsApp interface {
	Run() error
	Close()
}

type implWsApp struct {
	e *echo.Echo

	wsPool map[string]*websocket.Conn
	fileInfoPool map[string]*FileInfo
}

func (self *implWsApp) Run() error {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	pwd , _ := os.Getwd()
	fmt.Println(pwd)

	e.Static("/", pwd + "/public")
	e.GET("/ws/start/:fileName", self.startFfprobeConnection)
	e.GET("/get_data/:socketId/:fileName", self.getData)

	self.e = e

	e.Logger.Fatal(e.Start(":1323"))

	return nil
}

func (self *implWsApp) Close() {

	self.e.Close()
}

func (self *implWsApp) getData(c echo.Context) error {

	log.Info(`ffprobeからなんかきた`)


	rangeReq := c.Request().Header.Get(`Range`)

	//sId := c.Param(`socketId`)
	fName := c.Param(`fileName`)
	socketId := c.Param(`socketId`)

	ws  , ok  := self.wsPool[socketId]
	if !ok {
		return c.String(http.StatusBadRequest , `no conn`)
	}

	info  , ok  := self.fileInfoPool[socketId]
	if !ok {
		return c.String(http.StatusBadRequest , `no conn`)
	}

	log.Info(rangeReq)

	start , end , part , err := parseRangeHeader(rangeReq , info.Size)
	if err != nil {
		log.Error(err)
		return err
	}

	r := RangeRequest{
		SocketId:socketId,
		Range:rangeReq,
		FileName:fName,
		StartByte:start,
		EndByte:end,
	}
	j   , _ := json.Marshal(r)


	//  ブラウザにpush
	err = ws.WriteMessage(websocket.TextMessage, j)
	if err != nil {
		c.Logger().Error(err)
	}

	// ブラウザでの読み取りを待ち受け
	t, res, err := ws.ReadMessage()
	if err != nil {
		c.Logger().Error(err)
	}

	log.Info(t)

	/*
	res.set('Accept-Ranges', 'bytes');
	res.set('Connection', 'close');
	res.set('Content-Type', type);
	res.set('Content-Length', range.end - range.start);
	res.set('Content-Range', 'bytes ' + range.start + '-' + range.end + '/' + size);
	res.statusCode = 206;
	 */

	if part != len(res) {
		err = fmt.Errorf(`expected: %d , but: %d` , part , len(res))
		log.Error(`sizeちゃうな`)
		log.Error(err)
		return err
	}


	c.Response().Header().Set(`Accept-Ranges`, `bytes`)
	c.Response().Header().Set(`Connection`, `close`)
	c.Response().Header().Set(`Content-Type`, info.Type)
	c.Response().Header().Set(`Content-Length`, fmt.Sprint(len(res)))

	rangeResponse := fmt.Sprintf(`bytes %d-%d/%d` , start , start + len(res) , info.Size)
	fmt.Println(rangeResponse)
	c.Response().Header().Set(`Content-Range` , rangeResponse)
	c.Response().Status = http.StatusPartialContent

	_ , err = c.Response().Write(res)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func parseRangeHeader(rangeHeader string , size int) (start int , end int , total int , err error){

	defer func() {
		if onPanic := recover(); onPanic != nil {
			err = errors.New(fmt.Sprint(onPanic))
			log.Error(err)
		}
	}()


	b := strings.Split(rangeHeader , "=")
	b2 := strings.Split(b[1] , "-")

	_start := b2[0]
	//_end := b2[1] // ""

	s , err := strconv.Atoi(_start)
	if err != nil {
		log.Error(err)
		return 0, 0, 0 , err
	}

	const partSize = 1024 * 1024 * 5

	e := s + partSize
	if e >= size {
		e =  size
	}

	total = e - s

	return s , e , total , nil
}

type RangeRequest struct {
	SocketId  string `json:"socket_id"`
	Range string `json:"range"`
	FileName string `json:"file_name"`
	StartByte int `json:"start_byte"`
	EndByte int `json:"end_byte"`
}

type FileInfo struct {
	Size int
	Type string
}


func (self *implWsApp) startFfprobeConnection(c echo.Context) error {

	size   , _ := strconv.Atoi(c.QueryParam(`size`))
	fType := c.QueryParam(`file_type`)


	uniq := fmt.Sprint(time.Now().UnixNano())

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	m := new(sync.Mutex)
	m.Lock()
	self.wsPool[uniq] =  ws
	self.fileInfoPool[uniq]  = &FileInfo{
		Size: size,
		Type:fType,
	}
	m.Unlock()

	defer func() {
		ws.Close()
		m.Lock()
		self.wsPool[uniq] =  nil
		self.fileInfoPool[uniq] = nil
		m.Unlock()
	}()


	fname := c.Param(`fileName`)

	isComplete := false
	go func() {

		time.Sleep(time.Second)

		log.Info(`ffprobeを起動`)
		url := fmt.Sprintf( `http://127.0.0.1:1323/get_data/%s/%s` , uniq , fname)
		cmd := fmt.Sprintf(`ffprobe -i '%s' -v quiet -print_format json -show_format -show_streams -show_error -show_chapters` , url)

		fmt.Println(cmd)
		out , err := exec.Command(`sh` , `-c` , cmd).CombinedOutput()
		if err != nil {
			log.Error(err)
		}
		isComplete = true

		fmt.Println(string(out))
	}()

	for {
		time.Sleep(time.Second)

		if isComplete {
			log.Info(`complete or error!`)
			break
		}
	}

	return nil
}


