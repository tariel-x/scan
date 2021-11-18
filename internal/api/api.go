package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/labstack/echo/v4"
	"github.com/tariel-x/scan/internal"
	"github.com/tariel-x/scan/internal/scan"
	"go.uber.org/zap"
)

func NewApi(listen string, s *scan.Scan, l *zap.Logger, box packr.Box, p internal.ProjectManager) (*Api, error) {
	e := echo.New()
	e.Logger.SetOutput(ioutil.Discard)

	a := &Api{
		l:      l,
		e:      e,
		s:      s,
		p:      p,
		listen: listen,
		box:    box,
	}

	e.Use(a.LoggerMiddleware)
	e.GET("/api/devices", a.GetDevices)
	e.POST("/api/devices/refresh", a.GetDevicesRefresh)
	e.GET("/api/devices/:name/options", a.GetDevicesOptions)
	e.GET("/api/devices/:name/images", a.GetDevicesImages)
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
	l        *zap.Logger
	e        *echo.Echo
	s        *scan.Scan
	p        internal.ProjectManager
	exporter internal.Exporter
	listen   string
	box      packr.Box
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

func (a *Api) GetDevicesImages(c echo.Context) error {
	name := c.Param("name")
	scannerImages, err := a.p.Get(name)
	if err != nil {
		a.l.Error("can not get scanner images", zap.Error(err))
		return errors.New("can not get scanner images")
	}
	type image struct {
		Name string `json:"name"`
	}
	images := make([]image, 0, len(scannerImages))
	for _, scannerImage := range scannerImages {
		images = append(images, image{Name: scannerImage.Name})
		scannerImage.Image.Close()
	}
	return c.JSON(http.StatusOK, images)
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

	rawArguments := map[string]interface{}{}
	if err := c.Bind(&rawArguments); err != nil {
		a.l.Error("can not parse settings", zap.Error(err))
		return errors.New("can not parse settings")
	}

	logger := a.l.With(zap.String("scanner", name))

	arguments := make([]scan.Argument, 0, len(rawArguments))
	for key, value := range rawArguments {
		if key == "name" {
			continue
		}
		arguments = append(arguments, scan.Argument{
			Name:  key,
			Value: value,
		})
	}

	img, err := a.s.Scan(name, arguments)
	if err != nil {
		logger.Error("can not scan image", zap.Error(err))
		return errors.New("can not scan image")
	}

	if _, err := a.p.AddImage(name, img); err != nil {
		logger.Error("can not add image to the scanner project", zap.Error(err))
		return errors.New("can not scan image")
	}

	return c.Blob(http.StatusOK, "image/png", img)
}

func (a *Api) Static(c echo.Context) error {
	httpHandler := http.FileServer(a.box)
	httpHandler.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}
