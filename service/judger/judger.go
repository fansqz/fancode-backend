package judger

import (
	"FanCode/constants"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"os"
	"os/exec"
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
				Image: ImageName,
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

func (j *JudgeCore) Release() {
	for i := 0; i < j.poolSize; i++ {
		c := <-j.containerPool
		_ = j.cli.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
	}
}

// Compile 编译，编译时在容器外进行编译的
// language: 语言类型
// compileFiles: 需要编译文件列表
// outFilePath: 输出位置
// timeout: 限制编译时间
func (j *JudgeCore) Compile(language int, compileFiles []string, outFilePath string, timeout time.Duration) error {
	if language == constants.ProgramC {
		compileFiles = append([]string{"gcc", "-o", outFilePath}, compileFiles...)

		// 创建一个带有超时时间的上下文
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 执行编译命令
		cmd := exec.CommandContext(ctx, "gcc", compileFiles...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			// 如果是由于超时导致的错误，则返回自定义错误
			if ctx.Err() == context.DeadlineExceeded {
				return errors.New("编译超时")
			}
			return err
		}

		return nil
	} else {
		return errors.New("不支持该语言")
	}
}

func (j *JudgeCore) Execute(language int, execFile string, input <-chan string) (chan string, chan error, error) {
	execFileReader, err := os.Open(execFile)
	defer execFileReader.Close()
	if err != nil {
		return nil, nil, err
	}
	containerID := <-j.containerPool

	// 启动容器
	err = j.cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if err != nil {
		_ = j.releaseContainer(containerID)
		return nil, nil, err
	}

	// 将执行文件拷贝到容器内部的临时文件
	tmpFilePath := fmt.Sprintf("/tmp/execFile_%d", time.Now().UnixNano())
	err = j.simpleExec("mkdir "+tmpFilePath, containerID)
	if err != nil {
		_ = j.releaseContainer(containerID)
		return nil, nil, err
	}
	err = j.cli.CopyToContainer(context.Background(), containerID, tmpFilePath, execFileReader, types.CopyToContainerOptions{})
	if err != nil {
		_ = j.releaseContainer(containerID)
		return nil, nil, err
	}

	// 创建输出通道和错误通道
	output := make(chan string)
	errCh := make(chan error)

	// 根据扩展名设置执行命令
	cmd := []string{}
	switch language {
	case constants.ProgramC:
		cmd = []string{"sh", "-c", fmt.Sprintf(".%s", tmpFilePath)}
	case constants.ProgramJava:
		cmd = []string{"sh", "-c", fmt.Sprintf("java -jar %s", tmpFilePath)}
	default:
		return nil, nil, fmt.Errorf("不支持该语言")
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
		_ = j.releaseContainer(containerID)
		return nil, nil, err
	}

	hijack, err := j.cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		_ = j.releaseContainer(containerID)
		return nil, nil, err
	}
	defer hijack.Close()

	// 等待命令执行完成
	err = j.cli.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, nil, err
	}

	go func() {
		buf := make([]byte, 4096)
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

		// 删除临时文件
		_ = j.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})

		_ = j.releaseContainer(containerID)
		close(output)
	}()

	return output, errCh, nil
}

// releaseContainer 将容器放回容器池
func (j *JudgeCore) releaseContainer(containerID string) error {
	err := j.cli.ContainerStop(context.Background(), containerID, container.StopOptions{})
	j.containerPool <- containerID
	return err
}

func (j *JudgeCore) simpleExec(cmdStr string, containerID string) error {
	cmd := []string{"sh", "-c", cmdStr}

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
		_ = j.releaseContainer(containerID)
		return err
	}
	// 等待命令执行完成
	err = j.cli.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{})
	return err
}
