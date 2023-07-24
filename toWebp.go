package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	router := gin.Default()
	router.POST("/img", func(c *gin.Context) {
		body := c.Request.Body
		// 读取body
		data, err := io.ReadAll(body)
		if err != nil {
			log.Fatalln("err", err.Error())
			c.AbortWithError(500, err)
			return
		}
		// body 压缩
		img, err := Compress(data)
		if err != nil {
			log.Fatalln("err", err.Error())
			c.AbortWithError(500, err)
			return
		}
		c.Data(200, "image/webp", img)
	})
	router.GET("/url/*img", func(c *gin.Context) {
		img := c.Param("img")
		//去掉前面的/
		img = img[1:]
		rawPath := c.Request.URL.RawPath
		//rawPath = rawPath[4:]
		imgUrl := fmt.Sprintf("%s%s?%s", img, rawPath, c.Request.URL.RawQuery)
		remote, err := url.Parse(imgUrl)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = remote.Path
			req.URL.RawPath = remote.RawPath
			req.URL.RawQuery = remote.RawQuery
			req.Header.Set("referer", "https://www.toutiao.com/")
			//打印请求
		}
		proxy.ModifyResponse = func(res *http.Response) error {
			//缓存到/tmp
			if res.StatusCode == 200 {
				// 读取res.Body
				body, err := io.ReadAll(res.Body)
				if err != nil {
					log.Fatalln("Body:", err)
					return err
				}
				// body 压缩
				body, err = Compress(body)
				if err != nil {
					log.Fatalln("err", err.Error())
					return err
				}
				res.Header.Set("Content-Length", strconv.Itoa(len(body)))
				// 重置res.Body
				res.Body = io.NopCloser(bytes.NewReader(body))
			}
			return nil
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	srv := &http.Server{
		Addr:    ":8800",
		Handler: router,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func Compress(body []byte) ([]byte, error) {
	var buf []byte
	var intMinusOne vips.IntParameter
	intMinusOne.Set(-1)
	img, err := vips.LoadImageFromBuffer(body, &vips.ImportParams{
		NumPages: intMinusOne,
	})
	if err != nil {
		return nil, err
	}
	if img.Width() == 0 || img.Height() == 0 {
		return nil, fmt.Errorf("image width or height is 0")
	}
	buf, _, err = img.ExportWebp(vips.NewWebpExportParams())
	return buf, err
}
