package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tariel-x/scan/internal/scan"
	"go.uber.org/zap"
)

func NewApi(listen string, s *scan.Scan, l *zap.Logger, box packr.Box) (*Api, error) {
	e := echo.New()
	e.Logger.SetOutput(ioutil.Discard)

	a := &Api{
		l:      l,
		e:      e,
		s:      s,
		listen: listen,
		box:    box,
	}

	e.Use(a.LoggerMiddleware)
	//e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8085", "http://localhost:5000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.GET("/api/devices", a.GetDevices)
	e.POST("/api/devices/refresh", a.GetDevicesRefresh)
	e.GET("/api/devices/:name/options", a.GetDevicesOptions)
	e.POST("/api/devices/:name/scan", a.PostDevicesScan)
	e.GET("/*", a.Static)

	return a, nil
}

func (a *Api) LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		start := time.Now()
		if err = next(c); err != nil {
			c.Error(err)
		}
		stop := time.Now()

		p := req.URL.Path
		if p == "" {
			p = "/"
		}

		l := stop.Sub(start)

		a.l.Info("request", zap.String("path", p), zap.Int("status", res.Status), zap.Int64("latency", l.Milliseconds()))
		return
	}
}

type Api struct {
	l      *zap.Logger
	e      *echo.Echo
	s      *scan.Scan
	listen string
	box    packr.Box
}

func (a *Api) Run() error {
	a.l.Info("starting at " + a.listen)
	return a.e.Start(a.listen)
}

func (a *Api) Stop() error {
	return a.e.Close()
}

func (a *Api) GetDevices(c echo.Context) error {
	return c.JSON(http.StatusOK, a.s.ListDevices())
}

func (a *Api) GetDevicesRefresh(c echo.Context) error {
	if err := a.s.UpdateDevicesList(); err != nil {
		return errors.New("can not find devices")
	}
	return c.JSON(http.StatusOK, a.s.ListDevices())
}

func (a *Api) GetDevicesOptions(c echo.Context) error {
	name := c.Param("name")
	options, err := a.s.GetDeviceOptions(name)
	if err != nil {
		if err == scan.ErrNotFound {
			return err
		}
		a.l.Error("can not get device options", zap.Error(err))
		return errors.New("can not get device options")
	}
	return c.JSON(http.StatusOK, options)
}

func (a *Api) PostDevicesScan(c echo.Context) error {
	name := c.Param("name")
	img, err := a.s.Scan(name, nil)
	if err != nil {
		a.l.Error("can not scan image", zap.Error(err))
	}
	return c.Blob(http.StatusOK, "image/png", img)
}

func (a *Api) Static(c echo.Context) error {
	httpHandler := http.FileServer(a.box)
	httpHandler.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}
