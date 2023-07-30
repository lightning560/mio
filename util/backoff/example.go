package backoff

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func ExampleBackoff() {
	b := &Backoff{
		Max: 5 * time.Minute,
	}

	for {
		conn, err := net.Dial("tcp", "example.com:5309")
		if err != nil {
			d := b.Duration()
			fmt.Printf("%s, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}
		//connected
		b.Reset()
		conn.Write([]byte("hello world!"))
		// ... Read ... Write ... etc
		conn.Close()
		//disconnected
	}

}

func ExampleJitterBackoff() {
	b := &Backoff{
		Jitter: true,
	}

	rand.Seed(42)

	fmt.Printf("%s\n", b.Duration())
	fmt.Printf("%s\n", b.Duration())
	fmt.Printf("%s\n", b.Duration())

	fmt.Printf("Reset!\n")
	b.Reset()

	fmt.Printf("%s\n", b.Duration())
	fmt.Printf("%s\n", b.Duration())
	fmt.Printf("%s\n", b.Duration())
}
