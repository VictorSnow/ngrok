package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	smux "github.com/VictorSnow/smux"
)

const BUFF_SIZE = 8192

type Config struct {
	Mode        string
	Listen_addr string
	Local_addr  string
	Smux_addr   string
}

type Sconn struct {
	c net.Conn
}

// ignore timeout err
func (c Sconn) Write(buff []byte) (n int, err error) {
	n, err = c.c.Write(buff)
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return c.Write(buff)
	}
	return
}

// ignore timeout err
func (c Sconn) Read(buff []byte) (n int, err error) {
	n, err = c.c.Read(buff)
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return c.Read(buff)
	}
	return
}

var ServerConfig Config

func main() {
	config_file := flag.String("config", "", "config file")
	flag.Parse()

	if *config_file == "" {
		*config_file = "config.json"
	}

	f, e := os.OpenFile(*config_file, os.O_CREATE, os.ModePerm)
	if e != nil {
		log.Println("打开文件错误", *f)
		return
	}

	err := json.NewDecoder(f).Decode(&ServerConfig)

	if err != nil {
		log.Println("解析配置文件错误", err)
		return
	}

	go http.ListenAndServe(":9001", nil)

	if ServerConfig.Mode == "server" {
		server := smux.NewSmux(ServerConfig.Smux_addr, "server")
		go server.Start()

		// if you need a http server forward, not a tcp traffic forward
		// here is an example use of smux
		//
		//		// http server
		//		clientServer := ClientServer{
		//			addr: ServerConfig.Listen_addr,
		//			host: ServerConfig.Local_addr,
		//			dial: func(addr string) (net.Conn, error) {
		//				// ignore addr
		//				smuxConn, err := server.Dail()
		//				if err != nil {
		//					log.Println("dial error", err)
		//					return nil, err
		//				}
		//				dialConn := DialConn{smuxConn}
		//				// 开始分发数据
		//				return dialConn, nil
		//			},
		//		}

		//		clientServer.Listen()

		// server
		go func() {
			l, err := net.Listen("tcp", ServerConfig.Listen_addr)

			if err != nil {
				log.Println(err)
				return
			}

			for {
				conn, err := l.Accept()
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("handle new conn")
				go handleConn(server, conn)
			}
		}()
	} else {
		client := smux.NewSmux(ServerConfig.Smux_addr, "client")
		go client.Start()

		go func() {
			for {
				c := client.Accept()
				conn, err := net.Dial("tcp", ServerConfig.Local_addr)

				if err != nil {
					c.Close(true)
					log.Panicln(err)
					return
				}

				go pipe(c, conn)
			}
		}()
	}

	for {
		time.Sleep(time.Second)
	}
}

func handleConn(s *smux.Smux, c net.Conn) {
	conn, err := s.Dail()
	if err != nil {
		log.Println(err)
		c.Close()
		return
	}

	log.Println("dial success")
	pipe(conn, c)
}

func pipe(c1 *smux.Conn, c2 net.Conn) {

	sc := Sconn{c2}

	wg := sync.WaitGroup{}
	wg.Add(2)

	once := sync.Once{}
	closer := func() {
		c1.Close(true)
		c2.Close()
	}

	go func() {
		defer wg.Done()
		defer once.Do(closer)

		buff := make([]byte, BUFF_SIZE)
		for {
			n, err := c1.Read(buff)
			if err != nil {
				break
			}

			n, err = sc.Write(buff[:n])

			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer once.Do(closer)

		buff := make([]byte, BUFF_SIZE)
		for {
			n, err := sc.Read(buff)
			if err != nil {
				break
			}

			n, err = c1.Write(buff[:n])

			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
}
