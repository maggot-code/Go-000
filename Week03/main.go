/*
 * @Author: maggot-code
 * @Date: 2020-12-09 12:21:38
 * @LastEditors: maggot-code
 * @LastEditTime: 2020-12-09 21:37:25
 * @Description: 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够 一个退出，全部注销退出。
 */
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const PORT1 string = ":8848"
const PORT2 string = ":8899"

type IndexHandler struct {
	name string
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(h.name))
}

// CloseHandler 可触发http.Server Close
type CloseHandler struct {
	CloseChan chan error
}

func (h *CloseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("closing"))

	select {
	default:
		h.CloseChan <- errors.New("api shutdown")
	case <-h.CloseChan:
	}

}

func startServer(stopCh chan<- struct{}) {

	// index1 handler
	mux1 := http.NewServeMux()
	mux1.Handle("/", &IndexHandler{name: "index1"})
	closeHandler1 := &CloseHandler{}
	mux1.Handle("/close", closeHandler1)

	s1Ch := make(chan error, 1)

	// index2 handler
	mux2 := http.NewServeMux()
	mux2.Handle("/", &IndexHandler{name: "index2"})
	closeHandler2 := &CloseHandler{}
	mux2.Handle("/close", closeHandler1)

	s2Ch := make(chan error, 1)

	// 监听系统信号
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	ctx := context.Background()
	group, _ := errgroup.WithContext(ctx)

	server1 := &http.Server{Addr: PORT1, Handler: mux1}
	closeHandler1.CloseChan = s1Ch
	server2 := &http.Server{Addr: PORT2, Handler: mux2}
	closeHandler2.CloseChan = s2Ch

	group.Go(func() error {
		// 收到任何一个信号就关闭服务
		select {
		case <-ch:
			fmt.Println("receive close signal!")
		case err := <-s1Ch:
			fmt.Printf("receive server1 close! %+v\n", err)
		case err := <-s2Ch:
			fmt.Printf("receive server2 close! %+v\n", err)
		}

		signal.Stop(ch)
		close(s1Ch)
		close(s2Ch)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err1 := server1.Shutdown(ctx)
		fmt.Printf("server1 close %+v \n", err1)

		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err2 := server2.Shutdown(ctx)
		fmt.Printf("server2 close %+v \n", err2)

		return nil
	})

	group.Go(func() error {
		err := server1.ListenAndServe()
		select {
		default:
			s1Ch <- err
		case <-s1Ch:
		}
		fmt.Printf("server1 closed %+v \n", err)
		return nil
	})

	group.Go(func() error {
		err := server2.ListenAndServe()
		select {
		default:
			s1Ch <- err
		case <-s1Ch:
		}
		fmt.Printf("server2 closed %+v \n", err)
		return nil
	})

	err := group.Wait()
	stopCh <- struct{}{}
	fmt.Printf("group err %+v \n", err)
}

func main() {
	stopCh := make(chan struct{}, 1)
	go startServer(stopCh)
	<-stopCh
	close(stopCh)
}
