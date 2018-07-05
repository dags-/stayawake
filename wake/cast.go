package wake

import (
	"time"

	"github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/controllers"
	"golang.org/x/net/context"

	"github.com/barnybug/go-cast/discovery"
)

func PlayAudio(name, url string, vol float64) (error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()

	// get client
	client, media, err := getDevice(name, ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// set volume
	muted := false
	volume := &controllers.Volume{Level: &vol, Muted: &muted}
	if _, err = client.Receiver().SetVolume(ctx, volume); err != nil {
		return err
	}

	item := controllers.MediaItem{
		ContentId:   url,
		StreamType:  "BUFFERED",
		ContentType: "audio/mpeg",
	}

	if _, err = media.LoadMedia(ctx, item, 0, true, map[string]interface{}{}); err != nil {
		return err
	}

	return nil
}

func GetPlayerState(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()

	// get client
	client, err := connect(name, ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// find media status
	status, err := client.GetMediaStatus(ctx)
	if err != nil {
		if err.Error() == "no application" {
			return "IDLE", nil
		}
		return "", err
	}

	for _, m := range status {
		return m.PlayerState, nil
	}

	return "IDLE", nil
}

func getDevice(name string, ctx context.Context) (*cast.Client, *controllers.MediaController, error) {
	c, e := connect(name, ctx)
	if e != nil || c == nil {
		return nil, nil, e
	}
	m, e := c.Media(ctx)
	return c, m, e
}

func connect(name string, ctx context.Context) (*cast.Client, error) {
	var client *cast.Client
	service := discovery.NewService(ctx)
	go service.Run(ctx, 2*time.Second)

LOOP:
	for {
		select {
		case c := <-service.Found():
			if c.Name() == name {
				client = c
				break LOOP
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	err := client.Connect(ctx)
	if err != nil {
		return nil, ctx.Err()
	}

	return client, nil
}
