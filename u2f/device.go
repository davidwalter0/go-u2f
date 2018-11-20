package u2f

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/davidwalter0/go-u2f/cfg"
	"github.com/flynn/hid"
	"github.com/flynn/u2f/u2fhid"
	"github.com/flynn/u2f/u2ftoken"
)

// type Env struct {
// 	Tracing   bool
// 	Debugging bool
// 	File      string
// }

// type Tracer interface {
// 	Filename() string
// 	Trace(string) func()
// 	String() string
// }

const (
	Register                 = "Register"
	Authenticate             = "Authenticate"
	Registered               = "Registered"
	Authenticated            = "Authenticated"
	AuthenticationFailed     = "Authentication: Finalize"
	MissingKey               = "Missing Key"
	RegistrationFailed       = "Registration: Finalize"
	PressKeyToAuthenticate   = "Press key to authenticate"
	InsertSecurityKeyMessage = "Insert Security Key"
	PrimaryTitle             = "%20.20s Security Key U2F -- %s"
)

var (
	UnAuthenticatedTitle = fmt.Sprintf(PrimaryTitle, " ", "UnAuthenticated")
	AuthenticatedTitle   = fmt.Sprintf(PrimaryTitle, " ", "Authenticated")
	RegisteredTitle      = fmt.Sprintf(PrimaryTitle, " ", "Registered")
)

type Device struct {
	DeviceInfo *hid.DeviceInfo
	Device     *u2fhid.Device
	Token      *u2ftoken.Token
	Version    string
	Result     []byte
	Request    u2ftoken.AuthenticateRequest
	Registration
}

func NewDevice() *Device {
	devices, err := u2fhid.Devices()
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		// log.Fatal("no U2F tokens found")
		return nil
	}

	devInfo := devices[0]

	return &Device{
		DeviceInfo: devInfo,
	}
}

func (device *Device) String() string {
	return fmt.Sprintf("version %s, mfg = %q, prod = %q, vid = 0x%04x, pid = 0x%04x",
		device.DeviceInfo.Manufacturer,
		device.DeviceInfo.Product,
		device.DeviceInfo.ProductID,
		device.DeviceInfo.VendorID,
		device.Version)
}

func (device *Device) ChallengeInit() {
	device.Challenge = make([]byte, 32)
	device.Application = make([]byte, 32)
	io.ReadFull(rand.Reader, device.Challenge)
	io.ReadFull(rand.Reader, device.Application)
}

