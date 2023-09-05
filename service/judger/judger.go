package judger

import (
	"FanCode/constants"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"path/filepath"
	"time"
)

const (
	ImageName = "judge-docker-image"
)

type JudgeCore struct {
	cli      *client.Client
	poolSize int
	// 容器池
	containerPool chan string
}

func NewJudgeCore(poolSize int) (*JudgeCore, error) {
	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// 创建容器连接池
	containerPool := make(chan string, poolSize)
	for i := 0; i < poolSize; i++ {
		// 创建容器
		resp, err := cli.ContainerCreate(
			context.Background(),
			&container.Config{
				Image: "judge-docker-image",
			}, nil, nil, nil, "")

		if err != nil {
			return nil, err
		}

		// 将容器 ID 添加到连接池
		containerPool <- resp.ID
	}

	return &JudgeCore{
		cli:           cli,
		poolSize:      poolSize,
		containerPool: containerPool,
	}, nil
}

func (j *JudgeCore) RunCode(execFile io.Reader, input <-chan string, language int, timeout time.Duration) (chan string, chan error, error) {
	// 从连接池获取可用的容器
	containerID := <-j.containerPool

	// 启动容器
	err := j.cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 创建临时文件路径
	tmpFilePath := fmt.Sprintf("/tmp/execFile_%d", time.Now().UnixNano())

	// 将执行文件拷贝到容器内部的临时文件
	err = j.cli.CopyToContainer(context.Background(), containerID, tmpFilePath, execFile, types.CopyToContainerOptions{})
	if err != nil {
		return nil, nil, err
	}

	// 创建输出通道和错误通道
	output := make(chan string)
	errCh := make(chan error)

	// 获取文件名和扩展名
	fileName := filepath.Base(tmpFilePath)
	fileExt := filepath.Ext(fileName)

	// 根据扩展名设置执行命令
	cmd := []string{}
	switch language {
	case constants.ProgramC:
		cmd = []string{"sh", "-c", fmt.Sprintf(".%s", tmpFilePath)}
	case constants.ProgramJava:
		cmd = []string{"sh", "-c", fmt.Sprintf("java -jar %s", tmpFilePath)}
	default:
		return nil, nil, fmt.Errorf("Unsupported file extension: %s", fileExt)
	}

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Detach:       false,
	}
	resp, err := j.cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return nil, nil, err
	}

	hijack, err := j.cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, nil, err
	}
	defer hijack.Close()

	// 等待命令执行完成
	err = j.cli.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, nil, err
	}

	buf := make([]byte, 4096)

	go func() {
		for {
			inputItem := <-input
			if inputItem == "exit" {
				break
			}
			_, _ = hijack.Conn.Write([]byte(inputItem + "\n"))
			// 等待输出
			n, err := hijack.Reader.Read(buf)
			if err != nil {
				errCh <- err
				break
			}
			output <- string(buf[:n])
		}

		// 关闭容器
		_ = j.cli.ContainerStop(context.Background(), containerID, container.StopOptions{})

		// 删除临时文件
		_ = j.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})

		// 将容器 ID 放回连接池
		j.containerPool <- containerID

		close(output)
	}()

	return output, errCh, nil
}
