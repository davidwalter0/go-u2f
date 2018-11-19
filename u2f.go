package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/flynn/u2f/u2fhid"
	"github.com/flynn/u2f/u2ftoken"
)

func New() *Device {
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
	device.App = make([]byte, 32)
	io.ReadFull(rand.Reader, device.Challenge)
	io.ReadFull(rand.Reader, device.App)
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
		Application: device.App,
		KeyHandle:   device.KeyHandle,
	}
}

func (device *Device) Register() error {
	defer app.Trace("Register")()
	var err error
	device.Open()
	defer device.Close()

	var Result []byte
	if err = device.Registration.ReadFile(app.Filename); err != nil {
		device.ChallengeInit()
		for {
			Result, err = device.Token.Register(u2ftoken.RegisterRequest{
				Challenge:   device.Challenge,
				Application: device.App,
			})
			if err == u2ftoken.ErrPresenceRequired {
				time.Sleep(200 * time.Millisecond)
				continue
			} else if err != nil {
				mutex.Lock()
				Message <- fmt.Sprintf("Registration: Fail %s", err)
				mutex.Unlock()
				return err
			}
			break
		}
		Result = Result[66:]
		khLen := int(Result[0])
		Result = Result[1:]
		device.KeyHandle = Result[:khLen]
		device.WriteFile(app.Filename)
	} else {
		if app.Debug {
			device.Dump()
		}
	}
	log.Println(RegisteredTitle)
	Window.SetTitle(RegisteredTitle)
	return nil
}

func (device *Device) Authenticate() error {
	defer app.Trace("Authenticate")()

	var err error
	device.Open()
	if err = device.ReadFile(app.Filename); err != nil {
		log.Println(err)
		if err = device.Register(); err != nil {
			return err
		}
	}
	defer device.Close()
	device.SetRequest()
	if app.Debug {
		device.Dump()
	}

	if err = device.Token.CheckAuthenticate(device.Request); err != nil {
		mutex.Lock()
		Message <- fmt.Sprintf("Authenticating: Fail %s", err)
		mutex.Unlock()
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
		if app.Debug {
			log.Printf("counter = %d, signature = %x", Result.Counter, Result.Signature)
		}
		break
	}
	log.Println(AuthenticatedTitle)
	Window.SetTitle(AuthenticatedTitle)
	return nil
}

func (device *Device) Wink() error {
	defer app.Trace("Wink")()

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

func U2FAction(name string) error {
	switch name {
	case "Register":
		defer app.Trace("Action case: Register")()
		go func() {
			handle := New()
			if handle == nil {
				mutex.Lock()
				Message <- InsertSecurityKeyMessage
				mutex.Unlock()
				Action = RegistrationFailed
			} else {
				if err := handle.Register(); err != nil {
					Action = "Register"
					mutex.Lock()
					Message <- fmt.Sprintf("Registration Failed %s", err)
					mutex.Unlock()
				} else {
					Action = Registered
					mutex.Lock()
					Message <- Registered
					mutex.Unlock()
				}
			}
		}()
	case "Authenticate":
		defer app.Trace("Action case: Authenticate")()
		go func() {
			handle := New()
			if handle == nil {
				mutex.Lock()
				Message <- InsertSecurityKeyMessage
				mutex.Unlock()
			} else {
				if err := handle.Authenticate(); err != nil {
					mutex.Lock()
					Message <- fmt.Sprintf("Authentication Failed %s", err)
					mutex.Unlock()
				} else {
					Action = Authenticated
					mutex.Lock()
					Message <- Authenticated
					mutex.Unlock()
				}
			}
		}()
	default:
	}
	return nil
}
