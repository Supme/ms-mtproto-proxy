package main

import (
	"bufio"
	"fmt"
	"gopkg.in/fsnotify/fsnotify.v1"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
)

type mtprotoStruct struct {
	cmd *exec.Cmd
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Use command:")
		fmt.Println("\t", os.Args[0], "path/to/secret.conf path/to/mtproto-proxy params for mtproto-proxy")
	}

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("received an interrupt, stopping services...")
		close(cleanupDone)
	}()

	mtproto := new(mtprotoStruct)
	err := mtproto.start()
	if err != nil {
		log.Fatal(err)
	}
	defer mtproto.stop()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	doneWatcher := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == os.Args[1] {
					fmt.Println(event.Name)
					err := mtproto.stop()
					if err != nil {
						log.Fatal(err)
					}
					err = mtproto.start()
					if err != nil {
						log.Fatal(err)
					}
				}
			case err := <-watcher.Errors:
				log.Println("watcher error:", err)
				close(doneWatcher)
			}
		}
	}()
	err = watcher.Add(filepath.Dir(os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-doneWatcher:
		os.Exit(2)
	case <-cleanupDone:
		mtproto.stop()
		os.Exit(0)
	}
}

func (m *mtprotoStruct) stop() error {
	err := m.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}
	_, err = m.cmd.Process.Wait()
	if err != nil {
		return err
	}
	m.cmd = nil
	return nil
}

func (m *mtprotoStruct) start() error {
	secrets, err := m.parseSecrets(os.Args[1])
	if err != nil {
		return fmt.Errorf("parse config error: %s/n", err)
	}
	var (
		args []string
	)
	for k := range secrets {
		args = append(args, "-S")
		args = append(args, k)
	}
	for i := range os.Args {
		if i == 0 || i == 1 || i == 2 {
			continue
		}
		args = append(args, os.Args[i])
	}

	cmd := exec.Command(os.Args[2], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Execute:\n%+v\n",cmd.Args)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start %s error: %s/n", os.Args[2], err)
	}
	m.cmd = cmd
	return nil
}

func (m *mtprotoStruct) parseSecrets(secretFile string) (map[string]string, error) {
	file, err := os.OpenFile(secretFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	secrets := map[string]string{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(text, "#") {
			if fields := strings.Fields(text); len(fields) > 1 {
				secrets[fields[0]] = fields[1]
			}
		}
	}
	return secrets, nil
}
