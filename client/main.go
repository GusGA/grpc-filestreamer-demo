package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"

	pb "github.com/gusga/streamer/protos/file_streamer"
)

const (
    baseDirPath = "file_dir"
)

func listFiles(path string, log hclog.Logger) ([]string, error) {
    filesNames := make([]string, 0)

    files, err := ioutil.ReadDir(path)
    if err != nil {
        log.Error("Could not connect client", "error", err)
        return filesNames, err
    }

    for _, f := range files {
        log.Info("File found", "file", f.Name(), "size", f.Size())
        filesNames = append(filesNames, f.Name())
    }

    return filesNames, nil
}

func streamFile(client pb.FileStreamerClient, log hclog.Logger) {
    filesNames, err := listFiles(baseDirPath, log)
    if err != nil {
        os.Exit(1)
    }
    for _, file := range filesNames {
        log.Info("Streaming file content for file", "file", file)
        streamContent, err := client.StreamFile(context.Background(), &pb.StreamerFileRequest{
            FileName: file,
        })
        if err != nil {
            log.Error("Error streaming content for file", "file", file, "error", err)
            return
        }
        for {
            resp, err := streamContent.Recv()
            if err == io.EOF {
                log.Info("End of file reached", "file", file)
                break
            }
            if err != nil {
                log.Error("Error receiving streaming content for file", "file", file)
                break
            }
            fmt.Println(resp.FileLineContent)
        }
    }
}

func streamCounter(client pb.FileStreamerClient, log hclog.Logger) error {
    clientName := uuid.New().String()
    log.Info("Running new client", "client", clientName)
    sCounter, err := client.StreamCounter(context.Background(), &pb.StreamerCounterRequest{
        ClientName: clientName,
    })

    if err != nil {
        log.Error("Error receiving stream counter", "error", err)
        return err
    }
    var counter int32
    for {
        resp, err := sCounter.Recv()
        if err == io.EOF {
            log.Info("End of stream counter reached", "count", counter)
            break
        }
        if err != nil {
            return err
        }
        counter = resp.GetCounter()

        fmt.Printf("Counter goes %d\n", counter)
    }

    return nil
}

func main() {
    log := hclog.Default()

    log.Info("Running client")
    gClient, err := grpc.Dial("localhost:7000", grpc.WithInsecure())
    if err != nil {
        log.Error("Could not connect client", "error", err)
        os.Exit(1)
    }

    defer gClient.Close()

    client := pb.NewFileStreamerClient(gClient)

    streamCounter(client, log)
}


