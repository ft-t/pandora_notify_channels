package pandora_manager

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"os"
	"strings"
	"testing"
	"time"
)

var httpFunc func(ctx *fasthttp.RequestCtx)

func TestMain(m *testing.M) {
	_ = os.Setenv("encKey", "test_key")

	bindAddress := "127.0.0.1:8065"
	go func() {
		_ = fasthttp.ListenAndServe(bindAddress, func(ctx *fasthttp.RequestCtx) {
			httpFunc(ctx)
		})
	}()

	_ = os.Setenv("pandoraApi", fmt.Sprintf("http://%v/api", bindAddress))
	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	os.Exit(code)
}

func TestExecuteCommand(t *testing.T) {
	targetDeviceId := int64(235344213)
	sessionId := "4325tr4efdszcs"

	httpFunc = func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.URI().String()

		if strings.Contains(uri, "iamalive"){
			return
		}

		if !strings.Contains(uri, "/api/devices/command") {
			t.Error("invalid url")
		}

		args := ctx.Request.PostArgs()
		command := string(args.Peek("command"))

		assert.Equal(t, sessionId, string(ctx.Request.Header.Cookie("sid")))
		assert.Equal(t, fmt.Sprint(targetDeviceId), string(args.Peek("id")))
		assert.Equal(t, fmt.Sprint(Check), command)

		ctx.Response.SetStatusCode(200)
		ctx.Response.SetBody([]byte("{\"action_result\":{\"123\":\"sent\"}}"))
		return
	}

	mgr := NewPandoraAccountManager("", "", targetDeviceId, context.Background())
	mgr.sessionId = sessionId

	assert.Equal(t, nil, mgr.SendCommand(Check))
}

func TestKeepAlive(t *testing.T) {
	sessionId := "4325tr4efdszcs"

	httpFunc = func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.URI().String()

		if !strings.Contains(uri, "/iamalive") {
			t.Error("invalid url")
		}

		assert.Equal(t, sessionId, string(ctx.Request.Header.Cookie("sid")))

		ctx.Response.SetStatusCode(200)
	}

	mgr := NewPandoraAccountManager("", "", 0, context.Background())
	mgr.sessionId = sessionId

	assert.Equal(t, nil, mgr.sendKeepAlive())

	assert.True(t, mgr.lastKeepAlive.After(time.Now().UTC().Add(-1*(10*time.Second))))
}

func TestValidPassword(t *testing.T) {
	testLogin := "qwerty"
	encryptedPassword := "ieW1/Kk"
	expectedSid := "12345601"
	expectedExpires, _ := time.Parse(time.RFC1123, "Fri, 11 Sep 2020 18:36:28 UTC")
	httpFunc = func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.URI().String()

		if strings.Contains(uri, "iamalive") {
			return
		}

		if !strings.Contains(uri, "/api/users/login") {
			t.Error("invalid url")
		}

		ctx.Response.Header.Set("Set-Cookie", fmt.Sprintf("sid=%v; expires=Fri, 11 Sep 2020 18:36:28 GMT; path=/", expectedSid))
		ctx.Response.Header.Set("Set-Cookie", "lang=en; expires=Fri, 11 Sep 2020 18:36:28 GMT; path=/")
		ctx.Response.SetStatusCode(200)
		return
	}

	mgr := NewPandoraAccountManager(testLogin, encryptedPassword, 0, context.Background())

	assert.Equal(t, nil, mgr.Authorize())
	assert.Equal(t, expectedSid, mgr.sessionId)
	assert.Equal(t, expectedExpires.String(), mgr.sessionExpiresAt.String())
}

func TestInvalidPassword(t *testing.T) {
	testLogin := "qwerty"
	encryptedPassword := "ieW1/Kk"

	httpFunc = func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.URI().String()

		if !strings.Contains(uri, "/api/users/login") {
			t.Error("invalid url")
		}

		args := ctx.Request.PostArgs()

		assert.Equal(t, testLogin, string(args.Peek("login")))
		assert.Equal(t, "abcds", string(args.Peek("password")))
		assert.Equal(t, "en", string(args.Peek("lang")))

		ctx.Response.SetStatusCode(404)
		return
	}

	mgr := NewPandoraAccountManager(testLogin, encryptedPassword, 0, context.Background())

	assert.Equal(t, "invalid login or password", mgr.Authorize().Error())
}
