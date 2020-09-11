package pandora_manager

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"os"
	"pandora_notify_channels/pandora_manager/internal"
	"strings"
	"sync"
	"time"
)

var lang = "en"

type PandoraAccountManager struct {
	login             string
	encryptedPassword string
	sessionId         string
	deviceId          int64 // group -> info
	sessionExpiresAt  time.Time
	ctx               context.Context
	cancel            context.CancelFunc
	lastKeepAlive     time.Time
	keepAliveMutex    sync.Mutex
}

func NewPandoraAccountManager(login string, encryptedPassword string, deviceId int64, ctx context.Context) *PandoraAccountManager {
	innerCtx, cancel := context.WithCancel(ctx)

	mgr := &PandoraAccountManager{
		login:             login,
		encryptedPassword: encryptedPassword,
		ctx:               innerCtx,
		cancel:            cancel,
		deviceId:          deviceId,
	}

	mgr.initKeepAlive()

	return mgr
}

func (m *PandoraAccountManager) Close() {
	m.cancel()
}

func (m *PandoraAccountManager) initKeepAlive() {
	go func() {
		for m.ctx.Err() == nil {
			if len(m.sessionId) == 0 {
				time.Sleep(5 * time.Second)
				continue
			}

			err := m.sendKeepAlive() // todo log
			if err != nil {
				fmt.Println(err)
			}

			time.Sleep(27 * time.Second)
			continue
		}
	}()
}

//func (m *PandoraAccountManager) acquireUpdates() error {
//	if len(m.sessionId) == 0 {
//		return errors.New("invalid session id")
//	}
//
//	req := fasthttp.AcquireRequest()
//	defer fasthttp.ReleaseRequest(req)
//	resp := fasthttp.AcquireResponse()
//	defer fasthttp.ReleaseResponse(resp)
//
//	req.Header.SetMethod("GET")
//
//	now := time.Now() // !! exactly now, not utc. // currentTs = - 1 in some cases
//	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
//	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
//
//	req.SetRequestURI(fmt.Sprintf("%v/updates?ts=%v&from=%v&to=%v", getPandoraApiUrl(), now.Unix(),
//		from.Unix(), to.Unix()))
//
//	m.appendHttpHeaders(req, false, true)
//
//	if err := fasthttp.DoTimeout(req, resp, 3*time.Second); err != nil {
//		return err
//	}
//}

func (m *PandoraAccountManager) SendCommand(command PandoraCommand) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/devices/command", getPandoraApiUrl()))
	m.appendHttpHeaders(req, true, true)

	form := req.PostArgs()
	form.Add("id", fmt.Sprint(m.deviceId))
	form.Add("command", fmt.Sprint(command))

	if err := fasthttp.DoTimeout(req, resp, 3*time.Second); err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New("invalid keep alive response")
	}

	rawBody := string(resp.Body())

	if !strings.Contains(rawBody, "sent") {
		return errors.New(fmt.Sprintf("invalid command response. %v", rawBody))
	}

	return nil
}

func (m *PandoraAccountManager) sendKeepAlive() error {
	if len(m.sessionId) == 0 {
		return errors.New("invalid session id")
	}

	m.keepAliveMutex.Lock()
	defer m.keepAliveMutex.Unlock()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/iamalive", getPandoraApiUrl()))
	m.appendHttpHeaders(req, true, true)

	form := req.PostArgs()
	form.Add("num_click", "0")

	if err := fasthttp.DoTimeout(req, resp, 3*time.Second); err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New("invalid keep alive response")
	}

	m.lastKeepAlive = time.Now().UTC()

	return nil
}

func (m *PandoraAccountManager) Authorize() error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	decryptedPass, err := internal.Decrypt(m.encryptedPassword)

	if err != nil {
		return err
	}

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/users/login", getPandoraApiUrl()))

	m.appendHttpHeaders(req, true, false)

	form := req.PostArgs()

	form.Add("lang", "en")
	form.Add("login", m.login)
	form.Add("password", decryptedPass)

	if err = fasthttp.DoTimeout(req, resp, 3*time.Second); err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New("invalid login or password")
	}

	cook := fasthttp.AcquireCookie()
	cook.SetKey("sid")
	defer fasthttp.ReleaseCookie(cook)

	resp.Header.Cookie(cook)
	m.sessionId = string(cook.Value())
	m.sessionExpiresAt = cook.Expire().UTC()

	if len(m.sessionId) == 0 {
		return errors.New("invalid sid (problem with auth)")
	}

	return nil
}

func (m *PandoraAccountManager) appendHttpHeaders(req *fasthttp.Request, isPost bool, addAuth bool) {
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7,uk-UA;q=0.6,uk;q=0.5")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36")

	if isPost {
		req.Header.SetContentType("application/x-www-form-urlencoded; charset=UTF-8")
	}

	if addAuth {
		req.Header.SetCookie("sid", m.sessionId)
		req.Header.SetCookie("lang", lang)
	}
}

func getPandoraApiUrl() string {
	osUrl := os.Getenv("pandoraApi")

	if len(osUrl) != 0 {
		return osUrl
	}

	return "https://p-on.ru/api"
}
