package protocol

import (
	"fmt"
	"io"
	"net"
)

var sessionGenerator = 0

func NewSession(con net.Conn, infoChan chan<- string, errorChan chan<- error) Session {
	session := &session{
		id:         fmt.Sprintf("%d", sessionGenerator),
		connection: con,
		errorChan:  errorChan,
		infoChan:   infoChan,
	}

	sessionGenerator++
	return session
}

type Session interface {
	Id() string
	Close()
	Read(readChan chan []byte)
	Write(msg string) bool
	LogError(err error)
	LogInfo(info string)
}

type session struct {
	id         string
	connection net.Conn
	errorChan  chan<- error
	infoChan   chan<- string
}

func (s *session) Id() string {
	return s.id
}

func (s *session) Close() {
	err := s.connection.Close()
	if err != nil {
		s.LogError(err)
	}
}

func (s *session) Read(readChan chan []byte) {
	chunkBuffSize := 32
	chunkBuffer := make([]byte, 32)

	buffer := make([]byte, 0, 256)

	for {
		l, err := s.connection.Read(chunkBuffer)

		if err != nil {
			if err != io.EOF {
				s.LogError(fmt.Errorf("Cannot read message: %v", err))
			}
			close(readChan)
			return
		}

		buffer = append(buffer, chunkBuffer[:l]...)

		if l < chunkBuffSize {
			break
		}
	}

	readChan <- buffer
}

func (s *session) Write(msg string) bool {
	msgBytes := []byte(msg)
	l, err := s.connection.Write(msgBytes)

	if err != nil {
		s.LogError(fmt.Errorf("Could not send message '%s': %s", msg, err.Error()))
		return false
	}

	if l != len(msgBytes) {
		s.LogError(fmt.Errorf("Could not send message '%s' completely, just %d bytes", msg, l))
		return false
	}

	return true
}

func (s *session) LogError(err error) {
	s.errorChan <- fmt.Errorf("SESSION[%s]: %v", s.id, err)
}

func (s *session) LogInfo(info string) {
	s.infoChan <- fmt.Sprintf("SESSION[%s]:%s", s.id, info)
}
