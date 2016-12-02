package main

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	smux "github.com/VictorSnow/smux"
)

const BUFF_SIZE = 8192

func main() {

	go http.ListenAndServe(":9001", nil)

	server := smux.NewSmux("127.0.0.1:8090", "server")
	client := smux.NewSmux("127.0.0.1:8090", "client")

	go server.Start()
	go client.Start()

	// server
	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:9000")

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

	go func() {
		for {
			c := client.Accept()
			conn, err := net.Dial("tcp", "127.0.0.1:80")

			if err != nil {
				c.Close(true)
				log.Panicln(err)
				return
			}

			go pipe(c, conn)
		}
	}()

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

			n, err = c2.Write(buff[:n])

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
			n, err := c2.Read(buff)
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
