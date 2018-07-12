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

	client, err := connect(name, ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	err = setVolume(client, ctx, vol, false)
	if err != nil {
		return err
	}

	err = playMedia(client, ctx, url)
	if err != nil {
		return err
	}

	return setVolume(client, ctx, 0.5, false)
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

	return getStatus(client, ctx)
}

func setVolume(client *cast.Client, ctx context.Context, vol float64, muted bool) error {
	volume := &controllers.Volume{Level: &vol, Muted: &muted}
	_, err := client.Receiver().SetVolume(ctx, volume)
	return err
}

func playMedia(client *cast.Client, ctx context.Context, url string) error {
	media, err := client.Media(ctx)
	if err != nil {
		return err
	}

	audio := controllers.MediaItem{
		ContentId:   url,
		StreamType:  "BUFFERED",
		ContentType: "audio/mpeg",
	}

	_, err = media.LoadMedia(ctx, audio, 0, true, map[string]interface{}{})
	return err
}

func getStatus(client *cast.Client, ctx context.Context) (string, error) {
	status, err := client.GetMediaStatus(ctx)
	if err != nil {
		if err.Error() == "no media" {
			return "IDLE", nil
		}
		return "", err
	}

	for _, m := range status {
		return m.PlayerState, nil
	}
	return "IDLE", nil
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
