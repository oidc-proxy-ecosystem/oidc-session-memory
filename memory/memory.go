package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
)

func New() *memorySession {
	return newMemorySession()
}

type item struct {
	Value   string `json:"value"`
	Expires int64  `json:"expires"`
}

func newItem(value string, expiredTime int64) *item {
	return &item{
		Value:   value,
		Expires: expiredTime,
	}
}

func (i *item) ToJson() string {
	buf, _ := json.Marshal(i)
	return string(buf)
}

func (i *item) Expired(time int64) bool {
	if i.Expires == 0 {
		return true
	}
	return time > i.Expires
}

type memorySession struct {
	items  map[string]*item
	mu     sync.Mutex
	prefix string
	ttl    int
	writer io.Writer
	log    logger.ILogger
}

var _ session.Session = &memorySession{}

func (c *memorySession) Get(ctx context.Context, originalKey string) (string, error) {
	c.mu.Lock()
	key := path.Join(c.prefix, originalKey)
	var s string = ""
	if v, ok := c.items[key]; ok {
		s = v.Value
	}
	c.log.Debug(fmt.Sprintf("[GET] %s:%s", key, s))
	c.mu.Unlock()
	return s, nil
}

func (c *memorySession) Put(ctx context.Context, originalKey string, value string) error {
	c.mu.Lock()
	expiredTime := time.Now().Add(time.Duration(c.ttl) * time.Minute)
	key := path.Join(c.prefix, originalKey)
	c.items[key] = newItem(value, expiredTime.UnixNano())
	c.log.Debug(fmt.Sprintf("[PUT] %s:%s", key, value))
	c.mu.Unlock()
	return nil
}

func (c *memorySession) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	key = path.Join(c.prefix, key)
	c.log.Debug(fmt.Sprintf("[DEL] %s", key))
	delete(c.items, key)
	return nil
}

func (c *memorySession) Close(ctx context.Context) error {
	if c.writer != nil {
		if file, ok := c.writer.(*os.File); ok {
			file.Close()
		}
	}
	return nil
}

func (c *memorySession) Init(ctx context.Context, setting map[string]interface{}) error {
	if prefix, ok := setting["prefix"].(string); ok {
		c.prefix = prefix
	}
	var (
		filename      string
		ok            bool
		logLevel      string
		logFormat     logger.LogFormatType  = logger.FormatLong
		logDateFormat logger.TimeFormatType = logger.FormatDatetime
	)
	if filename, ok = setting["filename"].(string); !ok {
		filename = ""
	}
	if filename != "" {
		if file, err := os.Create(filename); err != nil {
			c.writer = os.Stdout
		} else {
			c.writer = file
		}
	} else {
		c.writer = os.Stdout
	}
	if logLevel, ok = setting["loglevel"].(string); !ok {
		logLevel = logger.Info.String()
	}
	if format, ok := setting["logformat"].(string); ok {
		switch strings.ToLower(format) {
		case "long":
			logFormat = logger.FormatLong
		case "std", "standard":
			logFormat = logger.FormatStandard
		case "short":
			logFormat = logger.FormatShort
		}
	}
	if dateFormat, ok := setting["logdateformat"].(string); ok {
		switch strings.ToLower(dateFormat) {
		case "date":
			logDateFormat = logger.FormatDate
		case "datetime":
			logDateFormat = logger.FormatDatetime
		case "time":
			logDateFormat = logger.FormatMillisec
		}
	}
	c.log = logger.New(c.writer, logger.Convert(logLevel), logFormat, logDateFormat)
	return nil
}

func newMemorySession() *memorySession {
	c := &memorySession{
		items:  make(map[string]*item),
		mu:     sync.Mutex{},
		prefix: "memory",
	}
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				c.mu.Lock()
				for k, v := range c.items {
					if v.Expired(time.Now().UnixNano()) {
						delete(c.items, k)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
	return c
}
