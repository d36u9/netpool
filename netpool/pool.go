package netpool

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type Pool interface {
	Get() (net.Conn, error)
	Close() error
	Destroy(conn net.Conn) error
	IsValid() bool
	Len() int
}

type ConnPool struct {
	lock        sync.Mutex
	conns       chan net.Conn
	initSize    int
	cnt         int
	valid       bool
	connCreator func() (net.Conn, error)
}

func GetTCPPool(host, port string, poolInit int) (*ConnPool, error) {
	return CreatePool(poolInit, func() (net.Conn, error) {
		return net.Dial("tcp", host+":"+port)
	})
}

func CreatePool(initSize int, connCreator func() (net.Conn, error)) (*ConnPool, error) {
	if initSize < 1 {
		return nil, errors.New("invalid pool size")
	}

	pool := &ConnPool{
		initSize:    initSize,
		connCreator: connCreator,
		conns:       make(chan net.Conn, initSize),
		valid:       true,
		cnt:         0,
	}

	for i := 0; i < initSize; i++ {
		conn, err := pool.createConn()
		if err != nil {
			return nil, err
		}
		pool.conns <- conn
	}
	return pool, nil
}

//will block if no idle conn in the pool
func (p *ConnPool) Get() (net.Conn, error) {
	if p.IsValid() == false {
		return nil, errors.New("Invalid Pool")
	}
	select {
	case conn := <-p.conns:
		return WrapConn(conn, p), nil
	}
}
func (p *ConnPool) Len() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	return len(p.conns)
}

// close the pool
func (p *ConnPool) Close() error {
	if p.IsValid() == false {
		return errors.New("Invalid Pool")
	}
	{
		p.lock.Lock()
		defer p.lock.Unlock()
		p.valid = false
		close(p.conns)
	}
	for conn := range p.conns {
		conn.Close()
	}
	return nil
}

func (p *ConnPool) Recycle(conn net.Conn) error {
	if p.IsValid() == false {
		return errors.New("Invalid Pool")
	}

	if conn == nil {
		p.lock.Lock()
		defer p.lock.Unlock()
		p.cnt--
		return errors.New("Closed Conn cannot be recycled")
	}
	fmt.Println("recycle connection")
	select {
	case p.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (p *ConnPool) IsValid() bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.valid
}

func (p *ConnPool) Destroy(conn net.Conn) error {
	if p.IsValid() == false {
		return errors.New("Closed Pool")
	}

	p.lock.Lock()
	p.cnt -= 1
	p.lock.Unlock()
	switch conn.(type) {
	case *Conn:
		return conn.(*Conn).Close()
	default:
		return conn.Close()
	}
	return nil
}

// use conncreator to create a new connection and add it to the pool
func (p *ConnPool) createConn() (net.Conn, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.cnt >= p.initSize {
		return nil, fmt.Errorf("Reached the cap limit of the pool %d", p.initSize)
	}
	conn, err := p.connCreator()
	if err != nil {
		return nil, err
	}
	p.cnt++
	return conn, nil
}
