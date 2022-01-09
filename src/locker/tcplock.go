package locker

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

var groups_lock sync.Mutex
var lock_num int64

type lock struct {
	sync.Mutex
	count uint32
}

var groups map[string]*lock

var MaxLockTime time.Duration = time.Second * 3

func HttpHandler(ctx *fasthttp.RequestCtx) {
	_name := ctx.FormValue("name")
	name := string(_name)

	atomic.AddInt64(&lock_num, 1)

	//取得锁
	lock := getLock(name)
	//放回锁
	defer closeLock(name, lock)

	//上锁
	lock.Lock()
	defer lock.Unlock()

	conn := ctx.Conn()
	conn.SetWriteDeadline(time.Now().Add(MaxLockTime))

	ctx.Write([]byte("end"))
}

//取得锁，没有时初史化
func getLock(name string) *lock {
	groups_lock.Lock()
	defer groups_lock.Unlock()

	l, ok := groups[name]

	if !ok {
		l = &lock{}
		groups[name] = l
	}

	l.count++

	return l
}

//放回锁，计数器为0时删除锁
func closeLock(name string, l *lock) {
	groups_lock.Lock()
	defer groups_lock.Unlock()

	l.count--

	if l.count == 0 {
		delete(groups, name)
	}
}

//定时打状态
func status() {
	c := time.Tick(5 * time.Minute)
	for _ = range c {
		groups_lock.Lock()
		num := len(groups)
		groups_lock.Unlock()

		log.Println("Crruent Lcoked:", num, "History Locked:", atomic.LoadInt64(&lock_num))
	}
}
