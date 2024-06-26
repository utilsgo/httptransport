package httptransport_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/utilsgo/httptransport"
	"github.com/utilsgo/httptransport/client"
	"github.com/utilsgo/httptransport/testdata/server/cmd/app/routes"
)

func BenchmarkHttpTransport(b *testing.B) {
	ht := httptransport.NewHttpTransport(func(server *http.Server) error {
		server.ReadTimeout = 15 * time.Second
		return nil
	})
	ht.SetDefaults()
	ht.Port = 8080
	go func() {
		_ = ht.Serve(routes.RootRouter)
	}()

	b.Run("request", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = http.Get("http://127.0.0.1:8080/demo/restful/123456")
		}
	})

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
}

func TestHttpTransport(t *testing.T) {
	ht := httptransport.NewHttpTransport(func(server *http.Server) error {
		server.ReadTimeout = 15 * time.Second
		return nil
	})
	ht.SetDefaults()
	ht.Port = 8080
	go func() {
		_ = ht.Serve(routes.RootRouter)
	}()

	time.Sleep(1 * time.Second)

	t.Run("request", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:8080/demo/restful/123456")
		NewWithT(t).Expect(err).To(BeNil())

		data, err := httputil.DumpResponse(resp, true)
		NewWithT(t).Expect(err).To(BeNil())
		fmt.Println(string(data))
	})

	t.Run("openapi", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:8080/demo")

		NewWithT(t).Expect(err).To(BeNil())

		data, err := httputil.DumpResponse(resp, true)
		NewWithT(t).Expect(err).To(BeNil())
		fmt.Println(string(data))
	})

	t.Run("proxy", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:8080/demo/proxy/v2")
		NewWithT(t).Expect(err).To(BeNil())

		data, err := httputil.DumpResponse(resp, true)
		NewWithT(t).Expect(err).To(BeNil())
		fmt.Println(string(data))
	})

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
}

func TestHttpTransportWithHTTP2(t *testing.T) {
	ht := httptransport.NewHttpTransport(
		func(server *http.Server) error {
			server.ReadTimeout = 15 * time.Second
			return nil
		},
	)

	ht.CertFile = "./testdata/certs/cert.pem"
	ht.KeyFile = "./testdata/certs/key.pem"

	rootCA, _ := ioutil.ReadFile(ht.CertFile)

	ht.Port = 8080

	go func() {
		_ = ht.Serve(routes.RootRouter)
	}()

	time.Sleep(500 * time.Millisecond)

	c := client.GetShortConnClient(10*time.Second, NewInsecureTLSTransport(rootCA))

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			_, err := c.Get("http://localhost:8080/demo/restful/123123123")
			NewWithT(t).Expect(err).To(BeNil())
		}()
	}

	wg.Wait()

	resp, err := c.Get("https://localhost:8080/demo/restful/123123123")
	NewWithT(t).Expect(err).To(BeNil())

	data, err := httputil.DumpResponse(resp, true)
	NewWithT(t).Expect(err).To(BeNil())
	fmt.Println(string(data))

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
}

func TestHttpTransportWithTLS(t *testing.T) {
	ht := httptransport.NewHttpTransport(func(server *http.Server) error {
		server.ReadTimeout = 15 * time.Second
		return nil
	})

	ht.CertFile = "./testdata/certs/cert.pem"
	ht.KeyFile = "./testdata/certs/key.pem"

	rootCA, _ := ioutil.ReadFile(ht.CertFile)

	ht.SetDefaults()
	ht.Port = 8081

	go func() {
		_ = ht.Serve(routes.RootRouter)
	}()

	time.Sleep(200 * time.Millisecond)

	req, err := http.NewRequest("GET", "https://localhost:8081/demo/restful/1", nil)
	NewWithT(t).Expect(err).To(BeNil())

	resp, err := client.GetShortConnClient(10*time.Second, NewInsecureTLSTransport(rootCA)).Do(req)
	NewWithT(t).Expect(err).To(BeNil())

	data, err := httputil.DumpResponse(resp, true)
	NewWithT(t).Expect(err).To(BeNil())
	fmt.Println(string(data))

	time.Sleep(2 * time.Second)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
}

func NewInsecureTLSTransport(rootCA []byte) client.HttpTransport {
	return func(rt http.RoundTripper) http.RoundTripper {
		if httpRt, ok := rt.(*http.Transport); ok {
			if httpRt.TLSClientConfig == nil {
				httpRt.TLSClientConfig = &tls.Config{}
			}
			httpRt.TLSClientConfig.RootCAs = rootCertPool(rootCA)
			return httpRt
		}
		return rt
	}
}

func rootCertPool(caData []byte) *x509.CertPool {
	if len(caData) == 0 {
		return nil
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return certPool
}
