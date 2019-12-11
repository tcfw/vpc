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
	"time"

	"github.com/tcfw/vpc/pkg/l2/transport/tap/protocol"

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
	in      chan *protocol.Packet
	port    int
}

type epConn struct {
	session quic.Session
	streams map[uint32]quic.Stream
}

//NewHandler inits a new QUIC protocol handler
func NewHandler() *Handler {
	return &Handler{
		epConns: map[string]*epConn{},
		in:      make(chan *protocol.Packet, 1000),
		port:    port,
	}
}

//Start opens the QUIC listener and starts waiting for sessions
func (p *Handler) Start() error {
	tlsConfig, err := p.tlsConfig()
	if err != nil {
		return err
	}

	lis, err := quic.ListenAddr(fmt.Sprintf(":%d", p.port), tlsConfig, nil)
	if err != nil {
		return err
	}

	p.quicSrv = lis

	go p.acceptSessions()

	return nil
}

//Stop closes all sessions and the quic server
func (p *Handler) Stop() error {
	for _, ep := range p.epConns {
		ep.session.Close()
	}
	return p.quicSrv.Close()
}

//Send sends a frame to an endpoint
func (p *Handler) Send(packet *protocol.Packet, addr net.Addr) error {
	epID := addr.String()
	conn, ok := p.epConns[epID]
	if !ok {
		if err := p.openConn(addr); err != nil {
			return err
		}
		conn = p.epConns[epID]
	}

	stream, ok := conn.streams[packet.VNID]

	if !ok {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		helloStream, err := conn.session.OpenUniStreamSync(ctx)
		if err != nil {
			return err
		}

		hello := make([]byte, 4)
		binary.BigEndian.PutUint32(hello, packet.VNID)
		helloStream.Write(hello)

		helloStream.Close()

		nStream, err := conn.session.OpenStreamSync(ctx)
		if err != nil {
			return err
		}

		conn.streams[packet.VNID] = nStream
		stream = nStream
	}
	_, err := stream.Write(packet.Frame)
	if err != nil {
		//Reopen session
		if err.Error() == "NO_ERROR: No recent network activity" {
			stream.Close()
			conn.session.Close()
			delete(p.epConns, epID)
			return p.Send(packet, addr)
		}
	}

	return err
}

//openConn opens a new connection to a remote endpoint
func (p *Handler) openConn(addr net.Addr) error {
	addrS := addr.String()

	rdst := (&net.UDPAddr{IP: net.ParseIP(addr.String()), Port: port}).String()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{nProto},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

//Recv waits for a new frame from a remote endpoint
func (p *Handler) Recv() (*protocol.Packet, error) {
	packet, ok := <-p.in
	if !ok {
		return nil, fmt.Errorf("receiving channel is closed")
	}

	return packet, nil
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
			streams: map[uint32]quic.Stream{},
		}

		go p.handleSession(sess)
	}
}

//handleSession processes session requsts
func (p *Handler) handleSession(sess quic.Session) {
	helloFrame, err := p.getHello(sess)
	if err != nil {
		sess.Close()
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

		vnid := binary.BigEndian.Uint32(buf[:n])

		frame <- &helloFrame{
			vnid: vnid,
		}
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
	buf := make([]byte, 1500)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			return
		}

		p.in <- &protocol.Packet{VNID: hello.vnid, Frame: buf[:n]}
	}
}