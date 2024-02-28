package tests

//go:generate tar cvf assets/docker.tar -C assets/docker/ .

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	ci "github.com/taubyte/go-simple-container"
	"github.com/taubyte/go-simple-container/gc"
)

func TestContainerBasicCommand(t *testing.T) {
	ci.ForceRebuild = true

	ctx := context.Background()

	cli, err := ci.New()
	if err != nil {
		t.Error(err)
		return
	}

	file, err := os.OpenFile(DockerTarBallPath, os.O_RDWR, 0444)
	if err != nil {
		t.Errorf("Opening docker tarball failed with: %s", err)
		return
	}

	defer file.Close()

	image, err := cli.Image(ctx, testCustomImage, ci.Build(file))
	if err != nil {
		t.Error(err)
		return
	}

	container, err := image.Instantiate(
		ctx,
		ci.Command(basicCommand),
	)
	if err != nil {
		t.Error(err)
		return
	}

	logs, err := container.Run(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(logs.Combined())
	out := buf.String()
	if !strings.Contains(out, message) {
		t.Error("Container output not the same as the given message")
		return
	}

	err = logs.Close()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestContainerCleanUpInterval(t *testing.T) {
	ci.ForceRebuild = true

	ctx := context.Background()
	cli, err := ci.New()
	if err != nil {
		t.Error(err)
		return
	}

	err = gc.Start(ctx, gc.Interval(20*time.Second), gc.MaxAge(10*time.Second))
	if err != nil {
		t.Error(err)
		return
	}

	file, err := os.OpenFile(DockerTarBallPath, os.O_RDWR, 0444)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	image, err := cli.Image(ctx, testCustomImage, ci.Build(file))
	if err != nil {
		t.Error(err)
		return
	}

	container, err := image.Instantiate(ctx, ci.Command(basicCommand))
	if err != nil {
		t.Error(err)
		return
	}

	logs, err := container.Run(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(logs.Combined())
	out := buf.String()
	if !strings.Contains(out, message) {
		t.Error("Container output not the same as the given message")
		return
	}

	err = logs.Close()
	if err != nil {
		t.Error(err)
		return
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{
		Filters: ci.NewFilter("reference", testCustomImage),
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(images) == 0 {
		t.Errorf("Expected to find docker image %s", testCustomImage)
		return
	}

	time.Sleep(20 * time.Second)

	images, err = cli.ImageList(ctx, types.ImageListOptions{
		Filters: ci.NewFilter("reference", testCustomImage),
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(images) > 0 {
		t.Error("Expected to find no containers after clean interval")
	}
}

func TestContainerMount(t *testing.T) {
	ci.ForceRebuild = true

	ctx := context.Background()
	cli, err := ci.New()
	if err != nil {
		t.Error(err)
		return
	}

	file, err := os.OpenFile(DockerTarBallPath, os.O_RDWR, 0444)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	image, err := cli.Image(ctx, testCustomImage, ci.Build(file))
	if err != nil {
		t.Error(err)
		return
	}

	container, err := image.Instantiate(
		ctx,
		ci.Volume(VolumePath, "/src"),
		ci.Command(testScriptCommand),
	)
	if err != nil {
		t.Error(err)
		return
	}
	logs, err := container.Run(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(logs.Combined())
	out := buf.String()
	if !strings.Contains(out, TestScriptMessage) {
		t.Error("Container output not the same as the given message")
		return
	}

	err = logs.Close()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestContainerBasicVariables(t *testing.T) {
	ci.ForceRebuild = true

	ctx := context.Background()
	cli, err := ci.New()
	if err != nil {
		t.Error(err)
		return
	}

	file, err := os.OpenFile(DockerTarBallPath, os.O_RDWR, 0444)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	image, err := cli.Image(ctx, testCustomImage, ci.Build(file))
	if err != nil {
		t.Error(err)
		return
	}

	vars := map[string]string{testEnv: testVal}
	container, err := image.Instantiate(
		ctx,
		ci.Command(testVarsCommand),
		ci.Volume(VolumePath, "/src"),
		ci.Variables(vars),
	)
	if err != nil {
		t.Error(err)
		return
	}

	logs, err := container.Run(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(logs.Combined())
	out := buf.String()
	if !strings.Contains(out, testVal) {
		t.Error("Container output not the same as the given message")
		return
	}

	err = logs.Close()
	if err != nil {
		t.Error(err)
		return
	}

}

func TestContainerParallel(t *testing.T) {
	var (
		wg    sync.WaitGroup
		count int = 4
	)

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			TestContainerMount(t)
			wg.Done()
		}()
	}

	wg.Wait()
}
