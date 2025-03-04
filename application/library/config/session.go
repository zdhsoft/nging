/*
Nging is a toolbox for webmasters
Copyright (C) 2018-present Wenhui Shen <swh@admpub.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package config

import (
	"path/filepath"
	"reflect"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/middleware/session/engine"
	"github.com/webx-top/echo/middleware/session/engine/cookie"
	"github.com/webx-top/echo/middleware/session/engine/file"
	"github.com/webx-top/echo/middleware/session/engine/redis"
)

var (
	SessionOptions *echo.SessionOptions
	CookieOptions  *echo.CookieOptions
	SessionEngine  = `file`
	SessionName    = `SID`

	sessionStoreCookieOptions *cookie.CookieOptions
	sessionStoreFileOptions   *file.FileOptions
	sessionStoreRedisOptions  *redis.RedisOptions
)

func InitSessionOptions(c *Config) {

	//==================================
	// session基础设置
	//==================================

	if len(c.Cookie.Path) == 0 {
		c.Cookie.Path = `/`
	}
	if len(c.Cookie.Prefix) == 0 {
		c.Cookie.Prefix = `Nging`
	}
	sessionName := c.Sys.SessionName
	sessionEngine := c.Sys.SessionEngine
	sessionConfig := c.Sys.SessionConfig
	if len(sessionName) == 0 {
		sessionName = SessionName
	}
	if len(sessionEngine) == 0 {
		sessionEngine = SessionEngine
	}
	if sessionConfig == nil {
		sessionConfig = echo.H{}
	}
	_cookieOptions := &echo.CookieOptions{
		Prefix:   c.Cookie.Prefix,
		Domain:   c.Cookie.Domain,
		Path:     c.Cookie.Path,
		MaxAge:   c.Cookie.MaxAge,
		HttpOnly: c.Cookie.HttpOnly,
		SameSite: c.Cookie.SameSite,
	}
	if CookieOptions == nil || SessionOptions == nil ||
		!reflect.DeepEqual(_cookieOptions, CookieOptions) ||
		(SessionOptions.Engine != sessionEngine || SessionOptions.Name != sessionName) {
		if SessionOptions != nil {
			*SessionOptions = *echo.NewSessionOptions(sessionEngine, sessionName, _cookieOptions)
		} else {
			SessionOptions = echo.NewSessionOptions(sessionEngine, sessionName, _cookieOptions)
		}
		if CookieOptions != nil {
			*CookieOptions = *_cookieOptions
		} else {
			CookieOptions = _cookieOptions
		}
	}

	//==================================
	// 注册session存储引擎
	//==================================

	//1. 注册默认引擎：cookie
	_sessionStoreCookieOptions := cookie.NewCookieOptions(c.Cookie.HashKey, c.Cookie.BlockKey)
	if sessionStoreCookieOptions == nil || !reflect.DeepEqual(_sessionStoreCookieOptions, sessionStoreCookieOptions) {
		cookie.RegWithOptions(_sessionStoreCookieOptions)
		sessionStoreCookieOptions = _sessionStoreCookieOptions
	}

	switch sessionEngine {
	case `file`: //2. 注册文件引擎：file
		fileOptions := &file.FileOptions{
			SavePath: sessionConfig.String(`savePath`),
			KeyPairs: sessionStoreCookieOptions.KeyPairs,
			MaxAge:   sessionConfig.Int(`maxAge`),
		}
		if len(fileOptions.SavePath) == 0 {
			fileOptions.SavePath = filepath.Join(echo.Wd(), `data`, `cache`, `sessions`)
		}
		if sessionStoreFileOptions == nil || !engine.Exists(`file`) || !reflect.DeepEqual(fileOptions, sessionStoreFileOptions) {
			file.RegWithOptions(fileOptions)
			engine.Del(`redis`)
			sessionStoreFileOptions = fileOptions
		}
	case `redis`: //3. 注册redis引擎：redis
		redisOptions := &redis.RedisOptions{
			Size:         sessionConfig.Int(`maxIdle`),
			Network:      sessionConfig.String(`network`),
			Address:      sessionConfig.String(`address`),
			Password:     sessionConfig.String(`password`),
			DB:           sessionConfig.Uint(`db`),
			KeyPairs:     sessionStoreCookieOptions.KeyPairs,
			MaxAge:       sessionConfig.Int(`maxAge`),
			MaxReconnect: sessionConfig.Int(`maxReconnect`),
		}
		if redisOptions.Size <= 0 {
			redisOptions.Size = 10
		}
		if len(redisOptions.Network) == 0 {
			redisOptions.Network = `tcp`
		}
		if len(redisOptions.Address) == 0 {
			redisOptions.Address = `127.0.0.1:6379`
		}
		if redisOptions.MaxReconnect <= 0 {
			redisOptions.MaxReconnect = 30
		}
		if sessionStoreRedisOptions == nil || !engine.Exists(`redis`) || !reflect.DeepEqual(redisOptions, sessionStoreRedisOptions) {
			redis.RegWithOptions(redisOptions)
			engine.Del(`file`)
			sessionStoreRedisOptions = redisOptions
		}
	}
}

func AutoSecure(ctx echo.Context, ses *echo.SessionOptions) {
	if !ses.Secure && ctx.IsSecure() {
		ses.Secure = true
	}
}
