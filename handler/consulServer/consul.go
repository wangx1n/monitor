package consulServerimport

import (
	"fmt"
	"google.golang.org/grpc/naming"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/consul/api"
)

type watchEntry struct {
	addr string
	modi uint64
	last uint64
}

type consulWatcher struct {
	down      int32
	c         *api.Client
	service   string
	mu        sync.Mutex
	watched   map[string]*watchEntry
	lastIndex uint64
}

func (w *consulWatcher) Close() {
	atomic.StoreInt32(&w.down, 1)
}

func (w *consulWatcher) Next() ([]*naming.Update, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	watched := w.watched
	lastIndex := w.lastIndex
retry:
	// 访问 Consul， 获取可用的服务列表
	services, meta, err := w.c.Catalog().Service(w.service, "", &api.QueryOptions{
		WaitIndex: lastIndex,
	})
	if err != nil {
		return nil, err
	}
	if lastIndex == meta.LastIndex {
		if atomic.LoadInt32(&w.down) != 0 {
			return nil, nil
		}
		goto retry
	}
	lastIndex = meta.LastIndex
	var updating []*naming.Update
	for _, s := range services {
		ws := watched[s.ServiceID]
		fmt.Println(s.ServiceAddress, s.ServicePort)
		if ws == nil {
			// 如果是新注册的服务
			ws = &watchEntry{
				addr: net.JoinHostPort(s.ServiceAddress, strconv.Itoa(s.ServicePort)),
				modi: s.ModifyIndex,
			}
			watched[s.ServiceID] = ws

			updating = append(updating, &naming.Update{
				Op:   naming.Add,
				Addr: ws.addr,
			})
		} else if ws.modi != s.ModifyIndex {
			// 如果是原来的服务
			updating = append(updating, &naming.Update{
				Op:   naming.Delete,
				Addr: ws.addr,
			})
			ws.addr = net.JoinHostPort(s.ServiceAddress, strconv.Itoa(s.ServicePort))
			ws.modi = s.ModifyIndex
			updating = append(updating, &naming.Update{
				Op:   naming.Add,
				Addr: ws.addr,
			})
		}
		ws.last = lastIndex
	}
	for id, ws := range watched {
		if ws.last != lastIndex {
			delete(watched, id)
			updating = append(updating, &naming.Update{
				Op:   naming.Delete,
				Addr: ws.addr,
			})
		}
	}
	w.watched = watched
	w.lastIndex = lastIndex
	return updating, nil
}

type consulResolver api.Client

func (r *consulResolver) Resolve(target string) (naming.Watcher, error) {
	return &consulWatcher{
		c:       (*api.Client)(r),
		service: target,
		watched: make(map[string]*watchEntry),
	}, nil
}

func ForConsul(reg *api.Client) naming.Resolver {
	return (*consulResolver)(reg)
}
