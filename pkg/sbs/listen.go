package sbs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

//Listen starts listening for remote connections
func (s *Server) Listen() error {
	return s.listen(s.listenPort)
}

//Listen starts listening
func (s *Server) listen(port int) error {
	s.log.Printf("PeerID: %s", s.peerID)

	tls, err := tlsConfig()
	if err != nil {
		return fmt.Errorf("failed to provision tls: %s", err)
	}

	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to open port: %s", err)
	}
	s.pConn = pc

	l, err := quic.Listen(s.pConn, tls, defaultQuicConfig())
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}

	s.listener = l

	s.log.WithFields(logrus.Fields{"port": port}).Info("Listening for peers")

	go s.accept()

	if s.controller != nil {
		err := s.controller.Register(s.peerID, s.listener.Addr().(*net.UDPAddr))
		if err != nil {
			s.log.WithError(err).Error("failed to register with controller")
		}
	}

	return nil
}

//defaultQuicConfig used for setting up peers
func defaultQuicConfig() *quic.Config {
	return &quic.Config{
		MaxIncomingStreams: 10000,
		KeepAlive:          true,
	}
}

//tlsConfig TODO(tcfw) use actual tls certs
func tlsConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"quic-vpc-sbs"},
	}, nil
}

//accept handles new QUIC connections - pre-handshake
func (s *Server) accept() {
	for {
		select {
		case <-s.shutdownCh:
			return
		default:
		}
		sess, err := s.listener.Accept(context.Background())
		if err != nil {
			s.log.WithFields(logrus.Fields{"err": err}).Warn("failed to accept conn to peer")
			continue
		}

		s.log.WithFields(logrus.Fields{"addr": sess.RemoteAddr().String()}).Debug("new connection")

		if sess.RemoteAddr().(*net.UDPAddr).IP.IsLoopback() {
			s.log.Warnf("connecting with local peer")
			go s.negotiateLocalPeer(sess)
		} else {
			go s.negotiate(sess, false)
		}
	}
}
