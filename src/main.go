package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
)

type myProgram struct {
	client  *ssh.Client
	chn     chan int
	workDir string
}

func (p *myProgram) Start(_ service.Service) error {
	log.Println("Service starting...")
	// 启动服务时执行的操作
	go p.run()
	return nil
}
func (p *myProgram) run() {
	// 在这里编写你的服务逻辑代码
	cfg, err := readConfig(p.workDir)
	if err != nil {
		log.Fatal(err)
	}

	outChn := make(chan int)
	p.chn = outChn

	var auth []ssh.AuthMethod
	if cfg.server.keyfile != nil {
		// connect to server
		key, err := os.ReadFile(*cfg.server.keyfile)
		if err != nil {
			log.Fatalf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if cfg.server.password != nil {
		auth = append(auth, ssh.Password(*cfg.server.password))
	}

	if len(auth) == 0 {
		log.Fatal("No authentication methods available")
	}

	sshClientConfig := &ssh.ClientConfig{
		User: cfg.server.user,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			log.Printf("Warning: Unknown host key for %s: %s", hostname, key)
			return nil
		},
		BannerCallback: func(message string) error {
			fmt.Println(message)
			return nil
		},
	}

	portsMap := make(map[int]bool, len(cfg.mappings))
	successForwardCount := 0

	loopFinishedChn := make(chan bool)
	clientCloseChn := make(chan bool)

	for index := range cfg.mappings {
		portsMap[index] = false
	}
	createClientTask := func() bool {
		client, err := ssh.Dial("tcp", cfg.server.addr, sshClientConfig)
		if err != nil {
			if errors.Is(err, ssh.ErrNoAuth) {
				log.Fatal("No authentication methods available")
			}
			return false
		}
		log.Print("Connected to server")
		p.client = client
		// reset ports map
		for index := range cfg.mappings {
			portsMap[index] = false
		}
		successForwardCount = 0
		go func() {
			_ = client.Wait()
			clientCloseChn <- true
		}()
		return true
	}
	mainTask := func() {
		// connect to server
		if p.client == nil {
			if !createClientTask() {
				log.Printf("Failed to connect to server, retry after 10s")
				time.Sleep(10 * time.Second)
				loopFinishedChn <- false
			}
		}

		select {
		case <-clientCloseChn:
			log.Printf("Client closed, retry after 10s")
			p.client = nil
			time.Sleep(10 * time.Second)
			loopFinishedChn <- false
		default:
			// forward ports
			if successForwardCount < len(cfg.mappings) {
				for index, portPair := range cfg.mappings {
					if portsMap[index] {
						continue
					}
					success := forward(p.client, portPair.remote, portPair.local, portPair.mode)
					if success {
						successForwardCount++
						portsMap[index] = true
					}
				}
			}
			// check if all ports forwarded
			if successForwardCount == len(cfg.mappings) {
				time.Sleep(10 * time.Second)
			} else {
				log.Printf("%d not forwarded, retry after 5s", len(cfg.mappings)-successForwardCount)
				time.Sleep(5 * time.Second)
			}
			loopFinishedChn <- true
		}
	}
	go mainTask()
	for {
		// debug
		select {
		case s := <-p.chn:
			log.Printf("Received signal: %d", s)
			return
		case <-loopFinishedChn:
			go mainTask()

		}
	}
}
func (p *myProgram) Stop(_ service.Service) error {
	if p.chn != nil {
		p.chn <- 1
	}
	if p.client != nil {
		err := p.client.Close()
		if err != nil {

		}
	}
	// 停止服务时执行的操作
	log.Println("Service stopped.")
	return nil
}

const RemoteMark = "[Remote]"
const LocalMark = "[Local]"

func main() {
	workDir, err := getWorkDir()
	if err != nil {
		log.Fatal("load executable workDir failed", err)
	}
	// 服务日志配置
	if !service.Interactive() {
		logConfig(workDir)
		log.Println("running as service")
	}

	svcFlag := flag.String("s", "", "Control the system service.")
	flag.Parse()

	svcConfig := &service.Config{
		Name:        "WinAutoSSH",
		DisplayName: "WinAutoSSH",
		Description: "AutoSSH for Windows, automatically forward ports from server to local",
	}

	prg := &myProgram{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	prg.workDir = workDir
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	// 以服务方式运行
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

func getWorkDir() (string, error) {
	fullExecPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir, _ := filepath.Split(fullExecPath)
	return dir, nil
}
