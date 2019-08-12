package netpool

import (
	"errors"
	"net"
)

type Conn struct {
	net.Conn
	pool *ConnPool
}

//Terminate the connection and remove it from the connection pool
func (conn *Conn) Terminate() error {
	if conn.pool == nil {
		conn.Close()
		return errors.New("Orphan connection")
	}
	err := conn.pool.Destroy(conn.Conn)
	if err != nil {
		return err
	}
	conn.pool = nil
	return nil
}

// the work is done, put the connection back to the pool
func (conn *Conn) Close() error {
	if conn.pool == nil {
		conn.Close()
	}
	return conn.pool.Recycle(conn.Conn)
}

func WrapConn(conn net.Conn, p *ConnPool) net.Conn {
	return &Conn{conn, p}
}
