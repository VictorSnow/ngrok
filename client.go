package main

import (
	"log"
	"net"
	"time"

	"github.com/VictorSnow/smux"
	"github.com/valyala/fasthttp"
)

type ClientServer struct {
	dial fasthttp.DialFunc
	addr string
	host string
}

/**
* http 转发服务
 */
func (c ClientServer) Listen() {
	server := fasthttp.Server{
		Handler: c.RequestHandler,
	}
	server.ListenAndServe(c.addr)
}

type DialConn struct {
	c *smux.Conn
}

func (c DialConn) Close() error {
	c.c.Close(true)
	return nil
}

func (c DialConn) LocalAddr() net.Addr {
	panic("not implemented")
	return nil
}

func (c DialConn) RemoteAddr() net.Addr {
	panic("not implemented")
	return nil
}

func (c DialConn) Read(buff []byte) (int, error) {
	return c.c.Read(buff)
}

func (c DialConn) Write(buff []byte) (int, error) {
	return c.c.Write(buff)
}

func (c DialConn) SetDeadline(t time.Time) error {
	return nil
}

func (c DialConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c DialConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c ClientServer) RequestHandler(ctx *fasthttp.RequestCtx) {
	dstReq := &fasthttp.Request{}
	dstResponse := &fasthttp.Response{}
	ctx.Request.CopyTo(dstReq)

	dstReq.SetHost(c.host)

	client := fasthttp.Client{
		Dial: c.dial,
	}
	err := client.Do(dstReq, dstResponse)
	if err != nil {
		ctx.Response.SetStatusCode(500)
		ctx.Response.SetBody([]byte("远程连接错误: " + err.Error()))
		log.Println("err in upstream", err)
		return
	}

	// copy
	dstResponse.CopyTo(&ctx.Response)
}
