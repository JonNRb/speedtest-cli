package speedtest

import (
	"context"
	"fmt"
	"io"
	"time"
)

const concurrentDownloadLimit = 6
const maxDownloadDuration = 10 * time.Second
const downloadBufferSize = 4096
const downloadRepeats = 5

var downloadImageSizes = []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

// Will probe download speed until enough samples are taken or ctx expires.
func (server Server) ProbeDownloadSpeed(ctx context.Context, client *Client) (BytesPerSecond, error) {
	pg := newProberGroup(concurrentDownloadLimit)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, size := range downloadImageSizes {
		for i := 0; i < downloadRepeats; i++ {
			url, err := server.RelativeURL(fmt.Sprintf("random%dx%d.jpg", size, size))
			if err != nil {
				return 0, fmt.Errorf("error parsing url for %v: %v", server, err)
			}

			pg.Add(func(url string) func() (bytesTransferred, error) {
				return func() (bytesTransferred, error) {
					return client.downloadFile(ctx, url)
				}
			}(url))
		}
	}

	return pg.Collect()
}

func (client *Client) downloadFile(ctx context.Context, url string) (bytesTransferred, error) {
	var t bytesTransferred

	// Check early failure where context is already canceled.
	select {
	case <-ctx.Done():
		return t, ctx.Err()
	default:
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resp, err := client.get(ctx, url)
	if err != nil {
		return t, err
	}

	go func() {
		<-ctx.Done()
		resp.Body.Close()
	}()

	var buf [downloadBufferSize]byte
	for {
		read, err := resp.Body.Read(buf[:])
		t += bytesTransferred(read)
		if err != nil {
			if err != io.EOF {
				return t, err
			}
			break
		}
	}
	return t, nil
}