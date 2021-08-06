package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	hvLayersCmd  *exec.Cmd
	dockerClient *client.Client
)

func init() {
	dclt, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		dockerClient = dclt
	} else {
		log.Fatal("Can't use docker..")
	}

	// check synerex-network

}

func runNodeSrv() {
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}
	cmdSlice := []string{"-addr", "0.0.0.0"}
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "synerex/nodeserv",
		Cmd:   cmdSlice,
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "geo")
	if err != nil {
		panic(err)
	}
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

}

func runSxServ() {

}

func runGeoDockerWithOSExec(cmds ...string) {
	baseCmd := []string{"run", "--network", "synerex-network", "synerex_geography", "-nodesrv", "nodeserv:9990"}
	cc := append(baseCmd, cmds...)
	geoCmd := exec.Command("docker", cc...)
	err := geoCmd.Start()
	if err != nil {
		log.Fatal("Can't start geo-provider docker", err)
	}
}

func runGeoDocker(cmds ...string) {

	baseCmd := []string{"run", "--network", "synerex-network", "synerex_geography", "-nodesrv", "nodeserv:9990"}
	cc := append(baseCmd, cmds...)
	geoCmd := exec.Command("docker", cc...)
	err := geoCmd.Start()
	if err != nil {
		log.Fatal("Can't start geo-provider docker", err)
	}
}

func runGeo(cmds ...string) {
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}
	cmdSlice := []string{"-nodesrv", "nodeserv:9990"}
	cmdSlice = append(cmdSlice, cmds...)
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "synerex/geo_with_data",
		Cmd:   cmdSlice,
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "")
	if err != nil {
		panic(err)
	}
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func runChRetrive(cmds ...string) {
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}

	cmdSlice := []string{"-nodesrv", "nodeserv:9990"}
	cmdSlice = append(cmdSlice, cmds...)
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:      "synerex/geo_with_data",
		Entrypoint: []string{"/sxbin/channel_retrieve"},
		Cmd:        cmdSlice,
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "")
	if err != nil {
		log.Print("docker create err:", err)
	} else {
		if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			log.Print("docker start err:", err)
		}
	}
}

func harmoVIS(mbtoken string) {
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}
	cmdSlice := []string{"-nodesrv", "nodeserv:9990", "-mapbox", mbtoken}
	portMap := nat.PortMap{}
	portSet := nat.PortSet{"10080/tcp": struct{}{}}
	portMap["10080/tcp"] = []nat.PortBinding{
		nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: "10090",
		},
	}
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image:        "synerex/harmovis_layers",
		Cmd:          cmdSlice,
		ExposedPorts: portSet,
	}, &container.HostConfig{
		AutoRemove:      true,
		PortBindings:    portMap,
		PublishAllPorts: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "harmovis_layers")
	if err != nil {
		panic(err)
	}
	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func runDemo(c echo.Context) error {
	//	log.Print("runDemo! %v",c)
	runGeo("-geojson", "higashiyama_facility.geojson", "-webmercator")
	time.Sleep(300 * time.Millisecond)
	//	runGeo("-lines", "higashiyama_line.geojson", "-webmercator")
	//	time.Sleep(300 * time.Millisecond)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")

	return c.Redirect(http.StatusMovedPermanently, "control.html")
}

func runDemo2(c echo.Context) error {
	//	uv, _ :=c.FormParams()
	log.Print("runDemo2! %v", c)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")

	runChRetrive("-channel", "13", "-sendfile", "higashi-sim.csv")
	time.Sleep(1200 * time.Millisecond)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")

	return c.Redirect(http.StatusMovedPermanently, "control.html")
}

func runCovid(c echo.Context) error {
	runChRetrive("-channel", "14", "-sendfile", "aichi-covid-19.csv", "-speed", "1.2")

	return c.Redirect(http.StatusMovedPermanently, "control.html")
}

func runMesh(c echo.Context) error {
	runChRetrive("-channel", "14", "-sendfile", "meshDemo.csv", "-speed", "-450")
	return c.Redirect(http.StatusMovedPermanently, "control.html")
}

func redirectControl(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "control.html")
}