func (device *Device) Open() error {
	var err error

	device.Device, err = u2fhid.Open(device.DeviceInfo)
	if err != nil {
		log.Fatal(err)
	}

	device.Token = u2ftoken.NewToken(device.Device)
	device.Version, err = device.Token.Version()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (device *Device) Close() {
	device.Device.Close()
}

func (device *Device) SetRequest() {
	device.Request = u2ftoken.AuthenticateRequest{
		Challenge:   device.Challenge,
		Application: device.Application,
		KeyHandle:   device.KeyHandle,
	}
}

func (device *Device) Register(Message chan string) error {
	defer cfg.Env.Trace("Register")()
	var err error
	device.Open()
	defer device.Close()

	var Result []byte
	if err = device.Registration.ReadFile(cfg.Env.Filename); err != nil {
		device.ChallengeInit()
		for {
			Result, err = device.Token.Register(u2ftoken.RegisterRequest{
				Challenge:   device.Challenge,
				Application: device.Application,
			})
			if err == u2ftoken.ErrPresenceRequired {
				time.Sleep(200 * time.Millisecond)
				continue
			} else if err != nil {
				Send(fmt.Sprintf("Registration: Fail %s", err), Message)
				return err
			}
			break
		}
		Result = Result[66:]
		khLen := int(Result[0])
		Result = Result[1:]
		device.KeyHandle = Result[:khLen]
		device.WriteFile(cfg.Env.Filename)
	} else {
		if cfg.Env.Debugging {
			device.Dump()
		}
	}
	return nil
}

func (device *Device) Authenticate(Message chan string) error {
	defer cfg.Env.Trace("Authenticate")()

	var err error
	device.Open()
	if err = device.ReadFile(cfg.Env.Filename); err != nil {
		log.Println(err)
		if err = device.Register(Message); err != nil {
			return err
		}
	}
	defer device.Close()
	device.SetRequest()
	if cfg.Env.Debugging {
		device.Dump()
	}

	if err = device.Token.CheckAuthenticate(device.Request); err != nil {
		return err
	}
	io.ReadFull(rand.Reader, device.Challenge)
	for {
		Result, err := device.Token.Authenticate(device.Request)
		if err == u2ftoken.ErrPresenceRequired {
			time.Sleep(200 * time.Millisecond)
			continue
		} else if err != nil {
			return err
		}
		if cfg.Env.Debugging {
			log.Printf("counter = %d, signature = %x", Result.Counter, Result.Signature)
		}
		break
	}
	return nil
}

func (device *Device) Wink() error {
	defer cfg.Env.Trace("Wink")()

	if device.Device.CapabilityWink {
		for i := 0; i < 3; i++ {
			for j := 1; j < 4; j++ {
				if err := device.Device.Wink(); err != nil {
					return err
				}
				time.Sleep(time.Second / time.Duration((j * 2)))
			}
		}
	} else {
		return fmt.Errorf("No Wink capability")
	}
	return nil
}

func Send(text string, Message chan string) {
	if Message != nil {
		Message <- text
	}
}

func U2FAction(name string, Message chan string) error {
	switch name {
	case Register:
		defer cfg.Env.Trace("Action case: Register")()

		handle := NewDevice()
		if handle == nil {
			Send(InsertSecurityKeyMessage, Message)
			return fmt.Errorf("Registration Failed %s", InsertSecurityKeyMessage)
		} else {
			if err := handle.Register(Message); err != nil {
				Send(fmt.Sprintf("Registration Failed %s", err), Message)
				return fmt.Errorf("Registration Failed %s", err)
			} else {
				Send(Registered, Message)
			}
		}

	case Authenticate:
		defer cfg.Env.Trace("Action case: Authenticate")()

		handle := NewDevice()
		if handle == nil {
			Send(InsertSecurityKeyMessage, Message)
			return fmt.Errorf("Registration Failed %s", InsertSecurityKeyMessage)
		} else {
			if err := handle.Authenticate(Message); err != nil {
				Send(fmt.Sprintf("Authentication Failed %s", err), Message)
				return fmt.Errorf("Authentication Failed %s", err)
			} else {
				Send(Authenticated, Message)
			}
		}

	default:
		return fmt.Errorf("U2FAction: Unknown Action argument %s", name)
	}
	return nil
}

func GTKU2FAction(name string, Message chan string, errors chan error) {
	switch name {
	case Register:
		go func() {
			defer cfg.Env.Trace("Action case: Register")()
			handle := NewDevice()
			if handle == nil {
				Send(InsertSecurityKeyMessage, Message)
				errors <- fmt.Errorf("Registration Failed %s", InsertSecurityKeyMessage)
			} else {
				if err := handle.Register(Message); err != nil {
					Send(fmt.Sprintf("Registration Failed %s", err), Message)
					errors <- fmt.Errorf("Registration Failed %s", err)
				} else {
					Send(Registered, Message)
				}
			}
		}()
	case Authenticate:
		go func() {
			defer cfg.Env.Trace("Action case: Authenticate")()
			handle := NewDevice()
			if handle == nil {
				Send(InsertSecurityKeyMessage, Message)
				errors <- fmt.Errorf("Registration Failed %s", InsertSecurityKeyMessage)
			} else {
				if err := handle.Authenticate(Message); err != nil {
					Send(fmt.Sprintf("Authentication Failed %s", err), Message)
					errors <- fmt.Errorf("Authentication Failed %s", err)
				} else {
					Send(Authenticated, Message)
				}
			}
		}()
	default:
		errors <- fmt.Errorf("U2FAction: Unknown Action argument %s", name)
	}
	errors <- nil
}
