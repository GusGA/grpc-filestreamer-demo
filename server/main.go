package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/gusga/streamer/protos/file_streamer"
)

const (
    baseDirPath = "file_dir"
)

var counter int32

type server struct {
    log hclog.Logger
    pb.UnimplementedFileStreamerServer
}

func (s *server) FileSize(ctx context.Context, sfr *pb.StreamerFileRequest) (*pb.FileSizeResponse, error) {
    s.log.Info("Requesting File size on base directory file_dir", "file", sfr.GetFileName() )
    file, err := os.Open(fmt.Sprintf("%s/%s", baseDirPath, sfr.GetFileName()))
    if err != nil {
        s.log.Error("Error opening file", "error", err)
        return &pb.FileSizeResponse{}, err
    }
    defer file.Close()

    fileInfo, err := file.Stat()
    if err != nil {
        s.log.Error("Error getting file info", "error", err)
        return &pb.FileSizeResponse{}, err
    }

    return &pb.FileSizeResponse{
        FileName: sfr.GetFileName(),
        FileSize: fileInfo.Size(),
    }, nil
}

func (s *server) StreamFile(sfr *pb.StreamerFileRequest, stream pb.FileStreamer_StreamFileServer) error {
    file, err := os.Open(fmt.Sprintf("%s/%s", baseDirPath, sfr.GetFileName()))
    if err != nil {
        s.log.Error("Error opening file", "error", err)
        return err
    }

    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        stream.Send(&pb.StreamerFileResponse{
           FileLineContent: scanner.Text(),
        })
        time.Sleep(200 * time.Millisecond)
    }

    if err := scanner.Err(); err != nil {
        s.log.Error("Error streaming file", "error", err)
    }
    return nil
}

func (s *server) StreamCounter(scr *pb.StreamerCounterRequest, stream pb.FileStreamer_StreamCounterServer) error {
    s.log.Info("Sending counter info to client", "client", scr.GetClientName())

    for counter < 1000 {
        stream.Send(&pb.StreamerCounterResponse{
            Counter: counter,
         })
         time.Sleep(100 * time.Millisecond)
         counter += 1
    }

    return nil
}

func main() {
    log := hclog.Default()

    s := &server{
        log,
        pb.UnimplementedFileStreamerServer{},
    }
    gServer := grpc.NewServer()
    reflection.Register(gServer)

    log.Info("Registering server")
    pb.RegisterFileStreamerServer(gServer, s)
    listener, err := net.Listen("tcp", "0.0.0.0:7000")
    if err != nil {
        log.Error("Error creating the net listener on port 7000", "error", err)
        os.Exit(1)
    }
    log.Info("Listener created")

    log.Info("Running gRPC Server on port 7000")
    if err := gServer.Serve(listener); err != nil {
        log.Error("failed to serve", err)
    }

}
