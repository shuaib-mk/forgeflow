package api

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Metrics struct{mu sync.RWMutex;started time.Time;requests map[string]uint64;duration map[string]time.Duration}
func NewMetrics()*Metrics{return &Metrics{started:time.Now(),requests:map[string]uint64{},duration:map[string]time.Duration{}}}
func(m *Metrics)Observe(method,path string,status int,duration time.Duration){key:=method+" "+path+" "+strconv.Itoa(status);m.mu.Lock();m.requests[key]++;m.duration[key]+=duration;m.mu.Unlock()}
func(m *Metrics)ServeHTTP(w http.ResponseWriter,_ *http.Request){m.mu.RLock();defer m.mu.RUnlock();w.Header().Set("Content-Type","text/plain; version=0.0.4");fmt.Fprintf(w,"# HELP forgeflow_uptime_seconds Process uptime.\n# TYPE forgeflow_uptime_seconds gauge\nforgeflow_uptime_seconds %.0f\n",time.Since(m.started).Seconds());fmt.Fprintln(w,"# HELP forgeflow_http_requests_total HTTP requests handled.\n# TYPE forgeflow_http_requests_total counter");for key,count:=range m.requests{fmt.Fprintf(w,"forgeflow_http_requests_total{request=%q} %d\n",key,count)}}

