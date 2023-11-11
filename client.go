package dockermgr

import (
	"context"
	"github.com/docker/docker/client"
	"log"
	"net"
)

func DockerClient(apiVersion string) *client.Client {
	if apiVersion == "" {
		apiVersion = "1.39"
	}
	dockerClient, err := client.NewClientWithOpts(
		//client.WithHost("tcp://192.168.124.84:2375"),
		client.WithDialContext(func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		}),
		client.WithVersion(apiVersion))
	if err != nil {
		log.Fatal(err)
	}
	ping, err := dockerClient.Ping(context.Background())
	//}))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ping.APIVersion)
	return dockerClient
}
