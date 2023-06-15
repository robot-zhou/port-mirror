package main

import (
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

var (
	SessionMutex sync.Mutex
	SessionCount int
)

func SessionIncCount() int {
	SessionMutex.Lock()
	SessionCount++
	SessionMutex.Unlock()
	return SessionCount
}

func SessionDecCount() int {
	SessionMutex.Lock()
	SessionCount--
	SessionMutex.Unlock()
	return SessionCount
}

type Session struct {
	MirrorConfig MirrorConfig
	Mutex        sync.Mutex
	end          bool

	Id               string
	DownConn         net.Conn
	LastDownDataTime int64
	DownDataSize     int64
	UpConn           net.Conn
	LastUpDataTime   int64
	UpDataSize       int64
}

func (sess *Session) IsAliveTimeout() bool {
	nowts := time.Now().Unix()
	if nowts-sess.LastDownDataTime > int64(sess.MirrorConfig.AliveTimeout) && nowts-sess.LastUpDataTime > int64(sess.MirrorConfig.AliveTimeout) {
		return true
	}
	return false
}

func (sess *Session) UpConnRoutine() {
	var (
		buf      [1024]byte
		readCnt  int
		readErr  error
		writeErr error
		aliveErr error
		exit     bool
	)

	defer sess.DownConn.Close()
	defer sess.UpConn.Close()

	for !exit {
		sess.UpConn.SetReadDeadline(time.Now().Add(time.Duration(sess.MirrorConfig.ReadTimeout) * time.Second))
		readCnt, readErr = sess.UpConn.Read(buf[:])
		if readErr != nil {
			if errors.Is(readErr, os.ErrDeadlineExceeded) {
				if !sess.IsAliveTimeout() {
					continue
				}

				aliveErr = errors.New("alive timeout")
				readErr = nil
			}

			exit = true
		}

		if readCnt > 0 {
			sess.LastUpDataTime = time.Now().Unix()
			sess.DownConn.SetWriteDeadline(time.Now().Add(time.Duration(sess.MirrorConfig.WriteTimeout) * time.Second))
			if _, writeErr = sess.DownConn.Write(buf[0:readCnt]); writeErr != nil {
				exit = true
			} else {
				sess.DownDataSize += int64(readCnt)
			}
		}
	}

	sess.Mutex.Lock()
	if !sess.end {
		SessionDecCount()
		if aliveErr != nil {
			log.Infof("[%s} session exit, alive time out", sess.Id)
		} else if readErr != nil {
			if readErr == io.EOF {
				log.Infof("[%s] session exit, up connect close", sess.Id)
			} else {
				log.Infof("[%s] session exit, up connect read error: %s", sess.Id, readErr.Error())
			}
		} else if writeErr != nil {
			log.Infof("[%s] session exit, down connect write error: %s", sess.Id, readErr.Error())
		}
		sess.end = true
	}
	sess.Mutex.Unlock()

	log.Infof("[%s] up connect exit, down data size: %d, session count(%d)", sess.Id, sess.DownDataSize, SessionCount)
}

func (sess *Session) ConnectTarget() error {
	var (
		err              error
		dailer           proxy.Dialer
		network, address string = ParseConfigAddress(sess.MirrorConfig.Target)
	)

	proxys := SplitStringTrim(sess.MirrorConfig.Proxy)

	if len(proxys) == 0 {
		dailer = proxy.Direct
	}

	sess.UpConn, err = dailer.Dial(network, address)
	return err
}

func (sess *Session) DownConnRoutine() {

	SessionIncCount()
	log.Infof("[%s] start session client(%s)->local(%s) target %s, session count(%d)", sess.Id, sess.DownConn.RemoteAddr().String(), sess.DownConn.LocalAddr().String(), sess.MirrorConfig.Target, SessionCount)

	if err := sess.ConnectTarget(); err != nil {
		SessionDecCount()
		sess.DownConn.Close()
		log.Errorf("[%s] session exit, dail %s fail: %v, session count(%d)", sess.Id, sess.MirrorConfig.Target, err, SessionCount)
		return
	} else {
		log.Infof("[%s] connect local(%s)->target(%s) success", sess.Id, sess.UpConn.LocalAddr().String(), sess.UpConn.RemoteAddr().String())
	}

	sess.LastDownDataTime = time.Now().Unix()
	sess.LastUpDataTime = sess.LastDownDataTime

	go sess.UpConnRoutine()

	var (
		buf      [1024]byte
		readCnt  int
		readErr  error
		writeErr error
		aliveErr error
		exit     bool
	)

	defer sess.DownConn.Close()
	defer sess.UpConn.Close()

	for !exit {
		sess.DownConn.SetReadDeadline(time.Now().Add(time.Duration(sess.MirrorConfig.ReadTimeout) * time.Second))
		readCnt, readErr = sess.DownConn.Read(buf[:])
		if readErr != nil {
			if errors.Is(readErr, os.ErrDeadlineExceeded) {
				if !sess.IsAliveTimeout() {
					continue
				}

				aliveErr = errors.New("alive timeout")
				readErr = nil
			}

			exit = true
		}

		if readCnt > 0 {
			sess.LastDownDataTime = time.Now().Unix()
			sess.UpConn.SetWriteDeadline(time.Now().Add(time.Duration(sess.MirrorConfig.WriteTimeout) * time.Second))
			if _, writeErr = sess.UpConn.Write(buf[0:readCnt]); writeErr != nil {
				exit = true
			} else {
				sess.UpDataSize += int64(readCnt)
			}
		}
	}

	sess.Mutex.Lock()
	if !sess.end {
		if aliveErr != nil {
			log.Infof("[%s} session exit, alive time out", sess.Id)
		} else if readErr != nil {
			if readErr == io.EOF {
				log.Infof("[%s] session exit, down connect close", sess.Id)
			} else {
				log.Infof("[%s] session exit, down connect read error: %s", sess.Id, readErr.Error())
			}
		} else if writeErr != nil {
			log.Infof("[%s] session exit, up connect write error: %s", sess.Id, readErr.Error())
		}
		sess.end = true
	}
	sess.Mutex.Unlock()

	log.Infof("[%s] down connect exit, up data size: %d, session count(%d)", sess.Id, sess.UpDataSize, SessionCount)
}

func (sess *Session) Start() {
	sess.Id = RandomString()
	go sess.DownConnRoutine()
}
