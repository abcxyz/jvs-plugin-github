// package main

// import (
// 	"fmt"
// 	"log"
// 	"regexp"
// )

// // // readPrivateKey reads a RSA encrypted private key using PEM encoding as a string and returns an RSA key.
// // // func readPrivateKey(rsaPrivateKeyPEM string) (*rsa.PrivateKey, error) {
// // // 	parsedKey, _, err := jwk.DecodePEM([]byte(rsaPrivateKeyPEM))
// // // 	if err != nil {
// // // 		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
// // // 	}
// // // 	privateKey, ok := parsedKey.(*rsa.PrivateKey)
// // // 	if !ok {
// // // 		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
// // // 	}
// // // 	return privateKey, nil
// // // }

// // // func main() {
// // // 	_, err := readPrivateKey("abcde")
// // // 	if err != nil {
// // // 		fmt.Println(err)
// // // 		return
// // // 	}
// // // 	fmt.Println("pass")
// // // }

// const (
// 	PEM = `-----BEGIN RSA PRIVATE KEY-----
// MIIEowIBAAKCAQEAqjKM0m+E4TKiCEO2sZINFpd3PrbjLK75PIjIdftEQlRKY+G5
// eNC/sX7zDyavDIu0Qyhkm48WKLTJfeHRabrJNFi145iYHXwpfCvwcrUhhcJouuXL
// 2MAjHnLDj4h3NG4gq7hKmV5P0ez1ZUFF/UuI/RtgWmql9lKmtPztC8URAKETjzlZ
// TO50fwp9eFi9Ix6jRVfjnwO2njjAGvELJi17a3PJ9Bu1NkATEEiTRBRwf0MxIuat
// xq/871QvyhssMa7Eyimbojfc44zSAUAi6DknVuCGlkJc9kb4kf+JYiS8uYtUKHv0
// M9QOmHRTzqN1o0AmUKbhNrjAPEwJKZnx6VUrnQIDAQABAoIBAB0crG3KWYZTrNeR
// DYzuGIMGwYTer5kTDNrH4tIbep+F33uaPqllu4RW3Kh4y3Rv6RObsynQnc+0rMp1
// d+aH5qanjeUyUsKoXEg7E9PrG6LPkC535BhdNSSeKMlCZHF5bOkyisAVG74itA7v
// zVL4OqRgrGiK2Xx6wr0ujjV3LeNXYqJaDvL1jVJT8MK27uZfMj1gYBXznwGK+kj+
// CFmM90AHMHDsPxInaXbdeVyojE9xOV2faJPtRFsP6hpdWelZtV1xzBwb67TFPtsA
// 78PdW+71sayE32+l1ZaQFydnBMkUghgaYPSmEn2K9I4YoUnZ5XR9Wx/0BF8PpX+e
// 3Yg3WYECgYEA0VwAXaTdGyUjXIe94TWVYFFg5A3cMkf3SJE1zOTh9Bv5fDDYSFyU
// FiZXnw2HhZj9yVE8+uXrUgEXgGnKEfT/U3Tl0w8G9Ek+jtaaLg6yGJBVnWDrsswc
// DItWDEGyNcHXhfQnNDuaGQ3P8PseFr/cp/U5vBOKj9f9du1/Rvx6q9ECgYEA0B0Z
// zdKfwKraaBWP0J3hH5avfGKeqEFiIJ88g6SDddcp6F1xAeth0qjVxeWqLBNoPGjq
// dY28wn2ZmqJT1irhnx4E2wd5oAMDaXshNjCxT3EkRKRQCwa93wAARmA0Igbh0UvX
// lI5sT55XDv3JnvUJiFnPK1GyFjI/k9mfsDLZ0g0CgYEAwQ7Hv5LR2cBLdX4vGMgi
// sSkZ4fLuBOfcHmzZYdIGkuZhD6azKzdDz5EX57HAMPA9xzFEvFDcyUf8dgwXrKtx
// 73GypQgMb6RDLdCzaJlgncorSO8hKkWR7/dlJ/RE89GGfx4AMOhtV4EnKZ9Hxc6z
// GabG0KpscezI7KxhXAJi1KECgYBwKL7XZkQimfHLVpODYxMI6zT4XE4Vb+dqnWcH
// q4oN4D/9sx5MYob9+W/8j6H+zxbGN+TkJdctGnPGGuYD7mhaUNtdD9JEolscZfeo
// NOXaYqehNszMpH1/yYhcZUyzafIZ0j4FGhzVbAiPU8dtm7HfgkdcmVLZE4ugKxEc
// 7MrnoQKBgHB7Qnz28YY4jy6LLMFHuIuJ8fTcf9hu7r0GRp6be9Ja88nZdQuyMMLn
// w5JQqSJhI2FPt6AhQRXISL34Gr0blEWVmJk9gl0R0H+mFHvVJqDEtI7kxtuu6ypb
// GwcRVVm6XCdenKI/bJw7/lskAybs7gcSjViV2uCPm10xVrGACqCm
// -----END RSA PRIVATE KEY-----
// `
// 	APP_ID_STRING          = "392946"
// 	INSTALLATION_ID_STRING = "42028900"
// )