func getMapboxToken(c echo.Context) error {
	mbtoken := c.FormValue("mbtoken")
	log.Printf("got mapbox: %s", mbtoken)

	// need to start harmovis-layers!
	if strings.HasPrefix(mbtoken, "pk.") && hvLayersCmd == nil {
		//		hvLayersCmd = exec.Command("./harmovis-layers","-port", "10090", "-mapbox", mbtoken)
		//        hvLayersCmd = exec.Command("docker","run","--rm","--network","synerex-network","-p","10090:10080","harmovis_layers","-nodesrv","nodeserv:9990","-mapbox",mbtoken)
		//		err := hvLayersCmd.Start()
		harmoVIS(mbtoken)
		//		if err != nil {
		//			log.Fatal("Can't start harmovis_layers docker",err)
		//		}
		return c.Redirect(http.StatusMovedPermanently, "control.html")
	}
	//	return	c.HTML(http.StatusOK, "<HTML></HTML>")
	return c.Redirect(http.StatusMovedPermanently, "/index.html")
}

func getNetworks() []types.NetworkResource {
	filtMap := map[string][]string{"name": {"synerex-network"}}
	filtBytes, _ := json.Marshal(filtMap)
	filt, err := filters.FromJSON(string(filtBytes))
	if err != nil {
		log.Fatalf("Can't filter networks: %v", err)
	}
	opts := types.NetworkListOptions{Filters: filt}
	nets, err := dockerClient.NetworkList(context.TODO(), opts)
	if err != nil {
		log.Fatal("Can't list networks: %v", err)
	}
	log.Printf("%v", nets)
	return nets
}

// create docker network...
func createSynerexNetwork() {
	// need to check there are already synerex-network

	opts := types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{},
		},
	}
	if _, err := dockerClient.NetworkCreate(context.TODO(), "synerex-network", opts); err != nil {
		log.Fatalf("Can't create synerex-network: %v", err)
	}
}

func isRunning(containerName string) []types.Container {
	filtMap := map[string][]string{"name": {containerName}}
	filtBytes, _ := json.Marshal(filtMap)
	filt, _ := filters.FromJSON(string(filtBytes))

	opts := types.ContainerListOptions{
		All:     false,
		Quiet:   false,
		Filters: filt,
	}
	resp, err := dockerClient.ContainerList(context.TODO(), opts)
	if err != nil {
		log.Fatalf("Can't obtain container list: %v", err)
	}
	return resp
}

func startNodeServ() {
	// check running container with same-name..
	resp0 := isRunning("nodeserv")
	if len(resp0) > 0 {
		log.Printf("nodeserv is running")
		return
	}

	ctx := context.Background()
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}

	portMap := nat.PortMap{}
	portSet := nat.PortSet{"9990/tcp": struct{}{}}
	portMap["9990/tcp"] = []nat.PortBinding{
		nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: "9990",
		},
	}

	

	cmdSlice := []string{"-addr", "0.0.0.0"}
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "synerex/nodeserv",
		Cmd:   cmdSlice,
		ExposedPorts: portSet,
	}, &container.HostConfig{
		AutoRemove: true,
		PortBindings:    portMap,
		PublishAllPorts: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "nodeserv")
	if err != nil {
		log.Print("docker create err:", err)
	} else {
		if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			log.Print("docker start err:", err)
		}
	}
}

func startContainer(contName string) {

}

func startSynerexServ() {
	// check running container with same-name..
	resp0 := isRunning("sxserv")
	if len(resp0) > 0 {
		log.Printf("sxserv is running")
		return
	}

	ctx := context.Background()
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}
	portMap := nat.PortMap{}
	portSet := nat.PortSet{"10000/tcp": struct{}{}}
	portMap["10000/tcp"] = []nat.PortBinding{
		nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: "10000",
		},
	}

	cmdSlice := []string{"-nodeaddr", "nodeserv", "-servaddr", "sxserv"}
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "synerex/sxserv",
		Cmd:   cmdSlice,
		ExposedPorts: portSet,
	}, &container.HostConfig{
		AutoRemove: true,
		PortBindings:    portMap,
		PublishAllPorts: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}, nil, "sxserv")
	if err != nil {
		log.Print("docker create err:", err)
	} else {
		if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			log.Print("docker start err:", err)
		}
	}
}

func main() {

	e := echo.New()

	nets := getNetworks()
	if len(nets) == 0 {
		createSynerexNetwork()
	}
	startNodeServ()

	startSynerexServ()

	// check is harmovis_layers is working..
	resp0 := isRunning("harmovis_layers")
	if len(resp0) > 0 {
		log.Printf("harmovis_layers is running")
		e.POST("/mapbox", redirectControl)
	}else{
		e.POST("/mapbox", getMapboxToken)		
	}

	e.Static("/", "static")	


	e.POST("/demo", runDemo)
	e.POST("/demo2", runDemo2)
	e.POST("/mesh", runMesh)
	e.POST("/covid", runCovid)

	e.Logger.Fatal(e.Start(":10101"))

}
