package main

import (
	"context"
	"net/http"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var (
	hvLayersCmd *exec.Cmd
)

func init(){
	dclt, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		dockerClient = dclt
	}else{
		log.Fatal("Can't use docker..")
	}

	// check synerex-network


}

func runGeoDockerWithOSExec(cmds ...string){
	baseCmd := []string{"run","--network","synerex-network","synerex_geography","-nodesrv","nodeserv:9990"}
	cc := append(baseCmd, cmds...)
	geoCmd := exec.Command("docker",cc...)
	err := geoCmd.Start()
	if err != nil {
		log.Fatal("Can't start geo-provider docker",err)
	}
}

func runGeoDocker(cmds ...string){

	baseCmd := []string{"run","--network","synerex-network","synerex_geography","-nodesrv","nodeserv:9990"}
	cc := append(baseCmd, cmds...)
	geoCmd := exec.Command("docker",cc...)
	err := geoCmd.Start()
	if err != nil {
		log.Fatal("Can't start geo-provider docker",err)
	}
}


func runGeo(cmds ...string){
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}
	cmdSlice := []string{"-nodesrv","nodeserv:9990"}
	cmdSlice = append(cmdSlice,cmds...)
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
			Image: "synerex/geo_with_data",
			Cmd: cmdSlice,
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
	
func runChRetrive(cmds ...string){
	ctx := context.Background()
	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig["synerex-network"] = &network.EndpointSettings{
		NetworkID: "synerex-network",
	}

	cmdSlice := []string{"-nodesrv","nodeserv:9990"}
	cmdSlice = append(cmdSlice,cmds...)
	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
			Image: "synerex/geo_with_data",
			Entrypoint: []string{"/sxbin/channel_retrieve"},
			Cmd: cmdSlice,
		}, &container.HostConfig{
			AutoRemove: true,
		}, &network.NetworkingConfig{
			EndpointsConfig: endpointsConfig,
		}, nil, "chan_retrieve")
	if err != nil {
		log.Print("docker create err:",err)
	}else{
		if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			log.Print("docker start err:",err)
		}
	}
}


func runDemo(c echo.Context)error{
//	log.Print("runDemo! %v",c)
	runGeo("-geojson", "higashiyama_facility.geojson", "-webmercator")
	time.Sleep(500* time.Millisecond)
	runGeo("-lines", "higashiyama_line.geojson", "-webmercator")
	time.Sleep(500* time.Millisecond)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")

	return	c.Redirect(http.StatusMovedPermanently, "control.html")
}

func runDemo2(c echo.Context)error{
	//	uv, _ :=c.FormParams()
	log.Print("runDemo2! %v",c)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")
	
	runChRetrive("-channel","13","-sendfile", "higashi-sim.csv")
	time.Sleep(500* time.Millisecond)
	runGeo("-viewState", "35.15596582695651,136.9783370942177,16")
	
	return	c.Redirect(http.StatusMovedPermanently, "control.html")
}

func runCovid(c echo.Context)error{
	runChRetrive("-channel","14","-sendfile", "aichi-covid-19.csv", "-speed", "1.2")
	
	return	c.Redirect(http.StatusMovedPermanently, "control.html")
}
	
func runMesh(c echo.Context)error{
	runChRetrive("-channel","14","-sendfile", "meshDemo.csv", "-speed", "-450")
	return	c.Redirect(http.StatusMovedPermanently, "control.html")
}
	

	



func getMapboxToken(c echo.Context)error{
	mbtoken := c.FormValue("mbtoken")
	log.Printf("got mapbox: %s",mbtoken)

	// need to start harmovis-layers!
	if strings.HasPrefix(mbtoken,"pk.") && hvLayersCmd == nil {
//		hvLayersCmd = exec.Command("./harmovis-layers","-port", "10090", "-mapbox", mbtoken)
        hvLayersCmd = exec.Command("docker","run","--rm","--network","synerex-network","-p","10090:10080","harmovis_layers","-nodesrv","nodeserv:9990","-mapbox",mbtoken)
		err := hvLayersCmd.Start()
		if err != nil {
			log.Fatal("Can't start harmovis_layers docker",err)
		}
		return	c.Redirect(http.StatusMovedPermanently, "control.html")
	}
//	return	c.HTML(http.StatusOK, "<HTML></HTML>")
	return	c.Redirect(http.StatusMovedPermanently, "/index.html")
}

func main(){

	e := echo.New()
	e.Static("/","static")

	e.POST("/mapbox", getMapboxToken)
	e.POST("/demo", runDemo)
	e.POST("/demo2", runDemo2)
	e.POST("/mesh", runMesh)
	e.POST("/covid", runCovid)

	e.Logger.Fatal(e.Start(":10101"))



}