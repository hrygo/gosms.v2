package server

import (
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"

	"github.com/hrygo/gosmsn/codec"
)

// 会话信息 gnet.Conn 的附加属性
type session struct {
	sync.Mutex
	id          uint64
	conn        gnet.Conn
	clientId    string    // 客户端识别号，由服务端分配
	serverName  string    // 连接的Server的name
	ver         byte      // 协议版本号
	stat        stat      // 会话状态
	nAt         byte      // 未接收到响应的心跳次数
	lastUseTime time.Time // 接收到客户端的 active/active_resp 或 mt 消息会更新该时间
	counter               // mt, dly, report 计数器
	createTime  time.Time
	window      chan struct{}   // 流控所需通道，登录成功后需设置此值，否则消息不能正常收发
	pool        *goroutine.Pool // 会话级别的线程池，登录成功后需设置此值，否则消息不能正常收发
}

type counter struct {
	mt, dly, report uint64 //  接收到的下行短信、发送的上行短信、发送的状态报告的数量
}

// 会话状态
type stat byte

const (
	StatConnect stat = iota
	StatLogin
	StatClosing
)

func createSession(c gnet.Conn) *session {
	se := &session{}
	se.id = uint64(codec.B64Seq.NextVal())
	se.stat = StatConnect
	se.conn = c
	se.createTime = time.Now()
	se.lastUseTime = time.Now()
	se.pool = createSessionSidePool(1) // 设置一个大小为1的默认pool
	return se
}

func createSessionSidePool(size int) *goroutine.Pool {
	var options = ants.Options{
		ExpiryDuration:   time.Minute, // 1 分钟内不被使用的worker会被清除
		Nonblocking:      false,       // 如果为true,worker池满了后提交任务会直接返回nil
		MaxBlockingTasks: size,        // blocking模式有效，否则worker池满了后提交任务会直接返回nil
		PreAlloc:         false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	var pool, _ = ants.NewPool(size, ants.WithOptions(options))
	return pool
}

// 关闭通道和线程池
func (s *session) closePoolChan() {
	if s == nil {
		return
	}

	if s.window != nil {
		close(s.window)
		s.window = nil
	}
	if s.pool != nil {
		s.pool.Release()
		s.pool = nil
	}
}

func (s *session) Window() chan struct{} {
	return s.window
}

func (s *session) Pool() *goroutine.Pool {
	return s.pool
}

func (s *session) Conn() gnet.Conn {
	return s.conn
}

func (s *session) Id() uint64 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *session) ClientId() string {
	return s.clientId
}

func (s *session) Ver() byte {
	return s.ver
}

func (s *session) Stat() stat {
	return s.stat
}

func (s *session) NAt() byte {
	return s.nAt
}

func (s *session) CreateTime() time.Time {
	return s.createTime
}

func (s *session) LastUseTime() time.Time {
	return s.lastUseTime
}

func (s *session) LogSession(cap ...int) []log.Field {
	if s == nil {
		return nil
	}
	var ret []log.Field
	if len(cap) == 0 {
		ret = make([]log.Field, 0, 8)
	} else if cap[0] > 8 {
		ret = make([]log.Field, 0, cap[0])
	}
	return append(ret,
		log.String(SrvName, s.serverName),
		log.Uint64(Sid, s.Id()),
		log.String(RemoteAddr, s.conn.RemoteAddr().String()))
}

func (s *session) LogCounter(cap ...int) []log.Field {
	var mt, dlv, rpt uint64 = 0, 0, 0
	if s != nil {
		mt, dlv, rpt = s.counter.mt, s.counter.dly, s.counter.report
	}
	return []log.Field{
		log.Uint64(LogKeyCounterMt, mt),
		log.Uint64(LogKeyCounterDlv, dlv),
		log.Uint64(LogKeyCounterRpt, rpt),
	}
}

func (s *session) ServerName() string {
	return s.serverName
}
