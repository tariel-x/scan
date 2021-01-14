package scan

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"strconv"
	"sync"
	"time"

	"github.com/tjgq/sane"
	"go.uber.org/zap"
)

func NewScan(l *zap.Logger) (*Scan, error) {
	if err := sane.Init(); err != nil {
		return nil, fmt.Errorf("can not init sane: %w", err)
	}

	s := &Scan{
		l:        l,
		devicesM: &sync.RWMutex{},
	}

	if err := s.findDevices(); err != nil {
		return nil, fmt.Errorf("can not list devices: %w", err)
	}

	go func() {
		for {
			time.Sleep(time.Minute * 30)
			if err := s.findDevices(); err != nil {
				s.l.Error("can not list devices", zap.Error(err))
			}
		}
	}()

	return s, nil
}

type Scan struct {
	devices  []Device
	devicesM *sync.RWMutex
	l        *zap.Logger
}

type Device struct {
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	Type   string `json:"type"`
}

func (s *Scan) Stop() {
	sane.Exit()
}

func (s *Scan) findDevices() error {
	devs, err := sane.Devices()
	if err != nil {
		return err
	}

	devices := make([]Device, 0, len(devs))
	for _, dev := range devs {
		devices = append(devices, Device{
			Name:   dev.Name,
			Vendor: dev.Vendor,
			Model:  dev.Model,
			Type:   dev.Type,
		})
	}

	s.devicesM.Lock()
	defer s.devicesM.Unlock()
	s.devices = devices
	s.l.Info("updated device list")
	return nil
}

func (s *Scan) openDevice(d Device) (*sane.Conn, error) {
	c, err := sane.Open(d.Name)
	if err != nil {
		return nil, fmt.Errorf("can not open device: %w", err)
	}
	return c, nil
}

func (s *Scan) findDevice(name string) (Device, error) {
	var found bool
	var d Device
	for _, device := range s.devices {
		if device.Name == name {
			d = device
			found = true
			break
		}
	}
	if !found {
		return d, ErrNotFound
	}

	return d, nil
}

func (s *Scan) ListDevices() []Device {
	s.devicesM.RLock()
	defer s.devicesM.RUnlock()

	devices := s.devices

	return devices
}

func (s *Scan) UpdateDevicesList() error {
	if err := s.findDevices(); err != nil {
		s.l.Error("can not list devices", zap.Error(err))
		return fmt.Errorf("can not list devices: %w", err)
	}
	return nil
}

var ErrNotFound = errors.New("device not found")

type Option struct {
	Name    string        `json:"name"`
	Group   string        `json:"group"`
	Title   string        `json:"title"`
	Desc    string        `json:"description"`
	Type    Type          `json:"type"`
	Length  int           `json:"length"`
	Set     []interface{} `json:"set"`
	Range   *Range        `json:"range"`
	Default interface{}   `json:"default"`
}

type Type int

const (
	TypeBool Type = iota
	TypeInt
	TypeFloat
	TypeString
	TypeButton
)

var fromSaneType = map[sane.Type]Type{
	sane.TypeBool:   TypeBool,
	sane.TypeInt:    TypeInt,
	sane.TypeFloat:  TypeFloat,
	sane.TypeString: TypeString,
	sane.TypeButton: TypeButton,
}

var toSaneType = map[Type]sane.Type{
	TypeBool:   sane.TypeBool,
	TypeInt:    sane.TypeInt,
	TypeFloat:  sane.TypeFloat,
	TypeString: sane.TypeString,
	TypeButton: sane.TypeButton,
}

type Range struct {
	Min   interface{} `json:"min"`   // minimum value
	Max   interface{} `json:"max"`   // maximum value
	Quant interface{} `json:"quant"` // quantization step
}

func (s *Scan) GetDeviceOptions(name string) ([]Option, error) {
	d, err := s.findDevice(name)
	if err != nil {
		return nil, err
	}

	conn, err := s.openDevice(d)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	lastGroup := ""

	var options []Option

	for _, opt := range conn.Options() {
		if !opt.IsSettable {
			continue
		}
		if opt.Group != lastGroup {
			lastGroup = opt.Group
		}
		val, _ := conn.GetOption(opt.Name)
		options = append(options, s.fromSaneOption(opt, val))
	}

	return options, nil
}

func (s *Scan) fromSaneOption(opt sane.Option, val interface{}) Option {
	option := Option{
		Name:    opt.Name,
		Group:   opt.Group,
		Title:   opt.Title,
		Desc:    opt.Desc,
		Type:    fromSaneType[opt.Type],
		Length:  opt.Length,
		Set:     opt.ConstrSet,
		Default: val,
	}

	if opt.ConstrRange != nil {
		option.Range = &Range{
			Min:   opt.ConstrRange.Min,
			Max:   opt.ConstrRange.Max,
			Quant: opt.ConstrRange.Quant,
		}
	}

	return option
}

type Argument struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

var ErrInvalidArgument = errors.New("invalid argument")

func (s *Scan) Scan(name string, arguments []Argument) ([]byte, error) {
	d, err := s.findDevice(name)
	if err != nil {
		return nil, err
	}

	conn, err := s.openDevice(d)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	opts := conn.Options()

	for _, arg := range arguments {
		if arg.Value == nil {
			continue
		}

		opt, err := s.findOption(opts, arg.Name)
		if err != nil {
			return nil, err
		}
		var v interface{}
		var ok bool

		switch opt.Type {
		case sane.TypeBool:
			if v, ok = arg.Value.(bool); !ok {
				return nil, fmt.Errorf("%s is %T not a bool: %w", v, arg.Name, ErrInvalidArgument) // not a bool
			}
		case sane.TypeInt:
			switch typedV := arg.Value.(type) {
			case int:
			case int32:
			case int64:
			case float64:
				v = int(typedV)
			case string:
				var err error
				v, err = strconv.ParseInt(typedV, 0, 32)
				if err != nil {
					return nil, fmt.Errorf("%s is not valid int: %w", typedV, ErrInvalidArgument)
				}
			default:
				return nil, fmt.Errorf("%s is %T not an int: %w", typedV, arg.Name, ErrInvalidArgument) // not an int
			}
		case sane.TypeFloat:
			if v, ok = arg.Value.(float64); !ok {
				return nil, fmt.Errorf("%s is %T not a float64: %w", v, arg.Name, ErrInvalidArgument) // not a float
			}
		case sane.TypeString:
			if v, ok = arg.Value.(string); !ok {
				return nil, fmt.Errorf("%s is %T not a string: %w", v, arg.Name, ErrInvalidArgument) // not a float
			}
		}

		if _, err := conn.SetOption(opt.Name, v); err != nil {
			return nil, fmt.Errorf("can not set option: %w", err)
		}
	}

	img, err := conn.ReadImage()
	if err != nil {
		return nil, fmt.Errorf("can not read image: %w", err)
	}

	out := &bytes.Buffer{}
	if err := png.Encode(out, img); err != nil {
		return nil, fmt.Errorf("can not decode image: %w", err)
	}

	return out.Bytes(), nil
}

var ErrNoOption = errors.New("no such option")

func (s *Scan) findOption(opts []sane.Option, name string) (*sane.Option, error) {
	for _, o := range opts {
		if o.Name == name {
			return &o, nil
		}
	}
	return nil, fmt.Errorf("can not find option %q: %w", name, ErrNoOption)
}
