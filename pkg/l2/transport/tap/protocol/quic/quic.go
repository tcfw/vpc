package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"
	"golang.org/x/sys/unix"

	quic "github.com/lucas-clemente/quic-go"
)

const (
	nProto = "tcfw-vpc"
	port   = 8443
)

//Handler provides QUIC streams between endpoints
type Handler struct {
	quicSrv quic.Listener
	epConns map[string]*epConn
	port    int
	recv    protocol.HandlerFunc
}

type epConn struct {
	session quic.Session
	streams map[uint32]quic.Stream
}

//NewHandler inits a new QUIC protocol handler
func NewHandler() *Handler {
	return &Handler{
		epConns: map[string]*epConn{},
		port:    port,
		recv:    func(_ []protocol.Packet) {},
	}
}

//SetHandler sets the callback for receiving packets
func (p *Handler) SetHandler(handle protocol.HandlerFunc) {
	p.recv = handle
}

//Start opens the QUIC listener and starts waiting for sessions
func (p *Handler) Start() error {
	tlsConfig, err := p.tlsConfig()
	if err != nil {
		return err
	}

	lisConfig := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_PRIORITY, 0)
				if err != nil {
					return
				}
			})
			return err
		},
	}

	conn, err := lisConfig.ListenPacket(context.Background(), "udp", fmt.Sprintf(":%d", p.port))
	if err != nil {
		return err
	}

	lis, err := quic.Listen(conn, tlsConfig, nil)
	if err != nil {
		return err
	}

	p.quicSrv = lis

	go p.acceptSessions()

	return nil
}

//Stop closes all sessions and the quic server
func (p *Handler) Stop() error {
	return p.quicSrv.Close()
}

//Send sends a frame to an endpoint
func (p *Handler) Send(packets []protocol.Packet, addr net.IP) (int, error) {
	addrID := addr.String()
	conn, ok := p.epConns[addrID]
	if !ok {
		if err := p.openConn(addrID); err != nil {
			return 0, err
		}
		conn = p.epConns[addrID]
	}

	var n int

	for _, packet := range packets {
		stream, ok := conn.streams[packet.VNID]
		if !ok {
			nStream, err := p.openStream(conn, packet.VNID)
			if err != nil {
				return 0, err
			}
			stream = nStream
		}

		for _, packet := range packets {
			ni, err := stream.Write(packet.Frame)
			if err != nil {
				//Reopen session
				if err.Error() == "NO_ERROR: No recent network activity" {
					stream.Close()
					delete(p.epConns, addrID)
					return p.Send(packets, addr)
				}
			}
			n += ni
		}
	}

	return n, nil
}

func (p *Handler) openStream(conn *epConn, vnid uint32) (quic.Stream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	helloStream, err := conn.session.OpenUniStreamSync(ctx)
	if err != nil {
		return nil, err
	}

	helloStream.Write(uint32Bytes(&vnid))

	helloStream.Close()

	nStream, err := conn.session.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}

	conn.streams[vnid] = nStream
	return nStream, nil
}

func uint32Bytes(data *uint32) []byte {
	return (*[4]byte)(unsafe.Pointer(data))[:]
}

//openConn opens a new connection to a remote endpoint
func (p *Handler) openConn(addr string) error {
	addrS := addr

	rdst := (&net.UDPAddr{IP: net.ParseIP(addr), Port: port}).String()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{nProto},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := quic.DialAddrContext(ctx, rdst, tlsConfig, nil)
	if err != nil {
		return err
	}

	p.epConns[addrS] = &epConn{
		session: conn,
		streams: map[uint32]quic.Stream{},
	}

	return nil
}

//tlsConfig generates a temporary TLS config
func (p *Handler) tlsConfig() (*tls.Config, error) {
	//TODO(tcfw) - taken from the quic-go echo example; need to implement reading certs from files
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{nProto},
	}, nil
}

//acceptSessions accepts the session requests and starts a session handler
func (p *Handler) acceptSessions() {
	for {
		sess, err := p.quicSrv.Accept(context.Background())
		if err != nil {
			log.Printf("Accept failed: %s", err)
			return
		}

		epID := sess.RemoteAddr().String()
		if _, ok := p.epConns[epID]; ok {
			sess.CloseWithError(0, "already connected")
			return
		}

		p.epConns[epID] = &epConn{
			session: sess,
			streams: make(map[uint32]quic.Stream),
		}

		go p.handleSession(sess)
	}
}

//handleSession processes session requsts
func (p *Handler) handleSession(sess quic.Session) {
	helloFrame, err := p.getHello(sess)
	if err != nil {
		return
	}

	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			return
		}

		go p.handleStream(helloFrame, stream)
	}
}

//getHello waits for a hello frame, with timeout
func (p *Handler) getHello(sess quic.Session) (*helloFrame, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	frame := make(chan *helloFrame)
	var err error

	go func() {
		stream, err := sess.AcceptUniStream(ctx)
		if err != nil {
			return
		}

		buf := make([]byte, 4)
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return
		}

		helloFrame := &helloFrame{
			vnid: binary.LittleEndian.Uint32(buf[:n]),
		}

		frame <- helloFrame
	}()

	select {
	case hello, ok := <-frame:
		if !ok {
			return nil, err
		}
		return hello, nil
	case <-time.After(3 * time.Second):
		log.Println("hello timeout")
		return nil, fmt.Errorf("hello timeout")
	}
}

//handleStream reads in frames from remote eps
func (p *Handler) handleStream(hello *helloFrame, stream quic.Stream) {
	buf := make([]byte, 81920)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			return
		}

		p.recv([]protocol.Packet{{VNID: hello.vnid, Frame: buf[:n]}})
	}
}