// func main() {
// 	tryParseStringWithRegExp()
// }

// func tryParseStringWithRegExp() {
// 	issueURLPatternRegExp := `^https:\/\/github.com\/(?P<owner>[a-zA-Z0-9-]*)\/(?P<repoName>[a-zA-Z0-9-]*)\/issues\/(?P<issue>[0-9]+$)`

// 	r, err := regexp.Compile(issueURLPatternRegExp)
// 	if err != nil {
// 		log.Fatal("Error compiling: ", err)
// 	}
// 	m := r.FindStringSubmatch("https://github.com/abcxyz/jvs-plugin-github/issues/4")
// 	result := make(map[string]string)
// 	for i, name := range r.SubexpNames() {
// 		if i != 0 && name != "" {
// 			result[name] = m[i]
// 		}
// 	}
// 	fmt.Println(result)

// }

// // // readPrivateKey reads a RSA encrypted private key using PEM encoding as a string
// // // and returns an RSA key.
// // // func readPrivateKey(rsaPrivateKeyPEM string) (*rsa.PrivateKey, error) {
// // // 	parsedKey, _, err := jwk.DecodePEM([]byte(rsaPrivateKeyPEM))
// // // 	if err != nil {
// // // 		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
// // // 	}
// // // 	privateKey, ok := parsedKey.(*rsa.PrivateKey)
// // // 	if !ok {
// // // 		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
// // // 	}
// // // 	return privateKey, nil
// // // }

// // func tryIfWork() {
// // 	cfg := &plugin.PluginConfig{
// // 		GitHubAppID:             APP_ID_STRING,
// // 		GitHubAppInstallationID: INSTALLATION_ID_STRING,
// // 		GitHubAppPrivateKeyPEM:  PEM,
// // 	}

// // 	ghClient := github.NewClient(nil)
// // 	pk, _ := readPrivateKey(cfg.GitHubAppPrivateKeyPEM)
// // 	ghAppCfg := githubapp.NewConfig(cfg.GitHubAppID, cfg.GitHubAppInstallationID, pk)
// // 	ghApp := githubapp.New(ghAppCfg)

// // 	v := plugin.NewValidator(ghClient, ghApp)
// // 	ctx := context.Background()
// // 	if err := v.MatchIssue(ctx, "https://github.com/qh-org/gh-app-test/issues/1"); err != nil {
// // 		fmt.Println(err)
// // 	} else {
// // 		fmt.Println("validation passed")
// // 	}
// // }

// // func fakeGet() {
// // 	fmt.Println("123")
// // }

// // func test() {
// // 	err := testFunc(nil)
// // 	if err != nil {
// // 		fmt.Errorf("error")
// // 	}
// // }

// // func testFunc(f func()) error {
// // 	if f == nil {
// // 		return fmt.Errorf("func is nil")
// // 	}
// // 	return nil
// // }

// // func test3() {
// // 	ctx := context.Background()
// // 	hc, done := newTestServer(testHandleObjectRead([]byte("test")))
// // 	defer done()
// // 	client := github.NewClient(hc)
// // 	client.Issues.Get(ctx, "qh-org", "gh-app-test", 1)
// // }

// // func newTestServer(handler func(w http.ResponseWriter, r *http.Request)) (*http.Client, func()) {
// // 	ts := httptest.NewTLSServer(http.HandlerFunc(handler))
// // 	// Need insecure TLS option for testing.
// // 	// #nosec G402
// // 	tlsConf := &tls.Config{InsecureSkipVerify: true}
// // 	tr := &http.Transport{
// // 		TLSClientConfig: tlsConf,
// // 		DialTLS: func(netw, addr string) (net.Conn, error) {
// // 			return tls.Dial("tcp", ts.Listener.Addr().String(), tlsConf)
// // 		},
// // 	}
// // 	return &http.Client{Transport: tr}, func() {
// // 		tr.CloseIdleConnections()
// // 		ts.Close()
// // 	}
// // }

// // func testHandleObjectRead(data []byte) func(w http.ResponseWriter, r *http.Request) {
// // 	return func(w http.ResponseWriter, r *http.Request) {
// // 		fmt.Println(r.URL.Path)
// // 		switch r.URL.Path {
// // 		// This is for getting object info
// // 		case "/foo/pmap-test/gh-prefix/dir1/dir2/bar":
// // 			_, err := w.Write(data)
// // 			if err != nil {
// // 				panic("failed to write response for object info")
// // 			}
// // 		default:
// // 			http.Error(w, "injected error", http.StatusNotFound)
// // 		}
// // 	}
// // }
