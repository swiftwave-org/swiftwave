package ssh_toolkit

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func testPrivateKey(t *testing.T) string {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(key, "")
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(block))
}

func poolSize() int {
	sshClientPool.mutex.RLock()
	defer sshClientPool.mutex.RUnlock()
	return len(sshClientPool.clients)
}

// waitForEmptyPool tolerates the asynchronous close of evicted entries
func waitForEmptyPool(t *testing.T) {
	t.Helper()
	for range 100 {
		if poolSize() == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("pool still holds %d entries", poolSize())
}

// listenAndIgnore accepts connections but never speaks ssh, which is how a server under
// load or a tarpitting firewall behaves. ssh.Dial would block on the handshake forever.
func listenAndIgnore(t *testing.T) (string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mutex sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			mutex.Lock()
			conns = append(conns, conn)
			mutex.Unlock()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		mutex.Lock()
		defer mutex.Unlock()
		for _, conn := range conns {
			_ = conn.Close()
		}
	})
	addr := listener.Addr().(*net.TCPAddr)
	return addr.IP.String(), addr.Port
}

func TestHandshakeDoesNotHang(t *testing.T) {
	original := sshTCPTimeoutSeconds
	sshTCPTimeoutSeconds = 1
	t.Cleanup(func() { sshTCPTimeoutSeconds = original })

	host, port := listenAndIgnore(t)
	start := time.Now()
	if _, err := getSSHClientWithOptions(host, port, "root", testPrivateKey(t), false); err == nil {
		t.Fatal("expected an error from a server that never completes the handshake")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("handshake took %s, it should be bounded by the tcp timeout", elapsed)
	}
	waitForEmptyPool(t)
}

// A stuck handshake used to hold a write lock that every other caller for the same host
// blocked on, and a concurrent eviction could stall the whole pool behind it.
func TestConcurrentCallersDoNotBlockOnStuckHandshake(t *testing.T) {
	original := sshTCPTimeoutSeconds
	sshTCPTimeoutSeconds = 1
	t.Cleanup(func() { sshTCPTimeoutSeconds = original })

	host, port := listenAndIgnore(t)
	key := testPrivateKey(t)
	done := make(chan struct{})
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_, _ = getSSHClientWithOptions(host, port, "root", key, false)
			DeleteSSHClient(host)
		})
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("concurrent callers deadlocked")
	}
	waitForEmptyPool(t)
}

// Rotating the key must not keep serving the connection opened with the old one.
func TestRotatedCredentialsReplacePoolEntry(t *testing.T) {
	host := "127.0.0.1"
	entry, owned := acquireSSHClientEntry(host, 22, "root", credentialHash("old-key"))
	if !owned {
		t.Fatal("expected to own a freshly created entry")
	}
	close(entry.ready) // stand in for a completed handshake

	same, owned := acquireSSHClientEntry(host, 22, "root", credentialHash("old-key"))
	if owned || same != entry {
		t.Fatal("unchanged credentials should reuse the pooled entry")
	}

	rotated, owned := acquireSSHClientEntry(host, 22, "root", credentialHash("new-key"))
	if !owned || rotated == entry {
		t.Fatal("rotated credentials should replace the pooled entry")
	}
	close(rotated.ready)
	DeleteSSHClient(host)
	waitForEmptyPool(t)
}

func TestDifferentPortReplacesPoolEntry(t *testing.T) {
	host := "127.0.0.1"
	first, _ := acquireSSHClientEntry(host, 22, "root", credentialHash("key"))
	close(first.ready)
	second, owned := acquireSSHClientEntry(host, 2222, "root", credentialHash("key"))
	if !owned || second == first {
		t.Fatal("a different port should not reuse the pooled entry")
	}
	close(second.ready)
	DeleteSSHClient(host)
	waitForEmptyPool(t)
}

func TestDeleteSSHClientKeepsNewerEntry(t *testing.T) {
	host := "127.0.0.1"
	stale, _ := acquireSSHClientEntry(host, 22, "root", credentialHash("key"))
	close(stale.ready)
	current, _ := acquireSSHClientEntry(host, 2222, "root", credentialHash("key"))
	close(current.ready)

	deleteSSHClientEntry(host, stale)
	if poolSize() != 1 {
		t.Fatal("deleting a stale entry should leave the newer one in place")
	}
	deleteSSHClientEntry(host, current)
	waitForEmptyPool(t)
}

func TestCredentialHashDetectsChange(t *testing.T) {
	if credentialHash("a") == credentialHash("b") {
		t.Fatal("hash should differ per key")
	}
}
