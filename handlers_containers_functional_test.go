// +build functional

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// NewMqttTestClientLocal creates mqtt client in localhost:1883
func NewMqttTestClientLocal() *MqttTestClient {
	uiOptions := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("UI")
	ui := mqtt.NewClient(uiOptions)
	if token := ui.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return &MqttTestClient{
		ui,
		"",
	}
}

type testStatus struct {
	appStatus *Status
	ui        *MqttTestClient
}

var ts testStatus

func TestMain(m *testing.M) {
	os.Exit(setupAndRun(m))
}

func setupAndRun(m *testing.M) int {
	ts.ui = NewMqttTestClientLocal()
	defer ts.ui.Close()

	ts.appStatus = NewStatus(program{}.Config, nil, nil, "")
	ts.appStatus.dockerClient, _ = docker.NewClientWithOpts(docker.WithVersion("1.38"))
	ts.appStatus.mqttClient = mqtt.NewClient(mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("arduino-connector"))

	defer ts.appStatus.mqttClient.Disconnect(100)

	return m.Run()
}

func TestDockerPsApi(t *testing.T) {
	if token := ts.appStatus.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	subscribeTopic(ts.appStatus.mqttClient, "0", "/containers/ps/post", ts.appStatus, ts.appStatus.ContainersPsEvent, false)
	resp := ts.ui.MqttSendAndReceiveTimeout(t, "/containers/ps", "{}", 50*time.Millisecond)

	// ask Docker about containers effectively running
	cmd := exec.Command("bash", "-c", "docker ps -a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
	// Remove the first line (command output header) and the last line (empty line)
	lines = lines[1 : len(lines)-1]

	// Take json without INFO tag
	resp = strings.TrimPrefix(resp, "INFO: ")
	resp = strings.TrimSuffix(resp, "\n\n")
	var result []types.Container
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(result), len(lines))
	for i, line := range lines {
		containerId := strings.Fields(line)[0]
		assert.True(t, strings.HasPrefix(result[i].ID, containerId))
	}
}

func TestDockerListImagesApi(t *testing.T) {
	if token := ts.appStatus.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	subscribeTopic(ts.appStatus.mqttClient, "0", "/containers/images/post", ts.appStatus, ts.appStatus.ContainersListImagesEvent, false)
	resp := ts.ui.MqttSendAndReceiveTimeout(t, "/containers/images", "{}", 50*time.Millisecond)

	// ask Docker about images effectively present
	cmd := exec.Command("bash", "-c", "docker images -a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
	// Remove the first line (command output header) and the last line (empty line)
	lines = lines[1 : len(lines)-1]

	// Take json without INFO tag
	resp = strings.TrimPrefix(resp, "INFO: ")
	resp = strings.TrimSuffix(resp, "\n\n")
	var result []types.ImageSummary
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(result), len(lines))
}

func TestDockerRenameApi(t *testing.T) {
	// create a test container through Docker API
	reader, err := ts.appStatus.dockerClient.ImagePull(context.Background(), "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(ioutil.Discard, reader)

	createContResp, err := ts.appStatus.dockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := ts.appStatus.dockerClient.ContainerRemove(context.Background(), createContResp.ID, types.ContainerRemoveOptions{}); err != nil {
			t.Fatal(err)
		}
	}()

	if token := ts.appStatus.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}

	cnPayload := ChangeNamePayload{
		ContainerID:   createContResp.ID,
		ContainerName: "newname",
	}
	data, err := json.Marshal(cnPayload)
	if err != nil {
		t.Fatal(err)
	}

	subscribeTopic(ts.appStatus.mqttClient, "0", "/containers/rename/post", ts.appStatus, ts.appStatus.ContainersRenameEvent, true)
	resp := ts.ui.MqttSendAndReceiveTimeout(t, "/containers/rename", string(data), 50*time.Millisecond)

	// ask Docker about containers
	cmd := exec.Command("bash", "-c", "docker container ls -a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
	// Remove the first line (command output header) and the last line (empty line)
	lines = lines[1 : len(lines)-1]

	// Take json without INFO tag
	resp = strings.TrimPrefix(resp, "INFO: ")
	resp = strings.TrimSuffix(resp, "\n\n")
	var result ChangeNamePayload
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, cnPayload, result)

	// find test container through its ID and check its name
	for _, line := range lines {
		tokens := strings.Fields(line)
		if strings.HasPrefix(result.ContainerID, tokens[0]) {
			assert.Equal(t, result.ContainerName, tokens[len(tokens)-1])
			return
		}
	}

	t.Fatalf("no container with ID %s has been found\n", result.ContainerID)
}